package gen

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	j "github.com/dave/jennifer/jen"
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
	out := j.NewFile(packageName)

	out.Commentf("//go:embed %s", path.Join(tfModuleEmbedDir, "*"))
	out.Var().Id("tfModule").Qual("embed", "FS")

	out.Type().Id("Variables").Struct(vars...).Line()

	// Generate module struct
	structName := utils.SnakeToCamel(packageName)
	out.Type().Id(structName).Struct(
		j.Id("V").Id("Variables"),
		j.Id("TF").Op("*").Qual("github.com/hashicorp/terraform-exec/tfexec", "Terraform"),
	)

	// Generate constructor
	out.Func().Id(fmt.Sprintf("New%s", structName)).Params(
		j.Id("workingDir").String(),
	).Op("*").Id(structName).Block(
		j.List(j.Id("execPath"), j.Err()).Op(":=").Qual("os/exec", "LookPath").Call(j.Lit("terraform")),
		j.If(
			j.Err().Op("==").Nil(),
		).Block(
			j.List(j.Id("execPath"), j.Err()).Op("=").Qual("path/filepath", "Abs").Call(j.Id("execPath")),
			j.If(
				j.Err().Op("!=").Nil(),
			).Block(
				j.Panic(j.Err()),
			),
		).Line(),

		j.Err().Op("=").Qual("os", "Mkdir").Call(j.Id("workingDir"), j.Qual("os", "ModePerm")),
		j.If(
			j.Err().Op("!=").Nil().Op("&&").Op("!os").Dot("IsExist").Call(j.Err()),
		).Block(
			j.Panic(j.Err()),
		).Line(),

		j.List(j.Id("tf"), j.Err()).Op(":=").Qual("github.com/hashicorp/terraform-exec/tfexec", "NewTerraform").Call(j.Id("workingDir"), j.Id("execPath")),
		j.If(
			j.Err().Op("!=").Nil(),
		).Block(
			j.Panic(j.Err()),
		).Line(),

		j.Id("tf").Dot("SetLogger").Call(j.Qual("log", "Default").Call()),

		j.Return(
			j.Op("&").Id(structName).Values(
				j.Dict{
					j.Id("TF"): j.Id("tf"),
				},
			),
		),
	).Line()

	// Generate Init()
	out.Func().Params(
		j.Id("m").Op("*").Id(structName),
	).Id("Init").Params(
		j.Id("ctx").Qual("context", "Context"),
		j.Id("opts").Op("...").Qual("github.com/hashicorp/terraform-exec/tfexec", "InitOption"),
	).Error().Block(
		j.List(j.Id("entries"), j.Err()).Op(":=").Id("tfModule").Dot("ReadDir").Call(j.Lit("terraform")),
		j.If(
			j.Err().Op("!=").Nil(),
		).Block(
			j.Panic(j.Err()),
		).Line(),

		j.For(j.List(j.Id("_"), j.Id("e"))).Op(":=").Range().Id("entries").Block(
			j.Id("fpath").Op(":=").Qual("path", "Join").Call(j.Lit("terraform"), j.Id("e").Dot("Name").Call()),
			j.List(j.Id("in"), j.Err()).Op(":=").Id("tfModule").Dot("ReadFile").Call(j.Id("fpath")),
			j.If(
				j.Err().Op("!=").Nil(),
			).Block(
				j.Panic(j.Err()),
			).Line(),

			j.Err().Op("=").Qual("os", "WriteFile").Call(
				j.Id("path").Dot("Join").Call(j.Id("m").Dot("TF").Dot("WorkingDir").Call(), j.Id("e").Dot("Name").Call()),
				j.Id("in"),
				j.Qual("os", "ModePerm"),
			),
			j.If(
				j.Err().Op("!=").Nil(),
			).Block(
				j.Panic(j.Err()),
			),
		).Line(),

		j.Return(j.Id("m").Dot("TF").Dot("Init").Call(j.Id("ctx"), j.Id("opts").Op("..."))),
	).Line()

	// Generate Apply()
	out.Func().Params(
		j.Id("m").Op("*").Id(structName),
	).Id("Apply").Params(
		j.Id("ctx").Qual("context", "Context"),
		j.Id("opts").Op("...").Qual("github.com/hashicorp/terraform-exec/tfexec", "ApplyOption"),
	).Error().Block(
		j.Return(j.Id("m").Dot("TF").Dot("Apply").Call(j.Id("ctx"), j.Id("opts").Op("..."))),
	).Line()

	// Generate Destroy()
	out.Func().Params(
		j.Id("m").Op("*").Id(structName),
	).Id("Destroy").Params(
		j.Id("ctx").Qual("context", "Context"),
		j.Id("opts").Op("...").Qual("github.com/hashicorp/terraform-exec/tfexec", "DestroyOption"),
	).Error().Block(
		j.Return(j.Id("m").Dot("TF").Dot("Destroy").Call(j.Id("ctx"), j.Id("opts").Op("..."))),
	).Line()

	// Generate Plan()
	out.Func().Params(
		j.Id("m").Op("*").Id(structName),
	).Id("Plan").Params(
		j.Id("ctx").Qual("context", "Context"),
		j.Id("opts").Op("...").Qual("github.com/hashicorp/terraform-exec/tfexec", "PlanOption"),
	).Parens(j.List(j.Bool(), j.Error())).Block(
		j.Return(j.Id("m").Dot("TF").Dot("Plan").Call(j.Id("ctx"), j.Id("opts").Op("..."))),
	).Line()

	err := os.MkdirAll(outPackageDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output dir: %v", err)
	}

	err = out.Save(path.Join(outPackageDir, fmt.Sprintf("%s.go", packageName)))
	if err != nil {
		return fmt.Errorf("failed to save module to file: %v", err)
	}

	// Copy the Terraform module into the go:embed path
	copyDirectory(tfModulePath, path.Join(outPackageDir, tfModuleEmbedDir))

	return nil
}

