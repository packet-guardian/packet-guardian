package common

import "testing"

type test struct {
	ua       string
	expected string
}

var userAgents = [...]test{
	test{
		ua:       "Mozilla/5.0 (Windows NT 5.1) AppleWebKit/535.1 (KHTML, like Gecko) Chrome/14.0.815.0 Safari/535.1",
		expected: "Windows XP",
	},
	test{
		ua:       "Mozilla/6.0 (Windows NT 6.2; WOW64; rv:16.0.1) Gecko/20121011 Firefox/16.0.1",
		expected: "Windows 8",
	},
	test{
		ua:       "Mozilla/5.0 (Windows; U; Windows NT 6.0; tr-TR) AppleWebKit/533.18.1 (KHTML, like Gecko) Version/5.0.2 Safari/533.18.5",
		expected: "Windows Vista",
	},
	test{
		ua:       "Mozilla/5.0 (Macintosh; U; PPC Mac OS X 10_4_11; en) AppleWebKit/528.4+ (KHTML, like Gecko) Version/4.0dp1 Safari/526.11.2",
		expected: "Mac OS X 10.4.11",
	},
	test{
		ua:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_6_8) AppleWebKit/537.13+ (KHTML, like Gecko) Version/5.1.7 Safari/534.57.2",
		expected: "Mac OS X 10.6.8 (Snow Leopard)",
	},
	test{
		ua:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_8) AppleWebKit/537.13+ (KHTML, like Gecko) Version/5.1.7 Safari/534.57.2",
		expected: "Mac OS X 10.11.8 (El Capitan)",
	},
	test{
		ua:       "Mozilla/5.0 (iPad; CPU OS 6_0 like Mac OS X) AppleWebKit/536.26 (KHTML, like Gecko) Version/6.0 Mobile/10A5355d Safari/8536.25",
		expected: "iPad iOS 6.0",
	},
	test{
		ua:       "Mozilla/5.0 (iPod touch; CPU iPhone OS 7_0_3) AppleWebKit/537.51.1 (KHTML, like Gecko) Version/7.0 Mobile/11B511 Safari/9537.53",
		expected: "iPod touch iOS 7.0.3",
	},
	test{
		ua:       "Mozilla/5.0 (iPhone; U; CPU iPhone OS 4_3_5) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8L1 Safari/6533.18.5",
		expected: "iPhone iOS 4.3.5",
	},
	test{
		ua:       "Mozilla/5.0 (X11; Ubuntu; Linux i686; rv:15.0) Gecko/20100101 Firefox/15.0.1",
		expected: "Ubuntu",
	},
	test{
		ua:       "Mozilla/5.0 (Fedora; Linux x86_64) AppleWebKit/602.1 (KHTML, like Gecko) Version/8.0 Safari/602.1",
		expected: "Fedora",
	},
	test{
		ua:       "Mozilla/5.0 (X11; CrOS i686 12.0.742.91) AppleWebKit/534.30 (KHTML, like Gecko) Chrome/12.0.742.93 Safari/534.30",
		expected: "Chromium OS",
	},
	test{
		ua:       "Mozilla/5.0 (X11; Linux i686) AppleWebKit/534.23 (KHTML, like Gecko) Chrome/11.0.686.3 Safari/534.23",
		expected: "Linux 32 bit",
	},
	test{
		ua:       "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
		expected: "Linux 64 bit",
	},
	test{
		ua:       "Mozilla/5.0 (Linux; U; Android 2.3; en-us) AppleWebKit/999+ (KHTML, like Gecko) Safari/999.9",
		expected: "Android 2.3",
	},
	test{
		ua:       "Mozilla/5.0 (Android 5.1.1; Tablet; rv:46.0) Gecko/46.0 Firefox/46.0",
		expected: "Android 5.1.1",
	},
}

func TestParseUserAgent(t *testing.T) {
	for i, uaTest := range userAgents {
		result := ParseUserAgent(uaTest.ua)
		if result != uaTest.expected {
			t.Errorf("Test %d: Incorrectly parsed user agent. Expected %s, got %s.", i, uaTest.expected, result)
		}
	}
}
