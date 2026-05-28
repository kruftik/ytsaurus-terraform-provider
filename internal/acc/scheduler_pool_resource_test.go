package acc

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"go.ytsaurus.tech/yt/go/yt"

	"terraform-provider-ytsaurus/internal/resource/acl"
	"terraform-provider-ytsaurus/internal/resource/schedulerpool"
)

func TestSchedulerPoolResourceMisconfigurations(t *testing.T) {
	resourceID := "fakepool"
	testSchedulerPoolName := resourceID
	testSchedulerPoolYTCypressPath := fmt.Sprintf("//sys/pool_trees/default/%s", testSchedulerPoolName)
	testPoolTree := "default"
	testMaxOperationCount := int64(10)
	testMaxRunningOperationCount := int64(10)

	configWithoutPoolTree := schedulerpool.SchedulerPoolModel{
		Name: types.StringValue(testSchedulerPoolName),
	}

	configMaxRunningOperationCountShouldBePositive := schedulerpool.SchedulerPoolModel{
		Name:                     types.StringValue(testSchedulerPoolName),
		PoolTree:                 types.StringValue(testPoolTree),
		MaxRunningOperationCount: types.Int64Value(0),
	}

	configMaxOperationCountShouldBePositive := schedulerpool.SchedulerPoolModel{
		Name:              types.StringValue(testSchedulerPoolName),
		PoolTree:          types.StringValue(testPoolTree),
		MaxOperationCount: types.Int64Value(0),
	}

	configParentNameNotEmpty := schedulerpool.SchedulerPoolModel{
		Name:       types.StringValue(testSchedulerPoolName),
		PoolTree:   types.StringValue(testPoolTree),
		ParentName: types.StringValue(`""`),
	}

	configWeightShouldBePositive := schedulerpool.SchedulerPoolModel{
		Name:     types.StringValue(testSchedulerPoolName),
		PoolTree: types.StringValue(testPoolTree),
		Weight:   types.Float64Value(0),
	}

	configModeOneOf := schedulerpool.SchedulerPoolModel{
		Name:     types.StringValue(testSchedulerPoolName),
		PoolTree: types.StringValue(testPoolTree),
		Mode:     types.StringValue("fake"),
	}

	configMaxOperationCountGreaterThenMaxRunningOperationCount := schedulerpool.SchedulerPoolModel{
		Name:                     types.StringValue(testSchedulerPoolName),
		PoolTree:                 types.StringValue(testPoolTree),
		MaxRunningOperationCount: types.Int64Value(testMaxRunningOperationCount),
		MaxOperationCount:        types.Int64Value(testMaxOperationCount - 1),
	}

	resourcesAttrTypes := map[string]attr.Type{
		"cpu":    types.Int64Type,
		"memory": types.Int64Type,
	}

	emptyResourceLimits, _ := types.ObjectValueFrom(ctx, resourcesAttrTypes, schedulerpool.SchedulerPoolResourcesModel{})

	configResourceLimitsNotEmpty := schedulerpool.SchedulerPoolModel{
		Name:           types.StringValue(testSchedulerPoolName),
		PoolTree:       types.StringValue(testPoolTree),
		ResourceLimits: emptyResourceLimits,
	}

	emptyStrongGuaranteeResources, _ := types.ObjectValueFrom(ctx, resourcesAttrTypes, schedulerpool.SchedulerPoolResourcesModel{})

	configStrongGuaranteeResourcesNotEmpty := schedulerpool.SchedulerPoolModel{
		Name:                     types.StringValue(testSchedulerPoolName),
		PoolTree:                 types.StringValue(testPoolTree),
		StrongGuaranteeResources: emptyStrongGuaranteeResources,
	}

	integralGuarantees, _ := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"guarantee_type": types.StringType,
		"resource_flow": types.ObjectType{
			AttrTypes: resourcesAttrTypes,
		},
		"burst_guarantee_resources": types.ObjectType{
			AttrTypes: resourcesAttrTypes,
		},
	}, schedulerpool.SchedulerPoolIntegralGuaranteesModel{
		GuaranteeType:           types.StringValue("fake"),
		ResourceFlow:            types.ObjectNull(resourcesAttrTypes),
		BurstGuaranteeResources: types.ObjectNull(resourcesAttrTypes),
	})

	configIntegralGuaranteesGuaranteeTypeOneOf := schedulerpool.SchedulerPoolModel{
		Name:               types.StringValue(testSchedulerPoolName),
		PoolTree:           types.StringValue(testPoolTree),
		IntegralGuarantees: integralGuarantees,
	}

	emptyIntegralGuarantees, _ := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"guarantee_type": types.StringType,
		"resource_flow": types.ObjectType{
			AttrTypes: resourcesAttrTypes,
		},
		"burst_guarantee_resources": types.ObjectType{
			AttrTypes: resourcesAttrTypes,
		},
	}, schedulerpool.SchedulerPoolIntegralGuaranteesModel{
		GuaranteeType:           types.StringNull(),
		ResourceFlow:            types.ObjectNull(resourcesAttrTypes),
		BurstGuaranteeResources: types.ObjectNull(resourcesAttrTypes),
	})

	configIntegralGuaranteesNotEmpty := schedulerpool.SchedulerPoolModel{
		Name:               types.StringValue(testSchedulerPoolName),
		PoolTree:           types.StringValue(testPoolTree),
		IntegralGuarantees: emptyIntegralGuarantees,
	}

	integralGuaranteesWithResourceFlow, _ := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"guarantee_type": types.StringType,
		"resource_flow": types.ObjectType{
			AttrTypes: resourcesAttrTypes,
		},
		"burst_guarantee_resources": types.ObjectType{
			AttrTypes: resourcesAttrTypes,
		},
	}, schedulerpool.SchedulerPoolIntegralGuaranteesModel{
		GuaranteeType:           types.StringNull(),
		ResourceFlow:            emptyResourceLimits,
		BurstGuaranteeResources: types.ObjectNull(resourcesAttrTypes),
	})

	configIntegralGuaranteesResourceFlowNotEmpty := schedulerpool.SchedulerPoolModel{
		Name:               types.StringValue(testSchedulerPoolName),
		PoolTree:           types.StringValue(testPoolTree),
		IntegralGuarantees: integralGuaranteesWithResourceFlow,
	}

	integralGuaranteesWithBurstGuaranteeResources, _ := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"guarantee_type": types.StringType,
		"resource_flow": types.ObjectType{
			AttrTypes: resourcesAttrTypes,
		},
		"burst_guarantee_resources": types.ObjectType{
			AttrTypes: resourcesAttrTypes,
		},
	}, schedulerpool.SchedulerPoolIntegralGuaranteesModel{
		GuaranteeType:           types.StringNull(),
		ResourceFlow:            types.ObjectNull(resourcesAttrTypes),
		BurstGuaranteeResources: emptyResourceLimits,
	})

	configIntegralGuaranteesBurstGuaranteeResourcesNotEmpty := schedulerpool.SchedulerPoolModel{
		Name:               types.StringValue(testSchedulerPoolName),
		PoolTree:           types.StringValue(testPoolTree),
		IntegralGuarantees: integralGuaranteesWithBurstGuaranteeResources,
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             accCheckYTsaurusObjectDestroyed(testSchedulerPoolYTCypressPath),
		Steps: []resource.TestStep{
			{
				Config:      accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configWithoutPoolTree),
				ExpectError: regexp.MustCompile("The argument \"pool_tree\" is required, but no definition was found."),
			},
			{
				Config:      accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configMaxRunningOperationCountShouldBePositive),
				ExpectError: regexp.MustCompile("Attribute max_running_operation_count value must be at least 1, got: 0"),
			},
			{
				Config:      accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configMaxOperationCountShouldBePositive),
				ExpectError: regexp.MustCompile("Attribute max_operation_count value must be at least 1, got: 0"),
			},
			{
				Config:      accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configParentNameNotEmpty),
				ExpectError: regexp.MustCompile("Attribute parent_name string length must be at least 1, got: 0"),
			},
			{
				Config:      accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configWeightShouldBePositive),
				ExpectError: regexp.MustCompile("Attribute weight value must be at least 1.000000, got: 0.000000"),
			},
			{
				Config:      accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configModeOneOf),
				ExpectError: regexp.MustCompile(`Attribute mode value must be one of: \["\\"fair_share\\"" "\\"fifo\\""\], got:`),
			},
			{
				Config:      accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configMaxOperationCountGreaterThenMaxRunningOperationCount),
				ExpectError: regexp.MustCompile("but 9 < 10"),
			},
			{
				Config:      accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configResourceLimitsNotEmpty),
				ExpectError: regexp.MustCompile("Attribute \"resource_limits.(cpu|memory)\" must be specified when \"resource_limits\""),
			},
			{
				Config:      accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configStrongGuaranteeResourcesNotEmpty),
				ExpectError: regexp.MustCompile("Attribute \"strong_guarantee_resources.(cpu|memory)\" must be specified when\n? ?\"strong_guarantee_resources\""),
			},
			{
				Config:      accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configIntegralGuaranteesGuaranteeTypeOneOf),
				ExpectError: regexp.MustCompile(`Attribute integral_guarantees.guarantee_type value must be one of:\n\["\\"burst\\"" "\\"relaxed\\"" "\\"none\\""\]`),
			},
			{
				Config:      accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configIntegralGuaranteesNotEmpty),
				ExpectError: regexp.MustCompile("Attribute \"integral_guarantees.(cpu|memory|resource_flow)\" must be specified\n? ?when\n? ?\"integral_guarantees\""),
			},
			{
				Config:      accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configIntegralGuaranteesResourceFlowNotEmpty),
				ExpectError: regexp.MustCompile("Attribute \"integral_guarantees.resource_flow.(cpu|memory)\" must be specified when\n\"integral_guarantees.resource_flow\" is specified"),
			},
			{
				Config:      accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configIntegralGuaranteesBurstGuaranteeResourcesNotEmpty),
				ExpectError: regexp.MustCompile("Attribute \"integral_guarantees.burst_guarantee_resources.(cpu|memory)\" must be\nspecified when \"integral_guarantees.burst_guarantee_resources\" is specified"),
			},
		},
	})
}

