package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/asashv/terraform-provider-asashv/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

var _ resource.Resource = &VMResource{}
var _ resource.ResourceWithConfigure = &VMResource{}
var _ resource.ResourceWithImportState = &VMResource{}

func NewVMResource() resource.Resource {
	return &VMResource{}
}

type VMResource struct {
	client *client.Client
}

type VMResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Description types.String   `tfsdk:"description"`
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
	UnmountCDROM types.Bool   `tfsdk:"unmount_cdrom"`
}

type DiskModel struct {
	Pool          types.String  `tfsdk:"pool"`
	SizeGB        types.Float64 `tfsdk:"size_gb"`
	Bus           types.String  `tfsdk:"bus"`
	Format        types.String  `tfsdk:"format"`
	StorageTarget types.String  `tfsdk:"storage_target"`
}

type CDROMModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (r *VMResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

func (r *VMResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a VM on ASASHV.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ASASHV VM ID.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "VM name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "VM description.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Default: stringdefault.StaticString("Created by Terraform"),
			},
			"vcpu": schema.Int64Attribute{
				Required:    true,
				Description: "Number of virtual CPUs.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"vcores": schema.Int64Attribute{
				Required:    true,
				Description: "Number of virtual cores.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"memory_gb": schema.Float64Attribute{
				Required:    true,
				Description: "Memory size in GB.",
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
				},
			},
			"boot_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "VM boot type: bios or uefi.",
				Default:     stringdefault.StaticString("uefi"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nics": schema.ListAttribute{
				Required:    true,
				Description: "List of VM networks/VLAN names, for example [\"default\"].",
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"boot_order": schema.ListAttribute{
				Optional:    true,
				Description: "Boot order, for example [\"disk0\", \"cdrom0\"].",
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"delete_disks": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Delete VM disks when destroying the VM.",
				Default:     booldefault.StaticBool(true),
			},
			"power_state": schema.StringAttribute{
				Computed:    true,
				Description: "Current VM power state.",
			},
			"ip": schema.StringAttribute{
				Computed:    true,
				Description: "VM IP address, if ASASHV reports it.",
			},
			"os": schema.StringAttribute{
				Computed:    true,
				Description: "Detected guest OS, if ASASHV reports it.",
			},
			"unmount_cdrom": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether ASASHV reports CDROM as unmounted.",
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
							Required:    true,
							Description: "Storage pool name, for example NFS-Storage.",
						},
						"size_gb": schema.Float64Attribute{
							Required:    true,
							Description: "Disk size in GB.",
						},
						"bus": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Disk bus.",
							Default:     stringdefault.StaticString("virtio"),
						},
						"format": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Disk format.",
							Default:     stringdefault.StaticString("qcow2"),
						},
						"storage_target": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Storage target.",
							Default:     stringdefault.StaticString("allocate"),
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
							Required:    true,
							Description: "ISO ID from /api/storage/iso.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "ISO name returned by ASASHV.",
						},
					},
				},
			},
		},
	}
}

func (r *VMResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VMResourceModel

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
		Disks:       expandDisks(plan.Disks),
		CDROM:       expandCDROM(plan.CDROM),
		NICs:        expandStringList(plan.NICs),
		BootOrder:   expandStringList(plan.BootOrder),
	}

	created, err := r.client.CreateVM(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating ASASHV VM", err.Error())
		return
	}

	vmID := created.ID

	if vmID == "" {
		vm, err := waitForVMByName(ctx, r.client, plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error finding created ASASHV VM", err.Error())
			return
		}
		vmID = vm.ID
	}

	vm, err := r.client.GetVM(vmID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading ASASHV VM after create", err.Error())
		return
	}

	applyVMToModel(&plan, vm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VMResourceModel

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

		resp.Diagnostics.AddError("Error reading ASASHV VM", err.Error())
		return
	}

	applyVMToModel(&state, vm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *VMResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VMResourceModel
	var state VMResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vm, err := r.client.GetVM(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading ASASHV VM during update", err.Error())
		return
	}

	plan.ID = state.ID
	applyVMToModel(&plan, vm)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VMResourceModel

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

		resp.Diagnostics.AddError("Error deleting ASASHV VM", err.Error())
		return
	}
}

func (r *VMResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	vm, err := r.client.GetVM(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing ASASHV VM", err.Error())
		return
	}

	var state VMResourceModel
	state.DeleteDisks = types.BoolValue(false)

	applyVMToModel(&state, vm)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func expandDisks(disks []DiskModel) []client.Disk {
	result := make([]client.Disk, 0, len(disks))

	for _, d := range disks {
		result = append(result, client.Disk{
			Pool:          d.Pool.ValueString(),
			SizeGB:        d.SizeGB.ValueFloat64(),
			Bus:           d.Bus.ValueString(),
			Format:        d.Format.ValueString(),
			StorageTarget: d.StorageTarget.ValueString(),
		})
	}

	return result
}

func expandCDROM(cdrom []CDROMModel) []client.CDROM {
	result := make([]client.CDROM, 0, len(cdrom))

	for _, c := range cdrom {
		result = append(result, client.CDROM{
			ID: c.ID.ValueString(),
		})
	}

	return result
}

func expandStringList(values []types.String) []string {
	result := make([]string, 0, len(values))

	for _, v := range values {
		if !v.IsNull() && !v.IsUnknown() {
			result = append(result, v.ValueString())
		}
	}

	return result
}

func flattenDisks(disks []client.Disk) []DiskModel {
	result := make([]DiskModel, 0, len(disks))

	for _, d := range disks {
		result = append(result, DiskModel{
			Pool:          types.StringValue(d.Pool),
			SizeGB:        types.Float64Value(d.SizeGB),
			Bus:           types.StringValue(d.Bus),
			Format:        types.StringValue(d.Format),
			StorageTarget: stringValueOrNull(d.StorageTarget),
		})
	}

	return result
}

func flattenCDROM(cdrom []client.CDROM) []CDROMModel {
	result := make([]CDROMModel, 0, len(cdrom))

	for _, c := range cdrom {
		result = append(result, CDROMModel{
			ID:   types.StringValue(c.ID),
			Name: stringValueOrNull(c.Name),
		})
	}

	return result
}

func flattenStringList(values []string) []types.String {
	result := make([]types.String, 0, len(values))

	for _, v := range values {
		result = append(result, types.StringValue(v))
	}

	return result
}

func stringValueOrNull(value string) types.String {
	if value == "" {
		return types.StringNull()
	}

	return types.StringValue(value)
}

func applyVMToModel(model *VMResourceModel, vm *client.VM) {
	model.ID = types.StringValue(vm.ID)
	model.Name = types.StringValue(vm.Name)
	model.Description = stringValueOrNull(vm.Description)
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
	model.UnmountCDROM = types.BoolValue(vm.UnmountCDROM)
}

func waitForVMByName(ctx context.Context, c *client.Client, name string) (*client.VM, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(60 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("timed out waiting for VM %q to appear", name)
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

func addDiagError(diags *diag.Diagnostics, summary, detail string) {
	diags.AddAttributeError(path.Root("id"), summary, detail)
}
