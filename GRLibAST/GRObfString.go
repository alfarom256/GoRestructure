package GRLibAST

import (
	"GoRestructure/GRLibUtil"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"go/ast"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"os"
	"strings"
)

/*

Okay so modules need to be able to do a few basic things

EACH module MUST implement the following methods:

1. ItemsFromFunction(fdecl *ast.FuncDecl) []*ast.Node
2. Transform() ([]*ast.Node, err)
3. Replace(fdecl *ast.FuncDecl, transformedNodes []*ast.Node) (*ast.FuncDecl, err)


4. GetImports() []*ast.GenDecl
Get the imports required by the transform function


5. InsertImports(*ast.File) (*ast.File, err)

*/

type stringObfStub struct {
	argc              uint8
	name              string
	function_stub     string
	function_call_fmt string
}

type xorStrStruct struct {
	Stub       string // function call Stub
	Key        []byte // xor Key
	Encoded    []byte // Encoded string
	original   []byte // original string
	tmpVarName string
}

func WriteStubToPackage(pkg GRPackage, outputPath string) bool {

	if string(outputPath[len(outputPath)-1]) != string(os.PathSeparator) {
		outputPath += string(os.PathSeparator)
	}

	fname := outputPath + pkg.Name + string(os.PathSeparator) + xorStub().name + ".go" // the whole file to write

	f, err := os.Create(fname)
	if err != nil {
		fmt.Printf("ERROR CREATING XOR DECODE STUB FILE")
		panic(err)
	}
	err = nil
	_, err = f.WriteString(GetStubAsText(pkg.Name))
	if err != nil {
		fmt.Printf("ERROR WRITING XOR DECODE STUB TO PKGs")
		panic(err)
	}
	_ = f.Close()
	return true
}

func GetStubAsText(pkgName string) string {
	retVal := "package " + pkgName + "\n"
	retVal += "import \"encoding/hex\"\n" + "import \"strings\"\n" + xorStub().function_stub
	return retVal
}

func xorStub() stringObfStub {
	// todo: convert this to AST
	codeFuncStub := `
func obfs(s []byte, k []byte) string {
	decoded_str, err := hex.DecodeString(string(s))
	decoded_key, err := hex.DecodeString(string(k))

	if err != nil {
		panic(err)
	}

	ret_val := make([]byte, len(s))
	for i := range decoded_key {
		ret_val[i] = decoded_str[i] ^ decoded_key[i]
	}
	retStr := strings.Trim(string(ret_val), string(0x0))
	return retStr
}
`
	// todo: convert this to AST
	function_call_fmt := "string(obfs([]byte(\"%x\"),[]byte(\"%s\")))"
	// this is fine
	name := "obfs"
	var argc uint8 = 2
	return stringObfStub{argc, name, codeFuncStub, function_call_fmt}
}

func AppendStub(fName string, outPath string) string {
	fset := token.NewFileSet()
	s := xorStub()
	inAst := GetASTFile(fName)
	tmpSource := *ParseNodeSource(inAst)

	var srcBuf bytes.Buffer
	printer.Fprint(&srcBuf, fset, inAst)
	strSrc := srcBuf.String()

	strSrc += s.function_stub
	for i := range tmpSource.FunctionDecl {
		if tmpSource.FunctionDecl[i].Name.Name == s.name {
			return outPath + fName
		}
	}
	f, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("ERROR IN APPENDING TO BUILD TARGET FILE")
		panic(err)
	}
	f.WriteString(strSrc)
	_ = f.Close()
	return outPath + fName
}

// Strings are literals
func StringsFromFunc(fDecl *ast.FuncDecl) []*ast.BasicLit {
	var StringLits []*ast.BasicLit

	// get the body of the function
	// this is a list of AST Statements
	fBody := fDecl.Body.List

	// walk through the body
	for i := range fBody {
		ast.Inspect(fBody[i], func(node ast.Node) bool {
			lit, ok := node.(*ast.BasicLit)
			if ok && lit.Kind == token.STRING {
				StringLits = append(StringLits, lit)
				fmt.Printf("LIT STRING, FUNC: %s | VAL: %s\n", fDecl.Name.Name, lit.Value)
				return true
			}
			return true
		})

	}

	return StringLits
}

