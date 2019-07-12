package schema

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/nirmata/kyverno/pkg/schema/utils"

	"github.com/golang/glog"
	client "github.com/nirmata/kyverno/pkg/dclient"
	"gopkg.in/yaml.v2"

	openapi_v2 "github.com/googleapis/gnostic/OpenAPIv2"
	"github.com/xeipuuv/gojsonschema"
)

type ClusterCache struct {
	data   map[string]*gojsonschema.Schema
	client *client.Client
	mu     sync.RWMutex
}

func NewClusterCacheFactory(client *client.Client) *ClusterCache {

	cache := &ClusterCache{
		// Key : apiversion kind
		data:   map[string]*gojsonschema.Schema{},
		client: client,
	}
	return cache
}

func (c *ClusterCache) loadAll(apidoc *openapi_v2.Document, forceLoad bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Load the schema from the kube-api server openAPI spec document
	for _, s := range apidoc.Definitions.AdditionalProperties {

		// // Reload-preloaded
		// if !forceLoad {
		// 	if _, ok := c.data[s.Name]; ok {
		// 		continue
		// 	}
		// }
		// Parse JSON
		// data, err := json.Marshal(s.Value.GetProperties())
		// if err != nil {
		// 	glog.Error("Unable to parse")
		// 	continue
		// }
		// Split Name
		name := strings.Split(s.Name, ".")
		kind := name[len(name)-1]
		// Load schema
		fmt.Println(s.Name)
		fmt.Println(kind)
		// schema, err := loadSchema(string(data))
		// if err != nil {
		// 	glog.Error("Unable to load Schema")
		// 	return nil
		// }

		// c.data[kind] = schema
	}
}

func lookUp(apidoc *openapi_v2.Document, kind string, apiversion string) *gojsonschema.Schema {
	for _, s := range apidoc.Definitions.AdditionalProperties {
		name := strings.Split(s.Name, ".")
		k := name[len(name)-1]
		v := name[len(name)-2]
		g := name[len(name)-3]
		apiv := strings.Split(apiversion, "/")
		group := apiv[0]
		version := apiv[1]
		if kind == k {
			fmt.Println(k)
			fmt.Println(group)
			fmt.Println(g)
			fmt.Println(version)
			fmt.Println(v)
		}
		if kind == k && g == group && v == version {
			data, err := json.Marshal(s.Value.GetProperties())
			if err != nil {
				glog.Error("Unable to parse")
				return nil
			}
			schema, err := utils.LoadSchema(string(data))
			if err != nil {
				glog.Error("Unable to load Schema")
				return nil
			}
			return schema
		}
	}
	return nil
}

func loadDocument(data []byte) (string, string, gojsonschema.JSONLoader, error) {
	var spec interface{}

	err := yaml.Unmarshal(data, &spec)
	if err != nil {
		glog.Error(err)
		return "", "", nil, err
	}
	body := utils.ConvertToStringKeys(spec)
	if body == nil {
		return "", "", nil, nil
	}
	cast, _ := body.(map[string]interface{})
	if len(cast) == 0 {
		return "", "", nil, fmt.Errorf("type mistmatch. expected map[string]interface{}")
	}
	documentLoader := gojsonschema.NewGoLoader(body)
	kind, err := utils.DetermineKind(body)
	if err != nil {
		return "", "", nil, err
	}
	apiversion, err := utils.DetermineAPIVersion(body)
	if err != nil {
		return "", "", nil, err
	}
	return kind, apiversion, documentLoader, nil
}

//TODO: return errors
func (c *ClusterCache) validate(data []byte) (bool, error) {
	kind, apiversion, documentLoader, err := loadDocument(data)
	if err != nil {
		return false, err
	}
	// Get schema
	schema := c.lookup(kind, apiversion, c.client.DiscoveryClient.OpenAPISchema())
	if schema == nil {
		glog.Info("Schema not found")
		return false, fmt.Errorf("Schema not found for document with kind %s", kind)
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

func (c *ClusterCache) lookup(kind string, apiversion string, apidoc *openapi_v2.Document) *gojsonschema.Schema {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var schema *gojsonschema.Schema
	apiv := strings.Split(apiversion, "/")
	group := apiv[0]
	version := apiv[1]
	key := kind + "." + group + "." + version
	schema, ok := c.data[key]
	if !ok {
		schema = lookUp(apidoc, kind, apiversion)
		if schema != nil {
			c.data[key] = schema
		}
	}
	return schema
}
