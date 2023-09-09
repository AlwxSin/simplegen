package codegen

import (
	"flag"
	"fmt"
	"github.com/AlwxSin/simplegen"
	"go/ast"
	"golang.org/x/tools/go/packages"
	"regexp"
	"strings"
)

var SorterTemplate = `
{{ range $key, $spec := .Specs }}
func {{$spec.Type.Name}}List{{$spec.Suffix}}SortByKeys(vs {{if not $spec.Type.IsSlice}}[]{{end}}{{$spec.Type}}, keys []{{$spec.FieldType}}) []{{$spec.Type}} {
	res := make([]{{$spec.Type}}, len(keys))
	for i, key := range keys {
		var appendable {{$spec.Type}}
		for _, v := range vs {
			if key == {{if $spec.FieldIsPtr}}*{{end}}v.{{$spec.FieldName}} {
				{{- if $spec.Type.IsSlice}}
				appendable = append(appendable, v)
				{{- else}}
				appendable = v
				break
				{{- end}}
			}
		}
		res[i] = appendable
	}
	return res
}

{{end}}
`

func Sorter(
	sg *simplegen.SimpleGenerator,
	pkg *packages.Package,
	node *ast.TypeSpec,
	comment *ast.Comment,
) (templateData simplegen.SpecData, imports []string, err error) {
	var (
		typeName  string
		fieldName string
		fieldType string
		suffix    string
	)

	fs := flag.FlagSet{}
	fs.StringVar(&typeName, "type", "", "Type for which need to generate")
	fs.StringVar(&suffix, "suffix", "", "Suffix for generated function")
	fs.StringVar(&fieldName, "fieldName", "ID", "Field name which be used as identifier")
	fs.StringVar(&fieldType, "fieldType", "int", "Field type")

	s := strings.TrimPrefix(comment.Text, "//")
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, simplegen.CmdKey+":sort-by-keys ")
	args := strings.Split(s, " ")

	err = fs.Parse(args)
	if err != nil {
		return nil, nil, err
	}

	parts := partsRe.FindStringSubmatch(typeName)
	if len(parts) != 4 {
		return nil, nil, fmt.Errorf("type must be in the form []*github.com/import/path.Name")
	}

	t := &goType{
		Modifiers:  parts[1],
		ImportPath: parts[2],
		Name:       strings.TrimPrefix(parts[3], "."),
	}

	if t.Name == "" {
		t.Name = t.ImportPath
		t.ImportPath = ""
	} else {
		imports = append(imports, t.ImportPath)
	}

	if t.ImportPath != "" {
		typePkg, err := sg.GetPackage(t.ImportPath)
		if err != nil {
			return nil, nil, err
		}
		t.ImportName = typePkg.Name
	}

	ssd := &SorterSpecData{
		Type:      t,
		Suffix:    suffix,
		FieldType: fieldType,
		FieldName: fieldName,
	}

	typePkg, err := sg.GetPackage(t.ImportPath)
	if err != nil {
		return nil, nil, err
	}

	structType, err := sg.GetStructType(typePkg, t.Name)
	if err != nil {
		return nil, nil, err
	}

	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		if field.Name() != fieldName {
			continue
		}
		if field.Type().String()[0] == '*' {
			ssd.FieldIsPtr = true
		}
		break
	}
	return ssd, imports, nil
}

type SorterSpecData struct {
	Type       *goType
	FieldType  string
	FieldName  string
	Suffix     string
	FieldIsPtr bool
}

type goType struct {
	Modifiers  string
	ImportPath string
	ImportName string
	Name       string
}

func (t *goType) IsSlice() bool {
	return strings.HasPrefix(t.Modifiers, "[]")
}

func (t *goType) String() string {
	if t.ImportName != "" {
		return t.Modifiers + t.ImportName + "." + t.Name
	}

	return t.Modifiers + t.Name
}

var partsRe = regexp.MustCompile(`^([\[\]\*]*)(.*?)(\.\w*)?$`)
