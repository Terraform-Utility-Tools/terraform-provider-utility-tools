// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ function.Function = &CombineFunction{}

type CombineFunction struct {
	defaultSeparator string
}

func NewCombineFunction(defaultSeparator string) function.Function {
	return &CombineFunction{defaultSeparator: defaultSeparator}
}

func (f *CombineFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "combine"
}

func (f *CombineFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "An opinionated setproduct for objects",
		Description: "Combine a map of lists or maps into a single map with an item for each combination of items in the lists or maps",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:        "input",
				Description: "Ordered list of {key, items} objects to combine. Each object has a 'key' string (the attribute name in the result) and an 'items' list or map. When 'items' is a map of objects, each result entry's dimension attribute is set to the map value merged with an 'id' field containing the map key.",
			},
		},
		VariadicParameter: &function.StringParameter{
			Name:        "separator",
			Description: "Separator string used to generate the keys in the combined map. Defaults to '/'.",
		},
		Return: function.DynamicReturn{},
	}
}

func (f *CombineFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var inputDyn types.Dynamic
	var separators []string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &inputDyn, &separators))
	if resp.Error != nil {
		return
	}

	keySeparator := f.defaultSeparator
	if keySeparator == "" {
		keySeparator = "/"
	}
	if len(separators) > 0 && separators[0] != "" {
		keySeparator = separators[0]
	}

	flatMap, funcErr := combineToAttrMap(ctx, inputDyn, keySeparator)
	if funcErr != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, funcErr)
		return
	}

	attrTypes := make(map[string]attr.Type, len(flatMap))
	for k, v := range flatMap {
		attrTypes[k] = v.Type(ctx)
	}
	outputObj, diags := types.ObjectValue(attrTypes, flatMap)
	if diags.HasError() {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, diags))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.DynamicValue(outputObj)))
}

func mergeKeyAttribute(ctx context.Context, key string, obj types.Object) attr.Value {
	existing := obj.Attributes()
	merged := make(map[string]attr.Value, len(existing)+1)
	for k, v := range existing {
		merged[k] = v
	}
	merged["id"] = types.StringValue(key)
	attrTypes := make(map[string]attr.Type, len(merged))
	for k, v := range merged {
		attrTypes[k] = v.Type(ctx)
	}
	result, _ := types.ObjectValue(attrTypes, merged)
	return result
}

func combineToAttrMap(ctx context.Context, inputDyn types.Dynamic, keySeparator string) (map[string]attr.Value, *function.FuncError) {
	var inputElements []attr.Value
	switch v := inputDyn.UnderlyingValue().(type) {
	case types.List:
		inputElements = v.Elements()
	case types.Tuple:
		inputElements = v.Elements()
	default:
		return nil, function.NewFuncError("input must be a list")
	}

	type entry struct {
		pathKey string
		value   attr.Value
	}
	type dimension struct {
		key     string
		entries []entry
	}

	dims := make([]dimension, 0, len(inputElements))
	for _, elem := range inputElements {
		obj, ok := elem.(types.Object)
		if !ok {
			return nil, function.NewFuncError("each input element must be an object")
		}
		attrs := obj.Attributes()
		keyStr, ok := attrs["key"].(types.String)
		if !ok {
			return nil, function.NewFuncError("each input element must have a string 'key' attribute")
		}
		key := keyStr.ValueString()

		var entries []entry
		switch v := attrs["items"].(type) {
		case types.List:
			for _, elem := range v.Elements() {
				sv, ok := elem.(types.String)
				if !ok {
					return nil, function.NewFuncError("list items must be strings")
				}
				entries = append(entries, entry{pathKey: sv.ValueString(), value: sv})
			}
		case types.Tuple:
			for _, elem := range v.Elements() {
				sv, ok := elem.(types.String)
				if !ok {
					return nil, function.NewFuncError("tuple items must be strings")
				}
				entries = append(entries, entry{pathKey: sv.ValueString(), value: sv})
			}
		case types.Map:
			elems := v.Elements()
			keys := make([]string, 0, len(elems))
			for k := range elems {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				val := elems[k]
				if obj, ok := val.(types.Object); ok {
					val = mergeKeyAttribute(ctx, k, obj)
				}
				entries = append(entries, entry{pathKey: k, value: val})
			}
		case types.Object:
			elems := v.Attributes()
			keys := make([]string, 0, len(elems))
			for k := range elems {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				val := elems[k]
				if obj, ok := val.(types.Object); ok {
					val = mergeKeyAttribute(ctx, k, obj)
				}
				entries = append(entries, entry{pathKey: k, value: val})
			}
		}
		if len(entries) == 0 {
			return nil, function.NewFuncError(
				"dimension \"" + key + "\" has no entries: 'items' must be a non-empty list or map",
			)
		}
		dims = append(dims, dimension{key: key, entries: entries})
	}

	type combo struct {
		pathKeys []string
		values   map[string]attr.Value
	}

	combos := []combo{{pathKeys: []string{}, values: map[string]attr.Value{}}}
	for _, dim := range dims {
		var expanded []combo
		for _, c := range combos {
			for _, e := range dim.entries {
				newValues := make(map[string]attr.Value, len(c.values)+1)
				for k, v := range c.values {
					newValues[k] = v
				}
				newValues[dim.key] = e.value
				expanded = append(expanded, combo{
					pathKeys: append(append([]string{}, c.pathKeys...), e.pathKey),
					values:   newValues,
				})
			}
		}
		combos = expanded
	}

	result := make(map[string]attr.Value, len(combos))
	for _, c := range combos {
		path := strings.Join(c.pathKeys, keySeparator)
		innerTypes := make(map[string]attr.Type, len(c.values))
		for k, v := range c.values {
			innerTypes[k] = v.Type(ctx)
		}
		comboObj, diags := types.ObjectValue(innerTypes, c.values)
		if diags.HasError() {
			return nil, function.FuncErrorFromDiags(ctx, diags)
		}
		result[path] = comboObj
	}

	return result, nil
}
