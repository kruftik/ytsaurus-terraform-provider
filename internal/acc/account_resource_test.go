package acc

import (
	"context"
	"fmt"
	"regexp"

	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"go.ytsaurus.tech/yt/go/yt"

	"terraform-provider-ytsaurus/internal/resource/account"
	"terraform-provider-ytsaurus/internal/resource/acl"
)

func TestAccountResourceCreateAndUpdate(t *testing.T) {
	ctx := context.TODO()

	resourceID := "testaccount"
	testAccountName := resourceID
	testAccountNameRenamed := "testaccount_renamed"
	testAccountYTCypressPath := fmt.Sprintf("//sys/accounts/%s", testAccountName)
	testAccountYTCypressPathRenamed := fmt.Sprintf("//sys/accounts/%s", testAccountNameRenamed)

	testNodeCount := int64(1000)
	testChunkCount := int64(1000)
	testTabletCount := int64(10)
	testTabletStaticMemory := int64(10000)
	testDefaultMedium := "default"
	testDefaultMediumSize := int64(1000000)
	testDefaultInheritAcl := true
	testInheritAcl := false

	testACL := []yt.ACE{
		{
			Action:          yt.ActionAllow,
			Subjects:        []string{"users"},
			Permissions:     []string{yt.PermissionUse},
			InheritanceMode: "object_and_descendants",
		},
	}

	configEmpty := account.AccountModel{}

	configWithoutResourceLimits := account.AccountModel{
		Name: types.StringValue(testAccountName),
	}

	configWithEmptyResourceLimits := account.AccountModel{
		Name: types.StringValue(testAccountName),
		ResourceLimits: types.ObjectValueMust(map[string]attr.Type{},
			map[string]attr.Value{},
		),
	}

	configCreate := account.AccountModel{
		Name: types.StringValue(testAccountName),
		ResourceLimits: types.ObjectValueMust(
			map[string]attr.Type{
				"node_count":            types.Int64Type,
				"chunk_count":           types.Int64Type,
				"tablet_count":          types.Int64Type,
				"tablet_static_memory":  types.Int64Type,
				"disk_space_per_medium": types.MapType{ElemType: types.Int64Type},
			},
			map[string]attr.Value{
				"node_count":           types.Int64Value(testNodeCount),
				"chunk_count":          types.Int64Value(testChunkCount),
				"tablet_count":         types.Int64Null(),
				"tablet_static_memory": types.Int64Null(),
				"disk_space_per_medium": types.MapValueMust(
					types.Int64Type,
					map[string]attr.Value{
						testDefaultMedium: types.Int64Value(testDefaultMediumSize),
					},
				),
			},
		),
	}

	configRename := account.AccountModel{
		Name: types.StringValue(testAccountNameRenamed),
		ResourceLimits: types.ObjectValueMust(
			map[string]attr.Type{
				"node_count":            types.Int64Type,
				"chunk_count":           types.Int64Type,
				"tablet_count":          types.Int64Type,
				"tablet_static_memory":  types.Int64Type,
				"disk_space_per_medium": types.MapType{ElemType: types.Int64Type},
			},
			map[string]attr.Value{
				"node_count":           types.Int64Value(testNodeCount),
				"chunk_count":          types.Int64Value(testChunkCount),
				"tablet_count":         types.Int64Null(),
				"tablet_static_memory": types.Int64Null(),
				"disk_space_per_medium": types.MapValueMust(
					types.Int64Type,
					map[string]attr.Value{
						testDefaultMedium: types.Int64Value(testDefaultMediumSize),
					},
				),
			},
		),
	}

	configUpdate := account.AccountModel{
		Name:       types.StringValue(testAccountName),
		InheritACL: types.BoolValue(testInheritAcl),
		ResourceLimits: types.ObjectValueMust(
			map[string]attr.Type{
				"node_count":            types.Int64Type,
				"chunk_count":           types.Int64Type,
				"tablet_count":          types.Int64Type,
				"tablet_static_memory":  types.Int64Type,
				"disk_space_per_medium": types.MapType{ElemType: types.Int64Type},
			},
			map[string]attr.Value{
				"node_count":           types.Int64Value(testNodeCount + 1),
				"chunk_count":          types.Int64Value(testChunkCount + 1),
				"tablet_count":         types.Int64Value(testTabletCount),
				"tablet_static_memory": types.Int64Value(testTabletStaticMemory),
				"disk_space_per_medium": types.MapValueMust(
					types.Int64Type,
					map[string]attr.Value{
						testDefaultMedium: types.Int64Value(testDefaultMediumSize + 1),
					},
				),
			},
		),
	}
	configUpdate.ACL, _ = acl.FlattenACL(ctx, testACL)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             accCheckYTsaurusObjectDestroyed(testAccountYTCypressPath),
		Steps: []resource.TestStep{
			{
				Config:      accGetYTLocalDockerProviderConfig() + accResourceYtsaurusAccountConfig(resourceID, configEmpty),
				ExpectError: regexp.MustCompile(`The argument "name" is required, but no definition was found.`),
			},
			{
				Config:      accGetYTLocalDockerProviderConfig() + accResourceYtsaurusAccountConfig(resourceID, configWithoutResourceLimits),
				ExpectError: regexp.MustCompile(`The argument "resource_limits" is required, but no definition was found.`),
			},
			{
				Config:      accGetYTLocalDockerProviderConfig() + accResourceYtsaurusAccountConfig(resourceID, configWithEmptyResourceLimits),
				ExpectError: regexp.MustCompile(`Inappropriate value for attribute "resource_limits": attributes\n"chunk_count", "disk_space_per_medium", and "node_count" are required.`),
			},
			{
				Config: accGetYTLocalDockerProviderConfig() + accResourceYtsaurusAccountConfig(resourceID, configCreate),
				Check: resource.ComposeAggregateTestCheckFunc(
					accCheckYTsaurusStringAttribute(testAccountYTCypressPath, "name", testAccountName),
					accCheckYTsaurusInt64Attribute(testAccountYTCypressPath, "resource_limits/node_count", testNodeCount),
					accCheckYTsaurusInt64Attribute(testAccountYTCypressPath, "resource_limits/chunk_count", testChunkCount),
					accCheckYTsaurusInt64Attribute(testAccountYTCypressPath, "resource_limits/disk_space_per_medium/default", testDefaultMediumSize),
					accCheckYTsaurusBoolAttribute(testAccountYTCypressPath, "inherit_acl", testDefaultInheritAcl),
				),
			},
			{
				Config: accGetYTLocalDockerProviderConfig() + accResourceYtsaurusAccountConfig(resourceID, configRename),
				Check: resource.ComposeAggregateTestCheckFunc(
					accCheckYTsaurusStringAttribute(testAccountYTCypressPathRenamed, "name", testAccountNameRenamed),
				),
			},
			{
				Config: accGetYTLocalDockerProviderConfig() + accResourceYtsaurusAccountConfig(resourceID, configUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					accCheckYTsaurusStringAttribute(testAccountYTCypressPath, "name", testAccountName),
					accCheckYTsaurusInt64Attribute(testAccountYTCypressPath, "resource_limits/node_count", testNodeCount+1),
					accCheckYTsaurusInt64Attribute(testAccountYTCypressPath, "resource_limits/chunk_count", testChunkCount+1),
					accCheckYTsaurusInt64Attribute(testAccountYTCypressPath, "resource_limits/disk_space_per_medium/default", testDefaultMediumSize+1),
					accCheckYTsaurusInt64Attribute(testAccountYTCypressPath, "resource_limits/tablet_count", testTabletCount),
					accCheckYTsaurusInt64Attribute(testAccountYTCypressPath, "resource_limits/tablet_static_memory", testTabletStaticMemory),
					accCheckYTsaurusACLAttribute(testAccountYTCypressPath, testACL),
					accCheckYTsaurusBoolAttribute(testAccountYTCypressPath, "inherit_acl", testInheritAcl),
				),
			},
		},
	})
}

