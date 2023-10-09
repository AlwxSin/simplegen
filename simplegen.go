/*
Package simplegen helps developer to generate code.

For codegen more complex than just generate small piece of code
usually need to parse source code and build ast.Tree.
Simplegen will take this part and give developer exact ast.Node to work with.

Also, simplegen provides tools for easy finding ast.Node.

Example:

	package main

	import (
		"flag"
		"fmt"
		"go/ast"
		"golang.org/x/tools/go/packages"

		"github.com/AlwxSin/simplegen"
	)

	var PaginatorTemplate = `
	{{ range $key, $struct := .Specs }}
	// {{$struct.Name}}ListPaginated represents {{$struct.Name}} list in a pagination container.
	type {{$struct.Name}}ListPaginated struct {
		CurrentCursor *string ` + "`json:\"currentCursor\"`\n" +
		`	NextCursor    *string ` + "`json:\"nextCursor\"`\n" +
		`	Results       []*{{$struct.Name}} ` + "`json:\"results\"`\n" +
		`
		isPaginated bool
		limit       int
		offset      int
	}
	`

	func Paginator(
		sg *simplegen.SimpleGenerator,
		pkg *packages.Package,
		node *ast.TypeSpec,
		comment *ast.Comment,
	) (templateData simplegen.SpecData, imports []string, err error) {
		imports = append(imports, "strconv")

		type PaginatorTypeSpec struct {
			Name string
		}

		tmplData := &PaginatorTypeSpec{
			Name: node.Name.Name,
		}
		return simplegen.SpecData(tmplData), imports, nil
	}

	// simplegen:paginator
	type User struct {
		Email     string
	}

	func main() {
		var pn simplegen.PackageNames

		flag.Var(&pn, "package", "Package where simplegen should find magic comments")
		flag.Parse()

		sg, err := simplegen.NewSimpleGenerator(pn, simplegen.GeneratorsMap{
			"paginator": simplegen.TemplateGenerator{
				Template:      PaginatorTemplate,
				GeneratorFunc: Paginator,
			},
		}, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = sg.Generate()
		if err != nil {
			fmt.Println(err)
		}
	}

In result simplegen will generate file with following content
Example:

	// UserListPaginated represents User list in a pagination container.
	type UserListPaginated struct {
		CurrentCursor *string `json:"currentCursor"`
		NextCursor    *string `json:"nextCursor"`
		Results       []*User `json:"results"`

		isPaginated bool
		limit       int
		offset      int
	}

See /examples dir for more detailed usage.
*/
package simplegen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"os"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

const packagesLoadMode = packages.NeedName |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo |
	packages.NeedImports

type SimpleGenerator struct {
	// pkgs collects all used packages for easy use
	pkgs map[pkgPath]*packages.Package

	generators GeneratorsMap
	cmdData    map[GeneratorName]map[*packages.Package]*cmdData

	tmplFuncMap template.FuncMap
}

func NewSimpleGenerator(pkgNames PackageNames, generators GeneratorsMap, tmplFuncMap template.FuncMap) (*SimpleGenerator, error) {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()
	cfg := &packages.Config{Fset: fset, Mode: packagesLoadMode, Dir: dir}
	pkgs, err := packages.Load(cfg,
		pkgNames...,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load packages %s: %w", pkgNames, err)
	}

	errors := sgErrors{}

	sg := &SimpleGenerator{
		generators:  generators,
		pkgs:        make(map[pkgPath]*packages.Package),
		cmdData:     make(map[GeneratorName]map[*packages.Package]*cmdData),
		tmplFuncMap: tmplFuncMap,
	}
	for _, pkg := range pkgs {
		sg.pkgs[pkgPath(pkg.PkgPath)] = pkg
	}

	if len(errors) > 0 {
		return nil, errors
	}

	return sg, nil
}

func (sg *SimpleGenerator) Generate() error {
	errors := sgErrors{}

	// first, inspect ast of loaded packages to find
	for _, pkg := range sg.pkgs {
		for _, fileAst := range pkg.Syntax {
			ast.Inspect(fileAst, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.GenDecl:
					copyGenDeclCommentsToSpecs(node)
				case *ast.TypeSpec:
					if node.Doc == nil {
						return true
					}
					for _, comment := range node.Doc.List {
						if strings.Contains(comment.Text, CmdKey) {
							for cmd, generator := range sg.generators {
								if strings.Contains(comment.Text, string(cmd)) {
									err := sg.add(cmd, pkg, node, comment, generator.GeneratorFunc)
									if err != nil {
										errors = append(errors, err)
									}
								}
							}
						}
					}
				}
				return true
			})
		}
	}
	if len(errors) > 0 {
		return errors
	}

	// second, write collected specs to files
	return sg.write()
}

