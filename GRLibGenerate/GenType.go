package GRLibGenerate

import (
	"fmt"
	"math/rand"
)

// ripped from
// https://github.com/golang/go/blob/master/src/reflect/type.go
type Kind uint

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Ptr
	Slice
	String
	Struct
	UnsafePointer
)

// 26 kinds [0-25)
var kindNames = []string{
	Invalid:       "invalid",
	Bool:          "bool",
	Int:           "int",
	Int8:          "int8",
	Int16:         "int16",
	Int32:         "int32",
	Int64:         "int64",
	Uint:          "uint",
	Uint8:         "uint8",
	Uint16:        "uint16",
	Uint32:        "uint32",
	Uint64:        "uint64",
	Uintptr:       "uintptr",
	Float32:       "float32",
	Float64:       "float64",
	Complex64:     "complex64",
	Complex128:    "complex128",
	Array:         "array", // 17
	Chan:          "chan",
	Func:          "func", // 19
	Interface:     "interface",
	Map:           "map", //21
	Ptr:           "ptr", // 22
	Slice:         "slice",
	String:        "string",
	Struct:        "struct",
	UnsafePointer: "unsafe.Pointer",
}

// end content https://github.com/golang/go/blob/master/src/reflect/type.go

type GeneratedType struct {
	Value            string
	TypeID           []int
	NeedsConstructor bool // follow up with new()
}

func GenerateTypeString(TypeID []int) string {
	retVal := ""
	// if TypeID is > 1, expand the type out
	// really this should only apply to compound or list types
	// i.e. map[int]string would reduce to map, int, string
	// array, pointer, struct
	nestedSep := "[%s]"
	nestCount := 0
	if len(TypeID) > 1 {
		for i := range TypeID {
			if nestCount > 0 {
				retVal += fmt.Sprintf(nestedSep, kindNames[TypeID[i]])
				nestCount--
				continue
			}
			if kindNames[TypeID[i]] == "map" {
				nestCount++
			} else if kindNames[TypeID[i]] == "array" {
				retVal += "[]"
				continue
			} else if kindNames[TypeID[i]] == "ptr" {
				retVal += "*"
				continue
			}

			retVal += kindNames[TypeID[i]]
		}
		// return the compound type
		return retVal
	}
	// get the index of the type
	// return it
	return kindNames[TypeID[0]]
}

func GetTypeFromIntID(id int) string {
	return kindNames[id]
}

func GenerateRandomType() *GeneratedType {
	retVal := new(GeneratedType)
	retVal.NeedsConstructor = false
	idx := 0
	if rand.Int()%50 >= 35 {
		length := rand.Intn(4) + 1 // MAX length of 3
		retVal.TypeID = make([]int, length)
		switch length {
		case 4:
			retVal.TypeID[0] = []int{17, 22}[rand.Int()%2]
			idx++
			retVal.NeedsConstructor = true
			retVal.TypeID[1] = 21
			idx++
		case 3:
			retVal.TypeID[0] = []int{17, 22}[rand.Int()%2]
			idx++
			retVal.NeedsConstructor = true
			retVal.TypeID[1] = 22
			idx++
		case 2:
			retVal.TypeID[0] = []int{17, 22}[rand.Int()%2]
			idx++

		}
		// if the length is 4, make it an array or pointer
		for i := idx; i < length; i++ {
			retVal.TypeID[i] = rand.Intn(16) + 1
		}
	} else {
		retVal.TypeID = []int{rand.Intn(16) + 1}
	}
	retVal.Value = GenerateTypeString(retVal.TypeID)
	return retVal
}
