package main

import (
	"fmt"
	"os"
)

func main() {
	jt := CreateTokenizer(os.Args[1])
	fmt.Println("test")
	for jt.HasMoreTokens() {
		tokenType := jt.TokenType()
		fmt.Println("<", tokenType, ">\t", jt.CurrentToken(), "\t</", tokenType, ">")
		jt.Advance()
	}
}
