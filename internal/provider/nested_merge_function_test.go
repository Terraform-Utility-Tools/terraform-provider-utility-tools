// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func buildObj(t *testing.T, attrs map[string]attr.Value) types.Object {
	t.Helper()
	attrTypes := make(map[string]attr.Type, len(attrs))
	for k, v := range attrs {
		attrTypes[k] = v.Type(context.Background())
	}
	obj, diags := types.ObjectValue(attrTypes, attrs)
	if diags.HasError() {
		t.Fatalf("failed to build object: %v", diags)
	}
	return obj
}

func TestNestedMergeTwo_ScalarOverride(t *testing.T) {
	ctx := context.Background()
	base := buildObj(t, map[string]attr.Value{
		"a": types.StringValue("old"),
		"b": types.StringValue("keep"),
	})
	override := buildObj(t, map[string]attr.Value{
		"a": types.StringValue("new"),
	})

	result, funcErr := nestedMergeTwo(ctx, base, override)
	if funcErr != nil {
		t.Fatalf("unexpected error: %v", funcErr)
	}
	obj := result.(types.Object)
	if v := obj.Attributes()["a"].(types.String).ValueString(); v != "new" {
		t.Errorf("expected a='new', got %q", v)
	}
	if v := obj.Attributes()["b"].(types.String).ValueString(); v != "keep" {
		t.Errorf("expected b='keep', got %q", v)
	}
}

func TestNestedMergeTwo_DeepMergeObjects(t *testing.T) {
	ctx := context.Background()
	inner1 := buildObj(t, map[string]attr.Value{
		"x": types.BoolValue(true),
		"z": types.BoolValue(true),
	})
	inner2 := buildObj(t, map[string]attr.Value{
		"x": types.BoolValue(false),
	})
	base := buildObj(t, map[string]attr.Value{"b": inner1})
	override := buildObj(t, map[string]attr.Value{"b": inner2})

	result, funcErr := nestedMergeTwo(ctx, base, override)
	if funcErr != nil {
		t.Fatalf("unexpected error: %v", funcErr)
	}
	obj := result.(types.Object)
	b := obj.Attributes()["b"].(types.Object)
	if v := b.Attributes()["x"].(types.Bool).ValueBool(); v != false {
		t.Error("expected b.x=false (overridden)")
	}
	if v := b.Attributes()["z"].(types.Bool).ValueBool(); v != true {
		t.Error("expected b.z=true (preserved from base)")
	}
}

func TestNestedMergeTwo_NewKeyAdded(t *testing.T) {
	ctx := context.Background()
	base := buildObj(t, map[string]attr.Value{"a": types.StringValue("1")})
	override := buildObj(t, map[string]attr.Value{"c": types.StringValue("3")})

	result, funcErr := nestedMergeTwo(ctx, base, override)
	if funcErr != nil {
		t.Fatalf("unexpected error: %v", funcErr)
	}
	obj := result.(types.Object)
	if len(obj.Attributes()) != 2 {
		t.Errorf("expected 2 attributes, got %d", len(obj.Attributes()))
	}
}

func TestNestedMergeTwo_NonObjectOverridesObject(t *testing.T) {
	ctx := context.Background()
	inner := buildObj(t, map[string]attr.Value{"z": types.BoolValue(true)})
	base := buildObj(t, map[string]attr.Value{"b": inner})
	override := buildObj(t, map[string]attr.Value{"b": types.StringValue("replaced")})

	result, funcErr := nestedMergeTwo(ctx, base, override)
	if funcErr != nil {
		t.Fatalf("unexpected error: %v", funcErr)
	}
	obj := result.(types.Object)
	if v, ok := obj.Attributes()["b"].(types.String); !ok || v.ValueString() != "replaced" {
		t.Error("expected b to be overridden with string 'replaced'")
	}
}

func TestNestedMergeTwo_ThreeWay(t *testing.T) {
	// Replicates the user's first example:
	// merge({a="b", b={z=true}}, {a=[1,2], c="z"}, {b={x=false}, d=3})
	// Result: a=[1,2], b={x=false, z=true}, c="z", d=3
	ctx := context.Background()

	listVal, diags := types.ListValue(types.NumberType, []attr.Value{
		types.NumberValue(big.NewFloat(1)),
		types.NumberValue(big.NewFloat(2)),
	})
	if diags.HasError() {
		t.Fatalf("failed to build list: %v", diags)
	}

	bInner1 := buildObj(t, map[string]attr.Value{"z": types.BoolValue(true)})
	bInner3 := buildObj(t, map[string]attr.Value{"x": types.BoolValue(false)})

	m1 := buildObj(t, map[string]attr.Value{"a": types.StringValue("b"), "b": bInner1})
	m2 := buildObj(t, map[string]attr.Value{"a": listVal, "c": types.StringValue("z")})
	m3 := buildObj(t, map[string]attr.Value{"b": bInner3, "d": types.NumberValue(big.NewFloat(3))})

	r12, funcErr := nestedMergeTwo(ctx, m1, m2)
	if funcErr != nil {
		t.Fatalf("merge m1+m2: %v", funcErr)
	}
	result, funcErr := nestedMergeTwo(ctx, r12, m3)
	if funcErr != nil {
		t.Fatalf("merge (m1+m2)+m3: %v", funcErr)
	}

	obj := result.(types.Object)

	// a should be list [1,2]
	if _, ok := obj.Attributes()["a"].(types.List); !ok {
		t.Error("expected a to be a list")
	}
	// b should be deep-merged: {x=false, z=true}
	b := obj.Attributes()["b"].(types.Object)
	if v := b.Attributes()["z"].(types.Bool).ValueBool(); v != true {
		t.Error("expected b.z=true (from m1)")
	}
	if v := b.Attributes()["x"].(types.Bool).ValueBool(); v != false {
		t.Error("expected b.x=false (from m3)")
	}
	// c and d should be present
	if _, ok := obj.Attributes()["c"]; !ok {
		t.Error("expected c in result")
	}
	if _, ok := obj.Attributes()["d"]; !ok {
		t.Error("expected d in result")
	}
}

func TestNestedMergeTwo_EmptyBase(t *testing.T) {
	ctx := context.Background()
	base := buildObj(t, map[string]attr.Value{})
	override := buildObj(t, map[string]attr.Value{"a": types.StringValue("1")})

	result, funcErr := nestedMergeTwo(ctx, base, override)
	if funcErr != nil {
		t.Fatalf("unexpected error: %v", funcErr)
	}
	obj := result.(types.Object)
	if len(obj.Attributes()) != 1 {
		t.Errorf("expected 1 attribute, got %d", len(obj.Attributes()))
	}
}
