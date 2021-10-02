package main

import (
	"flag"

	"github.com/reddec/go-congen"
)

func main() {
	inputFile := flag.String("i", "index.html", "input file")
	outputFile := flag.String("o", "stub.go", "output file")
	packageName := flag.String("p", "controller", "package name")
	flag.Parse()

	if err := congen.ProcessFile(*inputFile, *outputFile, *packageName); err != nil {
		panic(err)
	}
}
