package gen

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	j "github.com/dave/jennifer/jen"
	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/lolabyte/tf2go/terraform/ast"
	tfLexer "github.com/lolabyte/tf2go/terraform/lexer"
	tfParser "github.com/lolabyte/tf2go/terraform/parser"
	"github.com/lolabyte/tf2go/utils"
	cp "github.com/otiai10/copy"
)

func GenerateTFModulePackage(inputModulePath string, outPackageDir string, packageName string, embedDir string) error {
	dir, err := os.MkdirTemp("", packageName)
	if err != nil {
		return err
	}

	moduleDir, err := getInputModule(inputModulePath, dir)
	if err != nil {
		return fmt.Errorf("unable to get module from path %s: %v", dir, err)
	}
	defer os.RemoveAll(dir)

	module, diags := tfconfig.LoadModule(moduleDir)
	if diags.HasErrors() {
		return diags.Err()
	}

	out := j.NewFile(packageName)

	out.Commentf("//go:embed %s", path.Join(embedDir, "*"))
	out.Var().Id("tfModule").Qual("embed", "FS")

	generateVarStructs(out, module)
	generateOutputStruct(out, module)

	out.Func().Params(
		j.Id("v").Id("Variables"),
	).Id("WriteTFVarJSON").Params(
		j.Id("workingDir").String(),
	).Parens(j.List(j.String(), j.Error())).Block(
		j.List(j.Id("b"), j.Err()).Op(":=").Qual("encoding/json", "Marshal").Call(j.Id("v")),
		j.If(
			j.Err().Op("!=").Nil(),
		).Block(
			j.Return(j.Lit(""), j.Err()),
		).Line(),

		j.Id("outfile").Op(":=").Qual("path", "Join").Call(j.Id("workingDir"), j.Lit("terraform.tfvar.json")),
		j.Err().Op("=").Qual("os", "WriteFile").Call(j.Id("outfile"), j.Id("b"), j.Qual("os", "ModePerm")),
		j.If(
			j.Err().Op("!=").Nil(),
		).Block(
			j.Return(j.Lit(""), j.Qual("fmt", "Errorf").Call(j.Lit("failed to write terraform.tfvar.json: %v"), j.Err())),
		).Line(),

		j.Return(j.Id("outfile"), j.Nil()),
	).Line()

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

	// Generate Vars()
	out.Func().Params(
		j.Id("m").Op("*").Id(structName),
	).Id("Vars").Params().Qual("github.com/lolabyte/tf2go/terraform", "TFVars").Block(
		j.Return(j.Id("m").Dot("V")),
	)

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

	// Generate Output()
	out.Func().Params(
		j.Id("m").Op("*").Id(structName),
	).Id("Output").Params(
		j.Id("ctx").Qual("context", "Context"),
		j.Id("opts").Op("...").Qual("github.com/hashicorp/terraform-exec/tfexec", "OutputOption"),
	).Parens(j.List(j.Map(j.String()).Qual("github.com/hashicorp/terraform-exec/tfexec", "OutputMeta"), j.Error())).Block(
		j.Return(j.Id("m").Dot("TF").Dot("Output").Call(j.Id("ctx"), j.Id("opts").Op("..."))),
	).Line()

	// Copy the Terraform module to the go:embed path
	err = os.MkdirAll(path.Join(outPackageDir, embedDir), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output dir: %v", err)
	}

	err = out.Save(path.Join(outPackageDir, fmt.Sprintf("%s.go", packageName)))
	if err != nil {
		return fmt.Errorf("failed to save module to file: %v", err)
	}

	copyDirectory(moduleDir, path.Join(outPackageDir, embedDir))

	return nil
}

func getInputModule(src, dst string) (string, error) {
	_, err := os.Stat(src)
	if !os.IsNotExist(err) {
		return src, nil
	}

	client := getter.Client{
		Src:  src,
		Dst:  dst,
		Pwd:  dst,
		Mode: getter.ClientModeDir,
	}
	return dst, client.Get()
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

	opts := cp.Options{
		Skip: func(f os.FileInfo, src, dest string) (bool, error) {
			if f.IsDir() && f.Name() == ".terraform" {
				return true, nil
			}
			ok, err := path.Match("*/.teraform/*", src)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
			return false, nil
		},
	}
	err = cp.Copy(src, dest, opts)
	if err != nil {
		panic(err)
	}
}

func eval(src *j.File, node ast.Node, stmt *j.Statement, name string) *j.Statement {
	switch node := node.(type) {
	case *ast.Type:
		for _, s := range node.Statements {
			return eval(src, s.(*ast.ExpressionStatement).Expression, stmt, name)
		}
	case *ast.AnyTypeLiteral:
		return stmt.Interface()
	case *ast.BoolTypeLiteral:
		return stmt.Op("*").Bool()
	case *ast.NumberTypeLiteral:
		return stmt.Int64()
	case *ast.StringTypeLiteral:
		return stmt.String()
	case *ast.ListTypeLiteral:
		return eval(src, node.TypeExpression, stmt.Index(), name)
	case *ast.MapTypeLiteral:
		return eval(src, node.TypeExpression, stmt.Op("*").Map(j.String()), name)
	case *ast.ObjectTypeLiteral:
		var fields []j.Code

		for k, v := range node.ObjectSpec.(*ast.ObjectLiteral).KVPairs {
			structName := utils.SnakeToCamel(k.String())
			tag := structTagsForField(k.String())
			field := j.Id(structName)
			fields = append(fields, eval(src, v, field, k.String()).Tag(tag))
		}

		structName := utils.SnakeToCamel(name)
		src.Type().Id(structName).Struct(fields...).Line()
		return stmt.Op("*").Id(structName)
	case *ast.OptionalTypeLiteral:
		return eval(src, node.TypeExpression, stmt.Op("*"), name)
	}
	return nil
}

func astNodeType(v *tfconfig.Variable) ast.Node {
	lexer := tfLexer.New(v.Type)
	parser := tfParser.New(lexer)

	return parser.ParseType()
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

func structFieldNameForOutput(v *tfconfig.Output) string {
	return utils.SnakeToCamel(v.Name)
}

func generateVarStructs(src *j.File, mod *tfconfig.Module) {
	var defaultVarStructFields []j.Code
	for _, v := range mod.Variables {
		if v.Type == "" {
			v.Type = "string"
		}

		fieldName := structFieldNameForVar(v)
		tag := structTagsForField(v.Name)
		field := eval(src, astNodeType(v), j.Id(fieldName), v.Name).Tag(tag)
		if v.Description != "" {
			field = field.Comment(v.Description)
		}

		defaultVarStructFields = append(defaultVarStructFields, field)
	}

	src.Type().Id("Variables").Struct(defaultVarStructFields...).Line()
}

func generateOutputStruct(src *j.File, mod *tfconfig.Module) {
	var outputStructFields []j.Code
	for _, v := range mod.Outputs {
		fieldName := structFieldNameForOutput(v)
		tag := structTagsForField(v.Name)
		field := j.Id(fieldName).Op("json.RawMessage").Tag(tag)
		outputStructFields = append(outputStructFields, field)
	}
	src.Type().Id("Outputs").Struct(outputStructFields...).Line()
}
