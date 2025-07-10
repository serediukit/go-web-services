package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"text/template"
)

type tpl struct {
	FieldName string
}

var (
	intTpl = template.Must(template.New("intTpl").Parse(`
	// {{.FieldName}}
	{{.FieldName}}Raw, err := strconv.Atoi(params.Get("{{.FieldName}}"))
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid {{.FieldName}} - must be int")}
	}
	obj.{{.FieldName}} = {{.FieldName}}Raw
`))

	uint64Tpl = template.Must(template.New("int64Tpl").Parse(`
	// {{.FieldName}}
	{{.FieldName}}Raw, err := strconv.Atoi(params.Get("{{.FieldName}}"))
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid {{.FieldName}} - must be int")}
	}
	obj.{{.FieldName}} = uint64({{.FieldName}}Raw)
`))

	strTpl = template.Must(template.New("strTpl").Parse(`
	// {{.FieldName}}
	{{.FieldName}}Raw := params.Get("{{.FieldName}}")
	obj.{{.FieldName}} = {{.FieldName}}Raw
`))
)

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out)
	fmt.Fprintln(out, `import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)`)
	fmt.Fprintln(out)

	for _, f := range node.Decls {
		switch f.(type) {
		case *ast.FuncDecl:
			//fmt.Printf("%+v is *ast.FuncDecl\n", f)
		case *ast.GenDecl:
			g, _ := f.(*ast.GenDecl)
			//SPECS_LOOP:
			for _, spec := range g.Specs {
				currType, ok := spec.(*ast.TypeSpec)
				if !ok {
					fmt.Printf("SKIP %#T is not ast.TypeSpec\n", spec)
					continue
				}

				currStruct, ok := currType.Type.(*ast.StructType)
				if !ok {
					fmt.Printf("SKIP %#T is not ast.StructType\n", currStruct)
					continue
				}

				if currType.Name.Name == "ApiError" {
					fmt.Printf("SKIP struct %#v is ApiError\n", currType.Name.Name)
					continue
				}

				fmt.Printf("process struct %s\n", currType.Name.Name)
				fmt.Printf("\tgenerating Unpack method\n")

				fmt.Fprintln(out, "func (obj *"+currType.Name.Name+") Unpack(params url.Values) error {")
				//fmt.Fprintln(out, "	r := bytes.NewReader(data)")

				//FIELDS_LOOP:
				for _, field := range currStruct.Fields.List {
					fieldName := field.Names[0].Name
					fieldIdent, ok := field.Type.(*ast.Ident)
					if !ok {
						fmt.Printf("SKIP %#T is not ast.Ident\n", field.Type)
						continue
					}

					fieldType := fieldIdent.Name

					fmt.Printf("\tgenerating code for field %s.%s\n", currType.Name.Name, fieldName)

					switch fieldType {
					case "int":
						intTpl.Execute(out, tpl{fieldName})
					case "uint64":
						uint64Tpl.Execute(out, tpl{fieldName})
					case "string":
						strTpl.Execute(out, tpl{fieldName})
					case "error":
						continue
					default:
						log.Fatalln("unsupported", fieldType)
					}

					//
					//if field.Tag != nil {
					//	tags := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
					//}
				}

				fmt.Fprintln(out)
				fmt.Fprintln(out, "	return nil")
				fmt.Fprintln(out, "}")
				fmt.Fprintln(out)
			}
		default:
			fmt.Printf("SKIP %#T is not *ast.GenDecl or *ast.FuncDecl\n", f)
		}
	}
}
