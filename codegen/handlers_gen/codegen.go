package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
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
	fmt.Fprintln(out, `import "encoding/binary"`)
	fmt.Fprintln(out, `import "bytes"`)
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

				fmt.Printf("process struct %s\n", currType.Name.Name)
				fmt.Printf("\tgenerating Unpack method\n")

				fmt.Fprintln(out, "func (in *"+currType.Name.Name+") Unpack(data []byte) error {")
				fmt.Fprintln(out, "	r := bytes.NewReader(data)\n return nil}\n")
			}
		default:
			fmt.Printf("SKIP %#T is not *ast.GenDecl or *ast.FuncDecl\n", f)
		}
	}
}
