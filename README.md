# My Crawler (Manuel's Crawler)

`mcrawl` is an CLI utility to crawl a website and print all its links.

Implemented in Go for performance, it's mainly a set of packages for 
bootstraping custom crawling behaviors. 

The CLI is a show case on how to use the packages and as such it might
not be suited for production use.

This project is meant to be extended for specific or personal usages.

```
NAME:
   My Crawler - Crawl a website and print links

USAGE:
   My Crawler [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --url value    Root URL from which to crawl
   --save         Persist crawl results to a file (default: true)
   --help, -h     show help
   --version, -v  print the version

```

## Usage

You can get the help text by opening a terminal in the project directory
and executing:

```
go run .
```

To first build the binary and then crawl a specific site, you do:

```
go build
./mcrawl --url=example.com
```

## CLI usage of the package

Here's what the CLI does, and how it uses the `crawl` package and other supporting packages.

- Given a starting `root URL`, the crawler visits each URL found. 
- The crawler can limit the visited URLs via the `Skip` function.The CLI behavior is to limit the search to the subdomain of the root URL. 
    - e.g.: providing *https://example.com/* as root url, it would crawl all pages 
    within example.com, but not follow external links, for example to facebook.com
    or even subdomain.example.com

And that's it. I added a bit of file-based persistence for good measure, but other
than that it's close to a blank canvas.

Further extensions could be added to increase new-project development speed.
For example:
- Provide the option to store the entire page content in the persistence mechanism.
- Add sqlite as optional persistence mechanism.

Looking forward to discuss whatever catches your eye on this repo or whatever improvement
you can bring to the table.
