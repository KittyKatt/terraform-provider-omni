package provider

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/cosi-project/runtime/pkg/safe"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/siderolabs/gen/xslices"
	"github.com/siderolabs/omni/client/pkg/client"
	"github.com/siderolabs/omni/client/pkg/omni/resources"
	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"
)

const (
	controlPlaneNumericID = 1
	workerNumericID       = 2
)

type omniMachineStatusDataSource struct {
	omniClient *client.Client
}

type omniMachineStatusHardwareProcessor struct {
	CoreCount    types.Number `tfsdk:"corecount"`
	ThreadCount  types.Number `tfsdk:"threadcount"`
	Frequency    types.Number `tfsdk:"frequency"`
	Description  types.String `tfsdk:"description"`
	Manufacturer types.String `tfsdk:"manufacturer"`
}

type omniMachineStatusHardwareMemoryModules struct {
	SizeMB      types.Number `tfsdk:"sizemb"`
	Description types.String `tfsdk:"description"`
}

type omniMachineStatusHardwareBlockDevices struct {
	Size       types.Number `tfsdk:"size"`
	Model      types.String `tfsdk:"model"`
	LinuxName  types.String `tfsdk:"linuxname"`
	Name       types.String `tfsdk:"name"`
	Serial     types.String `tfsdk:"serial"`
	UUID       types.String `tfsdk:"uuid"`
	WWID       types.String `tfsdk:"wwid"`
	Type       types.String `tfsdk:"type"`
	BusPath    types.String `tfsdk:"buspath"`
	SystemDisk types.Bool   `tfsdk:"systemdisk"`
	ReadOnly   types.Bool   `tfsdk:"readonly"`
	Transport  types.String `tfsdk:"transport"`
}

type omniMachineStatusHardware struct {
	Processors    *[]omniMachineStatusHardwareProcessor     `tfsdk:"processors"`
	MemoryModules *[]omniMachineStatusHardwareMemoryModules `tfsdk:"memorymodules"`
	BlockDevices  *[]omniMachineStatusHardwareBlockDevices  `tfsdk:"blockdevices"`
	Arch          types.String                              `tfsdk:"arch"`
}

type omniMachineStatusNetwork struct {
	Hostname   types.String   `tfsdk:"hostname"`
	DomainName types.String   `tfsdk:"domainname"`
	Addresses  []types.String `tfsdk:"addresses"`
}

type omniMachineStatusSearchFilters struct {
	Labels      map[string]string `tfsdk:"labels"`
	ImageLabels []string          `tfsdk:"image_labels"`
	Cluster     types.String      `tfsdk:"cluster"`
	ID          types.String      `tfsdk:"id"`
	Connected   types.Bool        `tfsdk:"connected"`
	Maintenance types.Bool        `tfsdk:"maintenance"`
	Role        types.String      `tfsdk:"role"`
}

type machineInfo struct {
	Namespace         types.String               `tfsdk:"namespace"`
	Type              types.String               `tfsdk:"type"`
	ID                types.String               `tfsdk:"id"`
	Phase             types.String               `tfsdk:"phase"`
	Created           types.String               `tfsdk:"created"`
	Updated           types.String               `tfsdk:"updated"`
	Labels            types.Map                  `tfsdk:"labels"`
	TalosVersion      types.String               `tfsdk:"talosversion"`
	Hardware          *omniMachineStatusHardware `tfsdk:"hardware"`
	Network           *omniMachineStatusNetwork  `tfsdk:"network"`
	ManagementAddress types.String               `tfsdk:"managementaddress"`
	Connected         types.Bool                 `tfsdk:"connected"`
	Maintenance       types.Bool                 `tfsdk:"maintenance"`
	Cluster           types.String               `tfsdk:"cluster"`
	Role              types.String               `tfsdk:"role"`
	ImageLabels       types.Map                  `tfsdk:"imagelabels"`
}

type omniMachineStatusDataSourceModelV0 struct {
	Filters      *omniMachineStatusSearchFilters `tfsdk:"filters"`
	MachinesInfo []machineInfo                   `tfsdk:"machines"`
}

var _ datasource.DataSource = &omniMachineStatusDataSource{}

func NewOmniMachineStatusDataSource() datasource.DataSource {
	return &omniMachineStatusDataSource{}
}

func (d *omniMachineStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine_status"
}

