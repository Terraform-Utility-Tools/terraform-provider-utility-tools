// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNestedMinimal_RemovesNullLeaves(t *testing.T) {
	ctx := context.Background()

	leaf, _ := types.ObjectValue(
		map[string]attr.Type{"enabled": types.BoolType},
		map[string]attr.Value{"enabled": types.BoolValue(true)},
	)
	nullLeaf := types.ObjectNull(map[string]attr.Type{"enabled": types.BoolType})
	root := buildNestedObject(t, ctx, map[string]attr.Value{"a": leaf, "b": nullLeaf})

	flatAttrs := make(map[string]attr.Value)
	collapseValue("", root, "/", -1, flatAttrs)
	filtered := filterMapByPredicate(flatAttrs, func(v attr.Value) bool { return !v.IsNull() && !isEmptyValue(v) })

	result, err := expandAttrMap(ctx, filtered, "/")
	if err != nil {
		t.Fatalf("unexpected expand error: %v", err)
	}
	obj, ok := result.(types.Object)
	if !ok {
		t.Fatalf("expected types.Object, got %T", result)
	}
	if _, ok := obj.Attributes()["a"]; !ok {
		t.Error("expected key 'a' to be present")
	}
	if _, ok := obj.Attributes()["b"]; ok {
		t.Error("expected null key 'b' to be removed")
	}
}

func TestNestedMinimal_RemovesEmptyLeaves(t *testing.T) {
	ctx := context.Background()

	leaf, _ := types.ObjectValue(
		map[string]attr.Type{"x": types.StringType},
		map[string]attr.Value{"x": types.StringValue("hello")},
	)
	emptyLeaf, _ := types.ObjectValue(map[string]attr.Type{}, map[string]attr.Value{})
	root := buildNestedObject(t, ctx, map[string]attr.Value{"a": leaf, "b": emptyLeaf})

	flatAttrs := make(map[string]attr.Value)
	collapseValue("", root, "/", -1, flatAttrs)
	filtered := filterMapByPredicate(flatAttrs, func(v attr.Value) bool { return !v.IsNull() && !isEmptyValue(v) })

	result, err := expandAttrMap(ctx, filtered, "/")
	if err != nil {
		t.Fatalf("unexpected expand error: %v", err)
	}
	obj, ok := result.(types.Object)
	if !ok {
		t.Fatalf("expected types.Object, got %T", result)
	}
	if _, ok := obj.Attributes()["a"]; !ok {
		t.Error("expected key 'a' to be present")
	}
	if _, ok := obj.Attributes()["b"]; ok {
		t.Error("expected empty key 'b' to be removed by nestedMinimal")
	}
}

func TestNestedMinimal_DifferenceFromNestedCompact(t *testing.T) {
	ctx := context.Background()

	leaf, _ := types.ObjectValue(
		map[string]attr.Type{"x": types.StringType},
		map[string]attr.Value{"x": types.StringValue("hello")},
	)
	emptyLeaf, _ := types.ObjectValue(map[string]attr.Type{}, map[string]attr.Value{})

	root := buildNestedObject(t, ctx, map[string]attr.Value{"a": leaf, "b": emptyLeaf})

	flatAttrs := make(map[string]attr.Value)
	collapseValue("", root, "/", -1, flatAttrs)

	compactFiltered := filterMapByPredicate(flatAttrs, func(v attr.Value) bool { return !v.IsNull() })
	minimalFiltered := filterMapByPredicate(flatAttrs, func(v attr.Value) bool { return !v.IsNull() && !isEmptyValue(v) })

	// compact should keep "b" (empty but not null), minimal should drop it
	if _, ok := compactFiltered["b"]; !ok {
		t.Error("nestedCompact should keep empty (non-null) leaf 'b'")
	}
	if _, ok := minimalFiltered["b"]; ok {
		t.Error("nestedMinimal should remove empty leaf 'b'")
	}
}
