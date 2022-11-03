package test_package

import (
	"context"
	tfexec "github.com/hashicorp/terraform-exec/tfexec"
)

type Variables struct {
	Number                  int64    `json:"number,omitempty"`
	ListOfBoolWithDefault   []bool   `json:"list_of_bool_with_default,omitempty"`
	ListOfNumber            []int64  `json:"list_of_number,omitempty"`
	ListOfString            []string `json:"list_of_string,omitempty"`
	Bool                    bool     `json:"bool,omitempty"`
	BoolWithDefault         bool     `json:"bool_with_default,omitempty"`
	String                  string   `json:"string,omitempty"`
	StringWithDefault       string   `json:"string_with_default,omitempty"`
	SensitiveString         string   `json:"sensitive_string,omitempty"`
	NumberWithDefault       int64    `json:"number_with_default,omitempty"`
	ListOfBool              []bool   `json:"list_of_bool,omitempty"`
	ListOfStringWithDefault []string `json:"list_of_string_with_default,omitempty"`
	ListOfNumberWithDefault []int64  `json:"list_of_number_with_default,omitempty"`
}

type TestPackage struct {
	V  Variables
	TF *tfexec.Terraform
}

func NewTestPackage() *TestPackage {
	return &TestPackage{}
}
func (m *TestPackage) Init(ctx context.Context, opts ...tfexec.InitOption) error {
	return nil
}

func (m *TestPackage) Apply(ctx context.Context, opts ...tfexec.ApplyOption) error {
	return nil
}

func (m *TestPackage) Destroy(ctx context.Context, opts ...tfexec.DestroyOption) error {
	return nil
}

func (m *TestPackage) Plan(ctx context.Context, opts ...tfexec.PlanOption) error {
	return nil
}
