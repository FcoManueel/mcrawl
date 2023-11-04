package crawl

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"mcrawl/fetch"
	"mcrawl/urlparse"
)

const MAX_CRAWLERS = 200
const MAX_FETCHERS = 100

type crawler struct {
	// Callback is called for each crawled page, including a list of the pages it points to.
	Callback func(page *url.URL, links []*url.URL)
	// Skip is run before fetching a url. When true the url is skipped without being fetched.
	Skip func(u *url.URL) bool

	root         *url.URL
	queue        chan *url.URL
	visitedCache interface {
		LoadOrStore(key, value any) (actual any, loaded bool)
	}
	errs chan error

	fetcher *fetch.Client
	wg      *sync.WaitGroup
}

func NewCrawler(rootURL string) (*crawler, error) {
	root, err := urlparse.Normalize(rootURL)
	if err != nil {
		return nil, err
	}

	// http client used for fetch requests
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = MAX_FETCHERS
	t.MaxConnsPerHost = MAX_FETCHERS
	t.MaxIdleConnsPerHost = MAX_FETCHERS
	httpClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: t,
	}

	return &crawler{
		root:         root,
		queue:        make(chan *url.URL, MAX_CRAWLERS),
		visitedCache: &sync.Map{},
		errs:         make(chan error, 20),

		fetcher: fetch.NewClient(httpClient, MAX_FETCHERS),
		wg:      &sync.WaitGroup{},
	}, nil
}

func (c *crawler) Root() *url.URL {
	return c.root
}

func (c *crawler) Crawl(ctx context.Context) error {
	if c.root == nil {
		return errors.New("no page to crawl, make sure to call the 'NewCrawler' constructor with a valid url")
	}
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		c.enqueue(c.root)
		c.wg.Wait()
		cancel()
	}()

out:
	for {
		select {
		case <-ctx.Done():
			break out
		case err := <-c.errs:
			fmt.Println("AN ERROR OCCURRED:", err.Error())
			// bad urls exists, errors are expected, we can't stop the whole
			// thing because of one. Ideally we would enqueue the page so that
			// we can at least retry it later, so let's pretend we did that
			// instead of blatantly ignoring it.
			continue
		case page, ok := <-c.queue:
			if !ok {
				break out
			}
			go c.crawl(ctx, page)
		}
	}
	return nil
}

func (c *crawler) crawl(ctx context.Context, page *url.URL) {
	defer c.wg.Done()

	if c.skip(page) {
		return
	}
	links, err := c.fetcher.LinkedURLs(ctx, page)
	if err != nil {
		c.errs <- fmt.Errorf("while fetching page (%s) links: %v", page.String(), err)
	}

	c.Callback(page, links)

	for _, link := range links {
		c.enqueue(link)
	}
}

func (c *crawler) enqueue(page *url.URL) {
	c.wg.Add(1) // This will be resolved insider c.crawl
	c.queue <- page
}

func (c *crawler) cacheKey(page string) string {
	// drop schema from the cache entry, to prevent from visiting twice a page with http and https
	return strings.TrimPrefix(strings.TrimPrefix(page, "https://"), "http://")
}

func (c *crawler) skip(page *url.URL) bool {
	if c.Skip != nil && c.Skip(page) {
		return true
	}
	_, alreadyVisited := c.visitedCache.LoadOrStore(c.cacheKey(page.String()), true)
	return alreadyVisited
}
