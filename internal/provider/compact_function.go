// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ function.Function = &CompactFunction{}

type CompactFunction struct{}

func NewCompactFunction() function.Function {
	return &CompactFunction{}
}

func (f *CompactFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "compact"
}

func (f *CompactFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Remove null and empty string values from a list or map",
		Description: "Returns the input list or map with all null values and empty strings removed.",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:        "input",
				Description: "The list or map to compact.",
			},
		},
		Return: function.DynamicReturn{},
	}
}

func (f *CompactFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input types.Dynamic

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input))
	if resp.Error != nil {
		return
	}

	var result attr.Value

	notNullOrBlank := func(v attr.Value) bool { return !v.IsNull() && !isEmptyString(v) }

	switch v := input.UnderlyingValue().(type) {
	case types.List:
		list, diags := types.ListValue(v.ElementType(ctx), filterListByPredicate(v.Elements(), notNullOrBlank))
		if diags.HasError() {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, diags))
			return
		}
		result = list

	case types.Map:
		m, diags := types.MapValue(v.ElementType(ctx), filterMapByPredicate(v.Elements(), notNullOrBlank))
		if diags.HasError() {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, diags))
			return
		}
		result = m

	case types.Object:
		filteredVals := filterMapByPredicate(v.Attributes(), notNullOrBlank)
		filteredTypes := make(map[string]attr.Type, len(filteredVals))
		for k, val := range filteredVals {
			filteredTypes[k] = val.Type(ctx)
		}
		obj, diags := types.ObjectValue(filteredTypes, filteredVals)
		if diags.HasError() {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, diags))
			return
		}
		result = obj

	default:
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError("input must be a list, map, or object"))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.DynamicValue(result)))
}