func TestAccountResourceCreateWithAllOptions(t *testing.T) {

	resourceID := "testaccount"
	testAccountName := resourceID
	testAccountYTCypressPath := fmt.Sprintf("//sys/accounts/%s", testAccountName)

	testNodeCount := int64(1000)
	testChunkCount := int64(1000)
	testTabletCount := int64(10)
	testTabletStaticMemory := int64(10000)
	testDefaultMedium := "default"
	testDefaultMediumSize := int64(1000000)
	testInheritAcl := false

	testACL := []yt.ACE{
		{
			Action:          yt.ActionAllow,
			Subjects:        []string{"users"},
			Permissions:     []string{yt.PermissionUse},
			InheritanceMode: "object_and_descendants",
		},
	}

	configCreate := account.AccountModel{
		Name: types.StringValue(testAccountName),
		ResourceLimits: types.ObjectValueMust(
			map[string]attr.Type{
				"node_count":            types.Int64Type,
				"chunk_count":           types.Int64Type,
				"tablet_count":          types.Int64Type,
				"tablet_static_memory":  types.Int64Type,
				"disk_space_per_medium": types.MapType{ElemType: types.Int64Type},
			},
			map[string]attr.Value{
				"node_count":           types.Int64Value(testNodeCount),
				"chunk_count":          types.Int64Value(testChunkCount),
				"tablet_count":         types.Int64Value(testTabletCount),
				"tablet_static_memory": types.Int64Value(testTabletStaticMemory),
				"disk_space_per_medium": types.MapValueMust(
					types.Int64Type,
					map[string]attr.Value{
						testDefaultMedium: types.Int64Value(testDefaultMediumSize),
					},
				),
			},
		),
		InheritACL: types.BoolValue(testInheritAcl),
	}
	configCreate.ACL, _ = types.ListValueFrom(
		ctx,
		types.ObjectType{AttrTypes: map[string]attr.Type{
			"action":           types.StringType,
			"subjects":         types.SetType{ElemType: types.StringType},
			"permissions":      types.SetType{ElemType: types.StringType},
			"inheritance_mode": types.StringType,
		}},
		acl.ToACLModel(testACL),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             accCheckYTsaurusObjectDestroyed(testAccountYTCypressPath),
		Steps: []resource.TestStep{
			{
				Config: accGetYTLocalDockerProviderConfig() + accResourceYtsaurusAccountConfig(resourceID, configCreate),
				Check: resource.ComposeAggregateTestCheckFunc(
					accCheckYTsaurusStringAttribute(testAccountYTCypressPath, "name", testAccountName),
					accCheckYTsaurusInt64Attribute(testAccountYTCypressPath, "resource_limits/node_count", testNodeCount),
					accCheckYTsaurusInt64Attribute(testAccountYTCypressPath, "resource_limits/chunk_count", testChunkCount),
					accCheckYTsaurusInt64Attribute(testAccountYTCypressPath, "resource_limits/disk_space_per_medium/default", testDefaultMediumSize),
					accCheckYTsaurusInt64Attribute(testAccountYTCypressPath, "resource_limits/tablet_count", testTabletCount),
					accCheckYTsaurusInt64Attribute(testAccountYTCypressPath, "resource_limits/tablet_static_memory", testTabletStaticMemory),
					accCheckYTsaurusACLAttribute(testAccountYTCypressPath, testACL),
					accCheckYTsaurusBoolAttribute(testAccountYTCypressPath, "inherit_acl", testInheritAcl),
				),
			},
		},
	})
}

