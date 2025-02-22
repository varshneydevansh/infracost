package google

import (
	"github.com/infracost/infracost/internal/providers/terraform/provider_schemas"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

var DefaultProviderRegion = "us-central1"

func GetDefaultRefIDFunc(d *schema.ResourceData) []string {

	defaultRefs := []string{d.Get("id").String()}

	if d.Get("self_link").Exists() {
		defaultRefs = append(defaultRefs, d.Get("self_link").String())
	}

	return defaultRefs
}

func DefaultCloudResourceIDFunc(d *schema.ResourceData) []string {
	return []string{}
}

func GetSpecialContext(d *schema.ResourceData) map[string]interface{} {
	return map[string]interface{}{}
}

func GetResourceRegion(resourceType string, v gjson.Result) string {
	if v.Get("region").Exists() && v.Get("region").String() != "" {
		return v.Get("region").String()
	}

	return ""
}

func ParseTags(resourceType string, r gjson.Result) *map[string]string {
	_, supportsTags := provider_schemas.GoogleLabelsSupport[resourceType]
	rLabels := r.Get("labels").Map()
	if !supportsTags && len(rLabels) == 0 {
		return nil
	}

	tags := make(map[string]string)
	for k, v := range rLabels {
		tags[k] = v.String()
	}
	return &tags
}
