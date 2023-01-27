package main

import (
	"fmt"
	"os"
	"strings"
)

var subroutineCallName = ""
var subroutineCallArgs = 0
var currentSubroutineType string
var name = ""

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
	cEngine.vmw.Close()
}

func (cEngine *CompilationEngine) CompileClassVarDec() {
	symbolKind := cEngine.jt.CurrentToken()
	cEngine.jt.Advance() // type
	symbolType := cEngine.jt.CurrentToken()
	cEngine.jt.Advance() // name
	symbolName := cEngine.jt.CurrentToken()
	cEngine.classSymbolTable.Define(symbolName, symbolType, symbolKind)
	cEngine.jt.Advance()

	// check for more variables of same type in this line
	for cEngine.jt.CurrentToken() != ";" {
		cEngine.checkToken(",")
		cEngine.jt.Advance() // variable name
		cEngine.checkTokenType(IDENTIFIER)
		cEngine.classSymbolTable.Define(cEngine.jt.CurrentToken(), symbolType, symbolKind)
		cEngine.jt.Advance()
	}
	cEngine.checkToken(";")
	cEngine.jt.Advance()
}

func (cEngine *CompilationEngine) CompileSubroutine() {
	cEngine.subroutineSymbolTable.Reset()
	currentSubroutineType = cEngine.jt.CurrentToken()
	switch currentSubroutineType {
	case "constructor":
		{
			cEngine.jt.Advance()
			cEngine.checkTokenType("identifier")
			cEngine.currentSubroutine = cEngine.currentClass + ".new"
			cEngine.jt.Advance() // "new"
			cEngine.checkToken("new")

		}
	case "method":
		{
			cEngine.jt.Advance()                     //type
			cEngine.jt.Advance()                     // method/function name
			cEngine.subroutineSymbolTable.argIndex++ // add "this" as argument for a method
			cEngine.currentSubroutine = cEngine.currentClass + "." + cEngine.jt.CurrentToken()
			cEngine.checkTokenType("identifier")
		}
	case "function":
		{
			cEngine.jt.Advance() //type
			cEngine.jt.Advance() // method/function name
			cEngine.currentSubroutine = cEngine.currentClass + "." + cEngine.jt.CurrentToken()
			cEngine.checkTokenType("identifier")
		}
	}
	cEngine.jt.Advance() // "("
	cEngine.checkToken("(")
	cEngine.jt.Advance()

	// check for parameters
	cEngine.CompileParameterList()
	cEngine.checkToken(")")
	cEngine.jt.Advance()
	cEngine.CompileSubroutineBody()
}

func (cEngine *CompilationEngine) CompileSubroutineBody() {
	cEngine.checkToken("{")
	cEngine.jt.Advance()
	//check for var decs
	for cEngine.jt.CurrentToken() == "var" {
		cEngine.CompileVarDec()
	}
	cEngine.vmw.WriteFunction(cEngine.currentSubroutine, cEngine.subroutineSymbolTable.VarCount(VAR))
	switch currentSubroutineType {
	case "method":
		{
			cEngine.vmw.WritePush(ARG, 0)
			cEngine.vmw.WritePop(POINTER, 0)
		}
	case "constructor":
		{
			cEngine.vmw.WritePush(CONSTANT, cEngine.classSymbolTable.VarCount(FIELD))
			cEngine.vmw.WriteCall("Memory.alloc", 1)
			cEngine.vmw.WritePop(POINTER, 0)
		}
	}
	//check for statements
	if cEngine.isStatement(cEngine.jt.CurrentToken()) {
		cEngine.CompileStatements()
	}
	cEngine.checkToken("}")
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
	cEngine.jt.Advance() // var type
	symbolType := cEngine.jt.CurrentToken()
	cEngine.jt.Advance() //var name
	cEngine.checkTokenType(IDENTIFIER)
	cEngine.subroutineSymbolTable.Define(cEngine.jt.CurrentToken(), symbolType, VAR)
	cEngine.jt.Advance() // , or ;
	for cEngine.jt.CurrentToken() == "," {

		cEngine.jt.Advance() // var name
		cEngine.checkTokenType(IDENTIFIER)
		cEngine.subroutineSymbolTable.Define(cEngine.jt.CurrentToken(), symbolType, VAR)
		cEngine.jt.Advance()
	}
	cEngine.checkToken(";")

	cEngine.jt.Advance()
}

