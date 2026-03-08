// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// buildCombineInput constructs a types.Dynamic wrapping a list of {key, items} objects.
func buildCombineInput(t *testing.T, dims []struct {
	key    string
	values []string
}) types.Dynamic {
	t.Helper()
	itemType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"key":   types.StringType,
			"items": types.ListType{ElemType: types.StringType},
		},
	}
	items := make([]attr.Value, 0, len(dims))
	for _, d := range dims {
		strVals := make([]attr.Value, 0, len(d.values))
		for _, v := range d.values {
			strVals = append(strVals, types.StringValue(v))
		}
		valsList, diags := types.ListValue(types.StringType, strVals)
		if diags.HasError() {
			t.Fatalf("failed to build items list: %v", diags)
		}
		item, diags := types.ObjectValue(
			map[string]attr.Type{
				"key":   types.StringType,
				"items": types.ListType{ElemType: types.StringType},
			},
			map[string]attr.Value{"key": types.StringValue(d.key), "items": valsList},
		)
		if diags.HasError() {
			t.Fatalf("failed to build item object: %v", diags)
		}
		items = append(items, item)
	}
	list, diags := types.ListValue(itemType, items)
	if diags.HasError() {
		t.Fatalf("failed to build list: %v", diags)
	}
	return types.DynamicValue(list)
}

func TestCombineToAttrMap_SingleDimension(t *testing.T) {
	ctx := context.Background()
	input := buildCombineInput(t, []struct {
		key    string
		values []string
	}{
		{"env", []string{"dev", "staging", "prod"}},
	})

	result, funcErr := combineToAttrMap(ctx, input, "/")
	if funcErr != nil {
		t.Fatalf("unexpected error: %v", funcErr)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 entries, got %d", len(result))
	}
	for _, k := range []string{"dev", "staging", "prod"} {
		obj, ok := result[k].(types.Object)
		if !ok {
			t.Errorf("expected object for key %q", k)
			continue
		}
		v, ok := obj.Attributes()["env"].(types.String)
		if !ok || v.ValueString() != k {
			t.Errorf("expected env=%q in result[%q], got %v", k, k, obj.Attributes())
		}
	}
}

func TestCombineToAttrMap_TwoDimensions(t *testing.T) {
	ctx := context.Background()
	input := buildCombineInput(t, []struct {
		key    string
		values []string
	}{
		{"env", []string{"dev", "prod"}},
		{"region", []string{"us-east-1", "eu-west-1"}},
	})

	result, funcErr := combineToAttrMap(ctx, input, "/")
	if funcErr != nil {
		t.Fatalf("unexpected error: %v", funcErr)
	}

	expected := []string{"dev/us-east-1", "dev/eu-west-1", "prod/us-east-1", "prod/eu-west-1"}
	if len(result) != len(expected) {
		t.Errorf("expected %d combinations, got %d", len(expected), len(result))
	}
	for _, k := range expected {
		if _, ok := result[k]; !ok {
			t.Errorf("expected key %q in result", k)
		}
	}
}

func TestCombineToAttrMap_TwoDimensionsObjectAttributes(t *testing.T) {
	ctx := context.Background()
	// Each result object should contain both dimension attributes.
	input := buildCombineInput(t, []struct {
		key    string
		values []string
	}{
		{"env", []string{"dev"}},
		{"region", []string{"us-east-1"}},
	})

	result, funcErr := combineToAttrMap(ctx, input, "/")
	if funcErr != nil {
		t.Fatalf("unexpected error: %v", funcErr)
	}

	obj, ok := result["dev/us-east-1"].(types.Object)
	if !ok {
		t.Fatalf("expected object at key 'dev/us-east-1'")
	}
	if v, ok := obj.Attributes()["env"].(types.String); !ok || v.ValueString() != "dev" {
		t.Errorf("expected env='dev', got %v", obj.Attributes()["env"])
	}
	if v, ok := obj.Attributes()["region"].(types.String); !ok || v.ValueString() != "us-east-1" {
		t.Errorf("expected region='us-east-1', got %v", obj.Attributes()["region"])
	}
}

func TestCombineToAttrMap_ThreeDimensions(t *testing.T) {
	ctx := context.Background()
	input := buildCombineInput(t, []struct {
		key    string
		values []string
	}{
		{"a", []string{"1", "2"}},
		{"b", []string{"x", "y"}},
		{"c", []string{"p", "q"}},
	})

	result, funcErr := combineToAttrMap(ctx, input, "-")
	if funcErr != nil {
		t.Fatalf("unexpected error: %v", funcErr)
	}

	if len(result) != 8 { // 2*2*2
		t.Errorf("expected 8 combinations, got %d", len(result))
	}
	if _, ok := result["1-x-p"]; !ok {
		t.Error("expected key '1-x-p'")
	}
	if _, ok := result["2-y-q"]; !ok {
		t.Error("expected key '2-y-q'")
	}
}

