package GRLibGenerate

const (
	// format order
	// function name, args, return type/named return, body
	FuncStub = `
func %s(%s) %s{
%s
}
` // format order
	// variable name, type
	// x, int,
	FuncArgStub = `%s, %s,`

	// format order
	// LHS operands, RHS Operands
	// can be any type
	DeclStub = `%s := %s`
)

type GeneratedFunction struct {
	FuncName    string
	FuncBody    string
	FuncArgs    []GeneratedVariable
	NamedReturn bool
	ReturnType  *GeneratedVariable
}