func TestSchedulerPoolResourceCreateAndUpdate(t *testing.T) {
	resourceID := "fakepool"
	testSchedulerPoolName := resourceID
	testSchedulerPoolYTCypressPath := fmt.Sprintf("//sys/pool_trees/default/%s", testSchedulerPoolName)
	testPoolTree := "default"
	testForbidImmediateOperations := true
	testMaxOperationCount := int64(10)
	testMaxRunningOperationCount := int64(10)
	testSchedulerPoolResourcesModelCPU := int64(1)
	testSchedulerPoolResourcesModelMemory := int64(1 * 1024 * 1024)
	testGuaranteeTypeBurst := "burst"
	testGuaranteeTypeRelaxed := "relaxed"

	testACL := []yt.ACE{
		{
			Action:          yt.ActionAllow,
			Subjects:        []string{"users"},
			Permissions:     []string{yt.PermissionUse},
			InheritanceMode: "object_and_descendants",
		},
	}
	testACLModel, _ := acl.FlattenACL(ctx, testACL)

	configCreateWithMinimalOptions := schedulerpool.SchedulerPoolModel{
		Name:     types.StringValue(testSchedulerPoolName),
		PoolTree: types.StringValue(testPoolTree),
	}

	resourceLimits, _ := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"cpu":    types.Int64Type,
		"memory": types.Int64Type,
	}, schedulerpool.SchedulerPoolResourcesModel{
		CPU:    types.Int64Value(testSchedulerPoolResourcesModelCPU),
		Memory: types.Int64Value(testSchedulerPoolResourcesModelMemory),
	})

	integralGuarantees, _ := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"guarantee_type": types.StringType,
		"resource_flow": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"cpu":    types.Int64Type,
				"memory": types.Int64Type,
			},
		},
		"burst_guarantee_resources": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"cpu":    types.Int64Type,
				"memory": types.Int64Type,
			},
		},
	}, schedulerpool.SchedulerPoolIntegralGuaranteesModel{
		GuaranteeType: types.StringValue(testGuaranteeTypeBurst),
		ResourceFlow: types.ObjectNull(map[string]attr.Type{
			"cpu":    types.Int64Type,
			"memory": types.Int64Type,
		}),
		BurstGuaranteeResources: resourceLimits,
	})

	integralGuaranteesResourceFlow, _ := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"guarantee_type": types.StringType,
		"resource_flow": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"cpu":    types.Int64Type,
				"memory": types.Int64Type,
			},
		},
		"burst_guarantee_resources": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"cpu":    types.Int64Type,
				"memory": types.Int64Type,
			},
		},
	}, schedulerpool.SchedulerPoolIntegralGuaranteesModel{
		GuaranteeType: types.StringValue(testGuaranteeTypeRelaxed),
		ResourceFlow:  resourceLimits,
		BurstGuaranteeResources: types.ObjectNull(map[string]attr.Type{
			"cpu":    types.Int64Type,
			"memory": types.Int64Type,
		}),
	})

	configUpdateAllAttributes := schedulerpool.SchedulerPoolModel{
		Name:                      types.StringValue(testSchedulerPoolName),
		PoolTree:                  types.StringValue(testPoolTree),
		ACL:                       testACLModel,
		MaxRunningOperationCount:  types.Int64Value(testMaxRunningOperationCount),
		MaxOperationCount:         types.Int64Value(testMaxOperationCount),
		ForbidImmediateOperations: types.BoolValue(testForbidImmediateOperations),
		ResourceLimits:            resourceLimits,
		StrongGuaranteeResources:  resourceLimits,
	}

	configIntegralGuaranteesBurst := schedulerpool.SchedulerPoolModel{
		Name:                      types.StringValue(testSchedulerPoolName),
		PoolTree:                  types.StringValue(testPoolTree),
		ACL:                       testACLModel,
		MaxRunningOperationCount:  types.Int64Value(testMaxRunningOperationCount),
		MaxOperationCount:         types.Int64Value(testMaxOperationCount),
		ForbidImmediateOperations: types.BoolValue(testForbidImmediateOperations),
		ResourceLimits:            resourceLimits,
		StrongGuaranteeResources:  resourceLimits,
		IntegralGuarantees:        integralGuarantees,
	}

	configIntegralGuaranteesRelaxed := schedulerpool.SchedulerPoolModel{
		Name:                      types.StringValue(testSchedulerPoolName),
		PoolTree:                  types.StringValue(testPoolTree),
		ACL:                       testACLModel,
		MaxRunningOperationCount:  types.Int64Value(testMaxRunningOperationCount),
		MaxOperationCount:         types.Int64Value(testMaxOperationCount),
		ForbidImmediateOperations: types.BoolValue(testForbidImmediateOperations),
		ResourceLimits:            resourceLimits,
		StrongGuaranteeResources:  resourceLimits,
		IntegralGuarantees:        integralGuaranteesResourceFlow,
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             accCheckYTsaurusObjectDestroyed(testSchedulerPoolYTCypressPath),
		Steps: []resource.TestStep{
			{
				Config: accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configCreateWithMinimalOptions),
				Check: resource.ComposeAggregateTestCheckFunc(
					accCheckYTsaurusStringAttribute(testSchedulerPoolYTCypressPath, "name", testSchedulerPoolName),
				),
			},
			{
				Config: accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configUpdateAllAttributes),
				Check: resource.ComposeAggregateTestCheckFunc(
					accCheckYTsaurusStringAttribute(testSchedulerPoolYTCypressPath, "name", testSchedulerPoolName),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "max_operation_count", testMaxOperationCount),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "max_running_operation_count", testMaxRunningOperationCount),
					accCheckYTsaurusACLAttribute(testSchedulerPoolYTCypressPath, testACL),
					accCheckYTsaurusBoolAttribute(testSchedulerPoolYTCypressPath, "forbid_immediate_operations", testForbidImmediateOperations),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "resource_limits/cpu", testSchedulerPoolResourcesModelCPU),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "resource_limits/memory", testSchedulerPoolResourcesModelMemory),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "strong_guarantee_resources/cpu", testSchedulerPoolResourcesModelCPU),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "strong_guarantee_resources/memory", testSchedulerPoolResourcesModelMemory),
				),
			},
			{
				Config: accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configIntegralGuaranteesBurst),
				Check: resource.ComposeAggregateTestCheckFunc(
					accCheckYTsaurusStringAttribute(testSchedulerPoolYTCypressPath, "name", testSchedulerPoolName),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "max_operation_count", testMaxOperationCount),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "max_running_operation_count", testMaxRunningOperationCount),
					accCheckYTsaurusACLAttribute(testSchedulerPoolYTCypressPath, testACL),
					accCheckYTsaurusBoolAttribute(testSchedulerPoolYTCypressPath, "forbid_immediate_operations", testForbidImmediateOperations),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "resource_limits/cpu", testSchedulerPoolResourcesModelCPU),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "resource_limits/memory", testSchedulerPoolResourcesModelMemory),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "strong_guarantee_resources/cpu", testSchedulerPoolResourcesModelCPU),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "strong_guarantee_resources/memory", testSchedulerPoolResourcesModelMemory),
					accCheckYTsaurusStringAttribute(testSchedulerPoolYTCypressPath, "integral_guarantees/guarantee_type", testGuaranteeTypeBurst),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "integral_guarantees/burst_guarantee_resources/cpu", testSchedulerPoolResourcesModelCPU),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "integral_guarantees/burst_guarantee_resources/memory", testSchedulerPoolResourcesModelMemory),
				),
			},
			{
				Config: accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configIntegralGuaranteesRelaxed),
				Check: resource.ComposeAggregateTestCheckFunc(
					accCheckYTsaurusStringAttribute(testSchedulerPoolYTCypressPath, "name", testSchedulerPoolName),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "max_operation_count", testMaxOperationCount),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "max_running_operation_count", testMaxRunningOperationCount),
					accCheckYTsaurusACLAttribute(testSchedulerPoolYTCypressPath, testACL),
					accCheckYTsaurusBoolAttribute(testSchedulerPoolYTCypressPath, "forbid_immediate_operations", testForbidImmediateOperations),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "resource_limits/cpu", testSchedulerPoolResourcesModelCPU),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "resource_limits/memory", testSchedulerPoolResourcesModelMemory),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "strong_guarantee_resources/cpu", testSchedulerPoolResourcesModelCPU),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "strong_guarantee_resources/memory", testSchedulerPoolResourcesModelMemory),
					accCheckYTsaurusStringAttribute(testSchedulerPoolYTCypressPath, "integral_guarantees/guarantee_type", testGuaranteeTypeRelaxed),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "integral_guarantees/resource_flow/cpu", testSchedulerPoolResourcesModelCPU),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "integral_guarantees/resource_flow/memory", testSchedulerPoolResourcesModelMemory),
				),
			},
		},
	})
}

