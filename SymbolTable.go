package main

import "fmt"

const (
	STATIC = "static"
	FIELD  = "field"
	ARG    = "argument"
	VAR    = "var"
	NONE   = "none"
)

type SymbolTable struct {
	symbolMap   map[string]Symbol
	staticIndex int
	fieldIndex  int
	argIndex    int
	varIndex    int
}

func CreateSymbolTable() *SymbolTable {
	t := &SymbolTable{}
	return t
}

func (t *SymbolTable) Reset() {
	t.symbolMap = make(map[string]Symbol)
	t.staticIndex = 0
	t.fieldIndex = 0
	t.argIndex = 0
	t.varIndex = 0
}

func (t *SymbolTable) Define(name string, sType string, kind string) {
	symbol := Symbol{name: name, sType: sType, kind: kind}
	switch kind {
	case STATIC:
		{
			symbol.index = t.staticIndex
			t.staticIndex++
		}
	case FIELD:
		{
			symbol.index = t.fieldIndex
			t.fieldIndex++
		}
	case ARG:
		{
			symbol.index = t.argIndex
			t.argIndex++
		}
	case VAR:
		{
			symbol.index = t.varIndex
			t.varIndex++
		}
	}
	t.symbolMap[name] = symbol
}

func (t *SymbolTable) VarCount(kind string) int {
	res := 0
	switch kind {
	case STATIC:
		{
			res = t.staticIndex
		}
	case FIELD:
		{
			res = t.fieldIndex
		}
	case ARG:
		{
			res = t.argIndex
		}
	case VAR:
		{
			res = t.varIndex
		}
	}
	return res
}

func (t *SymbolTable) KindOf(name string) string {
	fmt.Println("inside KindOf() name is " + name)
	symbol, ok := t.symbolMap[name]
	if !ok {
		fmt.Println("inside KindOf() none is " + name)
		return NONE
	}
	return symbol.kind
}

func (t *SymbolTable) TypeOf(name string) string {
	return t.symbolMap[name].sType
}

func (t *SymbolTable) IndexOf(name string) int {
	return t.symbolMap[name].index
}
