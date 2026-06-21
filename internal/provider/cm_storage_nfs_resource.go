package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/asashv/terraform-provider-asashv/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CMStorageNFSResource{}
var _ resource.ResourceWithConfigure = &CMStorageNFSResource{}

func NewCMStorageNFSResource() resource.Resource {
	return &CMStorageNFSResource{}
}

type CMStorageNFSResource struct {
	client *client.Client
}

type CMStorageNFSResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	NFSTargetIP   types.String `tfsdk:"nfs_target_ip"`
	NFSRemotePath types.String `tfsdk:"nfs_remote_path"`

	Hostnames []types.String `tfsdk:"hostnames"`
	HostIDs   types.List     `tfsdk:"host_ids"`

	Type      types.String  `tfsdk:"type"`
	Protocol  types.String  `tfsdk:"protocol"`
	Status    types.String  `tfsdk:"status"`
	SizeGB    types.Float64 `tfsdk:"size_gb"`
	UsedGB    types.Float64 `tfsdk:"used_gb"`
	FreeGB    types.Float64 `tfsdk:"free_gb"`
	HostCount types.Int64   `tfsdk:"host_count"`
}

func (r *CMStorageNFSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cm_storage_nfs"
}

func (r *CMStorageNFSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an NFS storage pool in LejamCM.",
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
			"nfs_target_ip": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nfs_remote_path": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hostnames": schema.ListAttribute{
				Required:    true,
				Description: "LejamCM hostnames to attach this storage to, for example [\"VirtuWaHV-53\"].",
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"host_ids": schema.ListAttribute{
				Computed:    true,
				Description: "Resolved LejamCM host IDs.",
				ElementType: types.StringType,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"protocol": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"size_gb": schema.Float64Attribute{
				Computed: true,
			},
			"used_gb": schema.Float64Attribute{
				Computed: true,
			},
			"free_gb": schema.Float64Attribute{
				Computed: true,
			},
			"host_count": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (r *CMStorageNFSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CMStorageNFSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CMStorageNFSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostnames := expandStringList(plan.Hostnames)

	hostIDs, err := resolveCMHostnamesToIDs(r.client, hostnames)
	if err != nil {
		resp.Diagnostics.AddError("Error resolving LejamCM hostnames", err.Error())
		return
	}

	err = r.client.CreateCMStorageNFS(client.CreateCMStorageNFSRequest{
		Name:          plan.Name.ValueString(),
		Type:          "Shared",
		Protocol:      "NFS",
		NFSTargetIP:   plan.NFSTargetIP.ValueString(),
		NFSRemotePath: plan.NFSRemotePath.ValueString(),
		HostIDs:       hostIDs,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating LejamCM NFS storage", err.Error())
		return
	}

	pool, err := waitForCMStorageByName(ctx, r.client, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error finding created LejamCM NFS storage", err.Error())
		return
	}

	applyCMStorageNFSToModel(&plan, pool)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CMStorageNFSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CMStorageNFSResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pool, err := findCMStorage(r.client, state.ID.ValueString(), state.Name.ValueString())
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error reading LejamCM NFS storage", err.Error())
		return
	}

	applyCMStorageNFSToModel(&state, pool)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CMStorageNFSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CMStorageNFSResourceModel
	var state CMStorageNFSResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pool, err := findCMStorage(r.client, state.ID.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading LejamCM NFS storage during update", err.Error())
		return
	}

	plan.ID = state.ID
	applyCMStorageNFSToModel(&plan, pool)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CMStorageNFSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CMStorageNFSResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	storageID := state.ID.ValueString()

	removeHostIDs := stringListFromTypesList(state.HostIDs)
	if len(removeHostIDs) == 0 {
		if pool, err := findCMStorage(r.client, storageID, state.Name.ValueString()); err == nil {
			removeHostIDs = attachedHostIDs(pool.AttachedHosts)
		}
	}

	if len(removeHostIDs) > 0 {
		err := r.client.UpdateCMStorageHosts(storageID, client.UpdateCMStorageHostsRequest{
			NewName:       state.Name.ValueString(),
			AddHostIDs:    []string{},
			RemoveHostIDs: removeHostIDs,
		})
		if err != nil {
			resp.Diagnostics.AddError("Error detaching hosts from LejamCM NFS storage", err.Error())
			return
		}
	}

	if err := waitForCMStorageHostCount(ctx, r.client, storageID, state.Name.ValueString(), 0); err != nil {
		resp.Diagnostics.AddError("Error waiting for LejamCM NFS storage host detach", err.Error())
		return
	}

	err := r.client.DeleteCMStorage(storageID)
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not found") {
			return
		}

		resp.Diagnostics.AddError("Error deleting LejamCM NFS storage", err.Error())
		return
	}

	if err := waitForCMStorageDeleted(ctx, r.client, storageID, state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error waiting for LejamCM NFS storage delete", err.Error())
		return
	}
}

func applyCMStorageNFSToModel(model *CMStorageNFSResourceModel, pool *client.StoragePool) {
	model.ID = types.StringValue(pool.UUID)
	model.Name = types.StringValue(pool.Name)
	model.NFSTargetIP = types.StringValue(pool.Server)
	model.NFSRemotePath = types.StringValue(pool.Path)

	model.HostIDs = stringListToTypesList(attachedHostIDs(pool.AttachedHosts))
	model.Hostnames = flattenStringList(attachedHostnames(pool.AttachedHosts))

	model.Type = types.StringValue(pool.Type)
	model.Protocol = types.StringValue(pool.Protocol)
	model.Status = types.StringValue(pool.Status)
	model.SizeGB = types.Float64Value(pool.SizeGB)
	model.UsedGB = types.Float64Value(pool.UsedGB)
	model.FreeGB = types.Float64Value(pool.FreeGB)
	model.HostCount = types.Int64Value(pool.HostCount)
}

func resolveCMHostnamesToIDs(c *client.Client, hostnames []string) ([]string, error) {
	hosts, err := c.ListCMHosts()
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(hostnames))

	for _, wantedHostname := range hostnames {
		found := false

		for _, host := range hosts {
			if host.Hostname == wantedHostname {
				result = append(result, host.ID)
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("LejamCM host with hostname %q not found", wantedHostname)
		}
	}

	return result, nil
}

func stringListToTypesList(values []string) types.List {
	items := make([]attr.Value, 0, len(values))

	for _, value := range values {
		items = append(items, types.StringValue(value))
	}

	return types.ListValueMust(types.StringType, items)
}

func stringListFromTypesList(value types.List) []string {
	if value.IsNull() || value.IsUnknown() {
		return []string{}
	}

	result := make([]string, 0, len(value.Elements()))

	for _, item := range value.Elements() {
		str, ok := item.(types.String)
		if !ok || str.IsNull() || str.IsUnknown() {
			continue
		}

		result = append(result, str.ValueString())
	}

	return result
}

func attachedHostIDs(hosts []client.StorageAttachedHost) []string {
	result := make([]string, 0, len(hosts))

	for _, host := range hosts {
		if host.ID != "" {
			result = append(result, host.ID)
		}
	}

	return result
}

func attachedHostnames(hosts []client.StorageAttachedHost) []string {
	result := make([]string, 0, len(hosts))

	for _, host := range hosts {
		if host.Hostname != "" {
			result = append(result, host.Hostname)
		}
	}

	return result
}

func findCMStorage(c *client.Client, id string, name string) (*client.StoragePool, error) {
	pools, err := c.ListStoragePools()
	if err != nil {
		return nil, err
	}

	for _, pool := range pools {
		if id != "" && pool.UUID == id {
			return &pool, nil
		}

		if name != "" && pool.Name == name {
			return &pool, nil
		}
	}

	return nil, fmt.Errorf("LejamCM storage not found")
}

func waitForCMStorageByName(ctx context.Context, c *client.Client, name string) (*client.StoragePool, error) {
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("timed out waiting for LejamCM storage %q", name)
		case <-ticker.C:
			pool, err := findCMStorage(c, "", name)
			if err == nil {
				return pool, nil
			}
		}
	}
}

func waitForCMStorageHostCount(ctx context.Context, c *client.Client, id string, name string, wanted int64) error {
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timed out waiting for storage host_count=%d", wanted)
		case <-ticker.C:
			pool, err := findCMStorage(c, id, name)
			if err != nil {
				return err
			}

			if pool.HostCount == wanted {
				return nil
			}
		}
	}
}

func waitForCMStorageDeleted(ctx context.Context, c *client.Client, id string, name string) error {
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timed out waiting for LejamCM storage %q to be deleted", name)
		case <-ticker.C:
			_, err := findCMStorage(c, id, name)
			if err != nil && strings.Contains(strings.ToLower(err.Error()), "not found") {
				return nil
			}
		}
	}
}