func TestSchedulerPoolResourceCreateWithAllAttributesIntegralGuaranteesBurst(t *testing.T) {
	resourceID := "fakepool"
	testSchedulerPoolName := resourceID
	testSchedulerPoolYTCypressPath := fmt.Sprintf("//sys/pool_trees/default/%s", testSchedulerPoolName)
	testPoolTree := "default"
	testForbidImmediateOperations := true
	testMaxOperationCount := int64(10)
	testMaxRunningOperationCount := int64(10)
	testSchedulerPoolResourcesModelCPU := int64(1)
	testSchedulerPoolResourcesModelMemory := int64(1 * 1024 * 1024)
	testGuaranteeTypeBurst := "burst"

	testACL := []yt.ACE{
		{
			Action:          yt.ActionAllow,
			Subjects:        []string{"users"},
			Permissions:     []string{yt.PermissionUse},
			InheritanceMode: "object_and_descendants",
		},
	}
	testACLModel, _ := acl.FlattenACL(ctx, testACL)

	resourceLimits, _ := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"cpu":    types.Int64Type,
		"memory": types.Int64Type,
	}, schedulerpool.SchedulerPoolResourcesModel{
		CPU:    types.Int64Value(testSchedulerPoolResourcesModelCPU),
		Memory: types.Int64Value(testSchedulerPoolResourcesModelMemory),
	})

	integralGuarantees, _ := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"guarantee_type": types.StringType,
		"resource_flow": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"cpu":    types.Int64Type,
				"memory": types.Int64Type,
			},
		},
		"burst_guarantee_resources": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"cpu":    types.Int64Type,
				"memory": types.Int64Type,
			},
		},
	}, schedulerpool.SchedulerPoolIntegralGuaranteesModel{
		GuaranteeType: types.StringValue(testGuaranteeTypeBurst),
		ResourceFlow: types.ObjectNull(map[string]attr.Type{
			"cpu":    types.Int64Type,
			"memory": types.Int64Type,
		}),
		BurstGuaranteeResources: resourceLimits,
	})

	configIntegralGuaranteesBurst := schedulerpool.SchedulerPoolModel{
		Name:                      types.StringValue(testSchedulerPoolName),
		PoolTree:                  types.StringValue(testPoolTree),
		ACL:                       testACLModel,
		MaxRunningOperationCount:  types.Int64Value(testMaxRunningOperationCount),
		MaxOperationCount:         types.Int64Value(testMaxOperationCount),
		ForbidImmediateOperations: types.BoolValue(testForbidImmediateOperations),
		ResourceLimits:            resourceLimits,
		StrongGuaranteeResources:  resourceLimits,
		IntegralGuarantees:        integralGuarantees,
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             accCheckYTsaurusObjectDestroyed(testSchedulerPoolYTCypressPath),
		Steps: []resource.TestStep{
			{
				Config: accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configIntegralGuaranteesBurst),
				Check: resource.ComposeAggregateTestCheckFunc(
					accCheckYTsaurusStringAttribute(testSchedulerPoolYTCypressPath, "name", testSchedulerPoolName),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "max_operation_count", testMaxOperationCount),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "max_running_operation_count", testMaxRunningOperationCount),
					accCheckYTsaurusACLAttribute(testSchedulerPoolYTCypressPath, testACL),
					accCheckYTsaurusBoolAttribute(testSchedulerPoolYTCypressPath, "forbid_immediate_operations", testForbidImmediateOperations),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "resource_limits/cpu", testSchedulerPoolResourcesModelCPU),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "resource_limits/memory", testSchedulerPoolResourcesModelMemory),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "strong_guarantee_resources/cpu", testSchedulerPoolResourcesModelCPU),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "strong_guarantee_resources/memory", testSchedulerPoolResourcesModelMemory),
					accCheckYTsaurusStringAttribute(testSchedulerPoolYTCypressPath, "integral_guarantees/guarantee_type", testGuaranteeTypeBurst),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "integral_guarantees/burst_guarantee_resources/cpu", testSchedulerPoolResourcesModelCPU),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "integral_guarantees/burst_guarantee_resources/memory", testSchedulerPoolResourcesModelMemory),
				),
			},
		},
	})
}

