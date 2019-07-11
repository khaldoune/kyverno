package schema

import "testing"

func TestStaticDepoloymentSchemaLoad(t *testing.T) {
	schemaData := []byte(`{
		"x-kubernetes-group-version-kind": [
		  {
			"kind": "Deployment",
			"version": "v1beta1",
			"group": "apps"
		  }
		],
		"$schema": "http://json-schema.org/schema#",
		"type": "object",
		"description": "Deployment enables declarative updates for Pods and ReplicaSets.",
		"properties": {
		  "status": {
			"description": "Most recently observed status of the Deployment.",
			"$ref": "https://raw.githubusercontent.com/garethr/kubernetes-json-schema/master/v1.7.16/_definitions.json#/definitions/io.k8s.kubernetes.pkg.apis.apps.v1beta1.DeploymentStatus"
		  },
		  "kind": {
			"type": [
			  "string",
			  "null"
			],
			"description": "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds"
		  },
		  "spec": {
			"description": "Specification of the desired behavior of the Deployment.",
			"$ref": "https://raw.githubusercontent.com/garethr/kubernetes-json-schema/master/v1.7.16/_definitions.json#/definitions/io.k8s.kubernetes.pkg.apis.apps.v1beta1.DeploymentSpec"
		  },
		  "apiVersion": {
			"type": [
			  "string",
			  "null"
			],
			"description": "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources"
		  },
		  "metadata": {
			"description": "Standard object metadata.",
			"$ref": "https://raw.githubusercontent.com/garethr/kubernetes-json-schema/master/v1.7.16/_definitions.json#/definitions/io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta"
		  }
		}
	  }`)

	_, err := loadSchema(string(schemaData))
	if err != nil {
		t.Errorf("Unable to load schema. err %s ", err)
	}
}

func TestDynamicDepolymentLoad(t *testing.T) {
	//	schemaData := []byte(`{"description":"Deployment enables declarative updates for Pods and ReplicaSets.","type":{"value":["object"]},"properties":{"additional_properties":[{"name":"apiVersion","value":{"description":"APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources","type":{"value":["string"]}}},{"name":"kind","value":{"description":"Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds","type":{"value":["string"]}}},{"name":"metadata","value":{"_ref":"#/definitions/io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta","description":"Standard object metadata."}},{"name":"spec","value":{"_ref":"#/definitions/io.k8s.api.apps.v1.DeploymentSpec","description":"Specification of the desired behavior of the Deployment."}},{"name":"status","value":{"_ref":"#/definitions/io.k8s.api.apps.v1.DeploymentStatus","description":"Most recently observed status of the Deployment."}}]},"vendor_extension":[{"name":"x-kubernetes-group-version-kind","value":{"yaml":"- group: apps\n  kind: Deployment\n  version: v1\n"}}]}`)
	schemaData := []byte(`{"additional_properties":[{"name":"kind","value":{"description":"Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds","type":{"value":["string"]}}},{"name":"metadata","value":{"_ref":"#/definitions/io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta","description":"Standard object metadata."}},{"name":"spec","value":{"_ref":"#/definitions/io.k8s.api.apps.v1.DeploymentSpec","description":"Specification of the desired behavior of the Deployment."}},{"name":"status","value":{"_ref":"#/definitions/io.k8s.api.apps.v1.DeploymentStatus","description":"Most recently observed status of the Deployment."}},{"name":"apiVersion","value":{"description":"APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources","type":{"value":["string"]}}}]}`)
	_, err := loadSchema(string(schemaData))
	if err != nil {
		t.Errorf("Unable to load schema. err %s ", err)
	}
}
