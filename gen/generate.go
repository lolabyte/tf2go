package gen

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/dave/jennifer/jen"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/lolabyte/tf2go/terraform/ast"
	tfLexer "github.com/lolabyte/tf2go/terraform/lexer"
	tfParser "github.com/lolabyte/tf2go/terraform/parser"
	"github.com/lolabyte/tf2go/utils"
)

const tfModuleEmbedDir = "terraform"

func GenerateTFModulePackage(tfModulePath string, outPackageDir string, packageName string) error {
	module, diags := tfconfig.LoadModule(tfModulePath)
	if diags.HasErrors() {
		return diags.Err()
	}

	vars := gatherVars(module)
	out := jen.NewFile(packageName)

	out.Commentf("//go:embed %s", path.Join(tfModuleEmbedDir, "*"))
	out.Var().Id("tfModule").Qual("embed", "FS")

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

	err := os.MkdirAll(outPackageDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output dir: %v", err)
	}

	err = out.Save(path.Join(outPackageDir, fmt.Sprintf("%s.go", packageName)))
	if err != nil {
		return fmt.Errorf("failed to save module to file: %v", err)
	}

	copyDirectory(tfModulePath, path.Join(outPackageDir, tfModuleEmbedDir))

	return nil
}

func copyDirectory(srcDir string, destDir string) {
	cmd := exec.Command("cp", "-a", srcDir, destDir)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func eval(node ast.Node, stmt *jen.Statement, tfv *tfconfig.Variable) *jen.Statement {
	switch node := node.(type) {
	case *ast.Type:
		//fmt.Println("type")
		for _, s := range node.Statements {
			return eval(s.(*ast.ExpressionStatement).Expression, stmt, tfv)
		}
	case *ast.BoolTypeLiteral:
		//fmt.Println("bool")
		return stmt.Bool()
	case *ast.NumberTypeLiteral:
		//fmt.Println("number")
		return stmt.Int64()
	case *ast.StringTypeLiteral:
		//fmt.Println("string")
		return stmt.String()
	case *ast.ListTypeLiteral:
		//fmt.Println("list()")
		stmt = stmt.Index()
		return eval(node.TypeExpression, stmt, tfv)
	case *ast.MapTypeLiteral:
		//fmt.Println("map()")
		m := stmt.Map(jen.String())
		return eval(node.TypeExpression, m, tfv)
	case *ast.ObjectTypeLiteral:
		//fmt.Println("{}")
		var fields []jen.Code
		for k, v := range node.ObjectSpec.(*ast.ObjectLiteral).KVPairs {
			f := jen.Id(utils.SnakeToCamel(k.String()))
			fields = append(fields, eval(v, f, tfv))
		}
		return stmt.Struct(fields...)
	}
	return nil
}

func tfVarToStructField(stmt *jen.Statement, v *tfconfig.Variable) *jen.Statement {
	lexer := tfLexer.New(v.Type)
	parser := tfParser.New(lexer)

	T := parser.ParseType()
	return eval(T, stmt, v)
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
		// Default any non-declared types to string
		if v.Type == "" {
			v.Type = "string"
		}
		f := tfVarToStructField(jen.Id(structFieldNameForVar(v)), v)
		f = f.Tag(structTagsForVar(v))
		if v.Description != "" {
			f = f.Comment(v.Description)
		}
		vars = append(vars, f)
	}

	return vars
}
