package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// Type constants
const KEYWORD = "keyword"
const SYMBOL = "symbol"
const IDENTIFIER = "identifier"
const INT_CONST = "integerConstant"
const STRING_CONST = "stringConstant"

/*// Keyword constants
const CLASS = 5
const METHOD = 6
const FUNCTION = 7
const CONSTRUCTOR = 8
const INT = 9
const BOOLEAN = 10
const CHAR = 11
const VOID = 12
const VAR = 13
const STATIC = 14
const FIELD = 15
const LET = 16
const DO = 17
const IF = 18
const ELSE = 19
const WHILE = 20
const RETURN = 21
const TRUE = 22
const FALSE = 23
const NULL = 24
const THIS = 25*/

var tokenMap map[string]string

func init() {
	tokenMap = make(map[string]string)
	tokenMap["class"] = KEYWORD
	tokenMap["method"] = KEYWORD
	tokenMap["function"] = KEYWORD
	tokenMap["constructor"] = KEYWORD
	tokenMap["int"] = KEYWORD
	tokenMap["boolean"] = KEYWORD
	tokenMap["char"] = KEYWORD
	tokenMap["void"] = KEYWORD
	tokenMap["var"] = KEYWORD
	tokenMap["static"] = KEYWORD
	tokenMap["field"] = KEYWORD
	tokenMap["let"] = KEYWORD
	tokenMap["do"] = KEYWORD
	tokenMap["if"] = KEYWORD
	tokenMap["else"] = KEYWORD
	tokenMap["while"] = KEYWORD
	tokenMap["return"] = KEYWORD
	tokenMap["true"] = KEYWORD
	tokenMap["false"] = KEYWORD
	tokenMap["null"] = KEYWORD
	tokenMap["this"] = KEYWORD
	tokenMap["{"] = SYMBOL
	tokenMap["}"] = SYMBOL
	tokenMap["("] = SYMBOL
	tokenMap[")"] = SYMBOL
	tokenMap["["] = SYMBOL
	tokenMap["]"] = SYMBOL
	tokenMap["."] = SYMBOL
	tokenMap[","] = SYMBOL
	tokenMap[";"] = SYMBOL
	tokenMap["+"] = SYMBOL
	tokenMap["-"] = SYMBOL
	tokenMap["*"] = SYMBOL
	tokenMap["/"] = SYMBOL
	tokenMap["&"] = SYMBOL
	tokenMap["|"] = SYMBOL
	tokenMap["<"] = SYMBOL
	tokenMap[">"] = SYMBOL
	tokenMap["="] = SYMBOL
	tokenMap["~"] = SYMBOL
}

type Token struct {
	Type       string
	KeyWord    string
	Symbol     byte
	Identifier string
	IntVal     int
	StringVal  string
}

type JackTokenizer struct {
	scanner           *bufio.Scanner
	currentTokenIndex int
	tokens            []Token
}

func CreateTokenizer(inputFile *os.File) *JackTokenizer {
	jt := &JackTokenizer{currentTokenIndex: 0, tokens: make([]Token, 0)}
	jt.setInputFile(inputFile)
	jt.generateTokens()
	return jt
}

func (jt *JackTokenizer) setInputFile(inputFile *os.File) {
	scanner := bufio.NewScanner(inputFile)
	jt.scanner = scanner
}

func (jt *JackTokenizer) generateTokens() {
	for jt.scanner.Scan() {
		line := strings.TrimSpace(jt.scanner.Text())
		if len(jt.scanner.Text()) != 0 && !strings.HasPrefix(line, "/") && !strings.HasPrefix(line, "*") { // skip lines that are empty or comments
			line = strings.Split(line, "//")[0] // Remove comments from the line
			line = strings.TrimSpace(line)      //Remove remaining white spaces
			i := 0
			for i < len(line) {
				c := line[i]
				if isSymbol(c) { //Symbol
					token := Token{Type: SYMBOL, Symbol: c}
					jt.tokens = append(jt.tokens, token)
					i++
					continue
				}

				if c == '"' { // StringConstant
					j := i + 1
					str := ""
					for line[j] != '"' {
						str += string(line[j])
						j++
					}
					j++
					token := Token{Type: STRING_CONST, StringVal: str}
					jt.tokens = append(jt.tokens, token)
					i = j
					continue
				}

				if unicode.IsDigit(rune(c)) { // IntegerConstant
					str := string(c)
					j := i + 1
					for unicode.IsDigit(rune(line[j])) { // build the whole integer
						str += string(line[j])
						j++
					}
					num, _ := strconv.Atoi(str)
					token := Token{Type: INT_CONST, IntVal: num}
					jt.tokens = append(jt.tokens, token)
					i = j
					continue
				}

				if unicode.IsLetter(rune(c)) {
					str := string(c)
					j := i + 1
					for j < len(line) { // build the whole word
						if unicode.IsLetter(rune(line[j])) {
							str += string(line[j])
							j++
							continue
						}
						break
					}
					i = j
					token := Token{}
					if isKeyWord(str) { //KeyWord
						token.Type = KEYWORD
						token.KeyWord = str
					} else { // Identifier
						token.Type = IDENTIFIER
						token.Identifier = str
					}
					jt.tokens = append(jt.tokens, token)
					continue
				}
				i++
			}
		}
	}
}

func (jt *JackTokenizer) HasMoreTokens() bool {
	return jt.currentTokenIndex < len(jt.tokens)
}

// Advances to the next token
func (jt *JackTokenizer) Advance() {
	jt.currentTokenIndex++
}

func isSymbol(c byte) bool {
	val, ok := tokenMap[string(c)]
	if !ok {
		return false
	}
	if val == SYMBOL {
		return true
	}
	return false
}

func isKeyWord(s string) bool {
	val, ok := tokenMap[s]
	if !ok {
		return false
	}
	if val == KEYWORD {
		return true
	}
	return false
}

func (jt *JackTokenizer) CurrentToken() string {
	token := jt.tokens[jt.currentTokenIndex]
	res := ""
	switch token.Type {
	case KEYWORD:
		{
			res = token.KeyWord
		}
	case SYMBOL:
		{
			res = string(token.Symbol)
		}
	case IDENTIFIER:
		{
			res = token.Identifier
		}
	case INT_CONST:
		{
			res = strconv.Itoa(token.IntVal)
		}
	case STRING_CONST:
		{
			res = token.StringVal
		}
	}
	return res
}

func (jt *JackTokenizer) TokenType() string {
	return jt.tokens[jt.currentTokenIndex].Type
}

func (jt *JackTokenizer) KeyWord() string {
	return jt.tokens[jt.currentTokenIndex].KeyWord
}

func (jt *JackTokenizer) Symbol() byte {
	return jt.tokens[jt.currentTokenIndex].Symbol
}

func (jt *JackTokenizer) Identifier() string {
	return jt.tokens[jt.currentTokenIndex].Identifier
}

func (jt *JackTokenizer) IntVal() int {
	return jt.tokens[jt.currentTokenIndex].IntVal
}

func (jt *JackTokenizer) StringVal() string {
	return jt.tokens[jt.currentTokenIndex].StringVal
}