func TestSchedulerPoolResourceCreateWithAllAttributesIntegralGuaranteesRelaxed(t *testing.T) {
	resourceID := "fakepool"
	testSchedulerPoolName := resourceID
	testSchedulerPoolYTCypressPath := fmt.Sprintf("//sys/pool_trees/default/%s", testSchedulerPoolName)
	testPoolTree := "default"
	testForbidImmediateOperations := true
	testMaxOperationCount := int64(10)
	testMaxRunningOperationCount := int64(10)
	testSchedulerPoolResourcesModelCPU := int64(1)
	testSchedulerPoolResourcesModelMemory := int64(1 * 1024 * 1024)
	testGuaranteeTypeRelaxed := "relaxed"

	testACL := []yt.ACE{
		{
			Action:          yt.ActionAllow,
			Subjects:        []string{"users"},
			Permissions:     []string{yt.PermissionUse},
			InheritanceMode: "object_and_descendants",
		},
	}
	testACLModel, _ := acl.FlattenACL(ctx, testACL)

	resourceLimits, _ := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"cpu":    types.Int64Type,
		"memory": types.Int64Type,
	}, schedulerpool.SchedulerPoolResourcesModel{
		CPU:    types.Int64Value(testSchedulerPoolResourcesModelCPU),
		Memory: types.Int64Value(testSchedulerPoolResourcesModelMemory),
	})

	integralGuaranteesResourceFlow, _ := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"guarantee_type": types.StringType,
		"resource_flow": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"cpu":    types.Int64Type,
				"memory": types.Int64Type,
			},
		},
		"burst_guarantee_resources": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"cpu":    types.Int64Type,
				"memory": types.Int64Type,
			},
		},
	}, schedulerpool.SchedulerPoolIntegralGuaranteesModel{
		GuaranteeType: types.StringValue(testGuaranteeTypeRelaxed),
		ResourceFlow:  resourceLimits,
		BurstGuaranteeResources: types.ObjectNull(map[string]attr.Type{
			"cpu":    types.Int64Type,
			"memory": types.Int64Type,
		}),
	})

	configIntegralGuaranteesRelaxed := schedulerpool.SchedulerPoolModel{
		Name:                      types.StringValue(testSchedulerPoolName),
		PoolTree:                  types.StringValue(testPoolTree),
		ACL:                       testACLModel,
		MaxRunningOperationCount:  types.Int64Value(testMaxRunningOperationCount),
		MaxOperationCount:         types.Int64Value(testMaxOperationCount),
		ForbidImmediateOperations: types.BoolValue(testForbidImmediateOperations),
		ResourceLimits:            resourceLimits,
		StrongGuaranteeResources:  resourceLimits,
		IntegralGuarantees:        integralGuaranteesResourceFlow,
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             accCheckYTsaurusObjectDestroyed(testSchedulerPoolYTCypressPath),
		Steps: []resource.TestStep{
			{
				Config: accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configIntegralGuaranteesRelaxed),
				Check: resource.ComposeAggregateTestCheckFunc(
					accCheckYTsaurusStringAttribute(testSchedulerPoolYTCypressPath, "name", testSchedulerPoolName),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "max_operation_count", testMaxOperationCount),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "max_running_operation_count", testMaxRunningOperationCount),
					accCheckYTsaurusACLAttribute(testSchedulerPoolYTCypressPath, testACL),
					accCheckYTsaurusBoolAttribute(testSchedulerPoolYTCypressPath, "forbid_immediate_operations", testForbidImmediateOperations),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "resource_limits/cpu", testSchedulerPoolResourcesModelCPU),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "resource_limits/memory", testSchedulerPoolResourcesModelMemory),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "strong_guarantee_resources/cpu", testSchedulerPoolResourcesModelCPU),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "strong_guarantee_resources/memory", testSchedulerPoolResourcesModelMemory),
					accCheckYTsaurusStringAttribute(testSchedulerPoolYTCypressPath, "integral_guarantees/guarantee_type", testGuaranteeTypeRelaxed),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "integral_guarantees/resource_flow/cpu", testSchedulerPoolResourcesModelCPU),
					accCheckYTsaurusInt64Attribute(testSchedulerPoolYTCypressPath, "integral_guarantees/resource_flow/memory", testSchedulerPoolResourcesModelMemory),
				),
			},
		},
	})
}

