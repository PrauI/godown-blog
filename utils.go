package godown

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

//go:embed assets/navbar.tmpl
var navbar []byte

// errors
var errPageExists = errors.New("page exists")

func init() {
	// Überprüfen, ob die eingebettete Datei korrekt geladen wurde
	if len(navbar) == 0 {
		panic("Fehler: Die Datei 'assets/navbar.tmpl' konnte nicht geladen werden. Stellen Sie sicher, dass der Pfad relativ zur go.mod-Datei korrekt ist.")
	}

	// Debug-Ausgabe des Inhalts der eingebetteten Datei
	fmt.Println("Inhalt von navbar.tmpl:")
	fmt.Println(string(navbar))
}

type article struct {
	path  string
	Name  string
	Pages []page
}

type page struct {
	Name        string
	input_path  string
	output_path string
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

		pageFiles, err := ioutil.ReadDir(input_dir + "/" + file.Name())
		if err != nil {
			return nil, err
		}

		for _, pageFile := range pageFiles {
			if !strings.HasSuffix(pageFile.Name(), ".md") {
				continue
			}
			pages = append(pages, page{Name: strings.TrimSuffix(pageFile.Name(), ".md"), input_path: input_dir + "/" + file.Name(), output_path: output_dir + "/" + file.Name()})
		}
		current_article := article{Name: file.Name(), Pages: pages, path: input_dir + "/" + file.Name()}
		err = current_article.Render(output_dir)
		if err != nil {
			return nil, err
		}
		articles = append(articles, current_article)

	}

	return articles, nil
}

// read markdown in article, render to html and put in output dir
func (article *article) Render(output string) error {
	// check if directory exists

	err := os.Mkdir(output+"/"+article.Name, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	// create sidebar
	sidebar, err := article.genSidebar()
	if err != nil {
		return err
	}

	for _, page := range article.Pages {
		err = page.Render(sidebar)
		if err != nil && !errors.Is(err, errPageExists) {
			return err
		}
	}

	return nil
}

func (article *article) genSidebar() (*bytes.Buffer, error) {
	ts, err := template.New("sidebar").Parse(string(navbar))
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

func (page *page) Render(sidebar *bytes.Buffer) error {
	// check if page output file already exists
	if _, err := os.Stat(page.output_path + "/" + page.Name + ".layout.tmpl"); !errors.Is(err, os.ErrNotExist) {
		// file already exist
		return errPageExists
	}
	// read file
	dat, err := os.ReadFile(page.input_path + "/" + page.Name + ".md")
	if err != nil {
		return err
	}
	prefix := []byte("{{template \"article-base\" .}}\n{{define \"main\"}}\n")
	html := mdToHTML(dat)
	suffix := []byte(`{{end}}{{define "sidebar"}}`)
	sufsuffix := []byte(`{{end}}`)
	data := append(append(append(append(prefix, html...), suffix...), sidebar.Bytes()...), sufsuffix...)
	// write file
	err = os.WriteFile(page.output_path+"/"+page.Name+".layout.tmpl", data, 0644)

	return err
}

func mdToHTML(md []byte) []byte {
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
