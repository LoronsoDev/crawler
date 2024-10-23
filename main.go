package main

import (
	"fmt"
	"os"
)

func main() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) < 1 {
		fmt.Println("no website provided")
		os.Exit(1)
	}
	if len(argsWithoutProg) > 1 {
		fmt.Println("too many arguments provided")
		os.Exit(1)
	}
	url := argsWithoutProg[0]
	fmt.Println("starting crawl of: " + url)
	content, err := getHTML(url)
	if err != nil {
		fmt.Println("error while fetching url: " + err.Error())
		os.Exit(1)
	}
	fmt.Println(content)
	os.Exit(0)
}