func (d *omniMachineStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a list of Omni machine statuses.",
		Attributes: map[string]schema.Attribute{
			"filters": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"labels": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "A map of labels to filter machines by. Only machines with matching labels will be returned.",
					},
					"image_labels": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "A list of image label keys to filter machines by. Only machines with matching label keys will be returned.",
					},
					"cluster": schema.StringAttribute{
						Optional:    true,
						Description: "The cluster to filter machines by. Only machines in the specified cluster will be returned.",
					},
					"connected": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether to filter machines by their connected status. If true, only connected machines will be returned. If false, only disconnected machines will be returned.",
					},
					"maintenance": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether to filter machines by their maintenance status. If true, only machines in maintenance mode will be returned. If false, only machines not in maintenance mode will be returned.",
					},
					"id": schema.StringAttribute{
						Optional:    true,
						Description: "The ID to filter machines by. Only the machine with the specified ID will be returned.",
					},
					"role": schema.StringAttribute{
						Optional:    true,
						Description: "The role to filter machines by. Only machines with the specified role will be returned.",
						Validators: []validator.String{
							stringvalidator.OneOf("controlplane", "worker"),
						},
					},
				},
				Optional:    true,
				Description: "Filters to apply when retrieving machines. If not specified, all machines will be returned.",
			},
			"machines": schema.ListAttribute{
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"namespace":    types.StringType,
						"type":         types.StringType,
						"id":           types.StringType,
						"phase":        types.StringType,
						"created":      types.StringType,
						"updated":      types.StringType,
						"labels":       types.MapType{ElemType: types.StringType},
						"talosversion": types.StringType,
						"hardware": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"processors": types.ListType{ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"corecount":    types.NumberType,
										"threadcount":  types.NumberType,
										"frequency":    types.NumberType,
										"description":  types.StringType,
										"manufacturer": types.StringType,
									},
								}},
								"memorymodules": types.ListType{ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"sizemb":      types.NumberType,
										"description": types.StringType,
									},
								}},
								"blockdevices": types.ListType{ElemType: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"size":       types.NumberType,
										"model":      types.StringType,
										"linuxname":  types.StringType,
										"name":       types.StringType,
										"serial":     types.StringType,
										"uuid":       types.StringType,
										"wwid":       types.StringType,
										"type":       types.StringType,
										"buspath":    types.StringType,
										"systemdisk": types.BoolType,
										"readonly":   types.BoolType,
										"transport":  types.StringType,
									},
								}},
								"arch": types.StringType,
							},
						},
						"network": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"hostname":   types.StringType,
								"domainname": types.StringType,
								"addresses":  types.ListType{ElemType: types.StringType},
							},
						},
						"managementaddress": types.StringType,
						"connected":         types.BoolType,
						"maintenance":       types.BoolType,
						"cluster":           types.StringType,
						"role":              types.StringType,
						"imagelabels":       types.MapType{ElemType: types.StringType},
					},
				},
				Computed:    true,
				Description: "A list of machine statuses. Each machine status contains detailed information about the machine, including its hardware, network, and management details.",
			},
		},
	}
}

func (d *omniMachineStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	omniClient, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"failed to get image factory client",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.omniClient = omniClient
}

