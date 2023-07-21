package gen

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"

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

		j.Id("outfile").Op(":=").Qual("path", "Join").Call(j.Id("workingDir"), j.Lit("terraform.tfvars.json")),
		j.Err().Op("=").Qual("os", "WriteFile").Call(j.Id("outfile"), j.Id("b"), j.Qual("os", "ModePerm")),
		j.If(
			j.Err().Op("!=").Nil(),
		).Block(
			j.Return(j.Lit(""), j.Qual("fmt", "Errorf").Call(j.Lit("failed to write terraform.tfvars.json: %v"), j.Err())),
		).Line(),

		j.Return(j.Id("outfile"), j.Nil()),
	).Line()

	generateOutputStruct(out, module)

	out.Func().Params(
		j.Id("o").Id("Outputs"),
	).Id("WriteTFOutputJSON").Params(
		j.Id("workingDir").String(),
	).Parens(j.List(j.String(), j.Error())).Block(
		j.List(j.Id("b"), j.Err()).Op(":=").Qual("encoding/json", "Marshal").Call(j.Id("o")),
		j.If(
			j.Err().Op("!=").Nil(),
		).Block(
			j.Return(j.Lit(""), j.Err()),
		).Line(),

		j.Id("outfile").Op(":=").Qual("path", "Join").Call(j.Id("workingDir"), j.Lit("output.json")),
		j.Err().Op("=").Qual("os", "WriteFile").Call(j.Id("outfile"), j.Id("b"), j.Qual("os", "ModePerm")),
		j.If(
			j.Err().Op("!=").Nil(),
		).Block(
			j.Return(j.Lit(""), j.Qual("fmt", "Errorf").Call(j.Lit("failed to write output.json: %v"), j.Err())),
		).Line(),

		j.Return(j.Id("outfile"), j.Nil()),
	).Line()

	out.Func().Id("TFModuleEmbedFS").Params().Qual("embed", "FS").Block(
		j.Return(j.Id("tfModule")),
	)

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
	).Line()

	// Generate Init()
	out.Func().Params(
		j.Id("m").Op("*").Id(structName),
	).Id("Init").Params(
		j.Id("ctx").Qual("context", "Context"),
		j.Id("opts").Op("...").Qual("github.com/hashicorp/terraform-exec/tfexec", "InitOption"),
	).Error().Block(
		j.Err().Op(":=").Qual("github.com/lolabyte/tf2go/utils", "CopyDirFromEmbedFS").Call(j.Id("tfModule"), j.Lit(embedDir), j.Id("m").Dot("TF").Dot("WorkingDir").Call()),
		j.If(
			j.Err().Op("!=").Nil(),
		).Block(
			j.Panic(j.Err()),
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
	).Parens(j.List(j.Qual("github.com/lolabyte/tf2go/terraform", "TFOutput"), j.Error())).Block(
		j.List(j.Id("out"), j.Err()).Op(":=").Id("m").Dot("TF").Dot("Output").Call(j.Id("ctx"), j.Id("opts").Op("...")),
		j.If(j.Err().Op("!=").Nil()).Block(
			j.Return(j.Nil(), j.Err()),
		).Line(),

		j.Id("outJson").Op(":=").Make(j.Map(j.String()).Qual("encoding/json", "RawMessage")),
		j.For(j.List(j.Id("k"), j.Id("v"))).Op(":=").Range().Id("out").Block(
			j.Id("outJson").Index(j.Id("k")).Op("=").Id("v").Dot("Value"),
		).Line(),

		j.List(j.Id("b"), j.Err()).Op(":=").Qual("encoding/json", "Marshal").Call(j.Id("outJson")),
		j.If(j.Err().Op("!=").Nil()).Block(
			j.Return(j.Nil(), j.Err()),
		).Line(),

		j.Var().Id("outputs").Id("Outputs"),
		j.Err().Op("=").Qual("encoding/json", "Unmarshal").Call(j.Id("b"), j.Op("&").Id("outputs")),
		j.If(j.Err().Op("!=").Nil()).Block(
			j.Return(j.Nil(), j.Err()),
		).Line(),

		j.Return(j.Op("&").Id("outputs"), j.Nil()),
	).Line()

	// Generate Import()
	out.Func().Params(
		j.Id("m").Op("*").Id(structName),
	).Id("Import").Params(
		j.Id("ctx").Qual("context", "Context"),
		j.Id("address").String(),
		j.Id("id").String(),
		j.Id("opts").Op("...").Qual("github.com/hashicorp/terraform-exec/tfexec", "ImportOption"),
	).Parens(j.Error()).Block(
		j.Return(j.Id("m").Dot("TF").Dot("Import").Call(j.Id("ctx"), j.Id("address"), j.Id("id"), j.Id("opts").Op("..."))),
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

type kvpair struct {
	name  string
	value ast.Expression
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
		return eval(src, node.TypeExpression, stmt.Map(j.String()), name)
	case *ast.ObjectTypeLiteral:
		var fields []j.Code

		var kvpairs []kvpair
		for k, v := range node.ObjectSpec.(*ast.ObjectLiteral).KVPairs {
			kvpairs = append(kvpairs, kvpair{k.String(), v})
		}
		sort.Slice(kvpairs, func(i, j int) bool { return kvpairs[i].name < kvpairs[j].name })

		for _, kv := range kvpairs {
			structName := utils.SnakeToCamel(kv.name)
			tag := structTagsForField(kv.name)
			field := j.Id(structName)
			fields = append(fields, eval(src, kv.value, field, kv.name).Tag(tag))
		}

		structName := utils.SnakeToCamel(name)
		src.Type().Id(structName).Struct(fields...).Line()
		return stmt.Op("*").Id(structName)
	case *ast.OptionalTypeLiteral:
		switch node.TypeExpression.(type) {
		case *ast.ObjectLiteral:
			return eval(src, node.TypeExpression, stmt, name)
		case *ast.MapTypeLiteral:
			return eval(src, node.TypeExpression, stmt, name)
		case *ast.ListTypeLiteral:
			return eval(src, node.TypeExpression, stmt, name)
		default:
			return eval(src, node.TypeExpression, stmt.Op("*"), name)
		}
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

	// Sort alphabetically
	var variables []*tfconfig.Variable
	for _, v := range mod.Variables {
		variables = append(variables, v)
	}
	sort.Slice(variables, func(i, j int) bool { return variables[i].Name < variables[j].Name })

	for _, v := range variables {
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

	// Sort alphabetically
	var outputs []*tfconfig.Output
	for _, v := range mod.Outputs {
		outputs = append(outputs, v)
	}
	sort.Slice(outputs, func(i, j int) bool { return outputs[i].Name < outputs[j].Name })

	for _, v := range outputs {
		fieldName := structFieldNameForOutput(v)
		tag := structTagsForField(v.Name)
		field := j.Id(fieldName).Qual("encoding/json", "RawMessage").Tag(tag)
		outputStructFields = append(outputStructFields, field)
	}
	src.Type().Id("Outputs").Struct(outputStructFields...).Line()
}
