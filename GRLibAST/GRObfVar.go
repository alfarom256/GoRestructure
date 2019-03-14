package GRLibAST

import (
	"fmt"
	"go/ast"
	"go/token"
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ_"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

type ObfVar struct {
	varMap map[string]string
}

func VarsFromFunc(fDecl *ast.FuncDecl) []*ast.Ident {
	var decls []*ast.Ident
	funcFList := fDecl.Type.Params.List // this is a field list of the function  []*Field
	// we need to extract the parameters from the function declaration and rename them accordingly
	var funcParamNames []string
	for fList := range funcFList {
		tempNames := funcFList[fList].Names
		for i := range tempNames {
			if tempNames[i].Obj.Kind != ast.Typ && tempNames[i].Obj.Kind != ast.Pkg {
				fmt.Printf("FOUND PARAMETER FOR FUNC %s: %s\n", fDecl.Name, tempNames[i].Name)
				funcParamNames = append(funcParamNames, tempNames[i].Name)
				decls = append(decls, tempNames[i]) // add function parameters to all declaration idents in function
			}
		}
	}
	// now that we have a list of function parameters, let's add those to the map

	statementList := fDecl.Body.List
	for i := range statementList {
		tmp := statementList[i]
		assign, ok := tmp.(*ast.AssignStmt)
		if ok {
			if assign.Tok == token.DEFINE || assign.Tok == token.VAR {
				for i := range assign.Lhs {
					nilValChk := assign.Lhs[i].(*ast.Ident).Name
					if nilValChk != "_" {
						decls = append(decls, assign.Lhs[i].(*ast.Ident))
					} else {
						fmt.Printf("Ignoring underscore")
					}
				}
			}
		}
		// declStmt -> decl -> genDecl -> Specs[0] -> valueSpec -> (Names == []*ast.Ident) ... Jesus
		decl, ok := tmp.(*ast.DeclStmt)
		if ok {
			genDecl, ok := decl.Decl.(*ast.GenDecl)
			if ok {
				valSpec, ok := genDecl.Specs[0].(*ast.ValueSpec)
				if ok {
					for i := range valSpec.Names {
						nilValChk := valSpec.Names[i].Name
						if nilValChk != "_" {
							decls = append(decls, valSpec.Names[i])
						} else {
							fmt.Printf("Ignoring underscore")
						}
					}
				}
			}
		}
	}
	return decls
}

func ChangeVarsFuncAST(inAST *ast.File, varMap map[*ast.FuncDecl][]*ast.Ident) *ast.File {
	var fList []*ast.FuncDecl
	for k, v := range varMap {
		fList = append(fList, k)
		_ = v
	}
	ast.Inspect(inAST, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if ok && funcContains(fList, funcDecl) {
			changedVars := changeVarsInFunction(inAST, varMap[funcDecl])
			if changedVars == nil {
				return false
			}
			return true
		}
		return true
	})
	return inAST
}

func changeVarsInFunction(inAST *ast.File, identList []*ast.Ident) map[string]string {
	var identsToChange []*ast.Ident
	var retval = make(map[string]string, len(identList))
	for i := range identList {
		// if the thing we're changing is a type, ignore it.
		if identList[i].Obj.Kind != ast.Typ && identList[i].Obj.Kind != ast.Pkg {
			retval[identList[i].Name] = varString()
		}
	}
	ast.Inspect(inAST,
		func(n ast.Node) bool {
			ident, ok := n.(*ast.Ident)
			if ok && identContains(identList, ident) && ident.Name != "_" {
				identsToChange = append(identsToChange, ident)
				return true
			}
			return true
		})
	for i := range identsToChange {
		identsToChange[i].Name = retval[identsToChange[i].Name]
	}
	return retval
}

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func varString() string {
	return stringWithCharset(rand.Intn(7)+4, charset)
}

func identContains(nArr []*ast.Ident, n *ast.Ident) bool {
	for i := range nArr {
		if nArr[i].Name == n.Name {
			return true
		}
	}
	return false
}

func funcContains(nArr []*ast.FuncDecl, n *ast.FuncDecl) bool {
	for i := range nArr {
		if nArr[i].Name.Name == n.Name.Name {
			return true
		}
	}
	return false
}
