package main

import (
	"fmt"
	"os"
	"strconv"
)

var tab = 0

type CompilationEngine struct {
	jt                    *JackTokenizer
	vmw                   *VMWriter
	classSymbolTable      *SymbolTable
	subroutineSymbolTable *SymbolTable
	currentClass          string
	currentSubroutine     string
}

func CreateCompilationEngine(inputFile *os.File, outputFile *os.File) *CompilationEngine {
	cEngine := &CompilationEngine{jt: CreateTokenizer(inputFile), vmw: CreateVMWriter(outputFile)}
	return cEngine
}

func (cEngine *CompilationEngine) CompileClass() {
	cEngine.classSymbolTable = CreateSymbolTable()
	cEngine.subroutineSymbolTable = CreateSymbolTable()
	cEngine.classSymbolTable.Reset()
	cEngine.checkToken("class")

	tab++

	cEngine.jt.Advance() // class name
	cEngine.currentClass = cEngine.jt.CurrentToken()

	cEngine.jt.Advance()
	cEngine.checkToken("{")

	cEngine.jt.Advance()

	//check for field or static variables
	for cEngine.jt.CurrentToken() == "field" || cEngine.jt.CurrentToken() == "static" {
		cEngine.CompileClassVarDec()
	}

	//check for subroutines
	for cEngine.jt.CurrentToken() == "constructor" || cEngine.jt.CurrentToken() == "function" || cEngine.jt.CurrentToken() == "method" {
		cEngine.CompileSubroutine()
	}

	cEngine.checkToken("}")

	tab--

	cEngine.vmw.Close()
}

func (cEngine *CompilationEngine) CompileClassVarDec() {
	symbolKind := cEngine.jt.CurrentToken()

	tab++

	cEngine.jt.Advance() // type
	symbolType := cEngine.jt.CurrentToken()
	if cEngine.jt.TokenType() == KEYWORD { // primitive type

	} else { // class type

	}
	cEngine.jt.Advance() // name
	symbolName := cEngine.jt.CurrentToken()
	cEngine.classSymbolTable.Define(symbolName, symbolType, symbolKind)

	cEngine.jt.Advance()

	// check for more variables of same type in this line
	for cEngine.jt.CurrentToken() != ";" {
		cEngine.checkToken(",")

		cEngine.jt.Advance() // variable name
		cEngine.classSymbolTable.Define(cEngine.jt.CurrentToken(), symbolType, symbolKind)
		cEngine.writeVar()
	}
	cEngine.checkToken(";")

	cEngine.jt.Advance()
	tab--

}

func (cEngine *CompilationEngine) CompileSubroutine() {
	cEngine.subroutineSymbolTable.Reset()

	tab++
	cEngine.subroutineSymbolTable.Define("this", cEngine.currentClass, ARG)
	switch cEngine.jt.CurrentToken() {
	case "constructor":
		{
			cEngine.jt.Advance()
			cEngine.checkTokenType("identifier")
			cEngine.currentSubroutine = cEngine.currentClass + "." + cEngine.jt.CurrentToken()
			cEngine.vmw.WriteFunction(cEngine.currentSubroutine, cEngine.subroutineSymbolTable.argIndex)
			cEngine.vmw.WritePush(CONSTANT, cEngine.classSymbolTable.VarCount(FIELD))
			cEngine.vmw.WriteCall("Memory.alloc", 1)
			cEngine.vmw.WritePop(POINTER, 0)
			cEngine.jt.Advance() // "new"
			cEngine.checkToken("new")

		}
	case "method", "function":
		{
			cEngine.jt.Advance() //type
			cEngine.jt.Advance() // method/function name
			cEngine.currentSubroutine = cEngine.currentClass + "." + cEngine.jt.CurrentToken()
			cEngine.vmw.WriteFunction(cEngine.currentSubroutine, cEngine.subroutineSymbolTable.argIndex)
			if cEngine.jt.CurrentToken() == "method" {
				cEngine.vmw.WritePush(ARG, 0)
				cEngine.vmw.WritePop(POINTER, 0)
			}
			cEngine.checkTokenType("identifier")

		}
	}
	cEngine.jt.Advance() // "("
	cEngine.checkToken("(")

	cEngine.jt.Advance()
	// check for parameters

	tab++
	cEngine.CompileParameterList()
	tab--

	cEngine.checkToken(")")

	cEngine.jt.Advance()
	cEngine.CompileSubroutineBody()
	tab--

}

