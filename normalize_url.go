package main

import (
	"net/url"
	"strings"
)

func normalizeURL(rawUrl string) (string, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}
	return parsedUrl.Host + strings.TrimSuffix(parsedUrl.Path, "/"), err
}
