package godown

import (
	"fmt"
	"io/ioutil"
)

// input_dir: path to input directory, where markdowns are
// output_dir: path to output dir where tmpl should be placed
func New(input_dir string, output_dir string) error {
	files, err := ioutil.ReadDir(input_dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		fmt.Println(file.Name())
	}
	return nil
}
