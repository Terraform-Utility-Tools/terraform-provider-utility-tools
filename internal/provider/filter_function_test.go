// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- matchesAnyCondition ---

func TestMatchesAnyCondition_ObjectMatchesSingleAttr(t *testing.T) {
	obj, _ := types.ObjectValue(
		map[string]attr.Type{"env": types.StringType, "region": types.StringType},
		map[string]attr.Value{"env": types.StringValue("prod"), "region": types.StringValue("us-east-1")},
	)
	cond := []map[string]attr.Value{{"env": types.StringValue("prod")}}
	if !matchesAnyCondition(obj, cond) {
		t.Error("object should match condition {env: prod}")
	}
}

func TestMatchesAnyCondition_ObjectMatchesMultipleAttrs(t *testing.T) {
	obj, _ := types.ObjectValue(
		map[string]attr.Type{"env": types.StringType, "region": types.StringType},
		map[string]attr.Value{"env": types.StringValue("prod"), "region": types.StringValue("us-east-1")},
	)
	cond := []map[string]attr.Value{{
		"env":    types.StringValue("prod"),
		"region": types.StringValue("us-east-1"),
	}}
	if !matchesAnyCondition(obj, cond) {
		t.Error("object should match condition with both attrs")
	}
}

func TestMatchesAnyCondition_ObjectDoesNotMatchAttr(t *testing.T) {
	obj, _ := types.ObjectValue(
		map[string]attr.Type{"env": types.StringType},
		map[string]attr.Value{"env": types.StringValue("dev")},
	)
	cond := []map[string]attr.Value{{"env": types.StringValue("prod")}}
	if matchesAnyCondition(obj, cond) {
		t.Error("object with env=dev should not match condition {env: prod}")
	}
}

func TestMatchesAnyCondition_ObjectMissingConditionKey(t *testing.T) {
	obj, _ := types.ObjectValue(
		map[string]attr.Type{"region": types.StringType},
		map[string]attr.Value{"region": types.StringValue("us-east-1")},
	)
	cond := []map[string]attr.Value{{"env": types.StringValue("prod")}}
	if matchesAnyCondition(obj, cond) {
		t.Error("object missing condition key should not match")
	}
}

func TestMatchesAnyCondition_NonObjectWithCondition(t *testing.T) {
	cond := []map[string]attr.Value{{"env": types.StringValue("prod")}}
	if matchesAnyCondition(types.StringValue("prod"), cond) {
		t.Error("non-object value should not match a condition")
	}
}

func TestMatchesAnyCondition_PartialConditionMismatch(t *testing.T) {
	obj, _ := types.ObjectValue(
		map[string]attr.Type{"env": types.StringType, "region": types.StringType},
		map[string]attr.Value{"env": types.StringValue("prod"), "region": types.StringValue("eu-west-1")},
	)
	cond := []map[string]attr.Value{{
		"env":    types.StringValue("prod"),
		"region": types.StringValue("us-east-1"),
	}}
	if matchesAnyCondition(obj, cond) {
		t.Error("object should not match when one condition attribute differs")
	}
}

func TestMatchesAnyCondition_ORMatchesSecondCondition(t *testing.T) {
	obj, _ := types.ObjectValue(
		map[string]attr.Type{"env": types.StringType},
		map[string]attr.Value{"env": types.StringValue("staging")},
	)
	// First condition doesn't match, second does
	cond := []map[string]attr.Value{
		{"env": types.StringValue("dev")},
		{"env": types.StringValue("staging")},
	}
	if !matchesAnyCondition(obj, cond) {
		t.Error("object should match via second OR condition")
	}
}

func TestMatchesAnyCondition_ORNoConditionMatches(t *testing.T) {
	obj, _ := types.ObjectValue(
		map[string]attr.Type{"env": types.StringType},
		map[string]attr.Value{"env": types.StringValue("prod")},
	)
	cond := []map[string]attr.Value{
		{"env": types.StringValue("dev")},
		{"env": types.StringValue("staging")},
	}
	if matchesAnyCondition(obj, cond) {
		t.Error("object should not match when no OR condition applies")
	}
}

