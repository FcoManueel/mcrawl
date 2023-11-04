package fetch

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"mcrawl/urlparse"
)

type Client struct {
	c   *http.Client
	sem chan struct{}
}

func NewClient(c *http.Client, maxWorkers int) *Client {
	semaphore := make(chan struct{}, maxWorkers)
	return &Client{
		c:   c,
		sem: semaphore,
	}
}
func (c *Client) LinkedURLs(ctx context.Context, u *url.URL) ([]*url.URL, error) {
	// limit maximum amount of concurrent fetch calls
	c.sem <- struct{}{}
	defer func() {
		<-c.sem
	}()

	// create request with context
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	// fetch page
	resp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// parse linked pages
	var links []*url.URL
	tokenizer := html.NewTokenizer(resp.Body)
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken: // finished parsing the page
			return links, nil
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						href, err := url.Parse(attr.Val)
						if err != nil {
							return links, fmt.Errorf("error parsing href: %v", err)
						}
						href = u.ResolveReference(href) // in case href is relative
						// skip non-http(s) links (e.g. mailto, ftp, etc)
						if href.Scheme != "" && href.Scheme != "http" && href.Scheme != "https" {
							continue
						}
						// skip irrelevant files
						if strings.HasSuffix(href.Path, ".pdf") {
							continue
						}

						href, err = urlparse.Normalize(href.String())
						if err != nil {
							return links, fmt.Errorf("error parsing href: %v", err)
						}
						links = append(links, href)
					}
				}
			}
		}
	}
}
