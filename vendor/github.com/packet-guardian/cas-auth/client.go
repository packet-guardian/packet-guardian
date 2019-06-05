package cas

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

var (
	urlCleanParameters     = []string{"gateway", "renew", "service", "ticket"}
	redirectAttemptedError = errors.New("redirect")
	InvalidCredentials     = errors.New("Bad username or password")
)

// sanitisedURL cleans a URL of CAS specific parameters
func sanitisedURL(unclean *url.URL) *url.URL {
	// Shouldn't be any errors parsing an existing *url.URL
	u, _ := url.Parse(unclean.String())
	q := u.Query()

	for _, param := range urlCleanParameters {
		q.Del(param)
	}

	u.RawQuery = q.Encode()
	return u
}

// sanitisedURLString cleans a URL and returns its string value
func sanitisedURLString(unclean *url.URL) string {
	return sanitisedURL(unclean).String()
}

// serviceURL determines an absolute URL from the http.Request.
func (c *Client) serviceURL(r *http.Request) (*url.URL, error) {
	if c.ServiceURL != nil {
		return c.ServiceURL, nil
	}

	u, err := url.Parse(r.URL.String())
	if err != nil {
		return nil, err
	}

	u.Host = r.Host
	u.Scheme = "http"

	if scheme := r.Header.Get("X-Forwarded-Proto"); scheme != "" {
		u.Scheme = scheme
	} else if r.TLS != nil {
		u.Scheme = "https"
	}

	return u, nil
}

type Client struct {
	URL        *url.URL
	ServiceURL *url.URL
}

func (c *Client) AuthenticateUser(username, password string, r *http.Request) (*AuthenticationResponse, error) {
	lt, et, jsession, err := c.getLoginToken(r)
	if err != nil {
		return nil, err
	}
	if lt == "" {
		return nil, errors.New("Couldn't get a login token")
	}
	if jsession == nil {
		return nil, errors.New("Couldn't get server session cookie")
	}

	form := url.Values{}
	form.Add("username", username)
	form.Add("password", password)
	form.Add("lt", lt)
	form.Add("_eventId", "submit") // Not sure why this is needed, it's not in the spec
	if et != "" {
		form.Add("execution", et)
	}

	client := &http.Client{}
	// Force the client to never follow redirects
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return redirectAttemptedError
	}
	reqUrl, _ := c.loginUrlForRequestor(r)
	req, err := http.NewRequest("POST", reqUrl, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(jsession) // The login token is tied to our session
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()
	// Filter out our fake redirect error
	if urlError, ok := err.(*url.Error); ok && urlError.Err == redirectAttemptedError {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	location, err := resp.Location()
	if err != nil {
		return nil, InvalidCredentials
	}
	return c.validateTicket(location.Query().Get("ticket"), r)
}

func (c *Client) validateTicket(ticket string, r *http.Request) (*AuthenticationResponse, error) {
	// Validate the ticket
	reqUrl, _ := c.serviceValidateUrlForRequest(ticket, r)
	resp, err := http.Get(reqUrl)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Failed to verify ticket")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	return ParseServiceResponse(body)
}

func (c *Client) getLoginToken(r *http.Request) (string, string, *http.Cookie, error) {
	reqUrl, _ := c.loginUrlForRequestor(r)
	resp, err := http.Get(reqUrl)
	if err != nil {
		return "", "", nil, err
	}

	var jsession *http.Cookie
	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "JSESSIONID" {
			jsession = cookie
			break
		}
	}

	b := resp.Body
	defer b.Close()
	loginToken := ""
	executionToken := ""

	z := html.NewTokenizer(b)
tokenLoop:
	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			break tokenLoop
		case tt == html.SelfClosingTagToken:
			t := z.Token()

			// Check if the token is an <input> tag
			isInput := t.Data == "input"
			if !isInput {
				continue
			}

			// Iterate over all of the Token's attributes until we find an "id", "name", or "value"
			tokenName := ""
			tokenVal := ""
			for _, a := range t.Attr {
				if a.Key == "name" {
					tokenName = a.Val
				} else if a.Key == "value" {
					tokenVal = a.Val
				}
			}

			switch tokenName {
			case "lt":
				loginToken = tokenVal
			case "execution":
				executionToken = tokenVal
			}

			if loginToken != "" && executionToken != "" {
				break tokenLoop
			}
		}
	}
	return loginToken, executionToken, jsession, nil
}

func (c *Client) loginUrlForRequestor(r *http.Request) (string, error) {
	u, err := c.URL.Parse("login")
	if err != nil {
		return "", err
	}

	service, err := c.serviceURL(r)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Add("service", sanitisedURLString(service))
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (c *Client) validateUrlForRequest(ticket string, r *http.Request) (string, error) {
	u, err := c.URL.Parse("validate")
	if err != nil {
		return "", err
	}

	service, err := c.serviceURL(r)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Add("service", sanitisedURLString(service))
	q.Add("ticket", ticket)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (c *Client) serviceValidateUrlForRequest(ticket string, r *http.Request) (string, error) {
	u, err := c.URL.Parse("serviceValidate")
	if err != nil {
		return "", err
	}

	service, err := c.serviceURL(r)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Add("service", sanitisedURLString(service))
	q.Add("ticket", ticket)
	u.RawQuery = q.Encode()

	return u.String(), nil
}