func TestSchedulerPoolResourceCreateParentChild(t *testing.T) {
	resourceParentID := "fakepool_parent"
	resourceChildID := "fakepool_child"

	testParentSchedulerPoolName := resourceParentID
	testChildSchedulerPoolName := resourceChildID

	testSchedulerPoolParentYTCypressPath := fmt.Sprintf("//sys/pool_trees/default/%s", testParentSchedulerPoolName)
	testSchedulerPoolChildYTCypressPath := fmt.Sprintf("//sys/pool_trees/default/%s/%s", testParentSchedulerPoolName, testChildSchedulerPoolName)
	testSchedulerPoolRemovedParentYTCypressPath := fmt.Sprintf("//sys/pool_trees/default/%s", testChildSchedulerPoolName)
	testPoolTree := "default"
	testParentName := fmt.Sprintf("ytsaurus_scheduler_pool.%s.name", resourceParentID)

	configParent := schedulerpool.SchedulerPoolModel{
		Name:     types.StringValue(testParentSchedulerPoolName),
		PoolTree: types.StringValue(testPoolTree),
	}

	configChild := schedulerpool.SchedulerPoolModel{
		Name:       types.StringValue(testChildSchedulerPoolName),
		PoolTree:   types.StringValue(testPoolTree),
		ParentName: types.StringValue(testParentName),
	}

	configRemoveParent := schedulerpool.SchedulerPoolModel{
		Name:     types.StringValue(testChildSchedulerPoolName),
		PoolTree: types.StringValue(testPoolTree),
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             accCheckYTsaurusObjectDestroyed(testSchedulerPoolParentYTCypressPath),
		Steps: []resource.TestStep{
			{
				Config: accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceParentID, configParent) + accResourceYtsaurusSchedulerPoolConfig(resourceChildID, configChild),
				Check: resource.ComposeAggregateTestCheckFunc(
					accCheckYTsaurusStringAttribute(testSchedulerPoolParentYTCypressPath, "name", testParentSchedulerPoolName),
					accCheckYTsaurusStringAttribute(testSchedulerPoolChildYTCypressPath, "name", testChildSchedulerPoolName),
					accCheckYTsaurusStringAttribute(testSchedulerPoolChildYTCypressPath, "parent_name", testParentSchedulerPoolName),
				),
			},
			{
				Config: accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceParentID, configParent) + accResourceYtsaurusSchedulerPoolConfig(resourceChildID, configRemoveParent),
				Check: resource.ComposeAggregateTestCheckFunc(
					accCheckYTsaurusStringAttribute(testSchedulerPoolRemovedParentYTCypressPath, "name", testChildSchedulerPoolName),
				),
			},
		},
	})

}