func (cEngine *CompilationEngine) CompileSubroutineBody() {
	cEngine.checkToken("{")
	tab++

	cEngine.jt.Advance()
	//check for var decs
	for cEngine.jt.CurrentToken() == "var" {
		cEngine.CompileVarDec()
	}

	//check for statements
	if cEngine.isStatement(cEngine.jt.CurrentToken()) {
		cEngine.CompileStatements()
	}
	cEngine.checkToken("}")

	tab--

	cEngine.jt.Advance()
}

func (cEngine *CompilationEngine) CompileParameterList() {
	for cEngine.jt.TokenType() == KEYWORD || cEngine.jt.TokenType() == IDENTIFIER {
		symbolType := cEngine.jt.TokenType()
		symbolName := ""
		if cEngine.jt.TokenType() == KEYWORD {

			cEngine.jt.Advance() // parameter name
			symbolName = cEngine.jt.CurrentToken()
			cEngine.checkTokenType("identifier")

			cEngine.jt.Advance() // , or )
		} else {

			cEngine.jt.Advance() // parameter name
			symbolName = cEngine.jt.CurrentToken()
			cEngine.checkTokenType("identifier")

			cEngine.jt.Advance() // , or )
		}
		cEngine.subroutineSymbolTable.Define(symbolName, symbolType, ARG)
		//check for more parameters
		if cEngine.jt.CurrentToken() == "," {

			cEngine.jt.Advance()
			cEngine.CompileParameterList()
		}
	}
}

func (cEngine *CompilationEngine) CompileVarDec() {

	tab++

	cEngine.jt.Advance() // var type
	symbolType := cEngine.jt.CurrentToken()
	if cEngine.jt.TokenType() == KEYWORD { //primitive type

	} else { // class type

	}
	cEngine.jt.Advance() //var name
	cEngine.checkTokenType(IDENTIFIER)
	cEngine.subroutineSymbolTable.Define(cEngine.jt.CurrentToken(), symbolType, VAR)

	cEngine.jt.Advance() // , or ;
	for cEngine.jt.CurrentToken() == "," {

		cEngine.jt.Advance() // var name
		cEngine.subroutineSymbolTable.Define(cEngine.jt.CurrentToken(), symbolType, VAR)
		cEngine.writeVar()
	}
	cEngine.checkToken(";")

	tab--

	cEngine.jt.Advance()
}

func (cEngine *CompilationEngine) CompileStatements() {

	tab++
	for cEngine.isStatement(cEngine.jt.CurrentToken()) {
		switch cEngine.jt.CurrentToken() {
		case "let":
			{
				cEngine.CompileLet()
			}
		case "if":
			{
				cEngine.CompileIf()
			}
		case "while":
			{
				cEngine.CompileWhile()
			}
		case "do":
			{
				cEngine.CompileDo()
			}
		case "return":
			{
				cEngine.CompileReturn()
			}
		}
	}
	tab--

}

func (cEngine *CompilationEngine) CompileLet() {
	isArr := false

	tab++

	cEngine.jt.Advance() // var name
	cEngine.checkTokenType("identifier")
	symbolName := cEngine.jt.CurrentToken()
	fmt.Println("symbol name is " + symbolName)
	symbolKind := cEngine.subroutineSymbolTable.KindOf(symbolName)

	cEngine.jt.Advance() // "[" or "="
	if cEngine.jt.CurrentToken() == "[" {
		isArr = true

		cEngine.vmw.WritePush(cEngine.vmw.getSegmentOf(symbolKind), cEngine.subroutineSymbolTable.IndexOf(symbolName))
		cEngine.jt.Advance()
		cEngine.CompileExpression()
		cEngine.checkToken("]")
		cEngine.vmw.WriteArithmetic(ADD)

		cEngine.jt.Advance()
	}
	if cEngine.jt.CurrentToken() == "=" {

		cEngine.jt.Advance()
		cEngine.CompileExpression()
	}
	cEngine.checkToken(";")
	if isArr {
		cEngine.vmw.WritePop(TEMP, 0)
		cEngine.vmw.WritePop(POINTER, 1)
		cEngine.vmw.WritePop(TEMP, 0)
		cEngine.vmw.WritePop(THAT, 0)
	} else {
		cEngine.vmw.WritePop(cEngine.vmw.getSegmentOf(symbolKind), cEngine.subroutineSymbolTable.IndexOf(symbolName))
	}

	tab--

	cEngine.jt.Advance()
}