func TestAccountResourceCreateParentChild(t *testing.T) {

	resourceChildID := "testaccount"
	resourceParentID := "testaccount_father"

	testAccountChildName := resourceChildID
	testAccountParentName := resourceParentID

	testAccountChildYTCypressPath := fmt.Sprintf("//sys/accounts/%s", testAccountChildName)
	testAccountParentYTCypressPath := fmt.Sprintf("//sys/accounts/%s", testAccountParentName)

	testNodeCount := int64(1000)
	testChunkCount := int64(1000)
	testDefaultMedium := "default"
	testDefaultMediumSize := int64(1000000)
	testResourceLimits := types.ObjectValueMust(
		map[string]attr.Type{
			"node_count":            types.Int64Type,
			"chunk_count":           types.Int64Type,
			"tablet_count":          types.Int64Type,
			"tablet_static_memory":  types.Int64Type,
			"disk_space_per_medium": types.MapType{ElemType: types.Int64Type},
		},
		map[string]attr.Value{
			"node_count":           types.Int64Value(testNodeCount),
			"chunk_count":          types.Int64Value(testChunkCount),
			"tablet_count":         types.Int64Null(),
			"tablet_static_memory": types.Int64Null(),
			"disk_space_per_medium": types.MapValueMust(
				types.Int64Type,
				map[string]attr.Value{
					testDefaultMedium: types.Int64Value(testDefaultMediumSize),
				},
			),
		},
	)

	configParent := account.AccountModel{
		Name:           types.StringValue(testAccountParentName),
		ResourceLimits: testResourceLimits,
	}

	configChild := account.AccountModel{
		Name:           types.StringValue(testAccountChildName),
		ParentName:     types.StringValue(fmt.Sprintf("ytsaurus_account.%s.name", resourceParentID)),
		ResourceLimits: testResourceLimits,
	}

	configRemoveParentName := account.AccountModel{
		Name:           types.StringValue(testAccountChildName),
		ResourceLimits: testResourceLimits,
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             accCheckYTsaurusObjectDestroyed(testAccountParentYTCypressPath),
		Steps: []resource.TestStep{
			{
				Config: accGetYTLocalDockerProviderConfig() + accResourceYtsaurusAccountConfig(resourceParentID, configParent) + accResourceYtsaurusAccountConfig(resourceChildID, configChild),
				Check: resource.ComposeAggregateTestCheckFunc(
					accCheckYTsaurusStringAttribute(testAccountParentYTCypressPath, "name", testAccountParentName),
					accCheckYTsaurusStringAttribute(testAccountChildYTCypressPath, "name", testAccountChildName),
					accCheckYTsaurusStringAttribute(testAccountChildYTCypressPath, "parent_name", testAccountParentName),
				),
			},
			{
				Config: accGetYTLocalDockerProviderConfig() + accResourceYtsaurusAccountConfig(resourceParentID, configParent) + accResourceYtsaurusAccountConfig(resourceChildID, configRemoveParentName),
				Check: resource.ComposeAggregateTestCheckFunc(
					accCheckYTsaurusStringAttribute(testAccountParentYTCypressPath, "name", testAccountParentName),
					accCheckYTsaurusStringAttribute(testAccountChildYTCypressPath, "name", testAccountChildName),
					accCheckYTsaurusStringAttribute(testAccountChildYTCypressPath, "parent_name", "root"),
				),
			},
		},
	})
}

