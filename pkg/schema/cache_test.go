package schema

import (
	"fmt"
	"testing"

	"github.com/golang/glog"
	client "github.com/nirmata/kyverno/pkg/dclient"
	schemautils "github.com/nirmata/kyverno/pkg/schema/utils"
	"github.com/nirmata/kyverno/pkg/utils"
)

func TestStaticDepoloymentSchemaLoad(t *testing.T) {
	// schemaData := []byte(`{
	// 	"x-kubernetes-group-version-kind": [
	// 	  {
	// 		"kind": "Deployment",
	// 		"version": "v1beta1",
	// 		"group": "apps"
	// 	  }
	// 	],
	// 	"$schema": "http://json-schema.org/schema#",
	// 	"type": "object",
	// 	"description": "Deployment enables declarative updates for Pods and ReplicaSets.",
	// 	"properties": {
	// 	  "status": {
	// 		"description": "Most recently observed status of the Deployment.",
	// 		"$ref": "https://raw.githubusercontent.com/garethr/kubernetes-json-schema/master/v1.7.16/_definitions.json#/definitions/io.k8s.kubernetes.pkg.apis.apps.v1beta1.DeploymentStatus"
	// 	  },
	// 	  "kind": {
	// 		"type": [
	// 		  "string",
	// 		  "null"
	// 		],
	// 		"description": "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds"
	// 	  },
	// 	  "spec": {
	// 		"description": "Specification of the desired behavior of the Deployment.",
	// 		"$ref": "https://raw.githubusercontent.com/garethr/kubernetes-json-schema/master/v1.7.16/_definitions.json#/definitions/io.k8s.kubernetes.pkg.apis.apps.v1beta1.DeploymentSpec"
	// 	  },
	// 	  "apiVersion": {
	// 		"type": [
	// 		  "string",
	// 		  "null"
	// 		],
	// 		"description": "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources"
	// 	  },
	// 	  "metadata": {
	// 		"description": "Standard object metadata.",
	// 		"$ref": "https://raw.githubusercontent.com/garethr/kubernetes-json-schema/master/v1.7.16/_definitions.json#/definitions/io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta"
	// 	  }
	// 	}
	//   }`)
	schemaData := []byte(`{"additional_properties":[{"name":"kind","value":{"description":"Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds","type":{"value":["string"]}}},{"name":"metadata","value":{"_ref":"#/definitions/io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta","description":"Standard object metadata."}},{"name":"spec","value":{"_ref":"#/definitions/io.k8s.api.apps.v1.DeploymentSpec","description":"Specification of the desired behavior of the Deployment."}},{"name":"status","value":{"_ref":"#/definitions/io.k8s.api.apps.v1.DeploymentStatus","description":"Most recently observed status of the Deployment."}},{"name":"apiVersion","value":{"description":"APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources","type":{"value":["string"]}}}]}`)

	deploy := []byte(`{
		"apiVersion": "extensions/v1beta1",
		"kind": "Deployment",
		"metadata": {
			"annotations": {
				"deployment.kubernetes.io/revision": "1"
			},
			"generation": 1,
			"labels": {
				"app": "nginx_is_mutated",
				"cli": "test",
				"isMutated": "true"
			},
			"name": "nginx-deployment",
			"namespace": "default",
			"resourceVersion": "120124",
			"selfLink": "/apis/extensions/v1beta1/namespaces/default/deployments/nginx-deployment",
			"uid": "cc6e54a9-9c49-11e9-ad3c-0800273eb62d"
		},
		"spec": {
			"progressDeadlineSeconds": 600,
			"replicas": 1,
			"revisionHistoryLimit": 10,
			"selector": {
				"matchLabels": {
					"app": "nginx"
				}
			},
			"strategy": {
				"rollingUpdate": {
					"maxSurge": "25%",
					"maxUnavailable": "25%"
				},
				"type": "RollingUpdate"
			},
			"template": {
				"metadata": {
					"creationTimestamp": null,
					"labels": {
						"app": "nginx"
					}
				},
				"spec": {
					"containers": [
						{
							"image": "nginx:1.7.9",
							"imagePullPolicy": "Always",
							"name": "nginx",
							"ports": [
								{
									"containerPort": 80,
									"protocol": "TCP"
								}
							],
							"resources": {},
							"terminationMessagePath": "/dev/termination-log",
							"terminationMessagePolicy": "File"
						}
					],
					"dnsPolicy": "ClusterFirst",
					"restartPolicy": "Always",
					"schedulerName": "default-scheduler",
					"securityContext": {},
					"terminationGracePeriodSeconds": 30
				}
			}
		},
	}`)

	schema, err := schemautils.LoadSchema(string(schemaData))
	if err != nil {
		t.Errorf("Unable to load schema. err %s ", err)
	}
	_, _, jsonLoader, err := loadDocument(deploy)
	if err != nil {
		fmt.Println(err)
	}

	result, err := schema.Validate(jsonLoader)
	fmt.Println(result.Valid())
	fmt.Println(result.Errors())
}

