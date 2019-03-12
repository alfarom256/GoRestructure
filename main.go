package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"log"
	"os"
	"path/filepath"
	"strings"
	"GoRestructure/GRLibAST"
	"GoRestructure/GRLibFile"
)

func main() {
	argParser := argparse.NewParser("GoObfuscate", "A proof of concept golang obfuscation tool")

	// get the arguments we need from the user
	fPath := argParser.String("f", "file", &argparse.Options{Required: true, Help: "File to obfuscate"})

	outputPath := argParser.String("o", "output-path", &argparse.Options{Required: true, Help: "Output path"})
	err := argParser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
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

	pList := GRLibAST.InitLocalPackages(*fPath)
	// pList now contains an entire hierarchy of Golang Packages in the local directory
	// so for every package in the list, we're gonna generate a new source
	for i := range pList {
		tmp := pList[i]
		files := tmp.Files
		for f := range files {
			GRLibFile.GenSrcFromFile(files[f].Name, tmp.Name, *outputPath)
		}

		for j := range tmp.SubPackages {
			tmpSub := tmp.SubPackages[j]
			files := tmpSub.Files
			for f := range files {
				GRLibFile.GenSrcFromFile(files[f].Name, tmpSub.Name, *outputPath)
			}
		}
	}
	fSplit := strings.Split(*fPath, string(os.PathSeparator))
	fName := fSplit[len(fSplit)-1]
	// lastly, generate the source for the main file
	GRLibFile.GenSrcFromFile(*fPath, fName, *outputPath)
}

