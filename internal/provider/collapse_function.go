// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ function.Function = &CollapseFunction{}

type CollapseFunction struct {
	defaultSeparator string
}

func NewCollapseFunction(defaultSeparator string) function.Function {
	return &CollapseFunction{defaultSeparator: defaultSeparator}
}

func (f *CollapseFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "collapse"
}

func (f *CollapseFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Collapse a nested map into a flat map with path-based keys",
		Description: "Recursively flattens a nested map or object into a single-level map. Keys are joined with the separator. An optional depth limits how many levels are collapsed.",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:        "input",
				Description: "The nested map or object to collapse.",
			},
		},
		VariadicParameter: &function.DynamicParameter{
			Name:           "options",
			Description:    "Optional arguments: a string sets the separator, a number sets the max depth. Both can be passed in any order. Defaults to provider separator and unlimited depth.",
			AllowNullValue: true,
		},
		Return: function.DynamicReturn{},
	}
}

func (f *CollapseFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input types.Dynamic
	var options []types.Dynamic

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input, &options))
	if resp.Error != nil {
		return
	}

	separator := f.defaultSeparator
	if separator == "" {
		separator = "/"
	}
	maxDepth := -1 // -1 = unlimited

	for _, opt := range options {
		if opt.IsNull() || opt.IsUnknown() {
			continue
		}
		switch v := opt.UnderlyingValue().(type) {
		case types.String:
			separator = v.ValueString()
		case types.Number:
			bf := v.ValueBigFloat()
			d, _ := bf.Int64()
			maxDepth = int(d)
		}
	}

	result := make(map[string]attr.Value)
	collapseValue("", input.UnderlyingValue(), separator, maxDepth, result)

	// Build output as types.Object (keys are dynamic path strings)
	attrTypes := make(map[string]attr.Type, len(result))
	for k, v := range result {
		attrTypes[k] = v.Type(ctx)
	}

	outputObj, diags := types.ObjectValue(attrTypes, result)
	if diags.HasError() {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, diags))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.DynamicValue(outputObj)))
}

// isAllContainers returns true if v is a map/object whose children are all maps or objects.
func isAllContainers(v attr.Value) bool {
	var children map[string]attr.Value
	switch c := v.(type) {
	case types.Object:
		children = c.Attributes()
	case types.Map:
		children = c.Elements()
	default:
		return false
	}
	if len(children) == 0 {
		return false
	}
	for _, child := range children {
		switch child.(type) {
		case types.Object, types.Map:
		default:
			return false
		}
	}
	return true
}

func collapseValue(prefix string, v attr.Value, separator string, remainingDepth int, result map[string]attr.Value) {
	if remainingDepth == 0 {
		result[prefix] = v
		return
	}

	// In auto mode (-1), only recurse if all children are maps/objects.
	// This stops naturally at the deepest level of "container" objects.
	if remainingDepth == -1 && !isAllContainers(v) {
		result[prefix] = v
		return
	}

	nextDepth := remainingDepth - 1
	if remainingDepth == -1 {
		nextDepth = -1
	}

	var children map[string]attr.Value
	switch c := v.(type) {
	case types.Object:
		children = c.Attributes()
	case types.Map:
		children = c.Elements()
	default:
		result[prefix] = v
		return
	}

	for k, child := range children {
		childKey := k
		if prefix != "" {
			childKey = prefix + separator + k
		}
		collapseValue(childKey, child, separator, nextDepth, result)
	}
}

// collapseKeyToPath splits a collapsed key back into path segments.
func collapseKeyToPath(key, separator string) []string {
	return strings.Split(key, separator)
}
