package main

import (
	"encoding/json"
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

type enumTpl struct {
	FieldName       string
	LowerFieldName  string
	EnumFields      string
	EnumFieldsArray string
}

type minMaxTpl struct {
	FieldName      string
	LowerFieldName string
	Length         string
}

type funcTpl struct {
	ReceiverName string
	FuncName     string
}

type ValidatorRules struct {
	FieldType  string
	IsRequired bool
	ParamName  string
	Enum       []string
	Default    string
	HasDefault bool
	Min        int
	HasMin     bool
	Max        int
	HasMax     bool
}

func (vr ValidatorRules) HasValues() bool {
	return vr.IsRequired || len(vr.ParamName) > 0 || len(vr.Enum) > 0 || vr.HasDefault || vr.HasMin || vr.HasMax
}

type Templates struct {
	tpl         *template.Template
	requiredTpl *template.Template
	enumTpl     *template.Template
	defaultTpl  *template.Template
	minTpl      *template.Template
	maxTpl      *template.Template
}

type ApiMethodsJson struct {
	Url    string
	Auth   bool
	Method string
}

type FuncInTpl struct {
	StructInName string
	FuncName     string
}

var templates = map[string]*Templates{
	"int": &Templates{
		tpl: template.Must(template.New("intTpl").Parse(`
	// {{.FieldName}}
	{{.FieldName}}Raw, err := strconv.Atoi(params.Get("{{.RequestFieldName}}"))
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("{{.RequestFieldName}} must be int")}
	}
	obj.{{.FieldName}} = {{.FieldName}}Raw
`)),
		requiredTpl: template.Must(template.New("requiredIntTpl").Parse(`
	// {{.FieldName}} required
	if obj.{{.FieldName}} == 0 {
		return ApiError{http.StatusBadRequest, fmt.Errorf("{{.RequestFieldName}} must be not empty")}
	}
	`)),
		enumTpl: template.Must(template.New("enumIntTpl").Parse(`
	// {{.FieldName}} enum
	if !slices.Contains([]string{{"{"}}{{.EnumFields}}{{"}"}}, strconv.Itoa(obj.{{.FieldName}})) {
		return ApiError{http.StatusBadRequest, fmt.Errorf("{{.LowerFieldName}} must be one of [{{.EnumFieldsArray}}]")}
	}
`)),
		defaultTpl: template.Must(template.New("defaultIntTpl").Parse(`
	// {{.FieldName}} default
	if obj.{{.FieldName}} == 0 {
		obj.{{.FieldName}} = {{.RequestFieldName}}
	}
`)),
		minTpl: template.Must(template.New("minIntTpl").Parse(`
	// {{.FieldName}} min
	if obj.{{.FieldName}} < {{.Length}} {
		return ApiError{http.StatusBadRequest, fmt.Errorf("{{.LowerFieldName}} must be >= {{.Length}}")}
	}
`)),
		maxTpl: template.Must(template.New("maxIntTpl").Parse(`
	// {{.FieldName}} max
	if obj.{{.FieldName}} > {{.Length}} {
		return ApiError{http.StatusBadRequest, fmt.Errorf("{{.LowerFieldName}} must be <= {{.Length}}")}
	}
`)),
	},
	"uint64": &Templates{
		tpl: template.Must(template.New("int64Tpl").Parse(`
	// {{.FieldName}}
	{{.FieldName}}Raw, err := strconv.ParseUint(params.Get("{{.RequestFieldName}}"), 10, 64)
	if err != nil {
		return ApiError{http.StatusBadRequest, fmt.Errorf("{{.RequestFieldName}} must be uint64")}
	}
	obj.{{.FieldName}} = {{.FieldName}}Raw
`)),
		requiredTpl: template.Must(template.New("requiredUint64Tpl").Parse(`
	// {{.FieldName}} required
	if obj.{{.FieldName}} == uint64(0) {
		return ApiError{http.StatusBadRequest, fmt.Errorf("{{.RequestFieldName}} must be not empty")}
	}
	`)),
		enumTpl: template.Must(template.New("enumUint64Tpl").Parse(`
	// {{.FieldName}} enum
	if !slices.Contains([]string{{"{"}}{{.EnumFields}}{{"}"}}, strconv.Itoa(obj.{{.FieldName}})) {
		return ApiError{http.StatusBadRequest, fmt.Errorf("{{.LowerFieldName}} must be one of [{{.EnumFieldsArray}}]")}
	}
`)),
		defaultTpl: template.Must(template.New("defaultUint64Tpl").Parse(`
	// {{.FieldName}} default
	if obj.{{.FieldName}} == 0 {
		obj.{{.FieldName}} = {{.RequestFieldName}}
	}
`)),
		minTpl: template.Must(template.New("minUint64Tpl").Parse(`
	// {{.FieldName}} min
	if obj.{{.FieldName}} < {{.Length}} {
		return ApiError{http.StatusBadRequest, fmt.Errorf("{{.LowerFieldName}} must be >= {{.Length}}")}
	}
`)),
		maxTpl: template.Must(template.New("maxUint64Tpl").Parse(`
	// {{.FieldName}} max
	if obj.{{.FieldName}} > {{.Length}} {
		return ApiError{http.StatusBadRequest, fmt.Errorf("{{.LowerFieldName}} must be <= {{.Length}}")}
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
	// {{.FieldName}} required
	if obj.{{.FieldName}} == "" {
		return ApiError{http.StatusBadRequest, fmt.Errorf("{{.RequestFieldName}} must be not empty")}
	}
	`)),
		enumTpl: template.Must(template.New("enumStringTpl").Parse(`
	// {{.FieldName}} enum
	if !slices.Contains([]string{{"{"}}{{.EnumFields}}{{"}"}}, obj.{{.FieldName}}) {
		return ApiError{http.StatusBadRequest, fmt.Errorf("{{.LowerFieldName}} must be one of [{{.EnumFieldsArray}}]")}
	}
`)),
		defaultTpl: template.Must(template.New("defaultStringTpl").Parse(`
	// {{.FieldName}} default
	if obj.{{.FieldName}} == "" {
		obj.{{.FieldName}} = "{{.RequestFieldName}}"
	}
`)),
		minTpl: template.Must(template.New("minStringTpl").Parse(`
	// {{.FieldName}} min
	if len(obj.{{.FieldName}}) < {{.Length}} {
		return ApiError{http.StatusBadRequest, fmt.Errorf("{{.LowerFieldName}} len must be >= {{.Length}}")}
	}
`)),
		maxTpl: template.Must(template.New("maxStringTpl").Parse(`
	// {{.FieldName}} max
	if len(obj.{{.FieldName}}) > {{.Length}} {
		return ApiError{http.StatusBadRequest, fmt.Errorf("{{.LowerFieldName}} len must be <= {{.Length}}")}
	}
`)),
	},
}

var (
	funcHeaderTpl = template.Must(template.New("funcHeaderTpl").Parse(`
func (h {{.ReceiverName}}) wrapper{{.FuncName}}(w http.ResponseWriter, r *http.Request) (interface{{"{}"}}, error) {{"{"}}
`))
	funcParamsTpl = template.Must(template.New("inParamsTpl").Parse(`
	in := {{.StructInName}}{}
	err := in.Unpack(params)
	if err != nil {
		return nil, ApiError{http.StatusBadRequest, err}
	}

	err = in.Validate()
	if err != nil {
		return nil, ApiError{http.StatusBadRequest, err}
	}

	return h.{{.FuncName}}(r.Context(), in)
`))
	serveHTTPTpl = template.Must(template.New("serveHTTPTpl").Parse(`
func (h {{ .APIName }}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		res interface{}
	)

	switch r.URL.Path {
	{{- range .Methods }}
	case "{{ .Url }}":
		res, err = h.wrapper{{ .Method }}(w, r)
	{{- end }}
	default:
		err = ApiError{http.StatusNotFound, fmt.Errorf("unknown method")}
	}

	var response = struct {
		Error    string      ` + "`json:\"error\"`" + `
		Response interface{} ` + "`json:\"response,omitempty\"`" + `
	}{}

	if err == nil {
		response.Response = res
	} else {
		response.Error = err.Error()

		var errApi ApiError
		if errors.As(err, &errApi) {
			w.WriteHeader(errApi.HTTPStatus)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	responseJson, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJson)
}
`))
)

type ApiMethod struct {
	Url    string
	Method string
}

type ApiTpl struct {
	APIName string
	Methods ApiStruct
}

type ApiStruct []ApiMethod

var apisRoutes = make(map[string]ApiStruct)

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	createPackageAndImports(out, node.Name.Name)

	for _, f := range node.Decls {
		switch f.(type) {
		case *ast.FuncDecl:
			generateForFunc(out, f)
		case *ast.GenDecl:
			generateForType(out, f)
		default:
			fmt.Printf("SKIP %#T is not *ast.GenDecl or *ast.FuncDecl\n", f)
		}
	}

	for apiName, apiStruct := range apisRoutes {
		serveHTTPTpl.Execute(out, &ApiTpl{APIName: apiName, Methods: apiStruct})
	}
}

