// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ function.Function = &NestedFilterFunction{}

type NestedFilterFunction struct {
	defaultSeparator string
}

func NewNestedFilterFunction(defaultSeparator string) function.Function {
	return &NestedFilterFunction{defaultSeparator: defaultSeparator}
}

func (f *NestedFilterFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "nestedFilter"
}

func (f *NestedFilterFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Filter a nested map by leaf object attributes",
		Description: "Collapses a nested map, filters the leaves by the given condition(s), then expands back to the original structure. Multiple condition objects are OR'd together; attributes within one object are AND'd. At least one condition is required.",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:        "input",
				Description: "The nested map or object to filter.",
			},
		},
		VariadicParameter: &function.DynamicParameter{
			Name:           "condition",
			Description:    "One or more condition objects. A leaf is kept if it matches any one condition object (OR). Within each object all key=value pairs must match (AND).",
			AllowNullValue: true,
		},
		Return: function.DynamicReturn{},
	}
}

func (f *NestedFilterFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input types.Dynamic
	var conditionSlice []types.Dynamic

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input, &conditionSlice))
	if resp.Error != nil {
		return
	}

	separator := f.defaultSeparator
	if separator == "" {
		separator = "/"
	}

	conditions, funcErr := parseConditions(conditionSlice)
	if funcErr != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, funcErr)
		return
	}

	// Step 1: collapse (auto depth — stops at leaf objects)
	flatAttrs := make(map[string]attr.Value)
	collapseValue("", input.UnderlyingValue(), separator, -1, flatAttrs)

	// Step 2: filter
	filtered := filterAttrMap(flatAttrs, conditions)

	// Step 3: expand back to nested structure
	outputVal, funcErr := expandAttrMap(ctx, filtered, separator)
	if funcErr != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, funcErr)
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.DynamicValue(outputVal)))
}
