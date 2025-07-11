package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

type tpl struct {
	FieldName        string
	RequestFieldName string
}

type ValidatorRules struct {
	FieldType  string
	IsRequired bool
	ParamName  string
	Enum       []string
	Default    string
	HasDefault bool
	Min        int
	Max        int
}

type Templates struct {
	tpl         *template.Template
	requiredTpl *template.Template
	enumTpl     *template.Template
	defaultTpl  *template.Template
	minTpl      *template.Template
	maxTpl      *template.Template
}

var templates = map[string]*Templates{
	"int": &Templates{
		tpl: template.Must(template.New("intTpl").Parse(`
	// {{.FieldName}}
	{{.FieldName}}Raw, err := strconv.Atoi(params.Get("{{.RequestFieldName}}"))
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid {{.FieldName}} - must be int")}
	}
	obj.{{.FieldName}} = {{.FieldName}}Raw
`)),
		requiredTpl: template.Must(template.New("requiredIntTpl").Parse(`
	// required
	if {{.FieldName}} == 0 {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid {{.FieldName}} - field is required")}
	}
	`)),
	},
	"uint64": &Templates{
		tpl: template.Must(template.New("int64Tpl").Parse(`
	// {{.FieldName}}
	{{.FieldName}}Raw, err := strconv.ParseUint(params.Get("{{.RequestFieldName}}"), 10, 64)
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid {{.FieldName}} - must be uint64")}
	}
	obj.{{.FieldName}} = {{.FieldName}}Raw
`)),
		requiredTpl: template.Must(template.New("requiredUint64Tpl").Parse(`
	// required
	if {{.FieldName}} == uint64(0) {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid {{.FieldName}} - field is required")}
	}
	`)),
	},
	"string": &Templates{
		tpl: template.Must(template.New("strTpl").Parse(`
	// {{.FieldName}}
	{{.FieldName}}Raw := params.Get("{{.RequestFieldName}}")
	obj.{{.FieldName}} = {{.FieldName}}Raw
`)),
		requiredTpl: template.Must(template.New("requiredStringTpl").Parse(`
	// required
	if {{.FieldName}} == "" {
		return ApiError{http.StatusBadRequest, fmt.Errorf("invalid {{.FieldName}} - field is required")}
	}
	`)),
	},
}

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

				fieldsToValidate := make(map[string]*ValidatorRules)

				//FIELDS_LOOP:
				for _, field := range currStruct.Fields.List {
					fieldName := field.Names[0].Name
					fieldIdent, ok := field.Type.(*ast.Ident)
					if !ok {
						fmt.Printf("SKIP %#T is not ast.Ident\n", field.Type)
						continue
					}

					fieldType := fieldIdent.Name

					if field.Tag != nil {
						tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])

						if res, ok := tag.Lookup("apivalidator"); ok {
							rules := &ValidatorRules{FieldType: fieldType}

							tags := strings.Split(res, ",")

							for _, t := range tags {
								if t == "required" {
									rules.IsRequired = true
								} else {
									parts := strings.Split(t, "=")
									if len(parts) == 2 {
										switch parts[0] {
										case "paramname":
											rules.ParamName = parts[1]
										case "enum":
											rules.Enum = strings.Split(parts[1], "|")
										case "default":
											rules.Default = parts[1]
											rules.HasDefault = true
										case "min":
											rules.Min, _ = strconv.Atoi(parts[1])
										case "max":
											rules.Max, _ = strconv.Atoi(parts[1])
										}
									}
								}
							}

							fieldsToValidate[fieldName] = rules
							fmt.Println(rules)
						}
					}

					fmt.Printf("\tgenerating code for field %s.%s\n", currType.Name.Name, fieldName)

					lowerFieldName := strings.ToLower(fieldName)

					switch fieldType {
					case "int":
						fallthrough
					case "uint64":
						fallthrough
					case "string":
						if validatorRules, ok := fieldsToValidate[fieldName]; ok {
							if validatorRules.ParamName != "" {
								templates[fieldType].tpl.Execute(out, tpl{fieldName, strings.ToLower(validatorRules.ParamName)})
							} else {
								templates[fieldType].tpl.Execute(out, tpl{fieldName, lowerFieldName})
							}
						}
					default:
						log.Fatalln("unsupported", fieldType)
					}
				}

				fmt.Fprintln(out)
				fmt.Fprintln(out, "	return nil")
				fmt.Fprintln(out, "}")
				fmt.Fprintln(out)

				fmt.Fprintln(out, "func (obj *"+currType.Name.Name+") Validate() error {")

				for fieldName, validatorRules := range fieldsToValidate {
					switch validatorRules.FieldType {
					case "int":
						fallthrough
					case "uint64":
						fallthrough
					case "string":
						if validatorRules.IsRequired {
							templates[validatorRules.FieldType].requiredTpl.Execute(out, tpl{FieldName: fieldName})
						}
					default:
						log.Fatalln("unsupported", validatorRules.FieldType)
					}
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