func (cEngine *CompilationEngine) CompileStatements() {

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
}

func (cEngine *CompilationEngine) CompileLet() {
	isArr := false
	cEngine.jt.Advance() // var name
	cEngine.checkTokenType("identifier")
	symbolName := cEngine.jt.CurrentToken()
	symbolKind := cEngine.subroutineSymbolTable.KindOf(symbolName)
	symbolIndex := cEngine.subroutineSymbolTable.IndexOf(symbolName)
	if symbolKind == NONE { //try get it from class level
		symbolKind = cEngine.classSymbolTable.KindOf(symbolName)
		symbolIndex = cEngine.classSymbolTable.IndexOf(symbolName)
	}
	cEngine.jt.Advance() // "[" or "="
	if cEngine.jt.CurrentToken() == "[" {
		isArr = true
		cEngine.jt.Advance()
		cEngine.CompileExpression()
		cEngine.checkToken("]")
		cEngine.vmw.WritePush(cEngine.vmw.getSegmentOf(symbolKind), symbolIndex)
		cEngine.vmw.WriteArithmetic(ADD)
		cEngine.jt.Advance()
	}
	if cEngine.jt.CurrentToken() == "=" {
		//fmt.Println("var name: " + symbolName)
		cEngine.jt.Advance()
		cEngine.CompileExpression()
	}
	cEngine.checkToken(";")
	if isArr {
		cEngine.vmw.WritePop(TEMP, 0)
		cEngine.vmw.WritePop(POINTER, 1)
		cEngine.vmw.WritePush(TEMP, 0)
		cEngine.vmw.WritePop(THAT, 0)
	} else {
		cEngine.vmw.WritePop(cEngine.vmw.getSegmentOf(symbolKind), symbolIndex)
		//cEngine.vmw.WritePush(LOCAL, 0)
	}
	cEngine.jt.Advance()
}

func (cEngine *CompilationEngine) CompileIf() {
	cEngine.jt.Advance()
	cEngine.checkToken("(")
	cEngine.jt.Advance()
	cEngine.CompileExpression()
	if_true_label := "IF_" + cEngine.vmw.CreateLabel()
	if_false_label := "FALSEIF_" + cEngine.vmw.CreateLabel()
	if_continuation_label := "CONTIF_" + cEngine.vmw.CreateLabel()
	cEngine.vmw.WriteIf(if_true_label)
	cEngine.vmw.WriteGoTo(if_false_label)
	cEngine.vmw.WriteLabel(if_true_label)
	cEngine.checkToken(")")
	cEngine.jt.Advance()
	cEngine.checkToken("{")
	cEngine.jt.Advance()
	cEngine.CompileStatements()
	cEngine.checkToken("}")
	cEngine.jt.Advance() // else?
	elseFlag := false
	if cEngine.jt.CurrentToken() == "else" {
		elseFlag = true
		cEngine.vmw.WriteGoTo(if_continuation_label)
		cEngine.jt.Advance()
		cEngine.checkToken("{")
		cEngine.vmw.WriteLabel(if_false_label)
		cEngine.jt.Advance()
		cEngine.CompileStatements()
		cEngine.checkToken("}")
		cEngine.jt.Advance()
	} else {
		cEngine.vmw.WriteLabel(if_false_label)
	}
	if elseFlag {
		cEngine.vmw.WriteLabel(if_continuation_label)
	}
}

func (cEngine *CompilationEngine) CompileWhile() {

	cEngine.jt.Advance() // "("
	cEngine.checkToken("(")
	cEngine.jt.Advance()
	while_exp_label := "WHILE_EXP_" + cEngine.vmw.CreateLabel()
	while_end_label := "WHILE_END_" + cEngine.vmw.CreateLabel()
	cEngine.vmw.WriteLabel(while_exp_label)
	cEngine.CompileExpression()
	cEngine.checkToken(")")
	cEngine.jt.Advance() // "{"
	cEngine.checkToken("{")
	cEngine.vmw.WriteArithmetic(NOT)
	cEngine.vmw.WriteIf(while_end_label)
	cEngine.jt.Advance()
	cEngine.CompileStatements()
	cEngine.checkToken("}")
	cEngine.vmw.WriteGoTo(while_exp_label)
	cEngine.vmw.WriteLabel(while_end_label)
	cEngine.jt.Advance()
}

