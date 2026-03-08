// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpandAttrMap_Empty(t *testing.T) {
	ctx := context.Background()
	result, err := expandAttrMap(ctx, map[string]attr.Value{}, "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	obj, ok := result.(types.Object)
	if !ok {
		t.Fatalf("expected types.Object, got %T", result)
	}
	if len(obj.Attributes()) != 0 {
		t.Errorf("expected empty object, got %v", obj.Attributes())
	}
}

func TestExpandAttrMap_SingleFlatKey(t *testing.T) {
	ctx := context.Background()
	result, err := expandAttrMap(ctx, map[string]attr.Value{
		"key": types.StringValue("val"),
	}, "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	obj := result.(types.Object)
	v, ok := obj.Attributes()["key"].(types.String)
	if !ok || v.ValueString() != "val" {
		t.Errorf("expected key='val', got %v", obj.Attributes()["key"])
	}
}

func TestExpandAttrMap_TwoLevelKey(t *testing.T) {
	ctx := context.Background()
	result, err := expandAttrMap(ctx, map[string]attr.Value{
		"a/b": types.StringValue("hello"),
	}, "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	obj := result.(types.Object)
	inner, ok := obj.Attributes()["a"].(types.Object)
	if !ok {
		t.Fatalf("expected 'a' to be types.Object")
	}
	v, ok := inner.Attributes()["b"].(types.String)
	if !ok || v.ValueString() != "hello" {
		t.Errorf("expected a.b='hello', got %v", inner.Attributes()["b"])
	}
}

func TestExpandAttrMap_SharedPrefix(t *testing.T) {
	ctx := context.Background()
	result, err := expandAttrMap(ctx, map[string]attr.Value{
		"a/b": types.StringValue("b_val"),
		"a/c": types.StringValue("c_val"),
	}, "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	obj := result.(types.Object)
	inner, ok := obj.Attributes()["a"].(types.Object)
	if !ok {
		t.Fatalf("expected 'a' to be an object")
	}
	if _, ok := inner.Attributes()["b"]; !ok {
		t.Error("expected key 'b' under 'a'")
	}
	if _, ok := inner.Attributes()["c"]; !ok {
		t.Error("expected key 'c' under 'a'")
	}
	if len(inner.Attributes()) != 2 {
		t.Errorf("expected 2 keys under 'a', got %d", len(inner.Attributes()))
	}
}

func TestExpandAttrMap_MultipleTopLevelKeys(t *testing.T) {
	ctx := context.Background()
	result, err := expandAttrMap(ctx, map[string]attr.Value{
		"x/y": types.StringValue("1"),
		"p/q": types.StringValue("2"),
	}, "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	obj := result.(types.Object)
	if _, ok := obj.Attributes()["x"]; !ok {
		t.Error("expected top-level key 'x'")
	}
	if _, ok := obj.Attributes()["p"]; !ok {
		t.Error("expected top-level key 'p'")
	}
}

func TestExpandAttrMap_CustomSeparator(t *testing.T) {
	ctx := context.Background()
	result, err := expandAttrMap(ctx, map[string]attr.Value{
		"a.b.c": types.StringValue("deep"),
	}, ".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	obj := result.(types.Object)
	mid, ok := obj.Attributes()["a"].(types.Object)
	if !ok {
		t.Fatalf("expected 'a' to be object")
	}
	inner, ok := mid.Attributes()["b"].(types.Object)
	if !ok {
		t.Fatalf("expected 'b' to be object")
	}
	v, ok := inner.Attributes()["c"].(types.String)
	if !ok || v.ValueString() != "deep" {
		t.Errorf("expected a.b.c='deep', got %v", inner.Attributes()["c"])
	}
}

func TestExpandAttrMap_RoundTrip(t *testing.T) {
	ctx := context.Background()

	// Build nested structure, collapse it, then expand and verify
	inner, _ := types.ObjectValue(
		map[string]attr.Type{"x": types.StringType},
		map[string]attr.Value{"x": types.StringValue("hello")},
	)
	root, _ := types.ObjectValue(
		map[string]attr.Type{"a": inner.Type(ctx)},
		map[string]attr.Value{"a": inner},
	)

	flat := make(map[string]attr.Value)
	collapseValue(ctx, "", root, "/", 100, flat)

	expanded, err := expandAttrMap(ctx, flat, "/")
	if err != nil {
		t.Fatalf("expand error: %v", err)
	}
	expandedObj := expanded.(types.Object)
	innerResult, ok := expandedObj.Attributes()["a"].(types.Object)
	if !ok {
		t.Fatalf("expected 'a' to be object after round-trip")
	}
	v, ok := innerResult.Attributes()["x"].(types.String)
	if !ok || v.ValueString() != "hello" {
		t.Errorf("expected a.x='hello' after round-trip, got %v", innerResult.Attributes()["x"])
	}
}
