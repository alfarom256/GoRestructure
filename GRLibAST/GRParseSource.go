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

func GenSrcFromFile(fPath string, name string, outpath string, pList []*GRPackage) bool {
	// nab the file name
	fSplit := strings.Split(fPath, string(os.PathSeparator))
	fName := fSplit[len(fSplit)-1]
	// create the outpath files
	var newDir string
	if strings.Contains(name, ".go") {
		newDir = FixDirPath(outpath)
	} else {
		newDir = FixDirPath(outpath) + name + string(os.PathSeparator)
	}
	newFile := newDir + fName
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

	// uuuuuh
	// todo: find out why this is here??
	for i := range AllVars {
		ChangeVarsFuncAST(node, AllVars)
		_ = i
	}

	// now we need to get the strings in the file
	var AllStringASTLits []*ast.BasicLit
	for i := range mySource.FunctionDecl {
		tmp := StringsFromFunc(mySource.FunctionDecl[i])
		for j := range tmp {
			AllStringASTLits = append(AllStringASTLits, tmp[j])
		}
	}

	// make the function Stub handler
	funcStub := xorStub()
	// generate the string inlines
	res := *GenerateStrings(AllStringASTLits, funcStub)

	// swap all of the old values with the new ones
	for i := range res {
		tmpBasicLit := AllStringASTLits[i]
		tmpBasicLit.Kind = token.FUNC
		tmpBasicLit.Value = res[i].Stub
	}

	// now add the stub the package
	// if it's the main file

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
