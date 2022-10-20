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
	Plan(ctx context.Context, opts ...tfexec.PlanOption) error
	//Show(ctx context.Context, opts ...tfexec.ShowOption) (*tfjson.State, error)
	//ShowStateFile(ctx context.Context, statePath string, opts ...tfexec.ShowOption) (*tfjson.State, error)
	//ShowPlanFile(ctx context.Context, planPath string, opts ...tfexec.ShowOption) (*tfjson.Plan, error)
}
