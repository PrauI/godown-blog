package godown

import (
	"bytes"
	_ "embed"
	"errors"
	"html/template"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

//go:embed assets/navbar.tmpl
var navbarTemplate []byte

// Errors
var ErrPageAlreadyExists = errors.New("page already exists")

func init() {
	// Ensure the embedded file is loaded correctly
	if len(navbarTemplate) == 0 {
		panic("Error: The file 'assets/navbar.tmpl' could not be loaded. Ensure the path is correct relative to the go.mod file.")
	}

}

type Article struct {
	Path  string
	Name  string
	Pages []Page
}

type Page struct {
	Name       string
	InputPath  string
	OutputPath string
}

// NewArticles initializes articles from the input directory and renders them to the output directory.
func NewArticles(inputDir string, outputDir string) ([]Article, error) {
	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return nil, err
	}

	var articles []Article

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		var pages []Page
		pageFiles, err := ioutil.ReadDir(inputDir + "/" + file.Name())
		if err != nil {
			return nil, err
		}

		for _, pageFile := range pageFiles {
			if !strings.HasSuffix(pageFile.Name(), ".md") {
				continue
			}
			pages = append(pages, Page{
				Name:       strings.TrimSuffix(pageFile.Name(), ".md"),
				InputPath:  inputDir + "/" + file.Name(),
				OutputPath: outputDir + "/" + file.Name(),
			})
		}

		currentArticle := Article{
			Name:  file.Name(),
			Pages: pages,
			Path:  inputDir + "/" + file.Name(),
		}

		err = currentArticle.Render(outputDir)
		if err != nil {
			return nil, err
		}
		articles = append(articles, currentArticle)
	}

	return articles, nil
}

// Render processes the markdown files in an article and generates HTML output.
func (article *Article) Render(output string) error {
	err := os.Mkdir(output+"/"+article.Name, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	sidebar, err := article.genSidebar()
	if err != nil {
		return err
	}

	for _, page := range article.Pages {
		err = page.Render(sidebar)
		if err != nil && !errors.Is(err, ErrPageAlreadyExists) {
			return err
		}
	}

	return nil
}

func (article *Article) genSidebar() (*bytes.Buffer, error) {
	ts, err := template.New("sidebar").Parse(string(navbarTemplate))
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = ts.Execute(buf, article)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (page *Page) Render(sidebar *bytes.Buffer) error {
	outputFile := page.OutputPath + "/" + page.Name + ".layout.tmpl"
	if _, err := os.Stat(outputFile); err == nil {
		return ErrPageAlreadyExists
	}

	markdownData, err := os.ReadFile(page.InputPath + "/" + page.Name + ".md")
	if err != nil {
		return err
	}

	prefix := []byte("{{template \"article-base\" .}}\n{{define \"main\"}}\n")
	html := mdToHTML_(markdownData)
	suffix := []byte(`{{end}}{{define "sidebar"}}`)
	sidebarSuffix := []byte(`{{end}}`)

	data := append(append(append(append(prefix, html...), suffix...), sidebar.Bytes()...), sidebarSuffix...)
	err = os.WriteFile(outputFile, data, 0644)

	return err
}

func mdToHTML_(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}
