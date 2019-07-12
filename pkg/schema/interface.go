package schema

type Interface interface {
	// loadAll(apidoc *openapi_v2.Document, forceLoad bool)
	validate(data []byte) (bool, error)
}
