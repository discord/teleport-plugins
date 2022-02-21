// Code generated by _gen/main.go DO NOT EDIT
/*
Copyright 2015-2021 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/gravitational/teleport-plugins/terraform/tfschema"
	apitypes "github.com/gravitational/teleport/api/types"
)

// dataSourceTeleportGithubConnectorType is the data source metadata type
type dataSourceTeleportGithubConnectorType struct{}

// dataSourceTeleportGithubConnector is the resource
type dataSourceTeleportGithubConnector struct {
	p provider
}

// GetSchema returns the data source schema
func (r dataSourceTeleportGithubConnectorType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfschema.GenSchemaGithubConnectorV3(ctx)
}

// NewDataSource creates the empty data source
func (r dataSourceTeleportGithubConnectorType) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return dataSourceTeleportGithubConnector{
		p: *(p.(*provider)),
	}, nil
}

// Read reads teleport GithubConnector
func (r dataSourceTeleportGithubConnector) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var id types.String
	diags := req.Config.GetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("metadata").WithAttributeName("name"), &id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	githubConnectorI, err := r.p.Client.GetGithubConnector(ctx, id.Value, true)
	if err != nil {
		resp.Diagnostics.AddError("Error reading GithubConnector", err.Error())
		return
	}

    var state types.Object
	githubConnector := githubConnectorI.(*apitypes.GithubConnectorV3)
	diags = tfschema.CopyGithubConnectorV3ToTerraform(ctx, *githubConnector, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}