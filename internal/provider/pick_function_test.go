// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func buildStringList(t *testing.T, keys []string) types.List {
	t.Helper()
	elems := make([]attr.Value, len(keys))
	for i, k := range keys {
		elems[i] = types.StringValue(k)
	}
	list, diags := types.ListValue(types.StringType, elems)
	if diags.HasError() {
		t.Fatalf("failed to build keys list: %v", diags)
	}
	return list
}

func buildTestObject(t *testing.T, attrs map[string]string) types.Object {
	t.Helper()
	attrTypes := make(map[string]attr.Type, len(attrs))
	attrVals := make(map[string]attr.Value, len(attrs))
	for k, v := range attrs {
		attrTypes[k] = types.StringType
		attrVals[k] = types.StringValue(v)
	}
	obj, diags := types.ObjectValue(attrTypes, attrVals)
	if diags.HasError() {
		t.Fatalf("failed to build object: %v", diags)
	}
	return obj
}

// --- pick ---

func TestPick_KeepsSpecifiedKeys(t *testing.T) {
	ctx := context.Background()
	obj := buildTestObject(t, map[string]string{"env": "prod", "region": "eu", "service": "api"})
	keys := buildStringList(t, []string{"env", "service"})

	result, funcErr := applyKeySelection(ctx, types.DynamicValue(obj), func(k string) bool {
		set := map[string]struct{}{"env": {}, "service": {}}
		_, ok := set[k]
		return ok
	})
	if funcErr != nil {
		t.Fatalf("unexpected error: %v", funcErr)
	}
	out, ok := result.(types.Object)
	if !ok {
		t.Fatalf("expected object result")
	}
	_ = keys
	if len(out.Attributes()) != 2 {
		t.Errorf("expected 2 attributes, got %d", len(out.Attributes()))
	}
	if _, ok := out.Attributes()["env"]; !ok {
		t.Error("expected 'env' in result")
	}
	if _, ok := out.Attributes()["service"]; !ok {
		t.Error("expected 'service' in result")
	}
	if _, ok := out.Attributes()["region"]; ok {
		t.Error("'region' should be excluded")
	}
}

func TestPick_MissingKeyIgnored(t *testing.T) {
	ctx := context.Background()
	obj := buildTestObject(t, map[string]string{"env": "prod"})

	result, funcErr := applyKeySelection(ctx, types.DynamicValue(obj), func(k string) bool {
		return k == "env" || k == "nonexistent"
	})
	if funcErr != nil {
		t.Fatalf("unexpected error: %v", funcErr)
	}
	out := result.(types.Object)
	if len(out.Attributes()) != 1 {
		t.Errorf("expected 1 attribute, got %d", len(out.Attributes()))
	}
}

func TestPick_EmptyKeyList(t *testing.T) {
	ctx := context.Background()
	obj := buildTestObject(t, map[string]string{"env": "prod", "region": "eu"})

	result, funcErr := applyKeySelection(ctx, types.DynamicValue(obj), func(k string) bool {
		return false
	})
	if funcErr != nil {
		t.Fatalf("unexpected error: %v", funcErr)
	}
	out := result.(types.Object)
	if len(out.Attributes()) != 0 {
		t.Errorf("expected empty result, got %d attributes", len(out.Attributes()))
	}
}

func TestPick_MapInput(t *testing.T) {
	ctx := context.Background()
	m, diags := types.MapValue(types.StringType, map[string]attr.Value{
		"a": types.StringValue("1"),
		"b": types.StringValue("2"),
		"c": types.StringValue("3"),
	})
	if diags.HasError() {
		t.Fatalf("failed to build map: %v", diags)
	}

	result, funcErr := applyKeySelection(ctx, types.DynamicValue(m), func(k string) bool {
		return k == "a" || k == "c"
	})
	if funcErr != nil {
		t.Fatalf("unexpected error: %v", funcErr)
	}
	out, ok := result.(types.Map)
	if !ok {
		t.Fatalf("expected map result")
	}
	if len(out.Elements()) != 2 {
		t.Errorf("expected 2 elements, got %d", len(out.Elements()))
	}
	if _, ok := out.Elements()["b"]; ok {
		t.Error("'b' should be excluded")
	}
}

func TestPick_InvalidInput(t *testing.T) {
	ctx := context.Background()
	_, funcErr := applyKeySelection(ctx, types.DynamicValue(types.StringValue("bad")), func(k string) bool { return true })
	if funcErr == nil {
		t.Error("expected error for non-map/object input")
	}
}

// --- omit ---

func TestOmit_RemovesSpecifiedKeys(t *testing.T) {
	ctx := context.Background()
	obj := buildTestObject(t, map[string]string{"env": "prod", "region": "eu", "service": "api"})

	result, funcErr := applyKeySelection(ctx, types.DynamicValue(obj), func(k string) bool {
		return k != "region"
	})
	if funcErr != nil {
		t.Fatalf("unexpected error: %v", funcErr)
	}
	out := result.(types.Object)
	if len(out.Attributes()) != 2 {
		t.Errorf("expected 2 attributes, got %d", len(out.Attributes()))
	}
	if _, ok := out.Attributes()["region"]; ok {
		t.Error("'region' should be removed")
	}
}

func TestOmit_MissingKeyIgnored(t *testing.T) {
	ctx := context.Background()
	obj := buildTestObject(t, map[string]string{"env": "prod", "region": "eu"})

	result, funcErr := applyKeySelection(ctx, types.DynamicValue(obj), func(k string) bool {
		return k != "nonexistent"
	})
	if funcErr != nil {
		t.Fatalf("unexpected error: %v", funcErr)
	}
	out := result.(types.Object)
	if len(out.Attributes()) != 2 {
		t.Errorf("expected 2 attributes unchanged, got %d", len(out.Attributes()))
	}
}

// --- listToStringSet ---

func TestListToStringSet(t *testing.T) {
	ctx := context.Background()
	_ = ctx
	list, diags := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("a"),
		types.StringValue("b"),
		types.StringNull(),
	})
	if diags.HasError() {
		t.Fatalf("failed to build list: %v", diags)
	}
	set := listToStringSet(list)
	if _, ok := set["a"]; !ok {
		t.Error("expected 'a' in set")
	}
	if _, ok := set["b"]; !ok {
		t.Error("expected 'b' in set")
	}
	if len(set) != 2 {
		t.Errorf("expected 2 entries (null skipped), got %d", len(set))
	}
}