func (cEngine *CompilationEngine) CompileIf() {

	tab++

	cEngine.jt.Advance()
	cEngine.checkToken("(")

	cEngine.jt.Advance()
	cEngine.CompileExpression()
	label1 := cEngine.vmw.CreateLabel()
	label2 := cEngine.vmw.CreateLabel()
	cEngine.vmw.WriteArithmetic(NOT)
	cEngine.vmw.WriteIf(label1)
	cEngine.checkToken(")")

	cEngine.jt.Advance()
	cEngine.checkToken("{")

	cEngine.jt.Advance()
	cEngine.CompileStatements()
	cEngine.checkToken("}")

	cEngine.jt.Advance() // else?
	cEngine.vmw.WriteGoTo(label2)
	cEngine.vmw.WriteLabel(label1)
	if cEngine.jt.CurrentToken() == "else" {

		cEngine.jt.Advance()
		cEngine.checkToken("{")

		cEngine.jt.Advance()
		cEngine.CompileStatements()
		cEngine.checkToken("}")

		cEngine.jt.Advance()
	}
	cEngine.vmw.WriteLabel(label2)
	tab--

}

func (cEngine *CompilationEngine) CompileWhile() {

	tab++

	cEngine.jt.Advance() // "("
	cEngine.checkToken("(")

	cEngine.jt.Advance()
	label1 := cEngine.vmw.CreateLabel()
	label2 := cEngine.vmw.CreateLabel()
	cEngine.vmw.WriteLabel(label1)
	cEngine.CompileExpression()
	cEngine.checkToken(")")

	cEngine.jt.Advance() // "{"
	cEngine.checkToken("{")
	cEngine.vmw.WriteArithmetic(NOT)
	cEngine.vmw.WriteIf(label2)

	cEngine.jt.Advance()
	cEngine.CompileStatements()
	cEngine.checkToken("}")
	cEngine.vmw.WriteGoTo(label1)

	tab--

	cEngine.vmw.WriteLabel(label2)
	cEngine.jt.Advance()
}

func (cEngine *CompilationEngine) CompileDo() {

	tab++

	cEngine.jt.Advance() // subroutineName or className/varName
	cEngine.checkTokenType("identifier")
	cEngine.CompileSubroutineCall()
	cEngine.checkToken(";")
	cEngine.vmw.WritePop(TEMP, 0) // pop the return value

	tab--

	cEngine.jt.Advance()
}

func (cEngine *CompilationEngine) CompileReturn() {

	tab++

	cEngine.jt.Advance() // ; or experssion
	if cEngine.jt.CurrentToken() == ";" {
		cEngine.vmw.WritePush(CONSTANT, 0)

	} else {
		cEngine.CompileExpression()
		cEngine.checkToken(";")

	}
	cEngine.vmw.WriteReturn()
	tab--

	cEngine.jt.Advance()
}

func (cEngine *CompilationEngine) CompileExpression() {

	tab++
	cEngine.CompileTerm()
	for cEngine.isOp(cEngine.jt.CurrentToken()) {
		opCmd := cEngine.CompileOp()
		cEngine.CompileTerm()
		cEngine.vmw.outputFile.WriteString(opCmd + "\n")
	}
	tab--

}

func (cEngine *CompilationEngine) CompileTerm() {

	tab++
	if cEngine.jt.CurrentToken() == "~" || cEngine.jt.CurrentToken() == "-" { //unaryOp term

		cEngine.jt.Advance()
		cEngine.CompileTerm()
		if cEngine.jt.CurrentToken() == "~" {
			cEngine.vmw.WriteArithmetic(NEG)
		} else {
			cEngine.vmw.WriteArithmetic(NOT)
		}
	} else if cEngine.jt.TokenType() == STRING_CONST ||
		cEngine.jt.TokenType() == INT_CONST ||
		cEngine.jt.TokenType() == KEYWORD { //constant

		if cEngine.jt.TokenType() == INT_CONST {
			cEngine.vmw.WritePush(CONSTANT, cEngine.jt.IntVal())
		} else if cEngine.jt.TokenType() == STRING_CONST {
			str := cEngine.jt.StringVal()
			cEngine.vmw.WritePush(CONSTANT, len(str))
			cEngine.vmw.WriteCall("String.new", 1)
			for _, ch := range str {
				index, err := strconv.Atoi(strconv.QuoteRune(ch))
				if err != nil {
					cEngine.vmw.WritePush(CONSTANT, index)
					cEngine.vmw.WriteCall("String.appendChar", 2)
				}
			}
		}

		cEngine.jt.Advance()
	} else if varType := cEngine.jt.TokenType(); varType == "identifier" { //varName
		varName := cEngine.jt.CurrentToken()
		cEngine.vmw.WritePush(cEngine.vmw.getSegmentOf(cEngine.subroutineSymbolTable.KindOf(varName)), cEngine.subroutineSymbolTable.IndexOf(varName))
		cEngine.jt.Advance()
		if cEngine.jt.CurrentToken() == "(" || cEngine.jt.CurrentToken() == "." { // subroutineCall
			cEngine.CompileSubroutineCall()
		} else if cEngine.jt.CurrentToken() == "[" { // varName [experssion]

			cEngine.jt.Advance()
			cEngine.CompileExpression()
			cEngine.checkToken("]")
			cEngine.vmw.WriteArithmetic(ADD)
			cEngine.vmw.WritePop(POINTER, 1)
			cEngine.vmw.WritePush(THAT, 0)

			cEngine.jt.Advance()
		}
	} else if cEngine.jt.CurrentToken() == "(" { // (expressionList)

		cEngine.jt.Advance()
		cEngine.CompileExpression()
		cEngine.checkToken(")")

		cEngine.jt.Advance()
	}
	tab--

}

