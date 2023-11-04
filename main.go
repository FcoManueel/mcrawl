package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
	"mcrawl/crawl"
	"mcrawl/persist"
)

func main() {
	slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	var url string
	var save bool

	app := &cli.App{
		Name:    "My Crawler",
		Version: "1.0.0",
		Usage:   "Crawl a website and print links",
		Action: func(ctx *cli.Context) error {
			err := crawlerApp(ctx.Context, url, save)
			if err != nil {
				log.Println(err)
			}
			return err
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "url",
				Usage:       "Root URL from which to crawl",
				Required:    true,
				Destination: &url,
			},
			&cli.BoolFlag{
				Name:        "save",
				Usage:       "Persist crawl results to a file",
				Value:       true,
				Destination: &save,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func crawlerApp(ctx context.Context, root string, save bool) error {
	start := time.Now()

	// create a new crawler
	c, err := crawl.NewCrawler(root)
	if err != nil {
		return err
	}

	// handle persistence
	var persistence persist.Persistence
	if save {
		// start connection with persistence mechanism
		persistence = persist.Open(c.Root().String())

		// print path to persistence file
		defer func() {
			fmt.Println("Results stored at", persistence.Filename())
			persistence.Close()
		}()
	}

	// limit crawling to a single subdomain
	c.Skip = func(u *url.URL) bool {
		return u.Hostname() != c.Root().Hostname()
	}

	// handle each crawled page and the links it points to.
	// use a mutex to prevent printing interleaved results
	printMutex := &sync.Mutex{}
	c.Callback = func(page *url.URL, links []*url.URL) {
		printMutex.Lock()
		defer printMutex.Unlock()

		fmt.Println(page)
		if len(links) > 0 {
			for _, link := range links[:len(links)-1] {
				fmt.Println("├──", link)
			}
			fmt.Println("└──", links[len(links)-1])
		}
		fmt.Printf("%d links found on page (%q)\n\n", len(links), page.String())

		// persist the results for future reference
		persistence.Save(page, links)
	}

	err = c.Crawl(ctx)
	fmt.Printf("Finished crawling after %v. Have a nice day!\n", time.Now().Sub(start).String())
	return err
}
