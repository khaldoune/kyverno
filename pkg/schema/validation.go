package schema

import (
	"github.com/amzuko/gojsonschema"
	client "github.com/nirmata/kyverno/pkg/dclient"
)

type Validator struct {
	client *client.Client
	cache  Interface
}

func NewValidator(client *client.Client) *Validator {
	initializerLoader()

	return &Validator{
		client: client,
		cache:  NewCacheFactory()}
}

type Validate interface {
}

func (v *Validator) load(force bool) {
	// Load schema from the kube-api server regsitered schemas
	v.cache.loadAll(v.client.DiscoveryClient.OpenAPISchema(), force)
}

func initializerLoader() {
	// Without forcing these types the schema fails to load
	// Need to Work out proper handling for these types
	gojsonschema.FormatCheckers.Add("int64", ValidFormat{})
	gojsonschema.FormatCheckers.Add("byte", ValidFormat{})
	gojsonschema.FormatCheckers.Add("int32", ValidFormat{})
	gojsonschema.FormatCheckers.Add("int-or-string", ValidFormat{})
}
