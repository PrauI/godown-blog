package godown

import (
	"io/ioutil"
	"os"
	"strings"
)

type article struct {
	name  string
	pages []page
}

type page struct {
	name string
}

// input_dir: path to input directory, where markdowns are
// output_dir: path to output dir where tmpl should be placed
func New(input_dir string, output_dir string) ([]article, error) {
	files, err := ioutil.ReadDir(input_dir)
	if err != nil {
		return nil, err
	}

	var articles []article

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		// get all the pages of an article

		var pages []page

		pageFiles, err := ioutil.ReadDir(file.Name())
		if err != nil {
			return nil, err
		}

		for _, pageFile := range pageFiles {
			if !strings.HasSuffix(pageFile.Name(), ".md") {
				continue
			}
			pages = append(pages, page{name: strings.TrimSuffix(pageFile.Name(), ".md")})
		}
		articles = append(articles, article{name: file.Name(), pages: pages})
	}

	return articles, nil
}

// read markdown in article, render to html and put in output dir
func (article *article) Render(output string) error {
	err := os.Mkdir(output+"/"+article.name, 0755)
	if err != nil {
		return err
	}

	return nil
}
