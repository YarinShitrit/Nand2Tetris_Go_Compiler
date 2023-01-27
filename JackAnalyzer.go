package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No provided file or directory")
		os.Exit(1)
	}

	fileOrDir := os.Args[1]

	// This returns an *os.FileInfo type
	info, err := os.Stat(fileOrDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if info.IsDir() { // is directory

		files, err := os.ReadDir(fileOrDir)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// for each .jack file we generate its .xml output file
		for _, file := range files {
			fileStructure := strings.Split(file.Name(), ".")
			if fileStructure[1] == "jack" {
				input, _ := os.Open(fileOrDir + "/" + file.Name())
				output, _ := os.Create(fileOrDir + "/" + fileStructure[0] + "1.vm")
				cEngine := CreateCompilationEngine(input, output)
				cEngine.CompileClass()
			}
		}

	} else { // is file
		input, _ := os.Open(fileOrDir)
		output, _ := os.Create(strings.Split(fileOrDir, ".")[0] + "1.vm")
		cEngine := CreateCompilationEngine(input, output)
		cEngine.CompileClass()
	}
}
