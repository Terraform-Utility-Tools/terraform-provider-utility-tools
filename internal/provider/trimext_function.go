// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"path"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

func trimExtString(s string) string {
	base := path.Base(s)
	ext := path.Ext(base)
	if ext == base {
		// Hidden file with no further extension (e.g. ".hidden") — leave unchanged.
		return s
	}
	return s[:len(s)-len(ext)]
}

var _ function.Function = &TrimExtFunction{}

type TrimExtFunction struct{}

func NewTrimExtFunction() function.Function {
	return &TrimExtFunction{}
}

func (f *TrimExtFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "trimext"
}

func (f *TrimExtFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Remove the file extension from a path",
		Description: "Returns the path with its last extension removed. If the path has no extension it is returned unchanged.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "path",
				Description: "The file path or name to strip the extension from.",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *TrimExtFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input))
	if resp.Error != nil {
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, trimExtString(input)))
}
