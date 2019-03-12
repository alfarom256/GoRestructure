package GRLibUtil

import "go/ast"

func StrContains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}


func NodeContains(slice []*ast.Node, item *ast.Node, max int) bool {
	for i := 0; i < max; i++ {
		if *item == *slice[i] {
			return true
		}
	}
	return false
}
