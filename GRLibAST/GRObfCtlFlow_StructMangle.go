package GRLibAST

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"math/rand"
)

/*

type StructType struct {
        Struct     token.Pos  // position of "struct" keyword
        Fields     *FieldList // list of field declarations
        Incomplete bool       // true if (source) fields are missing in the Fields list
}

FIELDS
type FieldList struct {
        Opening token.Pos // position of opening parenthesis/brace, if any
        List    []*Field  // field list; or nil
        Closing token.Pos // position of closing parenthesis/brace, if any
}

LIST
type Field struct {
        Doc     *CommentGroup // associated documentation; or nil
        Names   []*Ident      // field/method/parameter names; or nil
        Type    Expr          // field/method/parameter type
        Tag     *BasicLit     // field tag; or nil
        Comment *CommentGroup // line comments; or nil
}

*/
const STRUCT_MAX_FIELDS = 20

type StructMangle struct {
	Original *ast.Node
	Mangled  *ast.Node

	/*
		This map will be structured like so:

		KEY:VAL

		NODE1 --> REPLACEMENT_NODE1
		NODE6 --> REPLACEMENT_NODE6
		NODE78 --> REPLACEMENT_NODE78
		...

		NODE## -- node of the original source to replace with
		REPLACEMENT_NODE##

	*/
	NodeMap map[*ast.Node]*ast.Node
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
		varType := "int"
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

/*
 type RandomAA struct {
    a int
    b byte
    c *string
    ...
}
type RandomAB struct {
    a rune
    b uint32
    c map[int]int
    ...
}

for control flow obfuscation

take a simple main funciton

before
func main(){
    fmt.Println("Hello")
    fmt.Println("World")
}

after


type RandomAA struct {...}
type RandomAB struct {...}
func main(){
   local_a := RandomAA{...} // fill with "random" junk
   local_b := RandomAB{...} // fill with "random" junk
   for iter := range ... {
         if (local_a.something < DO SOME MATH HERE > local_b.something){
             fmt.Println("Hello")
         } else if ( same thing as above ) {
             fmt.Println("World")
         }
    }
}
*/
