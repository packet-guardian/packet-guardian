package common

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmptyAPIRespEncode(t *testing.T) {
	resp := NewEmptyAPIResponse()

	if resp.Message != "" {
		t.Fatal("API Message should be empty")
	}
	if resp.Data != nil {
		t.Fatal("API Data should be nil")
	}

	expected := `{"Message":"","Data":null}`
	encoded := string(resp.Encode())
	if encoded != expected {
		t.Errorf("API testcase 'Empty Response' failed encode. Expected %s, got %s", expected, encoded)
	}
}

var apiRespEncodeTestCases = []struct {
	name     string
	message  string
	data     interface{}
	expected string
}{
	{
		name:     "No message no data",
		message:  "",
		data:     nil,
		expected: `{"Message":"","Data":null}`,
	},
	{
		name:     "Message no data",
		message:  "A message",
		data:     nil,
		expected: `{"Message":"A message","Data":null}`,
	},
	{
		name:    "Message and data",
		message: "A message",
		data: map[string]string{
			"Item1": "test item 1",
			"Item2": "test item 2",
		},
		expected: `{"Message":"A message","Data":{"Item1":"test item 1","Item2":"test item 2"}}`,
	},
}

func TestAPIRespEncodeWithMessage(t *testing.T) {
	for _, testcase := range apiRespEncodeTestCases {
		resp := NewAPIResponse(testcase.message, testcase.data)

		encoded := string(resp.Encode())
		if encoded != testcase.expected {
			t.Errorf("API testcase '%s' failed encode. Expected %s, got %s", testcase.name, testcase.expected, encoded)
		}
	}
}

var apiRespTestCases = []struct {
	name     string
	message  string
	data     interface{}
	len      int
	httpCode int
}{
	{
		name:     "No message no data",
		message:  "",
		data:     nil,
		len:      0,
		httpCode: http.StatusNoContent,
	},
	{
		name:     "Message no data",
		message:  "A message",
		data:     nil,
		len:      len(`{"Message":"A message","Data":null}`),
		httpCode: http.StatusOK,
	},
	{
		name:    "Message and data",
		message: "A message",
		data: map[string]string{
			"Item1": "test item 1",
			"Item2": "test item 2",
		},
		len:      len(`{"Message":"A message","Data":{"Item1":"test item 1","Item2":"test item 2"}}`),
		httpCode: http.StatusAccepted,
	},
}

func TestAPIHttpResp(t *testing.T) {
	for _, testcase := range apiRespTestCases {
		resp := NewAPIResponse(testcase.message, testcase.data)
		writter := httptest.NewRecorder()

		len, err := resp.WriteResponse(writter, testcase.httpCode)
		if err != nil {
			t.Errorf("Error writting api response, %q", err.Error())
			continue
		}

		if len != testcase.len {
			t.Errorf("API testcase '%s' wrong length. Expected %d, got %d", testcase.name, testcase.len, len)
			continue
		}

		contentType := writter.HeaderMap.Get("Content-Type")
		if contentType != ContentTypeJSON {
			t.Errorf("API testcase '%s' wrong content type. Expected %s, got %s", testcase.name, ContentTypeJSON, contentType)
		}
	}
}
