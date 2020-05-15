package client

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func NewConfig(clusterName, apiserverHost, apiserverToken string) (*rest.Config, error) {
	config := clientcmdapi.NewConfig()
	config.Clusters[clusterName] = &clientcmdapi.Cluster{Server: apiserverHost, InsecureSkipTLSVerify: true}
	config.AuthInfos[clusterName] = &clientcmdapi.AuthInfo{Token: apiserverToken}
	config.Contexts[clusterName] = &clientcmdapi.Context{
		Cluster:  clusterName,
		AuthInfo: clusterName,
	}
	config.CurrentContext = clusterName

	clientBuilder := clientcmd.NewNonInteractiveClientConfig(*config, clusterName, &clientcmd.ConfigOverrides{}, nil)

	return clientBuilder.ClientConfig()
}

func NewKubernetesClusterClient(name string, address string, token string) (*kubernetes.Clientset, error) {
	cfg, err := NewConfig(name, address, token)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
