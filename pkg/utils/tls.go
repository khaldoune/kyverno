package utils

import (
	"fmt"

	"github.com/golang/glog"
	client "github.com/nirmata/kyverno/pkg/dclient"
	"github.com/nirmata/kyverno/pkg/tls"
	"k8s.io/client-go/rest"
)

// Loads or creates PEM private key and TLS certificate for webhook server.
// Created pair is stored in cluster's secret.
// Returns struct with key/certificate pair.
func InitTLSPemPair(configuration *rest.Config, client *client.Client) (*tls.TlsPemPair, error) {
	certProps, err := client.GetTLSCertProps(configuration)
	if err != nil {
		return nil, err
	}
	tlsPair := client.ReadTlsPair(certProps)
	if tls.IsTlsPairShouldBeUpdated(tlsPair) {
		glog.Info("Generating new key/certificate pair for TLS")
		tlsPair, err = client.GenerateTlsPemPair(certProps)
		if err != nil {
			return nil, err
		}
		if err = client.WriteTlsPair(certProps, tlsPair); err != nil {
			return nil, fmt.Errorf("Unable to save TLS pair to the cluster: %v", err)
		}
		return tlsPair, nil
	}

	glog.Infoln("Using existing TLS key/certificate pair")
	return tlsPair, nil
}
