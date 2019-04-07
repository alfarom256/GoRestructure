package GRLibAST

import (
	"GoRestructure/GRLibGenerate"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"math/rand"
)

const STRUCT_MAX_FIELDS = 20

type StructMangle struct {
	Name         string
	FieldCount   int
	FieldTypeMap map[string]int
}

func GenerateRandomStruct() *ast.Decl {

	stringTemplate := `package fuckoff
type %s struct{
%s
}`

	// make sure it's between [1-20), just add one
	fieldCount := (rand.Int31() % STRUCT_MAX_FIELDS) + 1

	newFieldList := make([]*ast.Field, fieldCount)
	structName := StringWithCharset(10, charset)
	vars := ""
	for i := range newFieldList {
		_ = i
		varName := StringWithCharset(10, charset)
		varType := GRLibGenerate.GenerateRandomType().Value
		vars += varName + " " + varType + "\n"
	}
	structString := fmt.Sprintf(stringTemplate, structName, vars)
	// import the ast module to parse this
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "", structString, parser.AllErrors)
	if err != nil {
		log.Fatal("couldn't make struct, exiting...")
	}
	retVal := f.Decls[0]
	return &retVal
}

func AddToNodeTree(original *ast.File, nodeSwapMap map[*ast.Node]*ast.Node) *ast.File {
	decls := original.Decls
	// create an array of genDecls + 20 random structs to add
	newDecls := make([]ast.Decl, len(decls)+20)
	for i := 0; i < len(decls); i++ {
		newDecls[i] = decls[i]
	}

	for i := len(decls); i < len(decls)+20; i++ {
		temp := GenerateRandomStruct()
		newDecls[i] = *temp
	}

	original.Decls = newDecls
	return original
}
