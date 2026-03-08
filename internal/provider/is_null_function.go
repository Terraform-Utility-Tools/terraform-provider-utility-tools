// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ function.Function = &IsNullFunction{}

type IsNullFunction struct{}

func NewIsNullFunction() function.Function {
	return &IsNullFunction{}
}

func (f *IsNullFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "isNull"
}

func (f *IsNullFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Check if a value is null",
		Description: "Returns true if the given value is null, false otherwise.",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:           "input",
				Description:    "The value to check for null.",
				AllowNullValue: true,
			},
		},
		Return: function.BoolReturn{},
	}
}

func (f *IsNullFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input types.Dynamic

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input))
	if resp.Error != nil {
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, input.IsNull()))
}
