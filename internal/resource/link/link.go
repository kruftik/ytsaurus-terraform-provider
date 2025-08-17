package link

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"go.ytsaurus.tech/yt/go/ypath"
	"go.ytsaurus.tech/yt/go/yt"

	"terraform-provider-ytsaurus/internal/ytsaurus"
)

type linkResource struct {
	client yt.Client
}

var (
	_ resource.Resource                = &linkResource{}
	_ resource.ResourceWithConfigure   = &linkResource{}
	_ resource.ResourceWithImportState = &linkResource{}
)

type LinkModel struct {
	ID         types.String `tfsdk:"id"`
	Path       types.String `tfsdk:"path"`
	TargetPath types.String `tfsdk:"target_path"`
}

func toLinkModel(m ytsaurus.Link) LinkModel {
	return LinkModel{
		ID:         types.StringValue(m.ID),
		Path:       types.StringValue(m.Path),
		TargetPath: types.StringValue(m.TargetPath),
	}
}

func toYTsaurusLink(m LinkModel) ytsaurus.Link {
	return ytsaurus.Link{
		ID:         m.ID.ValueString(),
		Path:       m.Path.ValueString(),
		TargetPath: m.TargetPath.ValueString(),
	}
}

func NewLinkResource() resource.Resource {
	return &linkResource{}
}

func (r *linkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_link"
}

func (r *linkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ObjectID in the YTsaurus cluster, can be found in object's @id attribute.",
			},
			"path": schema.StringAttribute{
				Required:    true,
				Description: "Link's absolute path.",
			},
			"target_path": schema.StringAttribute{
				Required:    true,
				Description: "Link's target path.",
			},
		},
	}
}

func (r *linkResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(yt.Client)
}

func (r *linkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LinkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ytLink := toYTsaurusLink(plan)

	createOptions := &yt.LinkNodeOptions{
		Attributes: map[string]interface{}{
			"terraform_resource": true,
		},
	}

	p := ypath.Path(ytLink.Path)
	target := ypath.Path(ytLink.TargetPath)
	id, err := r.client.LinkNode(ctx, target, p, createOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating link",
			fmt.Sprintf(
				"Could not create link %q, unexpected error: %q",
				p.String(),
				err.Error(),
			),
		)
		return
	}
	ytLink.ID = id.String()

	resp.Diagnostics.Append(resp.State.Set(ctx, toLinkModel(ytLink))...)
}

func (r *linkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var objectID string
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &objectID)...)
	if resp.Diagnostics.HasError() {
		return
	}
	objectID = objectID + "&"

	var link ytsaurus.Link

	if err := ytsaurus.GetObjectByID(ctx, r.client, objectID, &link); err != nil {
		resp.Diagnostics.AddError(
			"Error reading link",
			fmt.Sprintf(
				"Could not read link with id %q, unexpected error: %q",
				objectID,
				err.Error(),
			),
		)
		return
	}

	p := ypath.Path(fmt.Sprintf("#%s&", objectID)).Attr("path")
	if err := r.client.GetNode(ctx, p, &link.Path, nil); err != nil {
		resp.Diagnostics.AddError(
			"Error reading link @path attribute",
			fmt.Sprintf(
				"Could not read link %q, unexpected error: %q",
				p.String(),
				err.Error(),
			),
		)
		return
	}

	state := toLinkModel(link)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *linkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LinkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var state LinkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Path.Equal(state.Path) {
		resp.Diagnostics.AddError(
			"Error updating link attributes",
			"Builtin attribute 'path' cannot be updated",
		)
		return
	}

	if !plan.TargetPath.Equal(state.TargetPath) {
		resp.Diagnostics.AddError(
			"Error updating link attributes",
			"Builtin attribute 'target_path' cannot be updated, please delete and recreate the link",
		)
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *linkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LinkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := ypath.Path(state.Path.ValueString() + "&")
	if err := r.client.RemoveNode(ctx, p, nil); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting link",
			fmt.Sprintf(
				"Could not delete link %q, unexpected error: %q",
				p.String(),
				err.Error(),
			),
		)
		return
	}
}

func (r *linkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
