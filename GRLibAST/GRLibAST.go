package GRLibAST

import (
	"go/ast"
	"GoRestructure/GRLibUtil"
)

type NodeSource struct {
	Assignments  []*ast.AssignStmt // :=
	Values       []*ast.ValueSpec  // consts, =
	Literals     []*ast.BasicLit   // "asdf", 1234, 0xFFFFD00D
	Imports      []*ast.ImportSpec // all the Imports
	FunctionDecl []*ast.FuncDecl // function declarations
	root *ast.Node // pointer to the root node of the AST
}



func ParseNodeSource(node ast.Node) *NodeSource {
	retVal := NodeSource{}

	// set the root node to the head of the AST
	retVal.root = &node

	// depth first iterate over each node in the AST Tree
	ast.Inspect(node, func(n ast.Node) bool {

		// if the node is an assignment
		assignments, ok := n.(*ast.AssignStmt)
		if ok {
			// add it to the list of Assignments
			retVal.Assignments = append(retVal.Assignments, assignments)
			return true
		}

		imports, ok := n.(*ast.ImportSpec)
		if ok {
			retVal.Imports = append(retVal.Imports, imports)
			return true
		}

		var importNames []string
		for i := range retVal.Imports {
			importNames = append(importNames, retVal.Imports[i].Path.Value)
		}
		// ditto
		values, ok := n.(*ast.ValueSpec)
		if ok {
			// ditto
			retVal.Values = append(retVal.Values, values)
			return true
		}

		// if the node is a literal
		vars, ok := n.(*ast.BasicLit)
		if ok && !GRLibUtil.StrContains(importNames, vars.Value) {
			// add it to our list of Literals
			retVal.Literals = append(retVal.Literals, vars)
			return true // our evaluation is done, don't recheck the same node
		}
		functionDecl, ok := n.(*ast.FuncDecl)
		if ok {
			// add it to our list of Literals
			retVal.FunctionDecl = append(retVal.FunctionDecl, functionDecl)
			return true // our evaluation is done, don't recheck the same node
		}
		return true
	})
	return &retVal
}

