package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const policyWorkQueueName = "policyworkqueue"

const policyWorkQueueRetryLimit = 5

const policyControllerWorkerCount = 2

type resourceInfo struct {
	resource *unstructured.Unstructured
	gvk      *metav1.GroupVersionKind
}
