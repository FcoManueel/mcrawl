package urlparse

import (
	"net/url"
	"strings"
)

// Normalize gets a url and returns a normalized version of it.
// Right now it has enough logic to get by, but it's not complete.
// For more info see https://datatracker.ietf.org/doc/html/rfc3986#section-6
func Normalize(s string) (*url.URL, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if !strings.HasPrefix(s, "http") {
		// schema has to be present in order for the url package to parse the hostname
		s = "http://" + s
	}
	s = strings.TrimSuffix(s, "/")

	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	u.Host = strings.TrimPrefix(u.Host, "www.")
	u.Fragment = ""
	return u, nil
}