func accResourceYtsaurusAccountConfig(id string, m account.AccountModel) string {

	config := fmt.Sprintf(`
	resource "ytsaurus_account" %q {`, id)

	if !m.Name.IsNull() {
		config += fmt.Sprintf(`
		name = %q`, m.Name.ValueString())
	}

	if !m.ParentName.IsNull() {
		config += fmt.Sprintf(`
		parent_name = %s`, m.ParentName.ValueString())
	}

	if !m.InheritACL.IsNull() {
		config += fmt.Sprintf(`
		inherit_acl = %t`, m.InheritACL.ValueBool())
	}

	if !m.ResourceLimits.IsNull() {
		var resourceLimitsModel account.AccountResourceLimitsModel
		m.ResourceLimits.As(ctx, &resourceLimitsModel, basetypes.ObjectAsOptions{})

		config += `
		resource_limits = {`

		if !resourceLimitsModel.NodeCount.IsNull() {
			config += fmt.Sprintf(`
			node_count = %d`, resourceLimitsModel.NodeCount.ValueInt64())
		}

		if !resourceLimitsModel.ChunkCount.IsNull() {
			config += fmt.Sprintf(`
			chunk_count = %d`, resourceLimitsModel.ChunkCount.ValueInt64())
		}

		if !resourceLimitsModel.TabletCount.IsNull() {
			config += fmt.Sprintf(`
			tablet_count = %d`, resourceLimitsModel.TabletCount.ValueInt64())
		}

		if !resourceLimitsModel.TabletStaticMemory.IsNull() {
			config += fmt.Sprintf(`
			tablet_static_memory = %d`, resourceLimitsModel.TabletStaticMemory.ValueInt64())
		}

		if len(resourceLimitsModel.DiskSpacePerMedium.Elements()) > 0 {
			config += `
			disk_space_per_medium = {`

			for k, v := range resourceLimitsModel.DiskSpacePerMedium.Elements() {

				config += fmt.Sprintf(`
				%q = %d`, k, v.(types.Int64).ValueInt64())
			}

			config += `
			}`
		}

		config += `
		}`
	}

	ctx := context.TODO()
	acl, _ := acl.ExpandACL(ctx, m.ACL)
	config += accAddACLConfig(acl)

	config += `
	}`

	return config
}