func TestMatchesAnyCondition_NestedObjectSubsetMatch(t *testing.T) {
	inner, _ := types.ObjectValue(
		map[string]attr.Type{"key": types.StringType, "tls": types.BoolType},
		map[string]attr.Value{"key": types.StringValue("api"), "tls": types.BoolValue(true)},
	)
	outer, _ := types.ObjectValue(
		map[string]attr.Type{"env": types.StringType, "service": inner.Type(context.Background())},
		map[string]attr.Value{"env": types.StringValue("prod"), "service": inner},
	)
	// Condition: service has key="api" (subset match — tls not required)
	condInner, _ := types.ObjectValue(
		map[string]attr.Type{"key": types.StringType},
		map[string]attr.Value{"key": types.StringValue("api")},
	)
	cond := []map[string]attr.Value{{"env": types.StringValue("prod"), "service": condInner}}
	if !matchesAnyCondition(outer, cond) {
		t.Error("should match: env=prod and service contains key=api")
	}
}

func TestMatchesAnyCondition_NestedObjectMismatch(t *testing.T) {
	inner, _ := types.ObjectValue(
		map[string]attr.Type{"key": types.StringType},
		map[string]attr.Value{"key": types.StringValue("web")},
	)
	outer, _ := types.ObjectValue(
		map[string]attr.Type{"env": types.StringType, "service": inner.Type(context.Background())},
		map[string]attr.Value{"env": types.StringValue("prod"), "service": inner},
	)
	condInner, _ := types.ObjectValue(
		map[string]attr.Type{"key": types.StringType},
		map[string]attr.Value{"key": types.StringValue("api")},
	)
	cond := []map[string]attr.Value{{"service": condInner}}
	if matchesAnyCondition(outer, cond) {
		t.Error("should not match: service.key=web does not match api")
	}
}

// --- filterAttrMap ---

func TestFilterAttrMap_ConditionFilter(t *testing.T) {
	devObj, _ := types.ObjectValue(
		map[string]attr.Type{"env": types.StringType},
		map[string]attr.Value{"env": types.StringValue("dev")},
	)
	prodObj, _ := types.ObjectValue(
		map[string]attr.Type{"env": types.StringType},
		map[string]attr.Value{"env": types.StringValue("prod")},
	)
	elems := map[string]attr.Value{"x": devObj, "y": prodObj}
	cond := []map[string]attr.Value{{"env": types.StringValue("prod")}}

	result := filterAttrMap(elems, cond)
	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
	if _, ok := result["y"]; !ok {
		t.Error("expected key 'y' (prod)")
	}
}

func TestFilterAttrMap_ORConditionFilter(t *testing.T) {
	devObj, _ := types.ObjectValue(
		map[string]attr.Type{"env": types.StringType},
		map[string]attr.Value{"env": types.StringValue("dev")},
	)
	prodObj, _ := types.ObjectValue(
		map[string]attr.Type{"env": types.StringType},
		map[string]attr.Value{"env": types.StringValue("prod")},
	)
	stagingObj, _ := types.ObjectValue(
		map[string]attr.Type{"env": types.StringType},
		map[string]attr.Value{"env": types.StringValue("staging")},
	)
	elems := map[string]attr.Value{"x": devObj, "y": prodObj, "z": stagingObj}
	cond := []map[string]attr.Value{
		{"env": types.StringValue("dev")},
		{"env": types.StringValue("prod")},
	}

	result := filterAttrMap(elems, cond)
	if len(result) != 2 {
		t.Errorf("expected 2 results (dev OR prod), got %d", len(result))
	}
	if _, ok := result["x"]; !ok {
		t.Error("expected key 'x' (dev)")
	}
	if _, ok := result["y"]; !ok {
		t.Error("expected key 'y' (prod)")
	}
	if _, ok := result["z"]; ok {
		t.Error("key 'z' (staging) should be excluded")
	}
}

func TestFilterAttrMap_EmptyInput(t *testing.T) {
	cond := []map[string]attr.Value{{"env": types.StringValue("prod")}}
	result := filterAttrMap(map[string]attr.Value{}, cond)
	if len(result) != 0 {
		t.Errorf("expected empty result for empty input")
	}
}

