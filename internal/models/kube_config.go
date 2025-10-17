package models

type KubeConfigCluster struct {
	Cluster struct {
		Server string `tfsdk:"server" yaml:"server"`
	}
	Name string `tfsdk:"name" yaml:"name"`
}

type KubeConfigContext struct {
	Context struct {
		Cluster   string `tfsdk:"cluster" yaml:"cluster"`
		Namespace string `tfsdk:"namespace" yaml:"namespace"`
		User      string `tfsdk:"user" yaml:"user"`
	}
	Name string `tfsdk:"name" yaml:"name"`
}

type KubeConfigUser struct {
	Name string `tfsdk:"name" yaml:"name"`
	User struct {
		Token string `tfsdk:"token" yaml:"token"`
	}
}

type KubeConfig struct {
	Clusters []KubeConfigCluster `tfsdk:"clusters" yaml:"clusters"`
	Contexts []KubeConfigContext `tfsdk:"contexts" yaml:"contexts"`
	Users    []KubeConfigUser    `tfsdk:"users" yaml:"users"`
}
