// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- pick ---

var _ function.Function = &PickFunction{}

type PickFunction struct{}

func NewPickFunction() function.Function {
	return &PickFunction{}
}

func (f *PickFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "pick"
}

func (f *PickFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Keep only specified keys from a map or object",
		Description: "Returns a new map or object containing only the keys listed in 'keys'. Keys that do not exist in the input are silently ignored.",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:        "input",
				Description: "The map or object to select keys from.",
			},
			function.ListParameter{
				Name:        "keys",
				Description: "The keys to keep.",
				ElementType: types.StringType,
			},
		},
		Return: function.DynamicReturn{},
	}
}

func (f *PickFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input types.Dynamic
	var keys types.List

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input, &keys))
	if resp.Error != nil {
		return
	}

	keySet := listToStringSet(keys)
	result, funcErr := applyKeySelection(ctx, input, func(k string) bool {
		_, ok := keySet[k]
		return ok
	})
	if funcErr != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, funcErr)
		return
	}
	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.DynamicValue(result)))
}

// --- omit ---

var _ function.Function = &OmitFunction{}

type OmitFunction struct{}

func NewOmitFunction() function.Function {
	return &OmitFunction{}
}

func (f *OmitFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "omit"
}

func (f *OmitFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Remove specified keys from a map or object",
		Description: "Returns a new map or object with the keys listed in 'keys' removed. Keys that do not exist in the input are silently ignored.",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:        "input",
				Description: "The map or object to remove keys from.",
			},
			function.ListParameter{
				Name:        "keys",
				Description: "The keys to remove.",
				ElementType: types.StringType,
			},
		},
		Return: function.DynamicReturn{},
	}
}

func (f *OmitFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input types.Dynamic
	var keys types.List

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input, &keys))
	if resp.Error != nil {
		return
	}

	keySet := listToStringSet(keys)
	result, funcErr := applyKeySelection(ctx, input, func(k string) bool {
		_, ok := keySet[k]
		return !ok
	})
	if funcErr != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, funcErr)
		return
	}
	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.DynamicValue(result)))
}

// --- shared helpers ---

func listToStringSet(list types.List) map[string]struct{} {
	set := make(map[string]struct{}, len(list.Elements()))
	for _, elem := range list.Elements() {
		s, ok := elem.(types.String)
		if ok && !s.IsNull() && !s.IsUnknown() {
			set[s.ValueString()] = struct{}{}
		}
	}
	return set
}

func applyKeySelection(ctx context.Context, input types.Dynamic, keep func(string) bool) (attr.Value, *function.FuncError) {
	switch v := input.UnderlyingValue().(type) {
	case types.Map:
		filtered := make(map[string]attr.Value)
		for k, val := range v.Elements() {
			if keep(k) {
				filtered[k] = val
			}
		}
		m, diags := types.MapValue(v.ElementType(ctx), filtered)
		if diags.HasError() {
			return nil, function.FuncErrorFromDiags(ctx, diags)
		}
		return m, nil

	case types.Object:
		filtered := make(map[string]attr.Value)
		for k, val := range v.Attributes() {
			if keep(k) {
				filtered[k] = val
			}
		}
		filteredTypes := make(map[string]attr.Type, len(filtered))
		for k, val := range filtered {
			filteredTypes[k] = val.Type(ctx)
		}
		obj, diags := types.ObjectValue(filteredTypes, filtered)
		if diags.HasError() {
			return nil, function.FuncErrorFromDiags(ctx, diags)
		}
		return obj, nil

	default:
		return nil, function.NewFuncError("input must be a map or object")
	}
}
