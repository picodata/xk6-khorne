package client

import (
	"xk6-khorne/pkg/k8s/config"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// New creates a new k8s client
func New() (*kubernetes.Clientset, error) {
	config := config.GetConfig()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
