package crawl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrawler(t *testing.T) {
	// Prepare test server to crawl. It'll return the same page for any path.
	// We'll test that the crawler considers the filter function (to avoid
	// crawling what we don't want) and the visited cache (to avoid infinite
	// loops).
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`<html>
<head></head>
<body>
	<h1>Page</h1>
	<a href="http://test.com/1">One</a>
	<a href="http://test.com/2">Two</a>
	<a href="http://test.com/shoe">Buckle my shoe</a>
	
	<a href="./about">About this page</a>
	<a href="./careers#go">Join the team</a>
</body>
</html>`))
	}))
	t.Cleanup(server.Close)
	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	c, err := NewCrawler(server.URL)
	require.NoError(t, err)

	c.Skip = func(u *url.URL) bool {
		// we'll only care about fetching two urls, the base url and ./careers,
		// ignore the rest
		return u.String() != serverURL.String() &&
			u.String() != serverURL.JoinPath("./careers").String()
	}

	// keep track of callbacks to make sure we only do it twice
	called := 0
	c.Callback = func(page *url.URL, links []*url.URL) {
		called++

		require.NotNil(t, page)
		switch page.String() {
		case server.URL, serverURL.JoinPath("./careers").String():
			require.Len(t, links, 5)
			require.Equal(t, "http://test.com/1", links[0].String())
			require.Equal(t, "http://test.com/2", links[1].String())
			require.Equal(t, "http://test.com/shoe", links[2].String())
			require.Equal(t, serverURL.JoinPath("./about").String(), links[3].String())
			require.Equal(t, serverURL.JoinPath("./careers").String(), links[4].String())
		default:
			require.FailNow(t, "unexpected page was fetch", "unexpected callback call", "page", page.String())
		}
	}

	err = c.Crawl(context.Background())
	require.NoError(t, err)
	require.Equal(t, 2, called, "expected crawler callback to be called exactly twice, once for the base url and once for ./careers")
}
