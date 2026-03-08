// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ function.Function = &NestedCompactFunction{}

type NestedCompactFunction struct {
	defaultSeparator string
}

func NewNestedCompactFunction(defaultSeparator string) function.Function {
	return &NestedCompactFunction{defaultSeparator: defaultSeparator}
}

func (f *NestedCompactFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "nestedCompact"
}

func (f *NestedCompactFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Remove null and empty string leaf entries from a nested map",
		Description: "Collapses a nested map to its leaf entries, removes null values and empty strings, then expands back to the original structure.",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:        "input",
				Description: "The nested map or object to process.",
			},
		},
		VariadicParameter: &function.StringParameter{
			Name:        "separator",
			Description: "Separator used when collapsing and expanding path keys. Defaults to provider separator.",
		},
		Return: function.DynamicReturn{},
	}
}

func (f *NestedCompactFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input types.Dynamic
	var separators []string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input, &separators))
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

	// Step 1: collapse (auto depth — stops at leaf objects)
	flatAttrs := make(map[string]attr.Value)
	collapseValue(ctx, "", input.UnderlyingValue(), separator, -1, flatAttrs)

	// Step 2: remove null and empty string entries
	filtered := filterMapByPredicate(flatAttrs, func(v attr.Value) bool { return !v.IsNull() && !isEmptyString(v) })

	// Step 3: expand back to nested structure
	outputVal, funcErr := expandAttrMap(ctx, filtered, separator)
	if funcErr != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, funcErr)
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.DynamicValue(outputVal)))
}