func (cEngine *CompilationEngine) CompileExpressionList() int {

	tab++
	sum := 0
	if cEngine.jt.CurrentToken() != ")" { // no more experssions
		cEngine.CompileExpression()
		for cEngine.jt.CurrentToken() == "," {

			cEngine.jt.Advance()
			cEngine.CompileExpression()
		}
	}
	tab--

	return sum
}

func (cEngine *CompilationEngine) CompileSubroutineCall() {
	if cEngine.jt.TokenType() == IDENTIFIER {

		cEngine.jt.Advance() // "." or subroutineName
	}
	if cEngine.jt.CurrentToken() == "." {

		cEngine.jt.Advance() // subrountineName

		cEngine.jt.Advance()
	}
	cEngine.checkToken("(")
	cEngine.vmw.WritePush(POINTER, 0)
	cEngine.jt.Advance()
	nArgs := cEngine.CompileExpressionList() + 1
	cEngine.checkToken(")")
	cEngine.vmw.WriteCall(cEngine.currentClass+"."+cEngine.currentSubroutine, nArgs)
	cEngine.jt.Advance()
}

func (cEngine *CompilationEngine) CompileOp() string {
	opCmd := ""
	switch cEngine.jt.CurrentToken() {
	case "+":
		{
			opCmd = "add"

		}
	case "-":
		{
			opCmd = "sub"

		}
	case "*":
		{
			opCmd = "call Math.multiply 2"

		}
	case "/":
		{
			opCmd = "call Math.divide 2"

		}
	case "&":
		{
			opCmd = "and"

		}
	case "|":
		{
			opCmd = "or"

		}
	case "<":
		{
			opCmd = "lt"

		}
	case ">":
		{
			opCmd = "gt"

		}
	case "=":
		{
			opCmd = "eq"

		}
	}
	cEngine.jt.Advance()
	return opCmd
}

func (cEngine *CompilationEngine) writeVar() {
	cEngine.checkTokenType(IDENTIFIER)

	cEngine.jt.Advance()
}

func (cEngine *CompilationEngine) checkToken(token string) {
	if cEngine.jt.CurrentToken() != token {
		fmt.Println("compilation error - expected " + token + " but recevied " + cEngine.jt.CurrentToken())
		panic(1)
	}
}

func (cEngine *CompilationEngine) checkTokenType(tokenType string) {
	if cEngine.jt.TokenType() != tokenType {
		fmt.Println("compilation error - expected type " + tokenType + " but recevied " + cEngine.jt.CurrentToken())
		panic(1)
	}
}

func (cEngine *CompilationEngine) isOp(token string) bool {
	if token == "+" ||
		token == "-" ||
		token == "*" ||
		token == "/" ||
		token == "&" ||
		token == "|" ||
		token == "<" ||
		token == ">" ||
		token == "=" {
		return true
	}
	return false
}

func (cEngine *CompilationEngine) isStatement(token string) bool {
	if cEngine.jt.CurrentToken() == "let" ||
		cEngine.jt.CurrentToken() == "if" ||
		cEngine.jt.CurrentToken() == "while" ||
		cEngine.jt.CurrentToken() == "do" ||
		cEngine.jt.CurrentToken() == "return" {
		return true
	}
	return false
}

func tabber(num int) string {
	tabber := ""
	for i := 0; i < num; i++ {
		tabber += "  "
	}
	return tabber
}