func (d *omniMachineStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.omniClient == nil {
		resp.Diagnostics.AddError("omni client is not configured", "Please report this issue to the provider developers.")
		return
	}

	st := d.omniClient.Omni().State()

	var config omniMachineStatusDataSourceModelV0

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	machines, err := safe.StateList[*omni.MachineStatus](ctx, st, omni.NewMachineStatus(resources.DefaultNamespace, "").Metadata())
	if err != nil {
		resp.Diagnostics.AddError("failed to get omni machine statuses", err.Error())
		return
	}

	var machinesSlice []*omni.MachineStatus
	machines.All()(func(m *omni.MachineStatus) bool {
		machinesSlice = append(machinesSlice, m)
		return true
	})

	tflog.Debug(ctx, fmt.Sprintf("number of machines found: %d", len(machinesSlice)))

	if config.Filters != nil {
		machinesSlice = xslices.Filter(machinesSlice, func(e *omni.MachineStatus) bool {
			if config.Filters.Labels != nil {
				for k, v := range config.Filters.Labels {
					labelValue, ok := e.Metadata().Labels().Get(k)
					if !ok || labelValue != v {
						return false
					}
				}
			}

			if config.Filters.ImageLabels != nil {
				imageLabels := e.TypedSpec().Value.ImageLabels
				for _, v := range config.Filters.ImageLabels {
					_, exists := imageLabels[v]
					if !exists {
						return false
					}
				}
			}

			if !config.Filters.Cluster.IsNull() && types.StringValue(e.TypedSpec().Value.GetCluster()) != config.Filters.Cluster {
				return false
			}

			if !config.Filters.Connected.IsNull() && e.TypedSpec().Value.Connected != config.Filters.Connected.ValueBool() {
				return false
			}

			if !config.Filters.Maintenance.IsNull() && e.TypedSpec().Value.Maintenance != config.Filters.Maintenance.ValueBool() {
				return false
			}

			if !config.Filters.ID.IsNull() && types.StringValue(e.Metadata().ID()) != config.Filters.ID {
				return false
			}

			if !config.Filters.Role.IsNull() {
				role := e.TypedSpec().Value.Role
				if config.Filters.Role.ValueString() == "controlplane" && role != controlPlaneNumericID {
					return false
				}
				if config.Filters.Role.ValueString() == "worker" && role != workerNumericID {
					return false
				}
			}

			return true
		})
	}

	tfMachinesInfo := xslices.Map(machinesSlice, func(e *omni.MachineStatus) machineInfo {
		return machineInfo{
			Namespace: types.StringValue(e.Metadata().Namespace()),
			Type:      types.StringValue(e.Metadata().Type()),
			ID:        types.StringValue(e.Metadata().ID()),
			Phase:     types.StringValue(e.Metadata().Phase().String()),
			Created:   types.StringValue(e.Metadata().Created().Format(time.RFC3339)),
			Updated:   types.StringValue(e.Metadata().Updated().Format(time.RFC3339)),
			Labels: func() types.Map {
				labelMap := make(map[string]attr.Value)
				labels := e.Metadata().Labels().Raw()
				for k, v := range labels {
					labelMap[k] = types.StringValue(v)
				}
				mapVal, _ := types.MapValue(types.StringType, labelMap)
				return mapVal
			}(),
			TalosVersion: types.StringValue(e.TypedSpec().Value.TalosVersion),
			Hardware: &omniMachineStatusHardware{
				Arch: types.StringValue(e.TypedSpec().Value.Hardware.Arch),
				Processors: func() *[]omniMachineStatusHardwareProcessor {
					src := e.TypedSpec().Value.Hardware.Processors
					if src == nil {
						return nil
					}
					out := make([]omniMachineStatusHardwareProcessor, len(src))
					for i, p := range src {
						out[i] = omniMachineStatusHardwareProcessor{
							CoreCount:    types.NumberValue(new(big.Float).SetFloat64(float64(p.CoreCount))),
							ThreadCount:  types.NumberValue(new(big.Float).SetFloat64(float64(p.ThreadCount))),
							Frequency:    types.NumberValue(new(big.Float).SetFloat64(float64(p.Frequency))),
							Description:  types.StringValue(p.Description),
							Manufacturer: types.StringValue(p.Manufacturer),
						}
					}
					return &out
				}(),
				MemoryModules: func() *[]omniMachineStatusHardwareMemoryModules {
					src := e.TypedSpec().Value.Hardware.MemoryModules
					if src == nil {
						return nil
					}
					out := make([]omniMachineStatusHardwareMemoryModules, len(src))
					for i, m := range src {
						out[i] = omniMachineStatusHardwareMemoryModules{
							SizeMB:      types.NumberValue(new(big.Float).SetFloat64(float64(m.SizeMb))),
							Description: types.StringValue(m.Description),
						}
					}
					return &out
				}(),
				BlockDevices: func() *[]omniMachineStatusHardwareBlockDevices {
					src := e.TypedSpec().Value.Hardware.Blockdevices
					if src == nil {
						return nil
					}
					out := make([]omniMachineStatusHardwareBlockDevices, len(src))
					for i, d := range src {
						out[i] = omniMachineStatusHardwareBlockDevices{
							Size:       types.NumberValue(new(big.Float).SetFloat64(float64(d.Size))),
							Model:      types.StringValue(d.Model),
							LinuxName:  types.StringValue(d.LinuxName),
							Name:       types.StringValue(d.Name),
							Serial:     types.StringValue(d.Serial),
							UUID:       types.StringValue(d.Uuid),
							WWID:       types.StringValue(d.Wwid),
							Type:       types.StringValue(d.Type),
							BusPath:    types.StringValue(d.BusPath),
							SystemDisk: types.BoolValue(d.SystemDisk),
							ReadOnly:   types.BoolValue(d.Readonly),
							Transport:  types.StringValue(d.Transport),
						}
					}
					return &out
				}(),
			},
			Network: &omniMachineStatusNetwork{
				Hostname:   types.StringValue(e.TypedSpec().Value.Network.Hostname),
				DomainName: types.StringValue(e.TypedSpec().Value.Network.Domainname),
				Addresses: func() []types.String {
					src := e.TypedSpec().Value.Network.Addresses
					out := make([]types.String, len(src))
					for i, addr := range src {
						out[i] = types.StringValue(addr)
					}
					return out
				}(),
			},
			ManagementAddress: types.StringValue(e.TypedSpec().Value.ManagementAddress),
			Connected:         types.BoolValue(e.TypedSpec().Value.Connected),
			Maintenance:       types.BoolValue(e.TypedSpec().Value.Maintenance),
			Cluster:           types.StringValue(e.TypedSpec().Value.Cluster),
			Role:              types.StringValue(fmt.Sprintf("%d", e.TypedSpec().Value.Role)),
			ImageLabels: func() types.Map {
				labelMap := make(map[string]attr.Value)
				for k, v := range e.TypedSpec().Value.ImageLabels {
					labelMap[k] = types.StringValue(v)
				}
				mapVal, _ := types.MapValue(types.StringType, labelMap)
				return mapVal
			}(),
		}
	})

	config.MachinesInfo = tfMachinesInfo

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
