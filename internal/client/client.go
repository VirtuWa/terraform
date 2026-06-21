package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	Endpoint   string
	Username   string
	Password   string
	Token      string
	HTTPClient *http.Client
}

type APIResponse[T any] struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type LoginData struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Role        string `json:"role"`
}

type CurrentUserData struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

type VM struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	PowerState   string   `json:"power_state"`
	BootType     string   `json:"boot_type"`
	Disks        []Disk   `json:"disks"`
	CDROM        []CDROM  `json:"cdrom"`
	NICs         []string `json:"nics"`
	VCPU         int64    `json:"vcpu"`
	VCores       int64    `json:"vcores"`
	MemoryGB     float64  `json:"memory_gb"`
	BootOrder    []string `json:"boot_order"`
	UnmountCDROM bool     `json:"unmount_cdrom"`

	IP          string  `json:"ip"`
	OS          string  `json:"os"`
	HostID      string  `json:"host_id"`
	HostName    string  `json:"host_name"`
	HostIP      string  `json:"host_ip"`
	ClusterID   string  `json:"cluster_id"`
	ClusterName string  `json:"cluster_name"`
	Status      string  `json:"status"`
	CPUUtilPct  float64 `json:"cpu_util_pct"`
	MemUtilPct  float64 `json:"mem_util_pct"`
	DiskGB      float64 `json:"disk_gb"`
	DiskUtilPct float64 `json:"disk_util_pct"`
}

type Disk struct {
	Pool          string  `json:"pool"`
	SizeGB        float64 `json:"size_gb"`
	Bus           string  `json:"bus"`
	Format        string  `json:"format"`
	StorageTarget string  `json:"storage_target,omitempty"`
	DiskImagePath *string `json:"disk_image_path,omitempty"`
	DiskImageID   *string `json:"disk_image_id,omitempty"`
}

type CDROM struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type CreateVMRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	VCPU        int64    `json:"vcpu"`
	VCores      int64    `json:"vcores"`
	MemoryGB    float64  `json:"memory_gb"`
	BootType    string   `json:"boot_type"`
	HostIDs     string   `json:"host_ids,omitempty"`
	ClusterID   string   `json:"cluster_id,omitempty"`
	Disks       []Disk   `json:"disks"`
	CDROM       []CDROM  `json:"cdrom"`
	NICs        []string `json:"nics"`
	BootOrder   []string `json:"boot_order,omitempty"`
}

type StorageListData struct {
	Pools []StoragePool `json:"pools"`
}

type StorageAttachedHost struct {
	ID        string `json:"id"`
	Hostname  string `json:"hostname"`
	IPAddress string `json:"ip_address"`
}

type StoragePool struct {
	Name          string                `json:"name"`
	UUID          string                `json:"uuid"`
	State         string                `json:"state"`
	Status        string                `json:"status"`
	Active        bool                  `json:"active"`
	Vendor        string                `json:"vendor"`
	Server        string                `json:"server"`
	Protocol      string                `json:"protocol"`
	Type          string                `json:"type"`
	SizeGB        float64               `json:"size_gb"`
	UsedGB        float64               `json:"used_gb"`
	FreeGB        float64               `json:"free_gb"`
	Path          string                `json:"path"`
	Usage         string                `json:"usage"`
	AttachedHosts []StorageAttachedHost `json:"attached_hosts"`
	HostCount     int64                 `json:"host_count"`
}

type ISOListData struct {
	ISOs []ISO `json:"isos"`
}

type ISO struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	SizeBytes int64   `json:"size_bytes"`
	SizeGB    float64 `json:"size_gb"`
}

type VLAN struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	VLANID           int64   `json:"vlan_id"`
	BridgeName       string  `json:"bridge_name"`
	ClusterID        *string `json:"cluster_id"`
	ClusterName      *string `json:"cluster_name"`
	PrimaryVSNetwork string  `json:"primary_vs_network"`
	VSNetworkCount   int64   `json:"vs_network_count"`
}

type CreateCMStorageNFSRequest struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Protocol      string   `json:"protocol"`
	NFSTargetIP   string   `json:"nfs_target_ip"`
	NFSRemotePath string   `json:"nfs_remote_path"`
	HostIDs       []string `json:"host_ids"`
}

type UpdateCMStorageHostsRequest struct {
	NewName       string   `json:"new_name"`
	AddHostIDs    []string `json:"add_host_ids"`
	RemoveHostIDs []string `json:"remove_host_ids"`
}

