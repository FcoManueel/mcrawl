package fetch

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFetch(t *testing.T) {
	start := time.Now()
	// test fetch client with a load bigger than its max worker count
	maxWorkers := 20
	concurrentFetches := 5 * maxWorkers
	c := NewClient(http.DefaultClient, maxWorkers)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// sleep to test that the client executes things in parallel
		// if client were to execute fetches sequentially this test would take
		// more than 10 seconds to complete (100ms * (20*5) fetches)
		time.Sleep(100 * time.Millisecond)

		w.WriteHeader(200)
		w.Write([]byte(`<html>
<head></head>
<body>
	<h1>Page</h1>
	<a href="http://test.com/1">One</a>
	<a href="http://test.com/2">Two</a>
	Some mal<for<mation>> <br /> <illegal /> <html> <h2>
	<a href="http://test.com/shoe">Buckle my shoe</a>
	
	<a href="./about">About this page</a>
	<a href="./careers#go">Join the team</a>
	
	<a href="mailto:help@test.com">Email support</a>
</body>
</html>`))
	}))

	t.Cleanup(func() {
		server.Close()
		require.WithinDuration(t, time.Now(), start, 10*time.Second, "expected test to take less than 10 seconds")
	})

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	for i := 0; i < concurrentFetches; i++ {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()

			urls, err := c.LinkedURLs(context.Background(), serverURL)
			require.NoError(t, err)
			require.NotNil(t, urls)
			require.Len(t, urls, 5)

			require.Equal(t, "http://test.com/1", urls[0].String())
			require.Equal(t, "http://test.com/2", urls[1].String())
			require.Equal(t, "http://test.com/shoe", urls[2].String())
			require.Equal(t, serverURL.JoinPath("./about").String(), urls[3].String(), "should resolve relative link")
			require.Equal(t, serverURL.JoinPath("./careers").String(), urls[4].String(), "should remove fragment")
		})
	}
}
