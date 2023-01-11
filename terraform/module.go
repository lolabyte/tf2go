package terraform

import (
	"context"

	"github.com/hashicorp/terraform-exec/tfexec"
)

// Module is the interface for defining a Terraform module as a go package
type Module interface {
	Init(ctx context.Context, opts ...tfexec.InitOption) error
	Apply(ctx context.Context, opts ...tfexec.ApplyOption) error
	Destroy(ctx context.Context, opts ...tfexec.DestroyOption) error
	Plan(ctx context.Context, opts ...tfexec.PlanOption) (bool, error)
	Output(ctx context.Context, opts ...tfexec.OutOption) (map[string]tfexec.OutputMeta, error)
	Vars() TFVars
}

type TFVars interface {
	WriteTFVarJSON(workingDir string) (string, error)
}
