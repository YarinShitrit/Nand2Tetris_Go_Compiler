package main

import (
	"os"
	"strconv"
)

// SEGMENT CONSTATNS
const (
	CONSTANT = "constant"
	ARGUMENT = "argument"
	LOCAL    = "local"
	THIS     = "this"
	THAT     = "that"
	POINTER  = "pointer"
	TEMP     = "temp"
)

// COMMAND CONSTATNS
const (
	ADD = "add"
	SUB = "sub"
	NEG = "neg"
	EQ  = "eq"
	GT  = "gt"
	LT  = "lt"
	AND = "and"
	OR  = "or"
	NOT = "not"
)

var labelIndex = 0

type VMWriter struct {
	outputFile *os.File
}

func CreateVMWriter(outputFile *os.File) *VMWriter {
	vmw := &VMWriter{outputFile: outputFile}

	return vmw
}

func (vmw *VMWriter) WritePush(segment string, index int) {
	vmw.outputFile.WriteString("push " + segment + " " + strconv.Itoa(index) + "\n")
}

func (vmw *VMWriter) WritePop(segment string, index int) {
	vmw.outputFile.WriteString("pop " + segment + " " + strconv.Itoa(index) + "\n")
}

func (vmw *VMWriter) WriteArithmetic(command string) {
	vmw.outputFile.WriteString(command + "\n")
}

func (vmw *VMWriter) WriteLabel(label string) {
	vmw.outputFile.WriteString("label " + label + "\n")
}

func (vmw *VMWriter) WriteGoTo(label string) {
	vmw.outputFile.WriteString("goto " + label + "\n")
}

func (vmw *VMWriter) WriteIf(label string) {
	vmw.outputFile.WriteString("if-goto " + label + "\n")
}

func (vmw *VMWriter) WriteCall(name string, nArgs int) {
	vmw.outputFile.WriteString("name " + name + " " + strconv.Itoa(nArgs) + "\n")
}

func (vmw *VMWriter) WriteFunction(name string, nArgs int) {
	vmw.outputFile.WriteString("function " + name + " " + strconv.Itoa(nArgs) + "\n")
}

func (vmw *VMWriter) WriteReturn() {
	vmw.outputFile.WriteString("return\n")
}

func (vmw *VMWriter) Close() {
	vmw.outputFile.Close()
}

func (vmw *VMWriter) getSegmentOf(kind string) string {
	switch kind {
	case FIELD:
		{
			return THIS
		}
	case STATIC:
		{
			return STATIC
		}
	case VAR:
		{
			return LOCAL
		}
	case ARG:
		{
			return ARG
		}
	default:
		{
			return NONE
		}
	}
}

func (vmw *VMWriter) CreateLabel() string {
	label := "LABEL_" + strconv.Itoa(labelIndex)
	labelIndex++
	return label
}
