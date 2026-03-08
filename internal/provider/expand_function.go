// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ function.Function = &ExpandFunction{}

type ExpandFunction struct {
	defaultSeparator string
}

func NewExpandFunction(defaultSeparator string) function.Function {
	return &ExpandFunction{defaultSeparator: defaultSeparator}
}

func (f *ExpandFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "expand"
}

func (f *ExpandFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Expand a flat path-keyed map into a nested map",
		Description: "Splits each key on the separator and rebuilds the nested object structure. Inverse of collapse.",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:        "input",
				Description: "The flat map or object with path-based keys to expand.",
			},
		},
		VariadicParameter: &function.StringParameter{
			Name:           "separator",
			Description:    "Separator used to split key segments. Defaults to '/'.",
			AllowNullValue: true,
		},
		Return: function.DynamicReturn{},
	}
}

func (f *ExpandFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input types.Dynamic
	var separators []types.String

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input, &separators))
	if resp.Error != nil {
		return
	}

	separator := f.defaultSeparator
	if separator == "" {
		separator = "/"
	}
	if len(separators) > 0 && !separators[0].IsNull() && !separators[0].IsUnknown() {
		separator = separators[0].ValueString()
	}

	var elems map[string]attr.Value
	switch v := input.UnderlyingValue().(type) {
	case types.Object:
		elems = v.Attributes()
	case types.Map:
		elems = v.Elements()
	default:
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError("input must be a map or object"))
		return
	}

	outputVal, funcErr := expandAttrMap(ctx, elems, separator)
	if funcErr != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, funcErr)
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.DynamicValue(outputVal)))
}

func expandAttrMap(ctx context.Context, elems map[string]attr.Value, separator string) (attr.Value, *function.FuncError) {
	type treeNode struct {
		leaf     attr.Value
		children map[string]*treeNode
	}

	root := &treeNode{children: make(map[string]*treeNode)}

	for key, val := range elems {
		parts := collapseKeyToPath(key, separator)
		current := root
		for i, part := range parts {
			if i == len(parts)-1 {
				current.children[part] = &treeNode{leaf: val}
			} else {
				if current.children[part] == nil {
					current.children[part] = &treeNode{children: make(map[string]*treeNode)}
				}
				current = current.children[part]
			}
		}
	}

	var buildErr *function.FuncError
	var buildObj func(node *treeNode) attr.Value
	buildObj = func(node *treeNode) attr.Value {
		if buildErr != nil {
			return nil
		}
		if node.leaf != nil {
			return node.leaf
		}
		attrTypes := make(map[string]attr.Type, len(node.children))
		attrVals := make(map[string]attr.Value, len(node.children))
		for k, child := range node.children {
			v := buildObj(child)
			if buildErr != nil {
				return nil
			}
			attrVals[k] = v
			attrTypes[k] = v.Type(ctx)
		}
		obj, diags := types.ObjectValue(attrTypes, attrVals)
		if diags.HasError() {
			buildErr = function.FuncErrorFromDiags(ctx, diags)
			return nil
		}
		return obj
	}

	result := buildObj(root)
	if buildErr != nil {
		return nil, buildErr
	}
	return result, nil
}
