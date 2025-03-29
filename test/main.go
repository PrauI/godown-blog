package main

import (
	"fmt"

	"github.com/PrauI/godown-blog"
)

func main() {
	_, err := godown.New("test/input", "test/output")

	if err != nil {
		fmt.Println(err)
		return
	}

}
