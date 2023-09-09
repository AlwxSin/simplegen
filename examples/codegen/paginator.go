package codegen

import (
	"github.com/AlwxSin/simplegen"
	"go/ast"
	"golang.org/x/tools/go/packages"
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

// New{{$struct.Name}}ListPaginated returns paginated {{$struct.Name}} list if able to parse PaginateOptions.
func New{{$struct.Name}}ListPaginated(paginateOptions PaginateOptions) (*{{$struct.Name}}ListPaginated, error) {
	offset := 0
	if paginateOptions.Cursor != nil {
		o, err := strconv.Atoi(*paginateOptions.Cursor)
		if err != nil {
			return nil, err
		}
		offset = o
	}
	return &{{$struct.Name}}ListPaginated{
		Results:       make([]*{{$struct.Name}}, 0),
		CurrentCursor: paginateOptions.Cursor,
		isPaginated:   paginateOptions.IsPaginated(),
		limit:         paginateOptions.Limit,
		offset:        offset,
	}, nil
}

{{ end }}
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
