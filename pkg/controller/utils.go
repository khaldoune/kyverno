package controller

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const policyWorkQueueName = "policyworkqueue"

const policyWorkQueueRetryLimit = 5

const policyControllerWorkerCount = 2

type resourceInfo struct {
	rawResource []byte
	gvk         *metav1.GroupVersionKind
}
