package models

type MachineSetYAML struct {
	Kind             string            `yaml:"kind"`
	SystemExtensions []string          `yaml:"systemExtensions,omitempty"`
	Name             string            `yaml:"name,omitempty"`
	Labels           map[string]string `yaml:"labels,omitempty"`
	Annotations      map[string]string `yaml:"annotations,omitempty"`
	Machines         []string          `yaml:"machines"`
	Patches          []PatchYAML       `yaml:"patches,omitempty"`
}
