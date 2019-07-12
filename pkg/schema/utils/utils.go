package utils

import (
	"errors"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

//kubeval
var DefaultSchemaLocation = "https://kubernetesjsonschema.dev"

// Based on https://stackoverflow.com/questions/40737122/convert-yaml-to-json-without-struct-golang
// We unmarshal yaml into a value of type interface{},
// go through the result recursively, and convert each encountered
// map[interface{}]interface{} to a map[string]interface{} value
// required to marshall to JSON.
func ConvertToStringKeys(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = ConvertToStringKeys(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = ConvertToStringKeys(v)
		}
	}
	return i
}

func LoadSchema(source string) (*gojsonschema.Schema, error) {
	body := ConvertToStringKeys(source)
	bodystr := fmt.Sprintf("%v", body)
	schemaLoader := gojsonschema.NewStringLoader(bodystr)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("Failed initalizing schema: err %s", err)
	}
	// cast, _ := body.(map[string]interface{})
	// if len(cast) == 0 {
	// 	return "", nil, nil
	// }
	// key, err := getKey(body)
	// if err != nil {
	// 	return "", nil, err
	// }

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

func DetermineKind(body interface{}) (string, error) {
	cast, _ := body.(map[string]interface{})
	if _, ok := cast["kind"]; !ok {
		return "", errors.New("Missing a kind key")
	}
	if cast["kind"] == nil {
		return "", errors.New("Missing a kind value")
	}
	return cast["kind"].(string), nil
}
func DetermineAPIVersion(body interface{}) (string, error) {
	cast, _ := body.(map[string]interface{})
	if _, ok := cast["apiVersion"]; !ok {
		return "", errors.New("Missing a apiVersion key")
	}
	if cast["apiVersion"] == nil {
		return "", errors.New("Missing a apiVersion value")
	}
	return cast["apiVersion"].(string), nil
}
