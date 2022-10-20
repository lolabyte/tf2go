package main

import (
	"flag"

	"github.com/lolabyte/tf2go/gen"
)

var (
	inputModulePath   string
	outputPackageName string
	outputDir         string
)

func init() {
	flag.StringVar(&inputModulePath, "module", "", "path to a TF module")
	flag.StringVar(&outputPackageName, "package", "", "name of the package to generate")
	flag.StringVar(&outputDir, "out", "", "path to output directory (will create if not exists)")
	flag.Parse()
}

func main() {
	err := gen.GenerateTFModulePackage(inputModulePath, outputDir, outputPackageName)
	if err != nil {
		panic(err)
	}
}
