package GRLibAST

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"strings"
)

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

func xorStub() stringObfStub {
	codeFuncStub := `
func obfs(s []byte, k []byte) []byte {
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
`
	function_call_fmt := "string(obfs([]byte(\"%x\"),[]byte(\"%s\")))"
	name := "obfs"
	var argc uint8 = 2
	return stringObfStub{argc, name, codeFuncStub, function_call_fmt}
}

func AppendStub(inAst *ast.File, fset *token.FileSet, s stringObfStub) *ast.File {
	var srcBuf bytes.Buffer
	printer.Fprint(&srcBuf, fset, inAst)
	strSrc := srcBuf.String()

	// check and see if the dec Stub is already in the file
	mySource := NodeSource{}
	mySource = *ParseNodeSource(inAst)
	for i := range mySource.FunctionDecl {
		localFName := mySource.FunctionDecl[i].Name.Name
		if localFName == s.name {
			return nil
		}
	}

	strSrc += s.function_stub
	ret_val, err := parser.ParseFile(fset, "", strSrc, parser.ParseComments)
	if err != nil {
		// freak the fuck out
		panic(err)
	}
	return ret_val
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
		if ok && literal.Kind == token.STRING && strContains(origXorStrings, strings.Trim(literal.Value, "\"")) {
			fmt.Printf("FOUND %s\n", literal.Value)
			stringNodes[idx] = &n
			idx++
			return true
		}

		varList, ok := n.(*ast.ValueSpec)
		if ok && strContains(varNameStrings, varList.Names[0].Name) {
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
			if ok && strContains(varNameStrings, assign.Names[0].Name) {
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

func strContains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}