type CreateVLANRequest struct {
	VLANID      int64  `json:"vlan_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type CMHost struct {
	ID                 string      `json:"id"`
	Hostname           string      `json:"hostname"`
	IPAddress          string      `json:"ipAddress"`
	ClusterID          *string     `json:"clusterId"`
	ClusterName        *string     `json:"clusterName"`
	Vendor             string      `json:"vendor"`
	Model              string      `json:"model"`
	CPUSockets         int64       `json:"cpuSockets"`
	CPUCoresPerSocket  int64       `json:"cpuCoresPerSocket"`
	RAMGB              float64     `json:"ramGb"`
	CPUUsagePercent    float64     `json:"cpuUsagePercent"`
	MemoryUsagePercent float64     `json:"memoryUsagePercent"`
	Status             string      `json:"status"`
	MaintenanceMode    bool        `json:"maintenance_mode"`
	Capabilities       interface{} `json:"capabilities"`
}

func New(endpoint, username, password, token string, insecure bool) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}, //nolint:gosec
	}

	return &Client{
		Endpoint: strings.TrimRight(endpoint, "/"),
		Username: username,
		Password: password,
		Token:    token,
		HTTPClient: &http.Client{
			Timeout:   60 * time.Second,
			Transport: tr,
		},
	}
}

func (c *Client) Login() error {
	if c.Token != "" {
		return nil
	}

	if c.Username == "" || c.Password == "" {
		return fmt.Errorf("username/password or token is required")
	}

	body := map[string]string{
		"username": c.Username,
		"password": c.Password,
	}

	var out APIResponse[LoginData]
	if err := c.doNoAuth(http.MethodPost, "/api/auth/login", body, &out); err != nil {
		return err
	}

	if out.Data.AccessToken == "" {
		return fmt.Errorf("login succeeded but access_token was empty")
	}

	c.Token = out.Data.AccessToken
	return nil
}

func (c *Client) Me() (*CurrentUserData, error) {
	var out APIResponse[CurrentUserData]
	err := c.do(http.MethodGet, "/api/auth/me", nil, &out)
	return &out.Data, err
}

func (c *Client) ListVMs() ([]VM, error) {
	var out APIResponse[[]VM]
	err := c.do(http.MethodGet, "/api/vms", nil, &out)
	return out.Data, err
}

func (c *Client) GetVM(id string) (*VM, error) {
	var out APIResponse[VM]
	err := c.do(http.MethodGet, "/api/vms/"+id, nil, &out)
	return &out.Data, err
}

func (c *Client) CreateVM(input CreateVMRequest) (*VM, error) {
	var out APIResponse[json.RawMessage]
	if err := c.do(http.MethodPost, "/api/vms", input, &out); err != nil {
		return nil, err
	}

	// LejamCM create response shape:
	// data: { status: 201, message: "...", data: { vm_id: "...", name: "...", host_id: "..." } }
	var nested struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
		Data    struct {
			VMID        string `json:"vm_id"`
			ID          string `json:"id"`
			Name        string `json:"name"`
			HostID      string `json:"host_id"`
			HostName    string `json:"host_name"`
			HostIP      string `json:"host_ip"`
			ClusterID   string `json:"cluster_id"`
			ClusterName string `json:"cluster_name"`
		} `json:"data"`
	}

	if err := json.Unmarshal(out.Data, &nested); err == nil {
		id := nested.Data.ID
		if id == "" {
			id = nested.Data.VMID
		}

		if id != "" {
			return &VM{
				ID:          id,
				Name:        nested.Data.Name,
				HostID:      nested.Data.HostID,
				HostName:    nested.Data.HostName,
				HostIP:      nested.Data.HostIP,
				ClusterID:   nested.Data.ClusterID,
				ClusterName: nested.Data.ClusterName,
				Status:      fmt.Sprintf("%d", nested.Status),
			}, nil
		}
	}

	// Direct response shape:
	// data: { vm_id: "..." } or data: { id: "..." }
	var direct struct {
		VMID        string `json:"vm_id"`
		ID          string `json:"id"`
		Name        string `json:"name"`
		HostID      string `json:"host_id"`
		HostName    string `json:"host_name"`
		HostIP      string `json:"host_ip"`
		ClusterID   string `json:"cluster_id"`
		ClusterName string `json:"cluster_name"`
	}

	if err := json.Unmarshal(out.Data, &direct); err == nil {
		id := direct.ID
		if id == "" {
			id = direct.VMID
		}

		if id != "" {
			return &VM{
				ID:          id,
				Name:        direct.Name,
				HostID:      direct.HostID,
				HostName:    direct.HostName,
				HostIP:      direct.HostIP,
				ClusterID:   direct.ClusterID,
				ClusterName: direct.ClusterName,
			}, nil
		}
	}

	// ASASHV direct hypervisor response shape:
	// data: { id: "...", name: "...", ...full VM... }
	var vm VM
	if err := json.Unmarshal(out.Data, &vm); err != nil {
		return nil, fmt.Errorf("failed to decode VM create response: %w; data: %s", err, string(out.Data))
	}

	if vm.ID == "" {
		return nil, fmt.Errorf("VM create response did not include id or vm_id; data: %s", string(out.Data))
	}

	return &vm, nil
}

func (c *Client) DeleteVM(id string, deleteDisks bool) error {
	path := "/api/vms/" + id
	if deleteDisks {
		path += "?delete_disks=true"
	}

	var out any
	return c.do(http.MethodDelete, path, nil, &out)
}

func (c *Client) ListStoragePools() ([]StoragePool, error) {
	var out APIResponse[StorageListData]
	err := c.do(http.MethodGet, "/api/storage", nil, &out)
	return out.Data.Pools, err
}

func (c *Client) CreateCMStorageNFS(input CreateCMStorageNFSRequest) error {
	var out any
	return c.do(http.MethodPost, "/api/storage", input, &out)
}

func (c *Client) UpdateCMStorageHosts(storageID string, input UpdateCMStorageHostsRequest) error {
	path := "/api/storage/" + storageID

	var out any
	return c.do(http.MethodPut, path, input, &out)
}

func (c *Client) DeleteCMStorage(storageID string) error {
	path := "/api/storage/" + storageID

	var out any
	return c.do(http.MethodDelete, path, nil, &out)
}

func (c *Client) ListISOs() ([]ISO, error) {
	var out APIResponse[ISOListData]
	err := c.do(http.MethodGet, "/api/storage/iso", nil, &out)
	return out.Data.ISOs, err
}

func (c *Client) ListCMISOs() ([]ISO, error) {
	var out APIResponse[ISOListData]
	err := c.do(http.MethodGet, "/api/storage/iso/", nil, &out)
	return out.Data.ISOs, err
}

func (c *Client) ListVLANs() ([]VLAN, error) {
	var out APIResponse[[]VLAN]
	err := c.do(http.MethodGet, "/api/network/vlans", nil, &out)
	return out.Data, err
}

func (c *Client) CreateCMVLAN(input CreateVLANRequest) (*VLAN, error) {
	var out APIResponse[VLAN]
	if err := c.do(http.MethodPost, "/api/network/vlans", input, &out); err != nil {
		return nil, err
	}

	return &out.Data, nil
}

func (c *Client) DeleteCMVLAN(id string) error {
	path := "/api/network/vlans/" + id

	var out any
	return c.do(http.MethodDelete, path, nil, &out)
}

func (c *Client) ListCMHosts() ([]CMHost, error) {
	var out APIResponse[[]CMHost]
	err := c.do(http.MethodGet, "/api/dashboard/hosts", nil, &out)
	return out.Data, err
}

func (c *Client) do(method, path string, body any, out any) error {
	if err := c.Login(); err != nil {
		return err
	}

	return c.doWithAuth(method, path, body, out, true)
}

func (c *Client) doNoAuth(method, path string, body any, out any) error {
	return c.doWithAuth(method, path, body, out, false)
}

func (c *Client) doWithAuth(method, path string, body any, out any, auth bool) error {
	var reqBody io.Reader

	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.Endpoint+path, reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if auth {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusTemporaryRedirect || resp.StatusCode == http.StatusPermanentRedirect {
		return fmt.Errorf("ASASHV API redirected from %s; use path without trailing slash", path)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ASASHV API error %d: %s", resp.StatusCode, string(respBody))
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("failed to decode ASASHV response from %s: %w; body: %s", path, err, string(respBody))
		}
	}

	return nil
}

type CMClusterNode struct {
	ID        string `json:"id"`
	Hostname  string `json:"hostname"`
	IPAddress string `json:"ip_address"`
	Status    string `json:"status"`
}

type CMCluster struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	VirtualIP          string          `json:"virtual_ip"`
	Status             string          `json:"status"`
	HAEnabled          bool            `json:"ha_enabled"`
	DRSEnabled         bool            `json:"drs_enabled"`
	SharedStorageID    string          `json:"shared_storage_id"`
	SharedStorageName  string          `json:"shared_storage_name"`
	Nodes              []CMClusterNode `json:"nodes"`
	HostCount          int64           `json:"host_count"`
	VMCount            int64           `json:"vm_count"`
	CPUUsagePercent    float64         `json:"cpu_usage_percent"`
	MemoryUsagePercent float64         `json:"memory_usage_percent"`
}

type CreateCMClusterRequest struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	VirtualIP       string   `json:"virtual_ip"`
	HostIDs         []string `json:"host_ids"`
	SharedStorageID string   `json:"shared_storage_id"`
	HAEnabled       bool     `json:"ha_enabled"`
	DRSEnabled      bool     `json:"drs_enabled"`
}

func (c *Client) ListCMClusters() ([]CMCluster, error) {
	var out APIResponse[[]CMCluster]
	err := c.do(http.MethodGet, "/api/ha/clusters", nil, &out)
	return out.Data, err
}

func (c *Client) CreateCMCluster(input CreateCMClusterRequest) (*CMCluster, error) {
	var out APIResponse[CMCluster]
	if err := c.do(http.MethodPost, "/api/ha/clusters", input, &out); err != nil {
		return nil, err
	}

	return &out.Data, nil
}

func (c *Client) DeleteCMCluster(clusterID string) error {
	path := "/api/ha/clusters/" + clusterID

	var out any
	return c.do(http.MethodDelete, path, nil, &out)
}

type CMVSNetworkHost struct {
	ID         string   `json:"id"`
	Hostname   string   `json:"hostname"`
	IPAddress  string   `json:"ip_address"`
	Interfaces []string `json:"interfaces"`
}

type CMVSNetwork struct {
	ID              string            `json:"id"`
	BridgeName      string            `json:"bridge_name"`
	BondMode        string            `json:"bond_mode"`
	BondName        string            `json:"bond_name"`
	ClusterID       string            `json:"cluster_id"`
	ClusterName     string            `json:"cluster_name"`
	Status          string            `json:"status"`
	Description     string            `json:"description"`
	Hosts           []CMVSNetworkHost `json:"hosts"`
	PhysicalUplinks []string          `json:"physical_uplinks"`
}

type CreateCMVSNetworkRequest struct {
	BridgeName  string              `json:"bridge_name"`
	BondMode    string              `json:"bond_mode"`
	ClusterID   string              `json:"cluster_id"`
	HostIDs     []string            `json:"host_ids"`
	Interfaces  map[string][]string `json:"interfaces"`
	Description string              `json:"description,omitempty"`
}

type CMVSNetworkValidationData struct {
	Valid            bool     `json:"valid"`
	Errors           []string `json:"errors"`
	Warnings         []string `json:"warnings"`
	SkipHVValidation bool     `json:"skip_hv_validation"`
}

func (c *Client) ListCMVSNetworks() ([]CMVSNetwork, error) {
	var out APIResponse[[]CMVSNetwork]
	err := c.do(http.MethodGet, "/api/network", nil, &out)
	return out.Data, err
}

func (c *Client) ValidateCMVSNetwork(input CreateCMVSNetworkRequest) (*CMVSNetworkValidationData, error) {
	var out APIResponse[CMVSNetworkValidationData]
	if err := c.do(http.MethodPost, "/api/network/validate", input, &out); err != nil {
		return nil, err
	}

	return &out.Data, nil
}

func (c *Client) CreateCMVSNetwork(input CreateCMVSNetworkRequest) (*CMVSNetwork, error) {
	var out APIResponse[CMVSNetwork]
	if err := c.do(http.MethodPost, "/api/network", input, &out); err != nil {
		return nil, err
	}

	return &out.Data, nil
}

func (c *Client) DeleteCMVSNetwork(networkID string) error {
	path := "/api/network/" + networkID

	var out any
	return c.do(http.MethodDelete, path, nil, &out)
}

type CreateCMHostRequest struct {
	IPAddress   string `json:"ipAddress"`
	SSHUsername string `json:"sshUsername"`
	SSHPassword string `json:"sshPassword"`
}

type CreateCMHostResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Host   CMHost `json:"host"`
}

func (c *Client) CreateCMHost(input CreateCMHostRequest) (*CMHost, error) {
	var out APIResponse[CreateCMHostResponse]
	if err := c.do(http.MethodPost, "/api/dashboard/hosts", input, &out); err != nil {
		return nil, err
	}

	if out.Data.Host.ID != "" {
		return &out.Data.Host, nil
	}

	return &CMHost{
		ID:        out.Data.ID,
		IPAddress: input.IPAddress,
		Status:    out.Data.Status,
	}, nil
}

func (c *Client) DeleteCMHost(hostID string) error {
	path := "/api/dashboard/hosts/" + hostID

	var out any
	return c.do(http.MethodDelete, path, nil, &out)
}
