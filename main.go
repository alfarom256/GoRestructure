package main

import (
	"GoRestructure/GRLibAST"
	"fmt"
	"github.com/akamensky/argparse"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	art := `

            ENTER THE WIRED

                GoNavi
`
	println(art)
	argParser := argparse.NewParser("GoNavi", "A proof of concept golang obfuscation tool")
	fPath := argParser.String("f", "file", &argparse.Options{Required: true, Help: "File to obfuscate"})
	outputPath := argParser.String("o", "output-path", &argparse.Options{Required: true, Help: "Output path"})
	err := argParser.Parse(os.Args)
	if err != nil {
		log.Fatal(argParser.Usage(err))
		os.Exit(1)
	}
	if outputPath == nil || fPath == nil {
		log.Fatal(argParser.Usage(err))
		os.Exit(2)
	}
	*fPath, err = filepath.Abs(*fPath)
	if err != nil {
		fmt.Printf("Can't find file %s on the system...", *fPath)
		panic(err)
	}
	*fPath = filepath.FromSlash(*fPath)
	*outputPath = filepath.FromSlash(*outputPath)
	pList := GRLibAST.InitLocalPackages(*fPath)
	for i := range pList {
		tmp := pList[i]
		files := tmp.Files
		for f := range files {
			GRLibAST.GenSrcFromFile(files[f].Name, tmp.Name, *outputPath, pList)
			GRLibAST.WriteStubToPackage(*tmp, *outputPath)
		}
		for j := range tmp.SubPackages {
			tmpSub := tmp.SubPackages[j]
			files := tmpSub.Files
			for f := range files {
				GRLibAST.GenSrcFromFile(files[f].Name, tmpSub.Name, *outputPath, pList)
				GRLibAST.RenderFunctions(pList)
				GRLibAST.WriteStubToPackage(tmpSub, *outputPath)
			}
		}
	}
	allFunctions := GRLibAST.RenderFunctions(pList)
	// this should work, just make sure you're parsing the right type of ast node,, inspect the AST for a callexpr again
	GRLibAST.FindAllUsagesInPackage(pList, allFunctions)
	fSplit := strings.Split(*fPath, string(os.PathSeparator))
	fName := fSplit[len(fSplit)-1]
	GRLibAST.AppendStub(*fPath, *fPath)
	GRLibAST.GenSrcFromFile(*fPath, fName, *outputPath, pList)
}
