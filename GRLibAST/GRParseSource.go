package GRLibAST

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func GetASTFile(fname string) *ast.File {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fname, nil, parser.ParseComments)
	if err != nil {
		log.Fatal("Fatal error parsing file\n" + err.Error())
		log.Fatal(err)
		log.Fatal(fname)
		panic(err)
	}
	return node
}

func GenSrcFromFile(fPath string, name string, outpath string) bool {
	// nab the file name
	fSplit := strings.Split(fPath, string(os.PathSeparator))
	fName := fSplit[len(fSplit)-1]
	// create the outpath files
	newDir := FixDirPath(outpath) + name
	newFile := newDir + string(os.PathSeparator) + fName
	_, err := os.Stat(newDir) // see if the file already exists
	if err != nil {
		os.MkdirAll(newDir, os.ModePerm)
	}

	err = nil
	_, err = os.Create(newFile)
	if err != nil {
		log.Fatal("Can't create file for writing...")
		log.Fatal(newFile)
		panic(err)
	}

	mySource := NodeSource{}
	fset := token.NewFileSet()

	node := GetASTFile(fPath)
	mySource = *ParseNodeSource(node)

	// carve out a map (table) that will store a list of Function nodes
	// the value will be all the ident variables in the function
	AllVars := make(map[*ast.FuncDecl][]*ast.Ident, len(mySource.FunctionDecl))
	for i := range mySource.FunctionDecl {
		tmp := VarsFromFunc(mySource.FunctionDecl[i])
		AllVars[mySource.FunctionDecl[i]] = tmp
	}

	for i := range AllVars {
		ChangeVarsFuncAST(node, AllVars)
		_ = i
	}
	// DEBUG
	printer.Fprint(os.Stdout, fset, node)
	// END DEBUG
	fmt.Printf("WRITING TO FILE: %s\n", newFile)

	var fWriteBuf bytes.Buffer
	printer.Fprint(&fWriteBuf, fset, node)
	// let's write to the file
	err = nil
	err = ioutil.WriteFile(newFile, fWriteBuf.Bytes(), 0644)
	if err != nil {
		fmt.Printf("Error writing file %s\n", newFile)
		panic(err)
	}
	return true
}