func (sg *SimpleGenerator) add(
	genName GeneratorName,
	pkg *packages.Package,
	node *ast.TypeSpec,
	comment *ast.Comment,
	genFunc GeneratorFunc,
) error {
	if _, ok := sg.cmdData[genName]; !ok {
		sg.cmdData[genName] = make(map[*packages.Package]*cmdData)
	}

	templateData, rawImports, err := genFunc(sg, pkg, node, comment)
	if err != nil {
		return err
	}

	importsMap := map[string]struct{}{}
	var imports []string
	for _, rawImp := range rawImports {
		if _, ok := importsMap[rawImp]; !ok {
			importsMap[rawImp] = struct{}{}
			imports = append(imports, rawImp)
		}
	}

	_, ok := sg.cmdData[genName][pkg]
	if !ok {
		sg.cmdData[genName][pkg] = newGeneratorData(pkg.Name, imports)
	}

	sg.cmdData[genName][pkg].add(templateData)
	return nil
}

func (sg *SimpleGenerator) write() error {
	errors := sgErrors{}

	for genName, genData := range sg.cmdData {
		cmdTemplate := sg.generators[genName].Template

		templateRaw := header + cmdTemplate
		tmpl := template.Must(template.New("").Funcs(sg.tmplFuncMap).Parse(templateRaw))

		for pkg, specs := range genData {
			buf := bytes.Buffer{}

			if err := tmpl.Execute(&buf, specs); err != nil {
				return err
			}

			content, err := format.Source(buf.Bytes())
			if err != nil {
				errors = append(errors, err)
				continue
			}
			parts := strings.SplitN(pkg.PkgPath, "/", 2)

			fName := fmt.Sprintf("/%s_gen.go", genName)
			err = writeFile(parts[1]+fName, content)
			if err != nil {
				errors = append(errors, err)
			}
		}
	}
	if len(errors) > 0 {
		return errors
	}
	return nil
}

// GetPackage returns packages.Package. It tries to load package if it didn't load before.
func (sg *SimpleGenerator) GetPackage(path string) (*packages.Package, error) {
	pkg, ok := sg.pkgs[pkgPath(path)]
	if !ok {
		pkgs, err := packages.Load(&packages.Config{Mode: packagesLoadMode}, path)
		if err != nil {
			return nil, err
		}
		if len(pkgs) != 1 {
			return nil, fmt.Errorf("too many packages found for path: %s", path)
		}
		pkg = pkgs[0]
		sg.pkgs[pkgPath(path)] = pkg
	}
	return pkg, nil
}

// GetObject tries to find type object in given package.
// In most cases you don't need it, use GetStructType instead.
func (sg *SimpleGenerator) GetObject(pkg *packages.Package, typeName string) (types.Object, error) {
	obj := pkg.Types.Scope().Lookup(typeName)
	if obj == nil {
		return nil, fmt.Errorf("%s not found in declared types of %s",
			typeName, pkg)
	}

	// check if it is a declared type
	if _, ok := obj.(*types.TypeName); !ok {
		return nil, fmt.Errorf("%v is not a named type", obj)
	}
	return obj, nil
}

// GetStructType tries to find type struct in given package.
func (sg *SimpleGenerator) GetStructType(pkg *packages.Package, typeName string) (*types.Struct, error) {
	obj, err := sg.GetObject(pkg, typeName)
	if err != nil {
		return nil, err
	}

	// expect the underlying type to be a struct
	structType, ok := obj.Type().Underlying().(*types.Struct)
	if !ok {
		return nil, fmt.Errorf("type %v is not a struct", obj)
	}

	return structType, nil
}

// writeFile (re)creates a new file and writes content into it.
func writeFile(fileName string, fileContent []byte) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(fileContent)
	return err
}

// copyDocsToSpecs will take the GenDecl level documents and copy them
// to the children Type and Value specs.  I think this is actually working
// around a bug in the AST, but it works for now.
func copyGenDeclCommentsToSpecs(x *ast.GenDecl) {
	// Copy the doc spec to the type or value spec
	// cause they missed this... whoops
	if x.Doc != nil {
		for _, spec := range x.Specs {
			if s, ok := spec.(*ast.TypeSpec); ok {
				if s.Doc == nil {
					s.Doc = x.Doc
				}
			}
		}
	}
}
