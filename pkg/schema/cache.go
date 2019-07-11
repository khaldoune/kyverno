package schema

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/golang/glog"

	openapi_v2 "github.com/googleapis/gnostic/OpenAPIv2"
	"github.com/xeipuuv/gojsonschema"
)

type Cache struct {
	data    map[string]*gojsonschema.Schema
	mu      sync.RWMutex
	cluster bool
}

type Interface interface {
	LoadAll(apidoc *openapi_v2.Document, forceLoad bool)
}

func (c Cache) LoadAll(apidoc *openapi_v2.Document, forceLoad bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Load the schema from the kube-api server openAPI spec document
	for _, s := range apidoc.Definitions.AdditionalProperties {
		// Reload-preloaded
		if !forceLoad {
			if _, ok := c.data[s.Name]; ok {
				continue
			}
		}
		// Parse JSON
		data, err := json.Marshal(s.Value.GetProperties())
		if err != nil {
			glog.Error("Unable to parse")
			continue
		}
		// Load schema
		schema, err := loadSchema(string(data))
		if err != nil {
			glog.Error("Unable to load Schema")
			continue
		}
		c.data[s.Name] = schema
	}
}

func loadSchema(source string) (*gojsonschema.Schema, error) {
	body := convertToStringKeys(source)
	bodystr := fmt.Sprintf("%v", body)
	// We need to remove the key:value pair //"type": "object"
	//	schemasource, err := removeType(bodystr)
	schemaLoader := gojsonschema.NewStringLoader(bodystr)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("Failed initalizing schema: err %s", err)
	}
	return schema, nil
}