func (cEngine *CompilationEngine) CompileDo() {
	cEngine.jt.Advance() // subroutineName or className/varName
	symbolName := cEngine.jt.CurrentToken()
	name = symbolName
	symbolType := cEngine.subroutineSymbolTable.TypeOf(symbolName)
	symbolKind := cEngine.subroutineSymbolTable.KindOf(symbolName)
	if symbolKind == NONE { //try get it from class level
		symbolType = cEngine.classSymbolTable.TypeOf(symbolName)
	}
	if symbolType == "" {
		subroutineCallName = symbolName
	} else {
		subroutineCallName = symbolType
	}
	cEngine.checkTokenType("identifier")
	if _, ok := cEngine.classSymbolTable.symbolMap[cEngine.jt.CurrentToken()]; ok {
		// if we call surboutine on field/static var it also gets "this" as argument
		subroutineCallArgs++
	} else if symbolType != "" {
		subroutineCallArgs++
	}
	if cEngine.jt.CurrentToken() == cEngine.currentClass && cEngine.currentClass != "Main" {
		subroutineCallArgs++
	}
	cEngine.jt.Advance()
	if cEngine.jt.CurrentToken() == "(" {
		subroutineCallArgs++
	}
	cEngine.CompileSubroutineCall()
	cEngine.checkToken(";")
	cEngine.vmw.WritePop(TEMP, 0) // pop the return value
	cEngine.jt.Advance()
}

func (cEngine *CompilationEngine) CompileReturn() {
	cEngine.jt.Advance() // ; or experssion
	if cEngine.jt.CurrentToken() == ";" {
		cEngine.vmw.WritePush(CONSTANT, 0)
	} else {
		switch cEngine.jt.CurrentToken() {
		case "this":
			{
				cEngine.vmw.WritePush(POINTER, 0)
				cEngine.jt.Advance()
			}
		case "null", "false":
			{
				cEngine.vmw.WritePush(CONSTANT, 0)
				cEngine.jt.Advance()
			}
		case "true":
			{
				cEngine.vmw.WritePush(CONSTANT, 1)
				cEngine.vmw.WriteArithmetic(NEG)
				cEngine.jt.Advance()
			}
		default:
			{
				cEngine.CompileExpression()
			}
		}
		cEngine.checkToken(";")
	}
	cEngine.vmw.WriteReturn()
	cEngine.jt.Advance()
}

func (cEngine *CompilationEngine) CompileExpression() {
	cEngine.CompileTerm()
	for cEngine.isOp(cEngine.jt.CurrentToken()) {
		opCmd := cEngine.CompileOp()
		cEngine.CompileTerm()
		cEngine.vmw.outputFile.WriteString(opCmd + "\n")
	}
}

