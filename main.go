package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) < 1 {
		fmt.Println("no website provided")
		os.Exit(1)
	}

	maxConcurrency := 3
	maxPages := 10
	if len(argsWithoutProg) > 1 {
		conc, err := strconv.Atoi(argsWithoutProg[1])
		if err != nil {
			fmt.Printf("Error - configure: %v", err)
			return
		}
		maxConcurrency = conc
	}
	if len(argsWithoutProg) > 2 {
		maxP, err := strconv.Atoi(argsWithoutProg[2])
		if err != nil {
			fmt.Printf("Error - configure: %v", err)
			return
		}
		maxPages = maxP
	}
	rawBaseURL := argsWithoutProg[0]
	fmt.Println("starting crawl of: " + rawBaseURL)
	cfg, err := configure(rawBaseURL, maxConcurrency, maxPages)
	if err != nil {
		fmt.Printf("Error - configure: %v", err)
		return
	}

	fmt.Printf("starting crawl of: %s...\n", rawBaseURL)

	cfg.wg.Add(1)
	go cfg.crawlPage(rawBaseURL)
	cfg.wg.Wait()

	printReport(cfg.pages, rawBaseURL)

	os.Exit(0)
}
