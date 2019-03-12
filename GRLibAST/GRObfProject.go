package GRLibAST

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Package struct {
	Name        string
	Path        string
	Files       []PackageFile
	Dirs        []string
	SubPackages []Package
}

type PackageFile struct {
	Name           string
	FileAST        *ast.File
	FileNodeSource *NodeSource
}

func InitLocalPackages(fPath string) []*Package {
	// this will hold all the package objects for each package we find during parsing
	var preParsePackages []*Package

	// this holds the
	// TODO: add an ignore option maybe?
	files, err := filesInDirectory(fPath)
	if err != nil {
		panic(err)
	}
	_ = files
	dirs, err := dirsInDirectory(fPath)
	if err != nil {
		panic(err)
	}

	// for all root level Dirs
	for _, dir := range dirs {
		// if it's a package
		if isPackage(dir) {
			// get a package object from it
			tmpPkg, isDir := packageFromDir(dir)
			if isDir && tmpPkg != nil { // if we found a directory that is a package
				preParsePackages = append(preParsePackages, tmpPkg) // jam it into our package array
			}
		}
	}

	return preParsePackages
}

// returns the full path of the file
func filesInDirectory(fPath string) ([]string, error) {
	var retVal []string
	fileSplit := strings.Split(fPath, "\\")
	// get the Name of the folder
	name := fileSplit[len(fileSplit)-2]
	name = strings.Trim(name, string(os.PathSeparator))
	root := filepath.Dir(fPath + string(os.PathSeparator)) // good golly miss molly I hate this
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		lPkg, _ := parsePackageFromFile(path)
		if !info.IsDir() && lPkg == name {
			retVal = append(retVal, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return retVal, nil
}

func isPackage(path string) bool {
	// get the package Name from the dir
	if string(path[len(path)-1]) != string(os.PathSeparator) {
		path += string(os.PathSeparator)
	}
	tmp := strings.Split(path, string(os.PathSeparator))
	pathName := tmp[len(tmp)-2]
	// holds all the go Files for testing in the local directory
	var goFiles []string

	// get all the Files in the directory
	files, err := filesInDirectory(path)
	if err != nil {
		log.Fatal(err)
		log.Fatal("fatal error in isPackage")
		return false
	}
	for i := range files {
		if filepath.Ext(files[i]) == ".go" {
			goFiles = append(goFiles, files[i])
		}
	}
	if len(goFiles) == 0 {
		return false
	}
	for i := range goFiles {
		filePackage, err := parsePackageFromFile(goFiles[i])
		if err != nil {
			fmt.Printf("Error in Parsing go file: %s", goFiles[i])
		}
		if filePackage == pathName {
			return true
		}
	}
	return false
}

// gets the package Name of the file from the AST
func parsePackageFromFile(file string) (string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, file, nil, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}
	tmp := node.Name.Name
	return tmp, nil
}

func packageFromDir(file string) (*Package, bool) {
	var tmpPackage *Package
	src, err := os.Stat(file)
	if err == nil {
		if src.IsDir() && file[0] != '.' {
			file = FixDirPath(file)
			fileSplit := strings.Split(file, string(os.PathSeparator))
			// get the Name of the folder
			name := fileSplit[len(fileSplit)-2]

			// make a new temp package object
			tmpPackage = &Package{name, file, nil, nil, nil}

			// get the Files in the directory
			tmpFiles, err := filesInDirectory(FixDirPath(file))
			if err != nil {
				fmt.Printf("error in FilesInDirectory ObfProject.go") // fix this shitty way to do this
				panic(err)
			}

			// get the directories in the directory
			tmpDirs, err := dirsInDirectory(file + string(os.PathSeparator)) // fix this shit
			if err != nil {
				fmt.Printf("error in DirsInDirectory ObfProject.go")
				panic(err)
			}

			/*
				Convert the array of string files into PackageFiles


				type PackageFile struct {
					Name           string
					FileAST        *ast.Node
					FileNodeSource *NodeSource
				}

			*/

			tmpPackage.Files = makePkgFilesFromPathList(tmpFiles)
			tmpPackage.Dirs = tmpDirs

			for i := range tmpPackage.Dirs {
				tmpPackage.Dirs[i] = FixDirPath(tmpPackage.Dirs[i])
			}

			// now that we have populated all the other information, let's scan each directory
			// and search for Files who's package names are equal to the directory Name
			// once we do this, we should end up with a tree-like structure of *Package
			var subPackages []Package

			// for every directory within
			for i := range tmpPackage.Dirs {
				iter_err := filepath.Walk(tmpPackage.Dirs[i], func(path string, f os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if f.IsDir() && isPackage(path) { // we got a winner

						// make a new package
						tmpSubPackage := Package{"", path, nil, nil, nil}
						tmp_name := strings.Split(path, string(os.PathSeparator))
						tmpSubPackage.Name = tmp_name[len(tmp_name)-2]
						tmpSubFiles, err := filesInDirectory(file + string(os.PathSeparator))

						if err != nil {
							fmt.Printf("error in FilesInDirectory ObfProject.go") // fix this shitty way to do this
							panic(err)
						}

						// get the directories in the directory
						tmpSubDirs, err := dirsInDirectory(file + string(os.PathSeparator)) // fix this shit
						if err != nil {
							fmt.Printf("error in DirsInDirectory ObfProject.go")
							panic(err)
						}
						tmpSubPackage.Dirs = tmpSubDirs
						for i := range tmpSubPackage.Dirs {
							tmpSubPackage.Dirs[i] = FixDirPath(tmpSubPackage.Dirs[i])
						}
						tmpSubPackage.Files = makePkgFilesFromPathList(tmpSubFiles)
						subPackages = append(subPackages, tmpSubPackage)
						return nil
					}
					return nil
				})

				if iter_err != nil {
					fmt.Printf("Fatal Error in walking filepath")
				}
			}
			tmpPackage.SubPackages = subPackages
		} else {
			return nil, true
		}
		return tmpPackage, true
	} else {
		log.Fatal(err)
		log.Fatal("Error parsing")
		return nil, false
	}
}

func packageFileInit(fname string) *PackageFile {
	retVal := PackageFile{fname, nil, nil}
	_ = retVal

	return nil
}

func FixDirPath(f string) string {
	if string(f[len(f)-1]) != string(os.PathSeparator) {
		return f + string(os.PathSeparator)
	}
	return f
}

func dirsInDirectory(fPath string) ([]string, error) {
	var retVal []string
	root := filepath.Dir(fPath)
	files, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}
	for i := range files {
		if files[i].IsDir() && files[i].Name()[0] != '.' {
			absPath := root + string(os.PathSeparator) + files[i].Name()
			if err != nil {
				fmt.Printf("Can't get abs Path of child file??? tf?")
			}
			retVal = append(retVal, absPath)
		}
	}
	return retVal, nil
}

func makePkgFilesFromPathList(tmpFiles []string) []PackageFile {
	tmpPkgFiles := make([]PackageFile, len(tmpFiles))
	for i := range tmpFiles {
		tmp := PackageFile{tmpFiles[i], nil, nil}
		tmp.FileAST = GetASTFile(tmp.Name)
		tmpPkgFiles[i] = tmp
	}
	return tmpPkgFiles
}
