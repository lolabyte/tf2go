package main

import (
	"os"

	"github.com/dave/jennifer/jen"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/lolabyte/tf2go/gen"
)

var (
	inputModulePath   string
	outputPackageName string
	outputDir         string
)

func tfVarToStructField(stmt *jen.Statement, v *tfconfig.Variable) *jen.Statement {
	switch v.Type {
	case "bool":
		return stmt.Bool()
	case "number":
		return stmt.Int64()
	case "string":
		return stmt.String()
	case "list(bool)":
		return stmt.List(jen.Bool())
	case "list(number)":
		return stmt.List(jen.Int64())
	case "list(string)":
		return stmt.List(jen.String())
	case "map(string)", "object":
		return stmt.Map(jen.String()).String()
	}
	return nil
}

func main() {
	err := gen.GenerateTFModulePackage(os.Args[1], os.Args[2], os.Args[3])
	if err != nil {
		panic(err)
	}
}
