package GRLibGenerate

import (
	"fmt"
	"strings"
)

/*

IMPORTANT:

In the scope of this file, declarations and variables are all
referred to by the shorthand "var"

*/

const (
	constructor = "%s %s %s" // LHS TOKEN RHS
)

type GeneratedVariable struct {
	VarString string
	VarNames  []string
	VarType   GeneratedType
	VarCount  int
	IsDecl    bool
}

func GenerateVariable(VarNames []string, VarType GeneratedType, IsDecl bool) GeneratedVariable {
	lhs, rhs, tok := "var ", "", ""

	if IsDecl {
		lhs = ""
		rhs = fmt.Sprintf("new(%s)", VarType.Value)
		tok = ":="
	} else {
		rhs = VarType.Value
	}

	if len(VarNames) > 1 { // if it's a compound assignment ( var a,b,c int ; a,b,c := 1,2,3
		lhs += strings.Join(VarNames, ", ")
	} else {
		lhs = VarNames[0]
	}

	retVal := new(GeneratedVariable)
	retVal.VarType = VarType
	retVal.VarNames = VarNames
	retVal.IsDecl = IsDecl
	retVal.VarCount = len(VarNames)
	retVal.VarString = fmt.Sprintf(constructor, lhs, tok, rhs)
	return *retVal
}
