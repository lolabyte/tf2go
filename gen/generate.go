package gen

import (
	"fmt"
	"os"
	"path"

	"github.com/dave/jennifer/jen"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/lolabyte/tf2go/terraform/ast"
	tfLexer "github.com/lolabyte/tf2go/terraform/lexer"
	tfParser "github.com/lolabyte/tf2go/terraform/parser"
	"github.com/lolabyte/tf2go/utils"
)

func GenerateTFModulePackage(tfModulePath string, goModuleOutDir string, packageName string) error {
	module, diags := tfconfig.LoadModule(tfModulePath)
	if diags.HasErrors() {
		return diags.Err()
	}

	vars := gatherVars(module)
	out := jen.NewFile(packageName)
	out.Type().Id("Variables").Struct(vars...).Line()

	// Generate module struct
	structName := utils.SnakeToCamel(packageName)
	out.Type().Id(structName).Struct(
		jen.Id("V").Id("Variables"),
		jen.Id("TF").Op("*").Qual("github.com/hashicorp/terraform-exec/tfexec", "Terraform"),
	)

	out.Func().Id(fmt.Sprintf("New%s", structName)).Params().Op("*").Id(structName).Block(
		jen.Return(jen.Op("&").Id(structName)).Values(),
	)

	// Generate Init()
	out.Func().Params(
		jen.Id("m").Op("*").Id(structName),
	).Id("Init").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("opts").Op("...").Qual("github.com/hashicorp/terraform-exec/tfexec", "InitOption"),
	).Error().Block(
		jen.Return(jen.Id("nil")),
	).Line()

	// Generate Apply()
	out.Func().Params(
		jen.Id("m").Op("*").Id(structName),
	).Id("Apply").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("opts").Op("...").Qual("github.com/hashicorp/terraform-exec/tfexec", "ApplyOption"),
	).Error().Block(
		jen.Return(jen.Id("nil")),
	).Line()

	// Generate Destroy()
	out.Func().Params(
		jen.Id("m").Op("*").Id(structName),
	).Id("Destroy").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("opts").Op("...").Qual("github.com/hashicorp/terraform-exec/tfexec", "DestroyOption"),
	).Error().Block(
		jen.Return(jen.Id("nil")),
	).Line()

	// Generate Plan()
	out.Func().Params(
		jen.Id("m").Op("*").Id(structName),
	).Id("Plan").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("opts").Op("...").Qual("github.com/hashicorp/terraform-exec/tfexec", "PlanOption"),
	).Error().Block(
		jen.Return(jen.Id("nil")),
	).Line()

	err := os.MkdirAll(goModuleOutDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output dir: %v", err)
	}

	err = out.Save(path.Join(goModuleOutDir, fmt.Sprintf("%s.go", packageName)))
	if err != nil {
		return fmt.Errorf("failed to save module to file: %v", err)
	}

	return nil
}

func eval(node ast.Node, stmt *jen.Statement) *jen.Statement {
	fmt.Println("call")
	switch node := node.(type) {
	case *ast.Type:
		for _, s := range node.Statements {
			return eval(s, stmt)
		}
	case *ast.BoolTypeLiteral:
		fmt.Println("bool")
		return stmt.Bool()
	case *ast.NumberTypeLiteral:
		fmt.Println("number")
		return stmt.Int64()
	case *ast.StringTypeLiteral:
		fmt.Println("string")
		return stmt.String()
	case *ast.ListTypeLiteral:
		fmt.Println("list()")
		stmt = stmt.Index()
		return eval(node.TypeExpression, stmt)
	case *ast.ObjectTypeLiteral:
		fmt.Println("object()")
		var fields []jen.Code
		for k, v := range node.ObjectSpec.(*ast.ObjectLiteral).KVPairs {
			f := jen.Id(utils.SnakeToCamel(k.String()))
			fields = append(fields, eval(v, f))
		}
		return jen.Struct(fields...)
	}

	return nil
}

func tfVarToStructField(stmt *jen.Statement, v *tfconfig.Variable) *jen.Statement {
	lexer := tfLexer.New(v.Type)
	parser := tfParser.New(lexer)

	T := parser.ParseType()
	return eval(T, stmt)

	//switch v.Type {
	//case "bool":
	//	return stmt.Bool()
	//case "number":
	//	return stmt.Int64()
	//case "string":
	//	return stmt.String()
	//case "list(bool)":
	//	return stmt.Index().Bool()
	//case "list(number)":
	//	return stmt.Index().Int64()
	//case "list(string)":
	//	return stmt.Index().String()
	//case "map(string)", "object":
	//	return stmt.Map(jen.String()).String()
	//default:
	//	//fmt.Println(v.Type)
	//}
	//return nil
}

func structTagsForVar(v *tfconfig.Variable) map[string]string {
	tags := map[string]string{
		"json": fmt.Sprintf("%s,omitempty", v.Name),
	}
	return tags
}

func structFieldNameForVar(v *tfconfig.Variable) string {
	return utils.SnakeToCamel(v.Name)
}

func gatherVars(mod *tfconfig.Module) []jen.Code {
	vars := make([]jen.Code, len(mod.Variables))
	for _, v := range mod.Variables {
		f := tfVarToStructField(jen.Id(structFieldNameForVar(v)), v)
		f = f.Tag(structTagsForVar(v))
		vars = append(vars, f)
	}
	return vars
}