func GenerateStrings(in_strings []*ast.BasicLit, s stringObfStub) *[]xorStrStruct {
	ret_val := make([]xorStrStruct, len(in_strings))
	for i := range in_strings {
		inString := strings.Trim(in_strings[i].Value, "\"")
		key := make([]byte, len(inString))
		rand.Read(key) // make a random Key

		inStringHex := make([]byte, hex.EncodedLen(len(inString)))
		hex.Encode(inStringHex, []byte(inString))
		keyHex := make([]byte, hex.EncodedLen(len(key)))
		hex.Encode(keyHex, []byte(key))

		encoded := xorString(inStringHex, keyHex)
		encoded = encoded[:len(key)]

		key = keyHex
		tStub := fmt.Sprintf(s.function_call_fmt, encoded, key)
		tmp := xorStrStruct{
			Stub:       tStub, // xorDecode("ASDF", "XYZ")
			Key:        key,
			Encoded:    encoded,
			original:   []byte(inString),
			tmpVarName: fmt.Sprintf("temp%d", i),
		}
		ret_val[i] = tmp
	}
	return &ret_val
}

func xorString(s []byte, k []byte) []byte {
	decoded_str, err := hex.DecodeString(string(s))
	decoded_key, err := hex.DecodeString(string(k))

	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	ret_val := make([]byte, len(s))
	for i := range decoded_key {
		ret_val[i] = decoded_str[i] ^ decoded_key[i]
	}
	return ret_val
}

func replaceTempVarStrings(node *ast.File, s []xorStrStruct) *ast.File {
	// ultimately this is where we will be looking up nodes to replace the RHS with
	nodeTable := make(map[string]*ast.Node)

	origXorStrings := make([]string, len(s))
	varNameStrings := make([]string, len(s))
	for i := range s {
		origXorStrings[i] = string(s[i].original)
	}
	for i := range s {
		varNameStrings[i] = string(s[i].tmpVarName)
	}
	// this will hold the nodes we want to swap the LHS with
	stringNodes := make([]*ast.Node, len(s))
	varNodes := make([]*ast.Node, len(s))

	idx := 0
	var_idx := 0
	// first we find all the temp# = function variable nodes
	//tempFuncNodeList := make([]*ast.BasicLit, len(s))
	ast.Inspect(node, func(n ast.Node) bool {
		literal, ok := n.(*ast.BasicLit)
		if ok && literal.Kind == token.STRING && GRLibUtil.StrContains(origXorStrings, strings.Trim(literal.Value, "\"")) {
			fmt.Printf("FOUND %s\n", literal.Value)
			stringNodes[idx] = &n
			idx++
			return true
		}

		varList, ok := n.(*ast.ValueSpec)
		if ok && GRLibUtil.StrContains(varNameStrings, varList.Names[0].Name) {
			fmt.Printf("FOUND CUSTOM VARIABLE NODE: %s\n", varList.Names[0].Name)
			varNodes[var_idx] = &n
			nodeTable[varList.Names[0].Name] = &n
			var_idx++
		}
		return true
	})

	str_idx := 0
	astutil.Apply(
		node,
		func(cursor *astutil.Cursor) bool {
			tmpNode := cursor.Node()
			replVars, ok := tmpNode.(*ast.BasicLit)
			if ok {
				if nodeContainsBasicLit(stringNodes, &tmpNode, idx) {
					fmt.Printf("\n got a node, to replace \n")
					spew.Dump(replVars)
					replVars.Kind = token.FUNC
					replVars.Value = s[str_idx].Stub
					str_idx++
					return true
				}
			}
			return true
		},
		func(cursor *astutil.Cursor) bool {
			return true
		},
	)

	astutil.Apply(
		node,
		func(cursor *astutil.Cursor) bool {
			tmpNode := cursor.Node()
			assign, ok := tmpNode.(*ast.ValueSpec)
			if ok && GRLibUtil.StrContains(varNameStrings, assign.Names[0].Name) {
				fmt.Printf("Cleaning up variable: %s\n", assign.Names[0].Name)
				return true
			}
			return true
		},
		func(cursor *astutil.Cursor) bool {
			return true
		},
	)
	var srcBuf bytes.Buffer
	fset := token.NewFileSet()
	printer.Fprint(&srcBuf, fset, node)
	strSrc := srcBuf.String()
	_ = strSrc
	return node
}

func nodeContainsBasicLit(slice []*ast.Node, item *ast.Node, max int) bool {
	for i := 0; i < max; i++ {
		if *item == *slice[i] {
			return true
		}
	}
	return false
}
