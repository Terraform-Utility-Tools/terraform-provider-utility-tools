// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure UtilityToolsProvider satisfies various provider interfaces.
var (
	_ provider.ProviderWithFunctions = &UtilityToolsProvider{}
)

// UtilityToolsProvider defines the provider implementation.
type UtilityToolsProvider struct {
	version          string
	defaultSeparator string
}

// UtilityToolsProviderModel describes the provider data model.
type UtilityToolsProviderModel struct {
	Separator types.String `tfsdk:"separator"`
}

func (p *UtilityToolsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "util"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *UtilityToolsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"separator": schema.StringAttribute{
				Optional:    true,
				Description: "Default separator used by collapse, expand, and combine. Defaults to '/'.",
			},
		},
	}
}

func (p *UtilityToolsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config UtilityToolsProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Separator.IsNull() && !config.Separator.IsUnknown() {
		p.defaultSeparator = config.Separator.ValueString()
	} else {
		p.defaultSeparator = "/"
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *UtilityToolsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// Resources defines the resources implemented in the provider.
func (p *UtilityToolsProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}

func (p *UtilityToolsProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		func() function.Function { return NewCollapseFunction(p.defaultSeparator) },
		func() function.Function { return NewCombineFunction(p.defaultSeparator) },
		NewCompactFunction,
		func() function.Function { return NewExpandFunction(p.defaultSeparator) },
		NewFilterFunction,
		NewIsNullFunction,
		NewIsNotNullFunction,
		NewMinimalFunction,
		NewOmitFunction,
		NewPickFunction,
		NewTrimExtFunction,
		func() function.Function { return NewNestedCombineFunction(p.defaultSeparator) },
		NewNestedMergeFunction,
		func() function.Function { return NewNestedCompactFunction(p.defaultSeparator) },
		func() function.Function { return NewNestedFilterFunction(p.defaultSeparator) },
		func() function.Function { return NewNestedMinimalFunction(p.defaultSeparator) },
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &UtilityToolsProvider{
			version:          version,
			defaultSeparator: "/",
		}
	}
}
