package models

type MachineIDList []string

type MachineInstall struct {
	Disk string `tfsdk:"disk" yaml:"disk"`
}

type MachinesYAML struct {
	Kind             string            `tfsdk:"kind" yaml:"kind"`
	SystemExtensions []string          `tfsdk:"system_extensions" yaml:"systemExtensions,omitempty"`
	Name             string            `tfsdk:"name" yaml:"name"`
	Labels           map[string]string `tfsdk:"labels" yaml:"labels,omitempty"`
	Annotations      map[string]string `tfsdk:"annotations" yaml:"annotations,omitempty"`
	Locked           bool              `tfsdk:"locked" yaml:"locked,omitempty"`
	Install          map[string]any    `tfsdk:"install" yaml:"install,omitempty"`
	Patches          []PatchYAML       `tfsdk:"patches" yaml:"patches,omitempty"`
}