func TestSchedulerPoolResourceCreateAndRename(t *testing.T) {
	resourceID := "fakepool"

	testSchedulerPoolName := resourceID
	testSchedulerPoolNameRenamed := fmt.Sprintf("%s_renamed", resourceID)
	testSchedulerPoolYTCypressPath := fmt.Sprintf("//sys/pool_trees/default/%s", testSchedulerPoolName)
	testSchedulerPoolYTCypressRenamed := fmt.Sprintf("//sys/pool_trees/default/%s", testSchedulerPoolNameRenamed)
	testPoolTree := "default"

	configCreate := schedulerpool.SchedulerPoolModel{
		Name:     types.StringValue(testSchedulerPoolName),
		PoolTree: types.StringValue(testPoolTree),
	}

	configRenamed := schedulerpool.SchedulerPoolModel{
		Name:     types.StringValue(testSchedulerPoolNameRenamed),
		PoolTree: types.StringValue(testPoolTree),
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             accCheckYTsaurusObjectDestroyed(testSchedulerPoolYTCypressRenamed),
		Steps: []resource.TestStep{
			{
				Config: accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configCreate),
				Check: resource.ComposeAggregateTestCheckFunc(
					accCheckYTsaurusStringAttribute(testSchedulerPoolYTCypressPath, "name", testSchedulerPoolName),
				),
			},
			{
				Config: accGetYTLocalDockerProviderConfig() + accResourceYtsaurusSchedulerPoolConfig(resourceID, configRenamed),
				Check: resource.ComposeAggregateTestCheckFunc(
					accCheckYTsaurusStringAttribute(testSchedulerPoolYTCypressRenamed, "name", testSchedulerPoolNameRenamed),
				),
			},
		},
	})
}

