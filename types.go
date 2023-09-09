package simplegen

import (
	"errors"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/packages"
)

const CmdKey = "simplegen"

type sgErrors []error

func (e sgErrors) Error() string {
	return errors.Join(e...).Error()
}

type pkgPath string

type GeneratorName string

// GeneratorFunc for generating template data from ast nodes.
// SimpleGenerator calls it with:
// sg -> SimpleGenerator instance (for useful methods, look into docs or examples to know how to use them).
// pkg -> packages.Package where magic comment was found.
// node -> ast.TypeSpec struct annotated with magic comment.
// comment -> ast.Comment magic comment itself.
type GeneratorFunc func(
	sg *SimpleGenerator,
	pkg *packages.Package,
	node *ast.TypeSpec,
	comment *ast.Comment,
) (templateData SpecData, imports []string, err error)

// TemplateGenerator contains raw template and GeneratorFunc to generate template data.
type TemplateGenerator struct {
	// Template is a string which contains full template in go style
	Template      string
	GeneratorFunc GeneratorFunc
}

// GeneratorsMap cmd_name -> func_to_generate_template_data
//
//	{
//	  "paginator": GeneratePaginatorData,
//	  "sorter": GenerateSorterData,
//	}
type GeneratorsMap map[GeneratorName]TemplateGenerator

// SpecData can be any struct. Will pass it to template.
type SpecData any

// cmdData internal struct. Will pass it to template.
type cmdData struct {
	PackageName string
	Imports     []string

	Specs []SpecData
}

func newGeneratorData(pkgName string, imports []string) *cmdData {
	return &cmdData{
		PackageName: pkgName,
		Imports:     imports,
		Specs:       make([]SpecData, 0),
	}
}

func (gd *cmdData) add(sd SpecData) {
	gd.Specs = append(gd.Specs, sd)
}

// PackageNames is a helper for flag.Parse
// Example:
// flag.Var(&pn, "package", "Package where simplegen should find magic comments").
type PackageNames []string

func (pn PackageNames) String() string {
	return strings.Join(pn, ",")
}

func (pn *PackageNames) Set(value string) error {
	*pn = append(*pn, value)
	return nil
}