func TestDynamicDepolymentLoad(t *testing.T) {
	schemaData := []byte(`{"additional_properties":[{"name":"kind","value":{"description":"Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds","type":{"value":["string"]}}},{"name":"metadata","value":{"_ref":"#/definitions/io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta","description":"Standard object metadata."}},{"name":"spec","value":{"_ref":"#/definitions/io.k8s.api.apps.v1.DeploymentSpec","description":"Specification of the desired behavior of the Deployment."}},{"name":"status","value":{"_ref":"#/definitions/io.k8s.api.apps.v1.DeploymentStatus","description":"Most recently observed status of the Deployment."}},{"name":"apiVersion","value":{"description":"APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources","type":{"value":["string"]}}}]}`)
	_, err := schemautils.LoadSchema(string(schemaData))
	if err != nil {
		t.Errorf("Unable to load schema. err %s ", err)
	}
}

// Needs cluster as it loads the registered schemas from kube-api server
func TestSchemaLoadCluster(t *testing.T) {
	deploy := []byte(`{
		"apiVersion": "extensions/v1beta1",
		"kind": "Deployment",
		"metadata": {
			"annotations": {
				"deployment.kubernetes.io/revision": "1"
			},
			"generation": 1,
			"labels": {
				"app": "nginx_is_mutated",
				"cli": "test",
				"isMutated": "true"
			},
			"name": "nginx-deployment",
			"namespace": "default",
			"resourceVersion": "120124",
			"selfLink": "/apis/extensions/v1beta1/namespaces/default/deployments/nginx-deployment",
			"uid": "cc6e54a9-9c49-11e9-ad3c-0800273eb62d"
		},
		"spec": {
			"progressDeadlineSeconds": 600,
			"replicas": 1,
			"revisionHistoryLimit": 10,
			"selector": {
				"matchLabels": {
					"app": "nginx"
				}
			},
			"strategy": {
				"rollingUpdate": {
					"maxSurge": "25%",
					"maxUnavailable": "25%"
				},
				"type": "RollingUpdate"
			},
			"template": {
				"metadata": {
					"creationTimestamp": null,
					"labels": {
						"app": "nginx"
					}
				},
				"spec": {
					"containers": [
						{
							"image": "nginx:1.7.9",
							"imagePullPolicy": "Always",
							"name": "nginx",
							"ports": [
								{
									"containerPort": 80,
									"protocol": "TCP"
								}
							],
							"resources": {},
							"terminationMessagePath": "/dev/termination-log",
							"terminationMessagePolicy": "File"
						}
					],
					"dnsPolicy": "ClusterFirst",
					"restartPolicy": "Always",
					"schedulerName": "default-scheduler",
					"securityContext": {},
					"terminationGracePeriodSeconds": 30
				}
			}
		},
	}`)

	clientConfig, err := utils.CreateClientConfig("/Users/shivd/.kube/config")
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v\n", err)
	}

	client, err := client.NewClient(clientConfig)
	if err != nil {
		glog.Fatalf("Error creating client: %v\n", err)
	}

	// Lets load the schema
	schemaValidator := NewValidator(client, true)
	result, err := schemaValidator.validate(deploy)
	if err != nil {
		t.Error(err)
	}
	if !result {
		t.Error("Document does not match the schema")
	}
	// load
	result, err = schemaValidator.validate(deploy)
	if err != nil {
		t.Error(err)
	}
	if !result {
		t.Error("Document does not match the schema")
	}

}

