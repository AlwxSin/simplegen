package codegen

import (
	"fmt"
	"github.com/AlwxSin/simplegen"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"strconv"
	"strings"
)

var SettableTemplate = `
// Settable acts like sql.NullString, sql.NullInt64 but generic.
// It allows to define was value set or it's zero value.
type Settable[T any] struct {
	Value T
	IsSet bool
}

// NewSettable returns set value.
func NewSettable[T any](value T) Settable[T] {
	return Settable[T]{
		Value: value,
		IsSet: true,
	}
}

{{ range $index, $struct := .Specs }}
// {{$struct.Name}}Settable allows to use {{$struct.Name}} with Settable fields 
type {{$struct.Name}}Settable struct {
	{{- range $index, $field := $struct.Fields }}
	{{$field.Name}} Settable[{{$field.TypeName}}] {{formatSettableTags $field.Tags}}
	{{- end }}
}

func (inp *{{$struct.Name}}) ToSettable(inputFields map[string]interface{}) *{{$struct.Name}}Settable {
	settable := &{{$struct.Name}}Settable{}
	{{ range $index, $field := $struct.Fields }}
	if _, ok := inputFields["{{$field.JSONTag}}"]; ok {
		settable.{{$field.Name}} = NewSettable(inp.{{$field.Name}})
	}
	{{end}}
	return settable
}

{{ end }}
`

type InputField struct {
	Name, TypeName, Tags, JSONTag string
}

type InputToSettableSpecData struct {
	Name   string
	Fields []*InputField
}

func Settable(
	sg *simplegen.SimpleGenerator,
	pkg *packages.Package,
	node *ast.TypeSpec,
	comment *ast.Comment,
) (templateData simplegen.SpecData, imports []string, err error) {
	return parseSettableStruct(sg, pkg, node.Name.Name)
}

func parseSettableStruct(
	sg *simplegen.SimpleGenerator,
	pkg *packages.Package,
	structName string,
) (specData *InputToSettableSpecData, imports []string, err error) {
	structType, err := sg.GetStructType(pkg, structName)
	if err != nil {
		return nil, nil, err
	}

	specData = &InputToSettableSpecData{Name: structName}

	// iterate over struct fields
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)

		// check if field is embedded
		// if so, extract field type from package
		// WARNING, there is no check if field type from another package, this is just an example
		if field.Embedded() {
			// field is embedded (usually common fields like ID or CreatedAt), go recursive
			embedInput, embedImports, tagErr := parseSettableStruct(sg, pkg, field.Name())
			if tagErr != nil {
				return nil, nil, tagErr
			}
			if len(embedInput.Fields) > 0 {
				specData.Fields = append(specData.Fields, embedInput.Fields...)
			}
			if len(embedImports) > 0 {
				imports = append(imports, embedImports...)
			}
			continue
		}

		// extract tags to use in settable structure
		tagValue := structType.Tag(i)
		jsonTagValue := parseJSONOrYamlTag(tagValue)
		if jsonTagValue == "" {
			return nil, nil, fmt.Errorf("type %v: field %s should has json/yaml value", structType, field.Name())
		}

		// store field type to use in settable type and collect all imports
		fieldType, fieldImports := extractFieldInfo(field.Type(), pkg)
		if fieldType != "" {
			specData.Fields = append(specData.Fields, &InputField{
				Name:     field.Name(),
				TypeName: fieldType,
				Tags:     tagValue,
				JSONTag:  jsonTagValue,
			})
		}
		if len(fieldImports) > 0 {

			imports = append(imports, fieldImports...)
		}
	}

	return specData, imports, nil
}

// extractFieldInfo returns type and imports
// time.Time, []string{time}
// models.User, []string{examples/my_project/models}
// *models.User, []string{examples/my_project/models}
// string, []string{}.
func extractFieldInfo(field types.Type, curPkg *packages.Package) (fieldTypeName string, fieldImports []string) {
	switch v := field.(type) {
	case *types.Named:
		typeName, fieldImport := getFieldInfo(v, curPkg)
		var imports []string
		if fieldImport != "" {
			imports = []string{fieldImport}
		}
		return typeName, imports
	case *types.Basic:
		return v.Name(), []string{}
	case *types.Interface:
		return v.String(), []string{}
	case *types.Pointer:
		typeName, imports := extractFieldInfo(v.Elem(), curPkg)
		return fmt.Sprintf("*%s", typeName), imports
	case *types.Slice:
		typeName, imports := extractFieldInfo(v.Elem(), curPkg)
		return fmt.Sprintf("[]%s", typeName), imports
	case *types.Map:
		keyInfo, keyImports := extractFieldInfo(v.Key(), curPkg)
		elemInfo, elemImports := extractFieldInfo(v.Elem(), curPkg)
		return fmt.Sprintf("map[%s]%s", keyInfo, elemInfo), append(keyImports, elemImports...)
	default:
		return "", nil
	}
}

// getFieldInfo returns type definition and import path related to package
// "models.User", "examples/my_project/models".
func getFieldInfo(field *types.Named, pkg *packages.Package) (fieldTypeName, fieldImport string) {
	if field.Obj().Pkg().Path() == pkg.PkgPath {
		return field.Obj().Name(), ""
	}
	return field.Obj().Pkg().Name() + "." + field.Obj().Name(), field.Obj().Pkg().Path()
}

// parseJSONOrYamlTag extract json/yaml value from tags
// `yaml:"phoneYaml" json:"phone"` - > "phone"
// `yaml:"phoneYaml"` - > "phoneYaml"
// `db:"phone"` - > "".
func parseJSONOrYamlTag(rawTags string) string {
	if rawTags == "" {
		return ""
	}
	tags := strings.Split(rawTags, " ")

	jsonTag := ""
	yamlTag := ""
	for _, tag := range tags {
		hasJSONTag := strings.Contains(tag, "json:")
		hasYamlTag := strings.Contains(tag, "yaml:")
		if !hasJSONTag && !hasYamlTag {
			continue
		}
		if hasJSONTag {
			t, qErr := strconv.Unquote(tag[5:])
			if qErr != nil {
				return ""
			}
			jsonTag = strings.Split(t, ",")[0]
			continue
		}
		if hasYamlTag {
			t, qErr := strconv.Unquote(tag[5:])
			if qErr != nil {
				return ""
			}
			yamlTag = strings.Split(t, ",")[0]
			continue
		}
	}
	if jsonTag != "" && jsonTag != "-" {
		return jsonTag
	}
	if yamlTag != "" && yamlTag != "-" {
		return yamlTag
	}
	return ""
}

func FormatSettableTags(tag string) string {
	if tag == "" {
		return tag
	}
	return fmt.Sprintf("`%s`", tag)
}
