package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "./testdata/config.go", nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	pvv := &configStructVisitor{}
	ast.Walk(pvv, f)
}

type configStructVisitor struct {
}

func (v *configStructVisitor) Visit(n ast.Node) (w ast.Visitor) {
	tspec, ok := n.(*ast.TypeSpec)
	if !ok {
		return v
	}

	if tspec.Name.Name == "Config" {
		fmt.Printf("config=%+v\n", 111)
	}

	if tspec.Name.Name == "ProducerConfig" {
		fmt.Printf("222=%+v\n", 222)
	}

	return v

	// ttype, ok := tspec.Type.(*ast.StructType)
	// if !ok {
	// 	return v
	// }

	// for _, field := range ttype.Fields.List {
	// 	fmt.Printf("field=%+v\n", field.Tag)

	// 	stype, ok := field.Type.(*ast.SelectorExpr)
	// 	if !ok {
	// 		return v
	// 	}
	// }
	// return v
}

type configBuilderVisitor struct {
	pkgPath string
}

func (v *configBuilderVisitor) Visit(n ast.Node) (w ast.Visitor) {
	return v
}
