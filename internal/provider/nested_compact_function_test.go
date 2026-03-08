// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// buildNestedInput constructs a two-level nested object suitable for nestedCompact/nestedMinimal testing.
// Structure: {outer: {inner1: leafObj1, inner2: leafObj2, ...}}.
func buildNestedObject(t *testing.T, ctx context.Context, leaves map[string]attr.Value) types.Object {
	t.Helper()
	leafTypes := make(map[string]attr.Type, len(leaves))
	for k, v := range leaves {
		leafTypes[k] = v.Type(ctx)
	}
	obj, diags := types.ObjectValue(leafTypes, leaves)
	if diags.HasError() {
		t.Fatalf("failed to build nested object: %v", diags)
	}
	return obj
}

func TestNestedCompact_RemovesNullLeaves(t *testing.T) {
	ctx := context.Background()

	leafA, _ := types.ObjectValue(
		map[string]attr.Type{"enabled": types.BoolType},
		map[string]attr.Value{"enabled": types.BoolValue(true)},
	)

	// Build: {a: leafA, b: null_object}
	nullLeaf := types.ObjectNull(map[string]attr.Type{"enabled": types.BoolType})
	root := buildNestedObject(t, ctx, map[string]attr.Value{
		"a": leafA,
		"b": nullLeaf,
	})

	flatAttrs := make(map[string]attr.Value)
	collapseValue("", root, "/", -1, flatAttrs)
	filtered := filterMapByPredicate(flatAttrs, func(v attr.Value) bool { return !v.IsNull() })

	result, err := expandAttrMap(ctx, filtered, "/")
	if err != nil {
		t.Fatalf("unexpected expand error: %v", err)
	}
	obj, ok := result.(types.Object)
	if !ok {
		t.Fatalf("expected types.Object, got %T", result)
	}
	if _, ok := obj.Attributes()["a"]; !ok {
		t.Error("expected key 'a' to be present after nestedCompact")
	}
	if _, ok := obj.Attributes()["b"]; ok {
		t.Error("expected null key 'b' to be removed by nestedCompact")
	}
}

func TestNestedCompact_PreservesAllNonNullLeaves(t *testing.T) {
	ctx := context.Background()

	leafA, _ := types.ObjectValue(
		map[string]attr.Type{"x": types.StringType},
		map[string]attr.Value{"x": types.StringValue("hello")},
	)
	leafB, _ := types.ObjectValue(
		map[string]attr.Type{"x": types.StringType},
		map[string]attr.Value{"x": types.StringValue("world")},
	)
	root := buildNestedObject(t, ctx, map[string]attr.Value{"a": leafA, "b": leafB})

	flatAttrs := make(map[string]attr.Value)
	collapseValue("", root, "/", -1, flatAttrs)
	filtered := filterMapByPredicate(flatAttrs, func(v attr.Value) bool { return !v.IsNull() })

	result, err := expandAttrMap(ctx, filtered, "/")
	if err != nil {
		t.Fatalf("unexpected expand error: %v", err)
	}
	obj, ok := result.(types.Object)
	if !ok {
		t.Fatalf("expected types.Object, got %T", result)
	}
	if len(obj.Attributes()) != 2 {
		t.Errorf("expected 2 keys, got %d", len(obj.Attributes()))
	}
}

func TestIsEmptyValue(t *testing.T) {
	ctx := context.Background()

	emptyList, _ := types.ListValue(types.StringType, []attr.Value{})
	nonEmptyList, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("x")})
	emptyMap, _ := types.MapValue(types.StringType, map[string]attr.Value{})
	nonEmptyMap, _ := types.MapValue(types.StringType, map[string]attr.Value{"k": types.StringValue("v")})
	emptyObj, _ := types.ObjectValue(map[string]attr.Type{}, map[string]attr.Value{})
	nonEmptyObj, _ := types.ObjectValue(map[string]attr.Type{"a": types.StringType}, map[string]attr.Value{"a": types.StringValue("v")})
	emptyTuple, _ := types.TupleValue([]attr.Type{}, []attr.Value{})
	_ = ctx

	cases := []struct {
		name  string
		val   attr.Value
		empty bool
	}{
		{"empty list", emptyList, true},
		{"non-empty list", nonEmptyList, false},
		{"empty map", emptyMap, true},
		{"non-empty map", nonEmptyMap, false},
		{"empty object", emptyObj, true},
		{"non-empty object", nonEmptyObj, false},
		{"empty tuple", emptyTuple, true},
		{"non-empty string", types.StringValue("x"), false},
		// isEmptyValue does not handle strings — empty string is handled by isEmptyString
		{"empty string", types.StringValue(""), false},
	}
	for _, tc := range cases {
		if got := isEmptyValue(tc.val); got != tc.empty {
			t.Errorf("isEmptyValue(%s): got %v, want %v", tc.name, got, tc.empty)
		}
	}
}

func TestIsEmptyString(t *testing.T) {
	cases := []struct {
		name  string
		val   attr.Value
		empty bool
	}{
		{"empty string", types.StringValue(""), true},
		{"non-empty string", types.StringValue("hello"), false},
		{"null string", types.StringNull(), false},
		{"bool", types.BoolValue(false), false},
		{"number zero", types.NumberUnknown(), false},
	}
	for _, tc := range cases {
		if got := isEmptyString(tc.val); got != tc.empty {
			t.Errorf("isEmptyString(%s): got %v, want %v", tc.name, got, tc.empty)
		}
	}
}