func TestSchemaLoadRepo(t *testing.T) {
	deploy := []byte(`{
		"apiVersion": "extensions/v1beta1",
		"kind": "Deployment",
		"metadata": {
			"annotations": {
				"deployment.kubernetes.io/revision": "1"
			},
			"generation": 1,
			"labels": {
				"app": "nginx_is_mutated",
				"cli": "test",
				"isMutated": "true"
			},
			"name": "nginx-deployment",
			"namespace": "default",
			"resourceVersion": "120124",
			"selfLink": "/apis/extensions/v1beta1/namespaces/default/deployments/nginx-deployment",
			"uid": "cc6e54a9-9c49-11e9-ad3c-0800273eb62d"
		},
		"spec": {
			"progressDeadlineSeconds": 600,
			"replicas": 1,
			"revisionHistoryLimit": 10,
			"selector": {
				"matchLabels": {
					"app": "nginx"
				}
			},
			"strategy": {
				"rollingUpdate": {
					"maxSurge": "25%",
					"maxUnavailable": "25%"
				},
				"type": "RollingUpdate"
			},
			"template": {
				"metadata": {
					"creationTimestamp": null,
					"labels": {
						"app": "nginx"
					}
				},
				"spec": {
					"containers": [
						{
							"image": "nginx:1.7.9",
							"imagePullPolicy": "Always",
							"name": "nginx",
							"ports": [
								{
									"containerPort": 80,
									"protocol": "TCP"
								}
							],
							"resources": {},
							"terminationMessagePath": "/dev/termination-log",
							"terminationMessagePolicy": "File"
						}
					],
					"dnsPolicy": "ClusterFirst",
					"restartPolicy": "Always",
					"schedulerName": "default-scheduler",
					"securityContext": {},
					"terminationGracePeriodSeconds": 30
				}
			}
		},
	}`)

	// clientConfig, err := utils.CreateClientConfig("/Users/shivd/.kube/config")
	// if err != nil {
	// 	glog.Fatalf("Error building kubeconfig: %v\n", err)
	// }

	// client, err := client.NewClient(clientConfig)
	// if err != nil {
	// 	glog.Fatalf("Error creating client: %v\n", err)
	// }

	// Lets load the schema
	schemaValidator := NewValidator(nil, false)
	result, err := schemaValidator.validate(deploy)
	if err != nil {
		t.Error(err)
	}
	if !result {
		t.Error("Document does not match the schema")
	}
	// load
	result, err = schemaValidator.validate(deploy)
	if err != nil {
		t.Error(err)
	}
	if !result {
		t.Error("Document does not match the schema")
	}

}

func TestPartial(t *testing.T) {
	deploy := []byte(`{
		"apiVersion": "extensions/v1beta1",
		"kind": "Deployment",
		"spec": {
			"progressDeadlineSeconds": 600,
			"replicas": 1,
			"revisionHistoryLimit": 10,
			"selector": {
				"matchLabels": {
					"app": "nginx"
				}
			},
				"spec": {
					"containers": [
						{
							"image": "nginx:1.7.9",
							"imagePullPolicy": "Always",
							"name": "nginx",
							"ports": [
								{
									"containerPort": 80,
									"protocol": "TCP"
								}
							],
							"resources": {},
							"terminationMessagePath": "/dev/termination-log",
							"terminationMessagePolicy": "File"
						}
					],
					"dnsPolicy": "ClusterFirst",
					"restartPolicy": "Always",
					"schedulerName": "default-scheduler",
					"securityContext": {},
					"terminationGracePeriodSeconds": 30
				}
			}
		},
	}`)
	// Lets load the schema
	schemaValidator := NewValidator(nil, false)
	result, err := schemaValidator.validate(deploy)
	if err != nil {
		t.Error(err)
	}
	if !result {
		t.Error("Document does not match the schema")
	}
}
