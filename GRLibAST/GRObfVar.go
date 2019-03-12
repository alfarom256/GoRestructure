package GRLibAST

import (
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
	statementList := fDecl.Body.List
	var decls []*ast.Ident
	for i := range statementList {
		tmp := statementList[i]
		assign, ok := tmp.(*ast.AssignStmt)
		if ok {
			if assign.Tok == token.DEFINE || assign.Tok == token.VAR{
				for i := range assign.Lhs{
					decls = append(decls, assign.Lhs[i].(*ast.Ident))
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
						decls = append(decls, valSpec.Names[i])
					}
				}
			}
		}
	}
	return decls
}

func ChangeVarsFuncAST(inAST *ast.File, varMap map[*ast.FuncDecl][]*ast.Ident) *ast.File{
	var fList []*ast.FuncDecl
	for k, v := range varMap{
		fList = append(fList, k)
		_ = v
	}
	ast.Inspect(inAST, func (n ast.Node) bool{
		funcDecl, ok := n.(*ast.FuncDecl)
		if ok && funcContains(fList, funcDecl){
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
		retval[identList[i].Name] = varString()
	}
	ast.Inspect(inAST,
		func (n ast.Node) bool{
			ident, ok := n.(*ast.Ident)
			if ok && identContains(identList, ident){
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
	return stringWithCharset(rand.Intn(7) + 1, charset)
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