package main

import (
	"examples/codegen"
	"flag"
	"fmt"
	"github.com/AlwxSin/simplegen"
	"text/template"
)

// main
// Example
// go run main.go -package examples/my_project/models -package examples/my_project/responses
func main() {
	var (
		help bool
		pn   simplegen.PackageNames
	)

	flag.BoolVar(&help, "h", false, "Show this help text")
	flag.BoolVar(&help, "help", false, "")
	flag.Var(&pn, "package", "Package where simplegen should find magic comments")
	flag.Parse()

	if help {
		flag.PrintDefaults()
		return
	}

	pn = simplegen.PackageNames{"examples/my_project/models", "examples/my_project/responses"}

	sg, err := simplegen.NewSimpleGenerator(pn, simplegen.GeneratorsMap{
		"paginator": simplegen.TemplateGenerator{
			Template:      codegen.PaginatorTemplate,
			GeneratorFunc: codegen.Paginator,
		},
		"settable-input": simplegen.TemplateGenerator{
			Template:      codegen.SettableTemplate,
			GeneratorFunc: codegen.Settable,
		},
		"sort-by-keys": simplegen.TemplateGenerator{
			Template:      codegen.SorterTemplate,
			GeneratorFunc: codegen.Sorter,
		},
	}, template.FuncMap{
		"formatSettableTags": codegen.FormatSettableTags,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	err = sg.Generate()
	if err != nil {
		fmt.Println(err)
	}
}
