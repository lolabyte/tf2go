package gen

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func GenerateTFModulePackage(tfModulePath string, goModuleOutDir string, packageName string) error {
	module, diags := tfconfig.LoadModule(tfModulePath)
	if diags.HasErrors() {
		return diags.Err()
	}

	vars := gatherVars(module)
	out := jen.NewFile(packageName)
	out.Type().Id("Variables").Struct(vars...)

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

func structTagsForVar(v *tfconfig.Variable) map[string]string {
	tags := map[string]string{
		"json": v.Name,
	}
	return tags
}

func structFieldNameForVar(v *tfconfig.Variable) string {
	var buf []byte
	strBytes := []byte(v.Name)
	for i, c := range strBytes {
		switch {
		case 'a' <= c && c <= 'z':
			if i == 0 || strBytes[i-1] == '-' || strBytes[i-1] == '_' {
				buf = append(buf, []byte(strings.ToUpper(string(c)))...)
			} else {
				buf = append(buf, c)
			}
		case c == '-' || c == '_':
			continue
		}
	}
	return string(buf)
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
