// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCollapseKeyToPath(t *testing.T) {
	tests := []struct {
		key       string
		separator string
		want      []string
	}{
		{"a/b/c", "/", []string{"a", "b", "c"}},
		{"a", "/", []string{"a"}},
		{"a.b.c", ".", []string{"a", "b", "c"}},
		{"a--b", "--", []string{"a", "b"}},
	}
	for _, tc := range tests {
		got := collapseKeyToPath(tc.key, tc.separator)
		if len(got) != len(tc.want) {
			t.Errorf("collapseKeyToPath(%q, %q) = %v, want %v", tc.key, tc.separator, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("collapseKeyToPath(%q, %q)[%d] = %q, want %q", tc.key, tc.separator, i, got[i], tc.want[i])
			}
		}
	}
}

func TestIsAllContainers_StringChild(t *testing.T) {
	obj, _ := types.ObjectValue(
		map[string]attr.Type{"a": types.StringType},
		map[string]attr.Value{"a": types.StringValue("x")},
	)
	if isAllContainers(obj) {
		t.Error("object with string child should return false")
	}
}

func TestIsAllContainers_ObjectChild(t *testing.T) {
	ctx := context.Background()
	inner, _ := types.ObjectValue(
		map[string]attr.Type{"x": types.StringType},
		map[string]attr.Value{"x": types.StringValue("v")},
	)
	outer, _ := types.ObjectValue(
		map[string]attr.Type{"a": inner.Type(ctx)},
		map[string]attr.Value{"a": inner},
	)
	if !isAllContainers(outer) {
		t.Error("object with object child should return true")
	}
}

func TestIsAllContainers_MixedChildren(t *testing.T) {
	ctx := context.Background()
	inner, _ := types.ObjectValue(
		map[string]attr.Type{"x": types.StringType},
		map[string]attr.Value{"x": types.StringValue("v")},
	)
	obj, _ := types.ObjectValue(
		map[string]attr.Type{"a": inner.Type(ctx), "b": types.StringType},
		map[string]attr.Value{"a": inner, "b": types.StringValue("y")},
	)
	if isAllContainers(obj) {
		t.Error("object with mixed children should return false")
	}
}

func TestIsAllContainers_EmptyObject(t *testing.T) {
	obj, _ := types.ObjectValue(map[string]attr.Type{}, map[string]attr.Value{})
	if isAllContainers(obj) {
		t.Error("empty object should return false")
	}
}

func TestIsAllContainers_NonContainer(t *testing.T) {
	if isAllContainers(types.StringValue("x")) {
		t.Error("string value should return false")
	}
}

func TestCollapseValue_AutoDepthStopsAtLeafObjects(t *testing.T) {
	ctx := context.Background()

	// {env: {dev: {region: "us-east-1"}, prod: {region: "eu-west-1"}}}
	devObj, _ := types.ObjectValue(
		map[string]attr.Type{"region": types.StringType},
		map[string]attr.Value{"region": types.StringValue("us-east-1")},
	)
	prodObj, _ := types.ObjectValue(
		map[string]attr.Type{"region": types.StringType},
		map[string]attr.Value{"region": types.StringValue("eu-west-1")},
	)
	envObj, _ := types.ObjectValue(
		map[string]attr.Type{"dev": devObj.Type(ctx), "prod": prodObj.Type(ctx)},
		map[string]attr.Value{"dev": devObj, "prod": prodObj},
	)
	root, _ := types.ObjectValue(
		map[string]attr.Type{"env": envObj.Type(ctx)},
		map[string]attr.Value{"env": envObj},
	)

	result := make(map[string]attr.Value)
	collapseValue("", root, "/", -1, result)

	// Auto mode stops at env/dev and env/prod because their children are strings.
	if len(result) != 2 {
		t.Errorf("expected 2 keys, got %d", len(result))
	}
	if _, ok := result["env/dev"]; !ok {
		t.Error("expected key 'env/dev'")
	}
	if _, ok := result["env/prod"]; !ok {
		t.Error("expected key 'env/prod'")
	}
}

func TestCollapseValue_AutoDepthSingleLevel(t *testing.T) {
	// {a: "x", b: "y"} — top level has string children, so nothing collapses
	root, _ := types.ObjectValue(
		map[string]attr.Type{"a": types.StringType, "b": types.StringType},
		map[string]attr.Value{"a": types.StringValue("x"), "b": types.StringValue("y")},
	)

	result := make(map[string]attr.Value)
	collapseValue("", root, "/", -1, result)

	// Root itself is a leaf (has non-container children), so result = {"": root}
	if _, ok := result[""]; !ok {
		t.Error("expected root to be stored as empty-key leaf")
	}
	if len(result) != 1 {
		t.Errorf("expected 1 key, got %d", len(result))
	}
}

func TestCollapseValue_ExplicitDepth1(t *testing.T) {
	ctx := context.Background()

	inner, _ := types.ObjectValue(
		map[string]attr.Type{"b": types.StringType},
		map[string]attr.Value{"b": types.StringValue("v")},
	)
	root, _ := types.ObjectValue(
		map[string]attr.Type{"a": inner.Type(ctx)},
		map[string]attr.Value{"a": inner},
	)

	result := make(map[string]attr.Value)
	collapseValue("", root, "/", 1, result)

	// Depth 1: expand one level, so {a: inner}
	if len(result) != 1 {
		t.Errorf("expected 1 key, got %d", len(result))
	}
	if _, ok := result["a"]; !ok {
		t.Error("expected key 'a'")
	}
}

func TestCollapseValue_ExplicitDepthUnlimited(t *testing.T) {
	ctx := context.Background()

	leaf, _ := types.ObjectValue(
		map[string]attr.Type{"c": types.StringType},
		map[string]attr.Value{"c": types.StringValue("val")},
	)
	mid, _ := types.ObjectValue(
		map[string]attr.Type{"b": leaf.Type(ctx)},
		map[string]attr.Value{"b": leaf},
	)
	root, _ := types.ObjectValue(
		map[string]attr.Type{"a": mid.Type(ctx)},
		map[string]attr.Value{"a": mid},
	)

	result := make(map[string]attr.Value)
	collapseValue("", root, "/", 100, result)

	if _, ok := result["a/b/c"]; !ok {
		t.Error("expected key 'a/b/c' with explicit depth 100")
	}
	if len(result) != 1 {
		t.Errorf("expected 1 key, got %d", len(result))
	}
}

func TestCollapseValue_CustomSeparator(t *testing.T) {
	ctx := context.Background()

	inner, _ := types.ObjectValue(
		map[string]attr.Type{"b": types.StringType},
		map[string]attr.Value{"b": types.StringValue("v")},
	)
	root, _ := types.ObjectValue(
		map[string]attr.Type{"a": inner.Type(ctx)},
		map[string]attr.Value{"a": inner},
	)

	result := make(map[string]attr.Value)
	collapseValue("", root, ".", 100, result)

	if _, ok := result["a.b"]; !ok {
		t.Errorf("expected key 'a.b' with dot separator")
	}
}