// --- filterMapByPredicate ---

func TestFilterMapByPredicate_KeepAll(t *testing.T) {
	elems := map[string]attr.Value{
		"a": types.StringValue("x"),
		"b": types.StringValue("y"),
	}
	result := filterMapByPredicate(elems, func(v attr.Value) bool { return true })
	if len(result) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result))
	}
}

func TestFilterMapByPredicate_KeepNone(t *testing.T) {
	elems := map[string]attr.Value{
		"a": types.StringValue("x"),
	}
	result := filterMapByPredicate(elems, func(v attr.Value) bool { return false })
	if len(result) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result))
	}
}

func TestFilterMapByPredicate_NonNullPredicate(t *testing.T) {
	elems := map[string]attr.Value{
		"a": types.StringValue("x"),
		"b": types.StringNull(),
	}
	result := filterMapByPredicate(elems, func(v attr.Value) bool { return !v.IsNull() })
	if len(result) != 1 {
		t.Errorf("expected 1 entry, got %d", len(result))
	}
	if _, ok := result["a"]; !ok {
		t.Error("expected key 'a'")
	}
}

// --- filterListByPredicate ---

func TestFilterListByPredicate_KeepAll(t *testing.T) {
	elems := []attr.Value{types.StringValue("a"), types.StringValue("b")}
	result := filterListByPredicate(elems, func(v attr.Value) bool { return true })
	if len(result) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result))
	}
}

func TestFilterListByPredicate_KeepNone(t *testing.T) {
	elems := []attr.Value{types.StringValue("a"), types.StringValue("b")}
	result := filterListByPredicate(elems, func(v attr.Value) bool { return false })
	if len(result) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result))
	}
}

func TestFilterListByPredicate_NonNullPredicate(t *testing.T) {
	elems := []attr.Value{types.StringValue("a"), types.StringNull(), types.StringValue("b")}
	result := filterListByPredicate(elems, func(v attr.Value) bool { return !v.IsNull() })
	if len(result) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result))
	}
}

// --- parseConditions ---

func TestParseConditions_NilSlice(t *testing.T) {
	_, err := parseConditions(nil)
	if err == nil {
		t.Error("expected error for nil conditions")
	}
}

func TestParseConditions_EmptySlice(t *testing.T) {
	_, err := parseConditions([]types.Dynamic{})
	if err == nil {
		t.Error("expected error for empty conditions")
	}
}

func TestParseConditions_NullCondition(t *testing.T) {
	_, err := parseConditions([]types.Dynamic{types.DynamicNull()})
	if err == nil {
		t.Error("expected error when all conditions are null")
	}
}

func TestParseConditions_ValidObject(t *testing.T) {
	obj, _ := types.ObjectValue(
		map[string]attr.Type{"env": types.StringType},
		map[string]attr.Value{"env": types.StringValue("prod")},
	)
	conds, err := parseConditions([]types.Dynamic{types.DynamicValue(obj)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conds) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(conds))
	}
	v, ok := conds[0]["env"].(types.String)
	if !ok || v.ValueString() != "prod" {
		t.Errorf("expected env='prod', got %v", conds[0]["env"])
	}
}

func TestParseConditions_TwoObjects(t *testing.T) {
	obj1, _ := types.ObjectValue(
		map[string]attr.Type{"env": types.StringType},
		map[string]attr.Value{"env": types.StringValue("dev")},
	)
	obj2, _ := types.ObjectValue(
		map[string]attr.Type{"env": types.StringType},
		map[string]attr.Value{"env": types.StringValue("prod")},
	)
	conds, err := parseConditions([]types.Dynamic{types.DynamicValue(obj1), types.DynamicValue(obj2)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conds) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(conds))
	}
}

func TestParseConditions_NonObjectCondition(t *testing.T) {
	_, err := parseConditions([]types.Dynamic{types.DynamicValue(types.StringValue("bad"))})
	if err == nil {
		t.Error("expected error for non-object condition")
	}
}