func accResourceYtsaurusSchedulerPoolConfig(id string, m schedulerpool.SchedulerPoolModel) string {

	config := fmt.Sprintf(`
	resource "ytsaurus_scheduler_pool" %q {
		name = %q`, id, m.Name.ValueString())

	if len(m.PoolTree.ValueString()) > 0 {
		config += fmt.Sprintf(`
		pool_tree = %q`, m.PoolTree.ValueString())
	}

	if !m.ParentName.IsNull() {
		config += fmt.Sprintf(`
		parent_name = %s`, m.ParentName.ValueString())
	}

	if !m.Weight.IsNull() {
		config += fmt.Sprintf(`
		weight = %f`, m.Weight.ValueFloat64())
	}

	if !m.Mode.IsNull() {
		config += fmt.Sprintf(`
		mode = %q`, m.Mode.ValueString())
	}

	if !m.MaxOperationCount.IsNull() {
		config += fmt.Sprintf(`
		max_operation_count = %d`, m.MaxOperationCount.ValueInt64())
	}

	if !m.MaxRunningOperationCount.IsNull() {
		config += fmt.Sprintf(`
		max_running_operation_count = %d`, m.MaxRunningOperationCount.ValueInt64())
	}

	if !m.ForbidImmediateOperations.IsNull() {
		config += fmt.Sprintf(`
		forbid_immediate_operations = %t`, m.ForbidImmediateOperations.ValueBool())
	}

	if !m.ResourceLimits.IsNull() {
		config += `
		resource_limits = {`

		var resourcesModel schedulerpool.SchedulerPoolResourcesModel
		_ = m.ResourceLimits.As(ctx, &resourcesModel, basetypes.ObjectAsOptions{})

		if !resourcesModel.CPU.IsNull() {
			config += fmt.Sprintf(`
			cpu = %d`, resourcesModel.CPU.ValueInt64())
		}

		if !resourcesModel.Memory.IsNull() {
			config += fmt.Sprintf(`
			memory = %d`, resourcesModel.Memory.ValueInt64())
		}

		config += `
		}`
	}

	if !m.StrongGuaranteeResources.IsNull() {
		config += `
		strong_guarantee_resources = {`

		var resourcesModel schedulerpool.SchedulerPoolResourcesModel
		_ = m.StrongGuaranteeResources.As(ctx, &resourcesModel, basetypes.ObjectAsOptions{})

		if !resourcesModel.CPU.IsNull() {
			config += fmt.Sprintf(`
			cpu = %d`, resourcesModel.CPU.ValueInt64())
		}

		if !resourcesModel.Memory.IsNull() {
			config += fmt.Sprintf(`
			memory = %d`, resourcesModel.Memory.ValueInt64())
		}

		config += `
		}`
	}

	if !m.IntegralGuarantees.IsNull() {
		config += `
		integral_guarantees = {`

		var integralGuaranteesModel schedulerpool.SchedulerPoolIntegralGuaranteesModel
		_ = m.IntegralGuarantees.As(ctx, &integralGuaranteesModel, basetypes.ObjectAsOptions{})

		if !integralGuaranteesModel.GuaranteeType.IsNull() {
			config += fmt.Sprintf(`
			guarantee_type = %q`, integralGuaranteesModel.GuaranteeType.ValueString())
		}

		if !integralGuaranteesModel.ResourceFlow.IsNull() {
			config += `
			resource_flow = {`

			var resourcesModel schedulerpool.SchedulerPoolResourcesModel
			_ = integralGuaranteesModel.ResourceFlow.As(ctx, &resourcesModel, basetypes.ObjectAsOptions{})

			if !resourcesModel.CPU.IsNull() {
				config += fmt.Sprintf(`
				cpu = %d`, resourcesModel.CPU.ValueInt64())
			}

			if !resourcesModel.Memory.IsNull() {
				config += fmt.Sprintf(`
				memory = %d`, resourcesModel.Memory.ValueInt64())
			}

			config += `
			}`
		}

		if !integralGuaranteesModel.BurstGuaranteeResources.IsNull() {
			config += `
			burst_guarantee_resources = {`

			var resourcesModel schedulerpool.SchedulerPoolResourcesModel
			_ = integralGuaranteesModel.BurstGuaranteeResources.As(ctx, &resourcesModel, basetypes.ObjectAsOptions{})

			if !resourcesModel.CPU.IsNull() {
				config += fmt.Sprintf(`
				cpu = %d`, resourcesModel.CPU.ValueInt64())
			}

			if !resourcesModel.Memory.IsNull() {
				config += fmt.Sprintf(`
				memory = %d`, resourcesModel.Memory.ValueInt64())
			}

			config += `
			}`
		}

		config += `
		}`
	}

	acl, _ := acl.ExpandACL(ctx, m.ACL)
	if len(acl) > 0 {
		config += accAddACLConfig(acl)
	}

	config += `
	}`

	return config
}