func createPackageAndImports(out *os.File, nodeName string) {
	fmt.Fprintln(out, `package `+nodeName)
	fmt.Fprintln(out)
	fmt.Fprintln(out, `import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
)`)
	fmt.Fprintln(out)
}

func generateForFunc(out *os.File, f ast.Decl) {
	g, _ := f.(*ast.FuncDecl)
	needCodegen := false
	if g.Doc != nil {
		var comment *ast.Comment
		for _, comment = range g.Doc.List {
			needCodegen = needCodegen || strings.HasPrefix(comment.Text, "// apigen:api")
		}
		if !needCodegen {
			fmt.Printf("SKIP func %#v doesnt have cgen mark\n", g.Name.Name)
			return
		}

		var receiverType string
		switch expr := g.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			if ident, ok := expr.X.(*ast.Ident); ok {
				receiverType = "*" + ident.Name
			}
		case *ast.Ident:
			receiverType = expr.Name
		default:
		}

		funcHeaderTpl.Execute(out, funcTpl{ReceiverName: receiverType, FuncName: g.Name.Name})

		jsonData := comment.Text[13:]
		apiData := &ApiMethodsJson{}
		if err := json.Unmarshal([]byte(jsonData), apiData); err != nil {
			fmt.Printf("SKIP func %#v has invalid json data\n", g.Name.Name)
		}

		apisRoutes[receiverType] = append(apisRoutes[receiverType], ApiMethod{Url: apiData.Url, Method: g.Name.Name})

		createInitFuncCode(out, apiData)

		var typeStr string
		switch t := g.Type.Params.List[1].Type.(type) {
		case *ast.Ident:
			typeStr = t.Name
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				typeStr = "*" + ident.Name
			}
		case *ast.SelectorExpr:
			if pkg, ok := t.X.(*ast.Ident); ok {
				typeStr = pkg.Name + "." + t.Sel.Name
			}
		default:
			typeStr = fmt.Sprintf("%T", t) //
		}

		funcParamsTpl.Execute(out, FuncInTpl{StructInName: typeStr, FuncName: g.Name.Name})

		fmt.Fprintln(out, "}")
		fmt.Fprintln(out)
	}
}

