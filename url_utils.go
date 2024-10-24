package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

func getURLsFromHTML(htmlBody string, baseURL *url.URL) ([]string, error) {
	htmlReader := strings.NewReader(htmlBody)
	doc, err := html.Parse(htmlReader)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse HTML: %v", err)
	}

	var urls []string
	var traverseNodes func(*html.Node)
	traverseNodes = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, anchor := range node.Attr {
				if anchor.Key == "href" {
					href, err := url.Parse(anchor.Val)
					if err != nil {
						fmt.Printf("couldn't parse href '%v': %v\n", anchor.Val, err)
						continue
					}

					resolvedURL := baseURL.ResolveReference(href)
					urls = append(urls, resolvedURL.String())
				}
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			traverseNodes(child)
		}
	}
	traverseNodes(doc)

	return urls, nil
}

func getHTML(rawURL string) (string, error) {
	res, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if err != nil {
		return "", err
	}
	if res.StatusCode >= 400 {
		return "", errors.New(rawURL + " returned error ")
	}
	if !strings.Contains(res.Header.Get("content-type"), "text/html") {
		return "", errors.New(rawURL + " returned " + res.Header.Get("content-type") + " instead of content-type: text/html")
	}
	bodyHTML, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}
	return string(bodyHTML), err
}

type config struct {
	pages              map[string]int
	baseURL            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	maxPages           int
}

func (cfg *config) addPageVisit(normalizedURL string) (isFirst bool) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if _, visited := cfg.pages[normalizedURL]; visited {
		cfg.pages[normalizedURL]++
		return false
	}

	cfg.pages[normalizedURL] = 1
	return true
}

func configure(rawBaseURL string, maxConcurrency, maxPages int) (*config, error) {
	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse base URL: %v", err)
	}

	return &config{
		pages:              make(map[string]int),
		baseURL:            baseURL,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxConcurrency),
		maxPages:           maxPages,
		wg:                 &sync.WaitGroup{},
	}, nil
}

func (cfg *config) crawlPage(rawCurrentURL string) {
	cfg.mu.Lock()
	if len(cfg.pages) >= cfg.maxPages {
		cfg.mu.Unlock()
		cfg.wg.Done()
		return
	}
	cfg.mu.Unlock()

	cfg.concurrencyControl <- struct{}{}
	defer func() {
		<-cfg.concurrencyControl
		cfg.wg.Done()
	}()

	currentURL, err := url.Parse(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error - crawlPage: couldn't parse URL '%s': %v\n", rawCurrentURL, err)
		return
	}

	// skip other websites
	if currentURL.Hostname() != cfg.baseURL.Hostname() {
		return
	}

	normalizedURL, err := normalizeURL(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error - normalizedURL: %v", err)
		return
	}

	isFirst := cfg.addPageVisit(normalizedURL)
	if !isFirst {
		return
	}

	fmt.Printf("crawling %s\n", rawCurrentURL)

	htmlBody, err := getHTML(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error - getHTML: %v", err)
		return
	}

	nextURLs, err := getURLsFromHTML(htmlBody, cfg.baseURL)
	if err != nil {
		fmt.Printf("Error - getURLsFromHTML: %v", err)
		return
	}

	for _, nextURL := range nextURLs {
		cfg.wg.Add(1)
		go cfg.crawlPage(nextURL)
	}
}

type Page struct {
	URL   string
	Count int
}

func printReport(pages map[string]int, baseURL string) {
	fmt.Printf(`
	=============================
	  REPORT for %s
	=============================
	`, baseURL)

	sortedPages := sortPages(pages)
	for _, page := range sortedPages {
		url := page.URL
		count := page.Count
		fmt.Printf("Found %d internal links to %s\n", count, url)
	}
}

func sortPages(pages map[string]int) []Page {
	pagesSlice := []Page{}
	for url, count := range pages {
		pagesSlice = append(pagesSlice, Page{URL: url, Count: count})
	}
	sort.Slice(pagesSlice, func(i, j int) bool {
		if pagesSlice[i].Count == pagesSlice[j].Count {
			return pagesSlice[i].URL < pagesSlice[j].URL
		}
		return pagesSlice[i].Count > pagesSlice[j].Count
	})
	return pagesSlice
}