func (cEngine *CompilationEngine) CompileTerm() {
	switch cEngine.jt.CurrentToken() {
	case "this":
		{
			cEngine.vmw.WritePush(POINTER, 0)
			cEngine.jt.Advance()
			return
		}
	case "null", "false":
		{
			cEngine.vmw.WritePush(CONSTANT, 0)
			cEngine.jt.Advance()
			return
		}
	case "true":
		{
			cEngine.vmw.WritePush(CONSTANT, 0)
			cEngine.vmw.WriteArithmetic(NOT)
			cEngine.jt.Advance()
			return
		}
	}
	if cEngine.jt.CurrentToken() == "~" || cEngine.jt.CurrentToken() == "-" { //unaryOp term
		op := cEngine.jt.CurrentToken()
		cEngine.jt.Advance()
		cEngine.CompileTerm()
		if op == "~" {
			cEngine.vmw.WriteArithmetic(NOT)
		} else {
			cEngine.vmw.WriteArithmetic(NEG)
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
				cEngine.vmw.WritePush(CONSTANT, int(ch))
				cEngine.vmw.WriteCall("String.appendChar", 2)
			}
		}
		cEngine.jt.Advance()
	} else if varType := cEngine.jt.TokenType(); varType == "identifier" { //varName or subroutineCall
		varName := cEngine.jt.CurrentToken()
		name = varName
		cEngine.jt.Advance()
		if cEngine.jt.CurrentToken() == "(" || cEngine.jt.CurrentToken() == "." { // subroutineCall
			subroutineCallName += varName
			symbolType := cEngine.subroutineSymbolTable.TypeOf(varName)
			symbolKind := cEngine.subroutineSymbolTable.KindOf(varName)
			if symbolKind == NONE { //try get it from class level
				symbolType = cEngine.classSymbolTable.TypeOf(varName)
			}
			if symbolType != "" {
				subroutineCallArgs++
			}
			cEngine.CompileSubroutineCall()
		} else {
			symbolKind := cEngine.subroutineSymbolTable.KindOf(varName)
			symbolIndex := cEngine.subroutineSymbolTable.IndexOf(varName)
			if symbolKind == NONE { //try get it from class level
				symbolKind = cEngine.classSymbolTable.KindOf(varName)
				symbolIndex = cEngine.classSymbolTable.IndexOf(varName)
			}
			if cEngine.jt.CurrentToken() == "[" { // varName [experssion]
				cEngine.jt.Advance()
				cEngine.CompileExpression()
				cEngine.checkToken("]")
				cEngine.vmw.WritePush(cEngine.vmw.getSegmentOf(symbolKind), symbolIndex)
				cEngine.vmw.WriteArithmetic(ADD)
				cEngine.vmw.WritePop(POINTER, 1)
				cEngine.vmw.WritePush(THAT, 0)
				cEngine.jt.Advance()
			} else {
				cEngine.vmw.WritePush(cEngine.vmw.getSegmentOf(symbolKind), symbolIndex)
			}
		}
	} else if cEngine.jt.CurrentToken() == "(" { // (expression)
		cEngine.jt.Advance()
		cEngine.CompileExpression()
		cEngine.checkToken(")")
		cEngine.jt.Advance()
	}
}

func (cEngine *CompilationEngine) CompileExpressionList() int {
	sum := 0
	if cEngine.jt.CurrentToken() != ")" { // no more experssions
		sum++
		cEngine.CompileExpression()
		for cEngine.jt.CurrentToken() == "," {
			cEngine.jt.Advance()
			cEngine.CompileExpression()
			sum++
		}
	}
	return sum
}

func (cEngine *CompilationEngine) CompileSubroutineCall() {
	pushPointerFlag := false
	pushThisFlag := false
	if cEngine.jt.CurrentToken() == "." {
		cEngine.jt.Advance() // subrountineName
		subroutineCallName = strings.Title(subroutineCallName) + "." + cEngine.jt.CurrentToken()
		cEngine.jt.Advance()
	} else {
		if cEngine.currentClass != "Main" {
			pushPointerFlag = true
		}
		pushThisFlag = true
		subroutineCallName = strings.Title(cEngine.currentClass) + "." + subroutineCallName
	}
	varKind := cEngine.subroutineSymbolTable.KindOf(name)
	varIndex := cEngine.subroutineSymbolTable.IndexOf(name)
	if varKind == NONE { // check if the var is in the class scope
		varIndex = cEngine.classSymbolTable.IndexOf(name)
		varKind = cEngine.classSymbolTable.KindOf(name)
	}
	if varKind != NONE {
		cEngine.vmw.WritePush(cEngine.vmw.getSegmentOf(varKind), varIndex)
	}
	cEngine.checkToken("(")
	cEngine.jt.Advance()
	if pushPointerFlag {
		cEngine.vmw.WritePush(POINTER, 0)
	} else if pushThisFlag {
		fmt.Println(cEngine.currentClass)
		if cEngine.currentClass != "Main" {
			subroutineCallArgs++ // add "this" as argument
		}
		cEngine.vmw.WritePush(THIS, 0)
	}
	subroutineCallArgs += cEngine.CompileExpressionList()
	cEngine.checkToken(")")
	cEngine.vmw.WriteCall(subroutineCallName, subroutineCallArgs)
	cEngine.jt.Advance()
	name = ""
	subroutineCallName = ""
	subroutineCallArgs = 0
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