func createInitFuncCode(out *os.File, apiData *ApiMethodsJson) {
	if apiData.Auth {
		fmt.Fprintln(out, "\tif r.Header.Get(\"X-Auth\") != \"100500\" {")
		fmt.Fprintln(out, "\t\treturn nil, ApiError{http.StatusForbidden, fmt.Errorf(\"unauthorized\")}")
		fmt.Fprintln(out, "\t}")
		fmt.Fprintln(out)
	}

	if apiData.Method != "" {
		fmt.Fprintln(out, "\tif r.Method != \""+apiData.Method+"\" {")
		fmt.Fprintln(out, "\t\treturn nil, ApiError{http.StatusNotAcceptable, fmt.Errorf(\"bad method\")}")
		fmt.Fprintln(out, "\t}")
		fmt.Fprintln(out)
	}

	fmt.Fprintln(out, "\tvar params url.Values")
	fmt.Fprintln(out, "\tif r.Method == \"GET\" {")
	fmt.Fprintln(out, "\t\tparams = r.URL.Query()")
	fmt.Fprintln(out, "\t} else {")
	fmt.Fprintln(out, "\t\terr := r.ParseForm()")
	fmt.Fprintln(out, "\t\tif err != nil {")
	fmt.Fprintln(out, "\t\t\treturn nil, ApiError{http.StatusBadRequest, fmt.Errorf(\"invalid request\")}")
	fmt.Fprintln(out, "\t\t}")
	fmt.Fprintln(out, "\t\tparams = r.PostForm")
	fmt.Fprintln(out, "\t}")
}

