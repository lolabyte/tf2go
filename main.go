package main

import (
	"os"

	"github.com/lolabyte/tf2go/gen"
)

var (
	inputModulePath   string
	outputPackageName string
	outputDir         string
)

func main() {
	err := gen.GenerateTFModulePackage(os.Args[1], os.Args[2], os.Args[3])
	if err != nil {
		panic(err)
	}
}
