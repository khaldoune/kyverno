package schema

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/golang/glog"
	"gopkg.in/yaml.v2"

	openapi_v2 "github.com/googleapis/gnostic/OpenAPIv2"
	"github.com/xeipuuv/gojsonschema"
)

type Cache struct {
	data    map[string]*gojsonschema.Schema
	mu      sync.RWMutex
	cluster bool
}

type Interface interface {
	loadAll(apidoc *openapi_v2.Document, forceLoad bool)
}

func NewCacheFactory() Interface {
	cache := &Cache{
		// Key : apiversion kind
		data:    map[string]*gojsonschema.Schema{},
		cluster: false,
	}
	return cache
}

func (c *Cache) loadAll(apidoc *openapi_v2.Document, forceLoad bool) {
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

//TODO: return errors
func (c *Cache) validate(data []byte) (bool, error) {
	var spec interface{}
	err := yaml.Unmarshal(data, &spec)
	if err != nil {
		glog.Error(err)
		return false, err
	}
	body := convertToStringKeys(spec)
	if body == nil {
		return false, nil
	}
	cast, _ := body.(map[string]interface{})
	if len(cast) == 0 {
		return false, nil
	}
	documentLoader := gojsonschema.NewGoLoader(body)

	// Kind
	kind, err := determineKind(body)
	if err != nil {
		return false, err
	}
	apiVersion, err := determineAPIVersion(body)
	if err != nil {
		return false, err
	}
	// Get schema
	schema := c.lookup(apiVersion + kind)
	if schema == nil {
		glog.Info("Schema not found")
		return false, fmt.Errorf("Schema not found for document with apiversion %s & kind %s", apiVersion, kind)
	}
	result, err := schema.Validate(documentLoader)
	if err != nil {
		return false, err
	}
	// display result
	for _, e := range result.Errors() {
		fmt.Println(e)
	}

	// get the result
	return result.Valid(), nil
}

func (c *Cache) lookup(avk string) *gojsonschema.Schema {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if schema, ok := c.data[avk]; ok {
		return schema
	}
	glog.Infof("schema not found for %s", avk)
	return nil
}
