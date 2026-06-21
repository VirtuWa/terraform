package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/asashv/terraform-provider-asashv/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CMVMResource{}
var _ resource.ResourceWithConfigure = &CMVMResource{}
var _ resource.ResourceWithImportState = &CMVMResource{}

func NewCMVMResource() resource.Resource {
	return &CMVMResource{}
}

type CMVMResource struct {
	client *client.Client
}

type CMVMResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Description types.String   `tfsdk:"description"`
	HostID      types.String   `tfsdk:"host_id"`
	ClusterID   types.String   `tfsdk:"cluster_id"`
	VCPU        types.Int64    `tfsdk:"vcpu"`
	VCores      types.Int64    `tfsdk:"vcores"`
	MemoryGB    types.Float64  `tfsdk:"memory_gb"`
	BootType    types.String   `tfsdk:"boot_type"`
	Disks       []DiskModel    `tfsdk:"disks"`
	CDROM       []CDROMModel   `tfsdk:"cdrom"`
	NICs        []types.String `tfsdk:"nics"`
	BootOrder   []types.String `tfsdk:"boot_order"`
	DeleteDisks types.Bool     `tfsdk:"delete_disks"`

	PowerState   types.String `tfsdk:"power_state"`
	IP           types.String `tfsdk:"ip"`
	OS           types.String `tfsdk:"os"`
	HostName     types.String `tfsdk:"host_name"`
	HostIP       types.String `tfsdk:"host_ip"`
	ClusterName  types.String `tfsdk:"cluster_name"`
	UnmountCDROM types.Bool   `tfsdk:"unmount_cdrom"`
}

func (r *CMVMResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cm_vm"
}

func (r *CMVMResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a VM through LejamCM.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("Created by Terraform through LejamCM"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host_id": schema.StringAttribute{
				Required:    true,
				Description: "LejamCM host ID, for example host-07fe88c2.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional LejamCM cluster ID. If omitted, LejamCM may return the selected host cluster ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vcpu": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"vcores": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"memory_gb": schema.Float64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
				},
			},
			"boot_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("uefi"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nics": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"boot_order": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"delete_disks": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"power_state": schema.StringAttribute{
				Computed: true,
			},
			"ip": schema.StringAttribute{
				Computed: true,
			},
			"os": schema.StringAttribute{
				Computed: true,
			},
			"host_name": schema.StringAttribute{
				Computed: true,
			},
			"host_ip": schema.StringAttribute{
				Computed: true,
			},
			"cluster_name": schema.StringAttribute{
				Computed: true,
			},
			"unmount_cdrom": schema.BoolAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"disks": schema.ListNestedBlock{
				Description: "VM disks.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"pool": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(""),
						},
						"size_gb": schema.Float64Attribute{
							Required: true,
						},
						"bus": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("virtio"),
						},
						"format": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("qcow2"),
						},
						"storage_target": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("allocate"),
						},
					},
				},
			},
			"cdrom": schema.ListNestedBlock{
				Description: "Attached ISO/CDROM list.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (r *CMVMResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *CMVMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CMVMResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateVMRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		VCPU:        plan.VCPU.ValueInt64(),
		VCores:      plan.VCores.ValueInt64(),
		MemoryGB:    plan.MemoryGB.ValueFloat64(),
		BootType:    plan.BootType.ValueString(),
		HostIDs:     plan.HostID.ValueString(),
		ClusterID:   "",
		Disks:       expandDisks(plan.Disks),
		CDROM:       expandCDROM(plan.CDROM),
		NICs:        expandStringList(plan.NICs),
		BootOrder:   expandStringList(plan.BootOrder),
	}

	if !plan.ClusterID.IsNull() && !plan.ClusterID.IsUnknown() {
		createReq.ClusterID = plan.ClusterID.ValueString()
	}

	created, err := r.client.CreateVM(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating LejamCM VM", err.Error())
		return
	}

	vmID := created.ID
	if vmID == "" {
		vm, err := waitForCMVMByName(ctx, r.client, plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error finding created LejamCM VM", err.Error())
			return
		}
		vmID = vm.ID
	}

	vm, err := r.client.GetVM(vmID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading LejamCM VM after create", err.Error())
		return
	}

	applyCMVMToModel(&plan, vm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CMVMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CMVMResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vm, err := r.client.GetVM(state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error reading LejamCM VM", err.Error())
		return
	}

	applyCMVMToModel(&state, vm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CMVMResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CMVMResourceModel
	var state CMVMResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vm, err := r.client.GetVM(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading LejamCM VM during update", err.Error())
		return
	}

	plan.ID = state.ID
	applyCMVMToModel(&plan, vm)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CMVMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CMVMResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteDisks := true
	if !state.DeleteDisks.IsNull() && !state.DeleteDisks.IsUnknown() {
		deleteDisks = state.DeleteDisks.ValueBool()
	}

	err := r.client.DeleteVM(state.ID.ValueString(), deleteDisks)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return
		}

		resp.Diagnostics.AddError("Error deleting LejamCM VM", err.Error())
		return
	}
}

func (r *CMVMResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	vm, err := r.client.GetVM(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing LejamCM VM", err.Error())
		return
	}

	var state CMVMResourceModel
	state.DeleteDisks = types.BoolValue(false)

	applyCMVMToModel(&state, vm)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func applyCMVMToModel(model *CMVMResourceModel, vm *client.VM) {
	model.ID = types.StringValue(vm.ID)
	model.Name = types.StringValue(vm.Name)
	model.Description = stringValueOrNull(vm.Description)
	if model.HostID.IsNull() || model.HostID.IsUnknown() || model.HostID.ValueString() == "" {
		model.HostID = stringValueOrNull(vm.HostID)
	}
	if model.ClusterID.IsNull() || model.ClusterID.IsUnknown() || model.ClusterID.ValueString() == "" {
		model.ClusterID = stringValueOrNull(vm.ClusterID)
	}
	model.VCPU = types.Int64Value(vm.VCPU)
	model.VCores = types.Int64Value(vm.VCores)
	model.MemoryGB = types.Float64Value(vm.MemoryGB)
	model.BootType = stringValueOrNull(vm.BootType)
	model.Disks = flattenDisks(vm.Disks)
	model.CDROM = flattenCDROM(vm.CDROM)
	model.NICs = flattenStringList(vm.NICs)
	model.BootOrder = flattenStringList(vm.BootOrder)
	model.PowerState = stringValueOrNull(vm.PowerState)
	model.IP = stringValueOrNull(vm.IP)
	model.OS = stringValueOrNull(vm.OS)
	model.HostName = stringValueOrNull(vm.HostName)
	model.HostIP = stringValueOrNull(vm.HostIP)
	model.ClusterName = stringValueOrNull(vm.ClusterName)
	model.UnmountCDROM = types.BoolValue(vm.UnmountCDROM)
}

func waitForCMVMByName(ctx context.Context, c *client.Client, name string) (*client.VM, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(90 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("timed out waiting for LejamCM VM %q to appear", name)
		case <-ticker.C:
			vms, err := c.ListVMs()
			if err != nil {
				return nil, err
			}

			for _, vm := range vms {
				if vm.Name == name {
					return &vm, nil
				}
			}
		}
	}
}
