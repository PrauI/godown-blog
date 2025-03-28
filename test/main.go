package main

import (
	"fmt"

	"github.com/PrauI/godown-blog"
)

func main() {
	articles, err := godown.New("test/input", "test/output")

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, article := range articles {
		fmt.Println(article.name)
		for _, page := range article.pages {
			fmt.Println(page.name)
		}
	}
}
