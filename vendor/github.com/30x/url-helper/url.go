package urlhelper

import (
	"net/http"
	"net/url"
	"path"
)

const (
	// XForwardedHost header for host forwarding
	XForwardedHost = "X-Forwarded-Host"
	// XForwardedProtocol header for http protocol forwarding
	XForwardedProtocol = "X-Forwarded-Proto"
)

// URLHelper struct to keep internal state of the request object
type URLHelper struct {
	base     *url.URL
	origBase string
}

// NewURLHelper creates a new URLHelper from the http.Request
func NewURLHelper(req *http.Request) (*URLHelper, error) {
	u, err := url.Parse(req.URL.RequestURI())
	if err != nil {
		return nil, err
	}

	// Set Host based on X-Forwarded-Host or default to req's Host
	if req.Header.Get(XForwardedHost) != "" {
		u.Host = req.Header.Get(XForwardedHost)
	} else {
		u.Host = req.Host
	}

	// Set protocol scheme based on X-Forwarded-Proto or default to used transport
	if req.Header.Get(XForwardedProtocol) != "" {
		u.Scheme = req.Header.Get(XForwardedProtocol)
	} else {
		if req.TLS != nil {
			u.Scheme = "https"
		} else {
			u.Scheme = "http"
		}
	}

	return &URLHelper{
		base:     u,
		origBase: u.Path,
	}, nil
}

/*
Join returns the absolute of the request joined with the inputed path segment.  Keeps query params.
*/
func (uh URLHelper) Join(pathname string) string {
	defer uh.reset()
	uh.base.Path = path.Join(uh.base.EscapedPath(), pathname)
	return uh.base.String()
}

/*
Path returns the absolute of the request reset to the inputed path segment. Removes all query params.
*/
func (uh URLHelper) Path(pathname string) string {
	return uh.path(pathname, nil)
}

/*
PathWithQuery returns the absolute of the request reset to the inputed path segment and uses the inputed query params.
*/
func (uh URLHelper) PathWithQuery(pathname string, query string) string {
	return uh.path(pathname, &query)
}

/*
Current retuns the current url with X-Forwarded-Host and X-Forwarded-Proto
*/
func (uh URLHelper) Current() string {
	return uh.base.String()
}

/*
reset sets the URL path to the original path
*/
func (uh URLHelper) reset() {
	uh.base.Path = uh.origBase
}

func (uh URLHelper) path(pathname string, query *string) string {
	defer uh.reset()
	origQuery := uh.base.RawQuery
	if query != nil {
		uh.base.RawQuery = *query
	} else {
		uh.base.RawQuery = ""
	}

	uh.base.Path = path.Clean(pathname)
	newPath := uh.base.String()
	uh.base.RawQuery = origQuery
	return newPath
}
