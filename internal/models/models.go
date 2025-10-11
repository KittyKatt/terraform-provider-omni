package models

type Metadata struct {
	Labels      Labels      `tfsdk:"labels"`
	Annotations Annotations `tfsdk:"annotations"`
}

type Labels map[string]string

type Annotations map[string]string

type PatchList []Patch
type Patch struct {
	IDOverride  string      `tfsdk:"id_override" yaml:"idOverride"`
	Labels      Labels      `tfsdk:"labels" yaml:"labels,omitempty"`
	Annotations Annotations `tfsdk:"annotations" yaml:"annotations,omitempty"`
	File        *string     `tfsdk:"file" yaml:"file,omitempty"`
	Inline      *string     `tfsdk:"inline" yaml:"inline,omitempty"`
}
type PatchYAML struct {
	IDOverride  string            `yaml:"idOverride"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
	File        *string           `yaml:"file,omitempty"`
	Inline      map[string]any    `yaml:"inline,omitempty"`
}

type SystemExtensions []string

type Model interface {
	isModel()
}
