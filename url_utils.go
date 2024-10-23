package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
	htmlReader := strings.NewReader(htmlBody)
	doc, err := html.Parse(htmlReader)
	if err != nil {
		return []string{}, err
	}
	var traverse func(*html.Node)
	var links []string
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			// Find href attribute
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					links = append(links, attr.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	for i, link := range links {
		url, _ := url.Parse(link)
		if url.Host == "" {
			links[i] = rawBaseURL + link
		}
	}
	return links, nil
}

func getHTML(rawURL string) (string, error) {
	res, err := http.Get(rawURL)
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
