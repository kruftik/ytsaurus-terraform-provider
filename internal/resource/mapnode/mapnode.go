package mapnode

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"go.ytsaurus.tech/yt/go/ypath"
	"go.ytsaurus.tech/yt/go/yt"

	"terraform-provider-ytsaurus/internal/resource/acl"
	"terraform-provider-ytsaurus/internal/ytsaurus"
)

type mapNodeResource struct {
	client yt.Client
}

var (
	_ resource.Resource                = &mapNodeResource{}
	_ resource.ResourceWithConfigure   = &mapNodeResource{}
	_ resource.ResourceWithImportState = &mapNodeResource{}
)

type MapNodeModel struct {
	ID         types.String `tfsdk:"id"`
	Path       types.String `tfsdk:"path"`
	Account    types.String `tfsdk:"account"`
	InheritACL types.Bool   `tfsdk:"inherit_acl"`
	ACL        types.List   `tfsdk:"acl"`
	Opaque     types.Bool   `tfsdk:"opaque"`
}

// flattenMapNode performs YT to Terraform conversion for MapNode resources.
// Direction: YT -> Terraform
func flattenMapNode(ctx context.Context, mapNode ytsaurus.MapNode) (MapNodeModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	acl, diags := acl.FlattenACL(ctx, mapNode.ACL)

	return MapNodeModel{
		ID:         types.StringValue(mapNode.ID),
		Path:       types.StringValue(mapNode.Path),
		Account:    types.StringValue(mapNode.Account),
		InheritACL: types.BoolValue(mapNode.InheritACL),
		Opaque:     types.BoolValue(mapNode.Opaque),
		ACL:        acl,
	}, diags
}

// func toMapNodeModel(m ytsaurus.MapNode) MapNodeModel {
// 	return MapNodeModel{
// 		ID:         types.StringValue(m.ID),
// 		Path:       types.StringValue(m.Path),
// 		Account:    types.StringValue(m.Account),
// 		InheritACL: types.BoolValue(m.InheritACL),
// 		Opaque:     types.BoolValue(m.Opaque),
// 		ACL:        acl.ToACLModel(m.ACL),
// 	}
// }

// expandAccount performs Terraform to YT conversion for MapNode resources.
// Direction: Terraform -> YT
func expandMapNode(ctx context.Context, mapNodeModel MapNodeModel) (ytsaurus.MapNode, diag.Diagnostics) {
	acl, diags := acl.ExpandACL(ctx, mapNodeModel.ACL)
	return ytsaurus.MapNode{
		Path:       mapNodeModel.Path.ValueString(),
		Account:    mapNodeModel.Account.ValueString(),
		InheritACL: mapNodeModel.InheritACL.ValueBool(),
		Opaque:     mapNodeModel.Opaque.ValueBool(),
		ACL:        acl,
	}, diags
}

func NewMapNodeResource() resource.Resource {
	return &mapNodeResource{}
}

func (r *mapNodeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_map_node"
}

func (r *mapNodeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ObjectID in the YTsaurus cluster, can be found in object's @id attribute.",
			},
			"path": schema.StringAttribute{
				Required:    true,
				Description: "Node absolute path.",
			},
			"account": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Account used to keep track of the resources being used by a specific node.",
			},
			"inherit_acl": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Enable or disable ACL inheritance from an object's parents.",
			},
			"acl": schema.ListNestedAttribute{
				Optional:     true,
				NestedObject: acl.ACLSchema,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Description: "A list of ACE records. More information: https://ytsaurus.tech/docs/en/user-guide/storage/access-control.",
			},
			"opaque": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Defines the 'transparency' of an object. Opaque will not show the contents of the object for implicit get requests if set to True.. False by default.",
			},
		},
	}
}

func (r *mapNodeResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(yt.Client)
}

func (r *mapNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MapNodeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ytMapNode, diags := expandMapNode(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createOptions := &yt.CreateNodeOptions{
		Attributes: map[string]interface{}{
			"opaque":             ytMapNode.Opaque,
			"terraform_resource": true,
		},
	}

	if !plan.ACL.IsNull() {
		createOptions.Attributes["acl"] = ytMapNode.ACL
	}
	if !plan.InheritACL.ValueBool() {
		createOptions.Attributes["inherit_acl"] = ytMapNode.InheritACL
	}
	if !plan.Account.IsNull() && !plan.Account.IsUnknown() {
		createOptions.Attributes["account"] = ytMapNode.Account
	}

	p := ypath.Path(ytMapNode.Path)
	id, err := r.client.CreateNode(ctx, p, yt.NodeMap, createOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating map_node",
			fmt.Sprintf(
				"Could not create map_node %q, unexpected error: %q",
				p.String(),
				err.Error(),
			),
		)
		return
	}
	ytMapNode.ID = id.String()

	if ytMapNode.Account == "" {
		if err := r.client.GetNode(ctx, p.Attr("account"), &ytMapNode.Account, nil); err != nil {
			resp.Diagnostics.AddError(
				"Error creating map_node",
				fmt.Sprintf(
					"Could not read 'account' attribute, unexpected error: %q",
					err.Error(),
				),
			)
			return
		}
	}

	state, diags := flattenMapNode(ctx, ytMapNode)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *mapNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var objectID string
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &objectID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var mapNode ytsaurus.MapNode
	if err := ytsaurus.GetObjectByID(ctx, r.client, objectID, &mapNode); err != nil {
		resp.Diagnostics.AddError(
			"Error reading map_node",
			fmt.Sprintf(
				"Could not read map_node with id %q, unexpected error: %q",
				objectID,
				err.Error(),
			),
		)
		return
	}

	p := ypath.Path(fmt.Sprintf("#%s", objectID)).Attr("path")
	if err := r.client.GetNode(ctx, p, &mapNode.Path, nil); err != nil {
		resp.Diagnostics.AddError(
			"Error reading map_node @path attribute",
			fmt.Sprintf(
				"Could not read map_node %q, unexpected error: %q",
				p.String(),
				err.Error(),
			),
		)
		return
	}

	state, diags := flattenMapNode(ctx, mapNode)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *mapNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MapNodeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var state MapNodeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Path.Equal(state.Path) {
		resp.Diagnostics.AddError(
			"Error updating map_node attributes",
			"Builtin attribute 'path' cannot be updated",
		)
		return
	}

	ytMapNode, diags := expandMapNode(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := ypath.Path(fmt.Sprintf("#%s", state.ID.ValueString()))
	attributeUpdates := map[string]interface{}{
		"account": ytMapNode.Account,
		"opaque":  ytMapNode.Opaque,
	}

	if !plan.ACL.Equal(state.ACL) {
		attributeUpdates["acl"] = ytMapNode.ACL
	}
	if !plan.InheritACL.Equal(state.InheritACL) {
		attributeUpdates["inherit_acl"] = ytMapNode.InheritACL
	}

	for k, v := range attributeUpdates {
		if err := r.client.SetNode(ctx, p.Attr(k), v, nil); err != nil {
			resp.Diagnostics.AddError(
				"Error updating map_node attributes",
				fmt.Sprintf(
					"Could not set node %q to '%v', unexpected error: %q",
					p.Attr(k).String(),
					v,
					err.Error(),
				),
			)
			return
		}
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *mapNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MapNodeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := ypath.Path(state.Path.ValueString())
	if err := r.client.RemoveNode(ctx, p, nil); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting map_node",
			fmt.Sprintf(
				"Could not delete map_node %q, unexpected error: %q",
				p.String(),
				err.Error(),
			),
		)
		return
	}
}

func (r *mapNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
