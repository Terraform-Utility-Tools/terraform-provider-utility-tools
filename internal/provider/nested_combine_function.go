// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ function.Function = &NestedCombineFunction{}

type NestedCombineFunction struct {
	defaultSeparator string
}

func NewNestedCombineFunction(defaultSeparator string) function.Function {
	return &NestedCombineFunction{defaultSeparator: defaultSeparator}
}

func (f *NestedCombineFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "nestedCombine"
}

func (f *NestedCombineFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Combine dimensions into a nested map",
		Description: "Produces the cartesian product of the input dimensions and returns the result as a nested map instead of a flat path-keyed map.",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:        "input",
				Description: "Ordered list of {key, values} items to combine. Each item is an object with a 'key' string and a 'values' list or map.",
			},
		},
		VariadicParameter: &function.StringParameter{
			Name:        "separator",
			Description: "Separator used to split path keys when expanding. Defaults to provider separator.",
		},
		Return: function.DynamicReturn{},
	}
}

func (f *NestedCombineFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var inputDyn types.Dynamic
	var separators []string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &inputDyn, &separators))
	if resp.Error != nil {
		return
	}

	separator := f.defaultSeparator
	if separator == "" {
		separator = "/"
	}
	if len(separators) > 0 && separators[0] != "" {
		separator = separators[0]
	}

	// Step 1: combine → flat path-keyed map
	flatMap, funcErr := combineToAttrMap(ctx, inputDyn, separator)
	if funcErr != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, funcErr)
		return
	}

	// Step 2: expand → nested structure
	outputVal, funcErr := expandAttrMap(ctx, flatMap, separator)
	if funcErr != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, funcErr)
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.DynamicValue(outputVal)))
}
