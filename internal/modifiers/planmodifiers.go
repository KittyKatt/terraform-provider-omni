package planmodifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gopkg.in/yaml.v3"
)

func CompareYAMLStringsModifier() planmodifier.String {
	return &compareYAMLStringsModifier{}
}

type compareYAMLStringsModifier struct {
}

func (d *compareYAMLStringsModifier) Description(ctx context.Context) string {
	return "Ensures that attribute_one and attribute_two attributes are kept synchronised."
}

func (d *compareYAMLStringsModifier) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d *compareYAMLStringsModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.State.Raw.IsNull() {
		return
	}

	if !req.PlanValue.IsUnknown() {
		return
	}

	if req.ConfigValue.IsUnknown() {
		return
	}

	var planValueYAMLDecoded map[string]any
	var stateValueYAMLDecoded map[string]any

	if err := yaml.Unmarshal([]byte(req.StateValue.ValueString()), &stateValueYAMLDecoded); err != nil {
		resp.Diagnostics.AddError("yaml unmarshal error", "")
	}
	if err := yaml.Unmarshal([]byte(resp.PlanValue.ValueString()), &planValueYAMLDecoded); err != nil {
		resp.Diagnostics.AddError("yaml unmarshal error", "")
	}

	tflog.Debug(ctx, "==================================")
	tflog.Debug(ctx, fmt.Sprintf("plan value unmarshaled:\n%s", planValueYAMLDecoded))
	tflog.Debug(ctx, fmt.Sprintf("state value unmarshaled:\n%s", stateValueYAMLDecoded))
	tflog.Debug(ctx, "==================================")
}
