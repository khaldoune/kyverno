package schema

import (
	"errors"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// Based on https://stackoverflow.com/questions/40737122/convert-yaml-to-json-without-struct-golang
// We unmarshal yaml into a value of type interface{},
// go through the result recursively, and convert each encountered
// map[interface{}]interface{} to a map[string]interface{} value
// required to marshall to JSON.
func convertToStringKeys(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convertToStringKeys(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convertToStringKeys(v)
		}
	}
	return i
}
func loadSchema(source string) (*gojsonschema.Schema, error) {
	body := convertToStringKeys(source)
	bodystr := fmt.Sprintf("%v", body)
	schemaLoader := gojsonschema.NewStringLoader(bodystr)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("Failed initalizing schema: err %s", err)
	}
	return schema, nil
}

// ValidFormat is a type for quickly forcing
// new formats on the gojsonschema loader
type ValidFormat struct{}

// IsFormat always returns true and meets the
// gojsonschema.FormatChecker interface
func (f ValidFormat) IsFormat(input string) bool {
	return true
}

func determineKind(body interface{}) (string, error) {
	cast, _ := body.(map[string]interface{})
	if _, ok := cast["kind"]; !ok {
		return "", errors.New("Missing a kind key")
	}
	if cast["kind"] == nil {
		return "", errors.New("Missing a kind value")
	}
	return cast["kind"].(string), nil
}
func determineAPIVersion(body interface{}) (string, error) {
	cast, _ := body.(map[string]interface{})
	if _, ok := cast["apiVersion"]; !ok {
		return "", errors.New("Missing a apiVersion key")
	}
	if cast["apiVersion"] == nil {
		return "", errors.New("Missing a apiVersion value")
	}
	return cast["apiVersion"].(string), nil
}
