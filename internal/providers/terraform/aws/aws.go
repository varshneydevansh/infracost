package aws

import (
	"github.com/infracost/infracost/internal/providers/terraform/provider_schemas"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

var DefaultProviderRegion = "us-east-1"
var arnAttributeMap = map[string]string{
	"aws_cloudwatch_dashboard":     "dashboard_arn",
	"aws_db_snapshot":              "db_snapshot_arn",
	"aws_db_cluster_snapshot":      "db_cluster_snapshot_arn",
	"aws_ecs_service":              "id",
	"aws_neptune_cluster_snapshot": "db_cluster_snapshot_arn",
	"aws_docdb_cluster_snapshot":   "db_cluster_snapshot_arn",
	"aws_dms_certificate":          "certificate_arn",
	"aws_dms_endpoint":             "endpoint_arn",
	"aws_dms_replication_instance": "replication_instance_arn",
	"aws_dms_replication_task":     "replication_task_arn",
}

func GetDefaultRefIDFunc(d *schema.ResourceData) []string {
	defaultRefs := []string{d.Get("id").String(), d.Get("name").String()}

	arnAttr, ok := arnAttributeMap[d.Type]
	if !ok {
		arnAttr = "arn"
	}

	if d.Get(arnAttr).Exists() {
		defaultRefs = append(defaultRefs, d.Get(arnAttr).String())
	}

	return defaultRefs
}

func DefaultCloudResourceIDFunc(d *schema.ResourceData) []string {
	var ids []string

	id := d.Get("id").String()
	if id != "" && id != "none" && !strings.HasPrefix(id, "hcl-") {
		ids = append(ids, id)
	}

	arnAttr, ok := arnAttributeMap[d.Type]
	if !ok {
		arnAttr = "arn"
	}

	arn := d.Get(arnAttr).String()
	if strings.HasPrefix(arn, "arn:aws:") && !strings.HasPrefix(arn, "arn:aws:hcl") {
		ids = append(ids, arn)
	}

	return ids
}

func GetSpecialContext(d *schema.ResourceData) map[string]interface{} {

	specialContexts := make(map[string]interface{})

	if strings.HasPrefix(d.Get("region").String(), "cn-") {
		specialContexts["isAWSChina"] = true
	}

	return specialContexts
}

func GetResourceRegion(resourceType string, v gjson.Result) string {
	// If a region key exists in the values use that
	if v.Get("region").Exists() && v.Get("region").String() != "" {
		return v.Get("region").String()
	}

	// Otherwise try and parse the ARN from the values
	arnAttr, ok := arnAttributeMap[resourceType]
	if !ok {
		arnAttr = "arn"
	}

	if !v.Get(arnAttr).Exists() {
		return ""
	}

	arn := v.Get(arnAttr).String()
	p := strings.Split(arn, ":")
	if len(p) < 4 {
		log.Debugf("Unexpected ARN format for %s", arn)
		return ""
	}

	return p[3]
}

func ParseTags(defaultTags *map[string]string, resourceType string, r gjson.Result) *map[string]string {
	_, supportsTags := provider_schemas.AWSTagsSupport[resourceType]
	_, supportsTagBlock := provider_schemas.AWSTagBlockSupport[resourceType]

	rTags := r.Get("tags").Map()
	if !supportsTags && !supportsTagBlock && len(rTags) == 0 {
		return nil
	}

	tags := make(map[string]string)

	_, supportsDefaultTags := provider_schemas.AWSTagsAllSupport[resourceType]
	if supportsDefaultTags && defaultTags != nil {
		for k, v := range *defaultTags {
			tags[k] = v
		}
	}

	if supportsTagBlock {
		for _, el := range r.Get("tag").Array() {
			k := el.Get("key").String()
			if k != "" {
				tags[k] = el.Get("value").String()
			}
		}
	}

	for k, v := range rTags {
		tags[k] = v.String()
	}
	return &tags
}