func generateForType(out *os.File, f ast.Decl) {
	g, _ := f.(*ast.GenDecl)
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

		fieldsToValidate := createUnpacking(out, currType.Name.Name, currStruct.Fields.List)

		createValidation(out, currType.Name.Name, fieldsToValidate)
	}
}

func createUnpacking(out *os.File, typeName string, fieldsList []*ast.Field) *map[string]*ValidatorRules {
	fmt.Printf("\tgenerating Unpack method\n")

	fmt.Fprintln(out, "func (obj *"+typeName+") Unpack(params url.Values) error {")

	fieldsToValidate := make(map[string]*ValidatorRules)

	for _, field := range fieldsList {
		fieldName := field.Names[0].Name
		fieldIdent, ok := field.Type.(*ast.Ident)
		if !ok {
			fmt.Printf("SKIP %#T is not ast.Ident\n", field.Type)
			continue
		}

		fieldType := fieldIdent.Name

		rules := getRules(field, fieldType)
		if rules != nil {
			fieldsToValidate[fieldName] = rules
		}

		fmt.Printf("\tgenerating code for field %s.%s\n", typeName, fieldName)

		switch fieldType {
		case "int":
			fallthrough
		case "uint64":
			fallthrough
		case "string":
			if validatorRules, ok := fieldsToValidate[fieldName]; ok {
				if validatorRules.ParamName != "" {
					templates[fieldType].tpl.Execute(out, tpl{FieldName: fieldName, RequestFieldName: strings.ToLower(validatorRules.ParamName)})
				} else {
					templates[fieldType].tpl.Execute(out, tpl{FieldName: fieldName, RequestFieldName: strings.ToLower(fieldName)})
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

	return &fieldsToValidate
}

func getRules(field *ast.Field, fieldType string) *ValidatorRules {
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
							rules.HasMin = true
						case "max":
							rules.Max, _ = strconv.Atoi(parts[1])
							rules.HasMax = true
						}
					}
				}
			}

			if rules.HasValues() {
				return rules
			}
		}
	}
	return nil
}

func createValidation(out *os.File, typeName string, fieldsToValidate *map[string]*ValidatorRules) {
	fmt.Printf("\tgenerating Validate method\n")

	fmt.Fprintln(out, "func (obj *"+typeName+") Validate() error {")

	for fieldName, validatorRules := range *fieldsToValidate {
		switch validatorRules.FieldType {
		case "int":
			fallthrough
		case "uint64":
			fallthrough
		case "string":
			if validatorRules.IsRequired {
				templates[validatorRules.FieldType].requiredTpl.Execute(out, tpl{FieldName: fieldName, RequestFieldName: strings.ToLower(fieldName)})
			}
			if validatorRules.HasDefault {
				templates[validatorRules.FieldType].defaultTpl.Execute(out, tpl{FieldName: fieldName, RequestFieldName: validatorRules.Default})
			}
			if len(validatorRules.Enum) > 0 {
				q := make([]string, len(validatorRules.Enum))
				for i, v := range validatorRules.Enum {
					q[i] = `"` + v + `"`
				}
				resEnum := strings.Join(q, ", ")

				templates[validatorRules.FieldType].enumTpl.Execute(out, enumTpl{FieldName: fieldName, LowerFieldName: strings.ToLower(fieldName), EnumFields: resEnum, EnumFieldsArray: strings.Join(validatorRules.Enum, ", ")})
			}
			if validatorRules.HasMin {
				templates[validatorRules.FieldType].minTpl.Execute(out, minMaxTpl{FieldName: fieldName, LowerFieldName: strings.ToLower(fieldName), Length: strconv.Itoa(validatorRules.Min)})
			}
			if validatorRules.HasMax {
				templates[validatorRules.FieldType].maxTpl.Execute(out, minMaxTpl{FieldName: fieldName, LowerFieldName: strings.ToLower(fieldName), Length: strconv.Itoa(validatorRules.Max)})
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
