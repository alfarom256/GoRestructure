package GRLibAST

import (
	"fmt"
	"go/ast"
	"unicode"
)

type GRFunc struct {
	Package      *GRPackage
	FunctionName string
	FunctionAST  *ast.FuncDecl
	isPublic     bool
}

/*
So this type is confusing...
What we're doing here is mapping OUR functions to a list of ALL places in the AST they're referenced

*/
type GRFuncMap map[GRFunc][]*ast.CallExpr

func GetAllFunctions(allPackages []*GRPackage) [][]*GRFunc {
	retVal := make([][]*GRFunc, len(allPackages))
	for i := range allPackages {
		retVal[i] = GetFunctionsFromPackage(allPackages[i])
	}
	return retVal
}

func GetFunctionsFromPackage(inPackage *GRPackage) []*GRFunc {
	var retVal []*GRFunc
	for i := range inPackage.Files {
		for j := range inPackage.Files[i].FileAST.Decls {
			decl := inPackage.Files[i].FileAST.Decls[j]
			funcDecl, err := decl.(*ast.FuncDecl)
			if err {
				funcDecls := funcDecl
				tempFunc := GRFunc{
					inPackage,
					funcDecls.Name.Name,
					funcDecls,
					!unicode.IsLower(rune(funcDecls.Name.Name[0])),
				}
				retVal = append(retVal, &tempFunc)
			}
		}

	}
	return retVal
}

func FindAllUsagesInPackage(inPackages []*GRPackage, allFunctionsList [][]*GRFunc) GRFuncMap {
	// go through all the functions we have listed and initialize a map with the string
	// name of the function as the keys
	retVal := GRFuncMap{}
	// flatten the 2d array
	var flatFuncList []*GRFunc
	for f := range allFunctionsList {
		flatFuncList = append(flatFuncList, allFunctionsList[f]...)
	}
	for i := range flatFuncList {
		retVal[*flatFuncList[i]] = nil
	}

	for i := range inPackages {
		files := inPackages[i].Files

		for j := range files {
			files[j].FileNodeSource = ParseNodeSource(files[j].FileAST)
			functions := files[j].FileNodeSource.FunctionDecl
			for k := range functions {
				ast.Inspect(functions[k].Body, func(node ast.Node) bool {
					call, ok := node.(*ast.CallExpr)

					if !ok {
						return false
					} else {
						funcIdent, ok := call.Fun.(*ast.Ident)
						if ok {
							fmt.Printf("funcIDENT: %s", funcIdent.Name)
						}
					}
					return false
				})
			}
		}
	}
	return nil
}

func CompareFuncName(grf GRFunc, funcName string) bool {
	return grf.FunctionName == funcName
}