func copyDirectory(srcDir string, destDir string) {
	src, err := filepath.Abs(srcDir)
	if err != nil {
		panic(err)
	}

	dest, err := filepath.Abs(destDir)
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("cp", "-a", src, dest)
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}

func eval(node ast.Node, stmt *j.Statement, tfv *tfconfig.Variable) *j.Statement {
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
		m := stmt.Map(j.String())
		return eval(node.TypeExpression, m, tfv)
	case *ast.ObjectTypeLiteral:
		//fmt.Println("{}")
		var fields []j.Code
		for k, v := range node.ObjectSpec.(*ast.ObjectLiteral).KVPairs {
			f := j.Id(utils.SnakeToCamel(k.String()))
			fields = append(fields, eval(v, f, tfv).Tag(structTagsForField(k.String())))
		}
		return stmt.Struct(fields...)
	}
	return nil
}

func tfVarToStructField(stmt *j.Statement, v *tfconfig.Variable) *j.Statement {
	lexer := tfLexer.New(v.Type)
	parser := tfParser.New(lexer)

	T := parser.ParseType()
	return eval(T, stmt, v)
}

func structTagsForField(name string) map[string]string {
	tags := map[string]string{
		"json": fmt.Sprintf("%s,omitempty", name),
	}
	return tags
}

func structFieldNameForVar(v *tfconfig.Variable) string {
	return utils.SnakeToCamel(v.Name)
}

func gatherVars(mod *tfconfig.Module) []j.Code {
	vars := make([]j.Code, len(mod.Variables))
	for _, v := range mod.Variables {
		// Default any non-declared types to string
		if v.Type == "" {
			v.Type = "string"
		}
		f := tfVarToStructField(j.Id(structFieldNameForVar(v)), v)
		f = f.Tag(structTagsForField(v.Name))
		if v.Description != "" {
			f = f.Comment(v.Description)
		}
		vars = append(vars, f)
	}

	return vars
}
