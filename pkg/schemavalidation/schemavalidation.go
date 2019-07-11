package schemavalidation

import client "github.com/nirmata/kyverno/pkg/dclient"

//SchemaValidator provides functionaly to valid resource against a schema
type SchemaValidator struct {
	client *client.Client
}

func NewSchemaValidator(client *client.Client) *SchemaValidator {
	return &SchemaValidator{
		client: client}
}

type Validation interface {
}
