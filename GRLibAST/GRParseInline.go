/*
//|Inline FILE FUNCTION ARGC ARGV
This command has a few parameters

FILE - File containing function you wish to call

FUNCTION - Function name you want to inline a call to

ARGC - Arg Count
The argument count

ARGV - Variable, literal, or mixed list.
ex:
//|Inline C:\test\mymath.go union 2 {[1,2,3,4,5,6,7,8], [0,3,4]}
this will insert an import for mymath and will call
mymath.union([1,2,3,4,5,6,7,8], [0,3,4])

//|Inline C:\test\myUtil.go readFile 2 {fileHandler.Open(), "myFile.txt", 3}

*/

package GRLibAST

import "go/ast"

type InlineTag struct {
	loc      *ast.Node
	fPath    string
	funcName string
	argc     uint
	argv     string
	// the arguments need to be reformatted from either variables or literals
	// we're going to add a function call like this
	// stringInline := packageName + '.' + funcName + '(' + argv + ')'
	// f, err := parser.ParseFile(fset, "src.go", src, 0)
	//	if err != nil {
	//		panic(err)
	// }
	// then we need to get the first node and swap the comment node pointer with the new AST Node
	// bingo bango bongo smango
	//

}
