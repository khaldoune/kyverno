package schema

import (
	"fmt"
	"strings"
	"sync"

	"github.com/golang/glog"

	"github.com/nirmata/kyverno/pkg/schema/utils"

	client "github.com/nirmata/kyverno/pkg/dclient"
	"github.com/xeipuuv/gojsonschema"
)

type RepoCache struct {
	data map[string]*gojsonschema.Schema
	mu   sync.RWMutex
}

func NewRepoCacheFactory(client *client.Client) *RepoCache {
	cache := &RepoCache{
		// Key : apiversion kind
		data: map[string]*gojsonschema.Schema{},
	}
	return cache
}

//TODO
// default url to rgithub repo
// build the url
// load the schema on request and cache
// lookup from the cache using the step above

func (c *RepoCache) validate(data []byte) (bool, error) {
	kind, apiversion, documentLoader, err := loadDocument(data)
	if err != nil {
		return false, err
	}
	// Get the schema
	schema := c.lookup(kind, apiversion)

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

func (c *RepoCache) lookup(kind string, apiversion string) *gojsonschema.Schema {
	var err error
	c.mu.RLock()
	defer c.mu.RUnlock()
	schemaRef := determineSchema(kind, apiversion)
	schema, ok := c.data[schemaRef]
	if !ok {
		schemaLoader := gojsonschema.NewReferenceLoader(schemaRef)
		schema, err = gojsonschema.NewSchema(schemaLoader)
		if err != nil {
			glog.Error(err)
			return nil
		}
	}
	return schema
}

func determineSchema(kind, apiVersion string) string {
	Version := ""
	if Version == "" {
		Version = "master"
	}
	normalisedVersion := Version
	if Version != "master" {
		normalisedVersion = "v" + normalisedVersion
	}
	var kindSuffix string
	strictSuffix := ""

	baseURL := utils.DefaultSchemaLocation
	groupParts := strings.Split(apiVersion, "/")
	versionParts := strings.Split(groupParts[0], ".")
	if len(groupParts) == 1 {
		kindSuffix = "-" + strings.ToLower(versionParts[0])
	} else {
		kindSuffix = fmt.Sprintf("-%s-%s", strings.ToLower(versionParts[0]), strings.ToLower(groupParts[1]))
	}
	return fmt.Sprintf("%s/%s-standalone%s/%s%s.json", baseURL, normalisedVersion, strictSuffix, strings.ToLower(kind), kindSuffix)

}