func TestCombineToAttrMap_MapValues(t *testing.T) {
	ctx := context.Background()

	// Build input with map items: {key: "env", items: {dev: "development", prod: "production"}}
	mapVal, diags := types.MapValue(types.StringType, map[string]attr.Value{
		"dev":  types.StringValue("development"),
		"prod": types.StringValue("production"),
	})
	if diags.HasError() {
		t.Fatalf("failed to build map: %v", diags)
	}
	item, diags := types.ObjectValue(
		map[string]attr.Type{
			"key":   types.StringType,
			"items": types.MapType{ElemType: types.StringType},
		},
		map[string]attr.Value{"key": types.StringValue("env"), "items": mapVal},
	)
	if diags.HasError() {
		t.Fatalf("failed to build item: %v", diags)
	}
	itemType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"key":   types.StringType,
			"items": types.MapType{ElemType: types.StringType},
		},
	}
	list, diags := types.ListValue(itemType, []attr.Value{item})
	if diags.HasError() {
		t.Fatalf("failed to build list: %v", diags)
	}

	result, funcErr := combineToAttrMap(ctx, types.DynamicValue(list), "/")
	if funcErr != nil {
		t.Fatalf("unexpected error: %v", funcErr)
	}
	// Keys come from map keys (dev, prod), sorted
	if len(result) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result))
	}
	if _, ok := result["dev"]; !ok {
		t.Error("expected key 'dev'")
	}
	if _, ok := result["prod"]; !ok {
		t.Error("expected key 'prod'")
	}
}

func TestCombineToAttrMap_MapObjectValues(t *testing.T) {
	ctx := context.Background()

	// Build input with object map items: {key: "env", items: {dev: {debug: true}, prod: {debug: false}}}
	devObj, diags := types.ObjectValue(
		map[string]attr.Type{"debug": types.BoolType},
		map[string]attr.Value{"debug": types.BoolValue(true)},
	)
	if diags.HasError() {
		t.Fatalf("failed to build dev object: %v", diags)
	}
	prodObj, diags := types.ObjectValue(
		map[string]attr.Type{"debug": types.BoolType},
		map[string]attr.Value{"debug": types.BoolValue(false)},
	)
	if diags.HasError() {
		t.Fatalf("failed to build prod object: %v", diags)
	}
	itemsObj, diags := types.ObjectValue(
		map[string]attr.Type{
			"dev":  types.ObjectType{AttrTypes: map[string]attr.Type{"debug": types.BoolType}},
			"prod": types.ObjectType{AttrTypes: map[string]attr.Type{"debug": types.BoolType}},
		},
		map[string]attr.Value{"dev": devObj, "prod": prodObj},
	)
	if diags.HasError() {
		t.Fatalf("failed to build items object: %v", diags)
	}
	item, diags := types.ObjectValue(
		map[string]attr.Type{
			"key": types.StringType,
			"items": types.ObjectType{AttrTypes: map[string]attr.Type{
				"dev":  types.ObjectType{AttrTypes: map[string]attr.Type{"debug": types.BoolType}},
				"prod": types.ObjectType{AttrTypes: map[string]attr.Type{"debug": types.BoolType}},
			}},
		},
		map[string]attr.Value{"key": types.StringValue("env"), "items": itemsObj},
	)
	if diags.HasError() {
		t.Fatalf("failed to build item: %v", diags)
	}
	list, diags := types.ListValue(item.Type(ctx), []attr.Value{item})
	if diags.HasError() {
		t.Fatalf("failed to build list: %v", diags)
	}

	result, funcErr := combineToAttrMap(ctx, types.DynamicValue(list), "/")
	if funcErr != nil {
		t.Fatalf("unexpected error: %v", funcErr)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result))
	}

	// Result for "dev" should have env = {id: "dev", debug: true}
	devResult, ok := result["dev"].(types.Object)
	if !ok {
		t.Fatalf("expected object at key 'dev'")
	}
	envAttr, ok := devResult.Attributes()["env"].(types.Object)
	if !ok {
		t.Fatalf("expected env attribute to be an object in dev result")
	}
	if idAttr, ok := envAttr.Attributes()["id"].(types.String); !ok || idAttr.ValueString() != "dev" {
		t.Errorf("expected env.id='dev', got %v", envAttr.Attributes()["id"])
	}
	if debugAttr, ok := envAttr.Attributes()["debug"].(types.Bool); !ok || !debugAttr.ValueBool() {
		t.Errorf("expected env.debug=true for dev, got %v", envAttr.Attributes()["debug"])
	}
}

func TestCombineToAttrMap_NonListInput(t *testing.T) {
	ctx := context.Background()
	input := types.DynamicValue(types.StringValue("not a list"))
	_, funcErr := combineToAttrMap(ctx, input, "/")
	if funcErr == nil {
		t.Error("expected error for non-list input")
	}
}

func TestCombineToAttrMap_EmptyDimension(t *testing.T) {
	ctx := context.Background()
	input := buildCombineInput(t, []struct {
		key    string
		values []string
	}{
		{"env", []string{}},
	})
	_, funcErr := combineToAttrMap(ctx, input, "/")
	if funcErr == nil {
		t.Error("expected error for empty dimension values")
	}
}
