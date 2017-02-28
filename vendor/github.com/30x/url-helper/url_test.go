package urlhelper

import (
	"net/http"
	"net/url"
	"testing"
)

func validate(expected, actual, msg string, t *testing.T) {
	if expected != actual {
		t.Fatalf("URL did not match %s != %s: %s", expected, actual, msg)
	}
}

func makeReq(path, host string, xfh, xfp *string) *http.Request {
	u, _ := url.Parse(path)
	req := http.Request{
		URL:  u,
		Host: host,
		Header: http.Header{
			"Host": []string{host},
		},
	}

	if xfh != nil {
		req.Header.Set(XForwardedHost, *xfh)
	}

	if xfp != nil {
		req.Header.Set(XForwardedProtocol, *xfp)
	}

	return &req
}

func TestJoin(t *testing.T) {
	helper, _ := NewURLHelper(makeReq("/some/path?test=123", "1.2.3.4", nil, nil))
	validate(helper.Join("new"), "http://1.2.3.4/some/path/new?test=123", "Should add new path to end of /some/path", t)
	validate(helper.Join(".."), "http://1.2.3.4/some?test=123", "Relitive back should work", t)

	xfh := "api.example.dev"
	helperWithXfh, _ := NewURLHelper(makeReq("/some/path?test=123", "1.2.3.4", &xfh, nil))
	validate(helperWithXfh.Join("new"), "http://api.example.dev/some/path/new?test=123", "Should add new path to end of /some/path", t)
	validate(helperWithXfh.Join(".."), "http://api.example.dev/some?test=123", "Relitive back should work", t)

	xfp := "https"
	helperWithXfp, _ := NewURLHelper(makeReq("/some/path?test=123", "1.2.3.4", &xfh, &xfp))
	validate(helperWithXfp.Join("new"), "https://api.example.dev/some/path/new?test=123", "Should add new path to end of /some/path", t)
	validate(helperWithXfp.Join(".."), "https://api.example.dev/some?test=123", "Relitive back should work", t)
}

func TestPath(t *testing.T) {
	helper, _ := NewURLHelper(makeReq("/some/path?test=123", "1.2.3.4", nil, nil))
	validate(helper.Path("/some/new"), "http://1.2.3.4/some/new", "Should have new path", t)
	validate(helper.Path("/"), "http://1.2.3.4/", "Should have new base path", t)
	validate(helper.PathWithQuery("/", "page=123"), "http://1.2.3.4/?page=123", "Should have new base path", t)
	validate(helper.Join("new"), "http://1.2.3.4/some/path/new?test=123", "Should have new base path", t)

	xfh := "api.example.dev"
	helperWithXfh, _ := NewURLHelper(makeReq("/some/path?test=123", "1.2.3.4", &xfh, nil))
	validate(helperWithXfh.Path("/"), "http://api.example.dev/", "Should have new base path", t)

	xfp := "https"
	helperWithXfp, _ := NewURLHelper(makeReq("/some/path?test=123", "1.2.3.4", &xfh, &xfp))
	validate(helperWithXfp.Path("/"), "https://api.example.dev/", "Should have new base path", t)
}

func TestCurrent(t *testing.T) {
	helper, _ := NewURLHelper(makeReq("/some/path?test=123", "1.2.3.4", nil, nil))
	validate(helper.Current(), "http://1.2.3.4/some/path?test=123", "Current matches the same url", t)

	xfh := "api.example.dev"
	helperWithXfh, _ := NewURLHelper(makeReq("/some/path?test=123", "1.2.3.4", &xfh, nil))
	validate(helperWithXfh.Current(), "http://api.example.dev/some/path?test=123", "Current matches the same url forwarded host", t)

	xfp := "https"
	helperWithXfp, _ := NewURLHelper(makeReq("/some/path?test=123", "1.2.3.4", &xfh, &xfp))
	validate(helperWithXfp.Current(), "https://api.example.dev/some/path?test=123", "Current matches the same url with forwarded host and proto", t)
}
