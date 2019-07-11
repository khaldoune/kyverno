package utils

import (
	"github.com/golang/glog"
	"github.com/nirmata/kyverno/pkg/version"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func Contains(list []string, element string) bool {
	for _, e := range list {
		if e == element {
			return true
		}
	}
	return false
}
func printVersionInfo() {
	v := version.GetVersion()
	glog.Infof("Kyverno version: %s\n", v.BuildVersion)
	glog.Infof("Kyverno BuildHash: %s\n", v.BuildHash)
	glog.Infof("Kyverno BuildTime: %s\n", v.BuildTime)
}

func CreateClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig == "" {
		glog.Info("Using in-cluster configuration")
		return rest.InClusterConfig()
	}
	glog.Infof("Using configuration from '%s'", kubeconfig)
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}
