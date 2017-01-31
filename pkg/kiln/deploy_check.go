package kiln

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/labels"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClusterConfig is the config used to interact with Shipyard's in-cluster artifacts
type ClusterConfig struct {
	Client       *kubernetes.Clientset
	OrgLabelName string
	AppLabelName string
}

// NewClusterConfig creates a new k8s in-cluster client
func NewClusterConfig() (*ClusterConfig, error) {
	clusterConfig := &ClusterConfig{}

	clusterConfig.OrgLabelName = os.Getenv("ORG_LABEL")
	if clusterConfig.OrgLabelName == "" {
		return nil, fmt.Errorf("Missing required ORG_LABEL environment variable")
	}

	clusterConfig.AppLabelName = os.Getenv("APP_NAME_LABEL")
	if clusterConfig.OrgLabelName == "" {
		return nil, fmt.Errorf("Missing required APP_NAME_LABEL environment variable")
	}

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clusterConfig.Client, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clusterConfig, nil
}

// NewLocalClusterConfig creates a new k8s in-cluster client
func NewLocalClusterConfig() (*ClusterConfig, error) {
	clusterConfig := &ClusterConfig{}

	clusterConfig.OrgLabelName = os.Getenv("ORG_LABEL")
	if clusterConfig.OrgLabelName == "" {
		return nil, fmt.Errorf("Missing required ORG_LABEL environment variable")
	}

	clusterConfig.AppLabelName = os.Getenv("APP_NAME_LABEL")
	if clusterConfig.OrgLabelName == "" {
		return nil, fmt.Errorf("Missing required APP_NAME_LABEL environment variable")
	}

	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	// uses the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", path.Join(usr.HomeDir, ".kube/config"))
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clusterConfig.Client, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clusterConfig, nil
}

// CheckActiveDeployments checks if a deployment for appName by org is currently active
func (clusterConfig *ClusterConfig) CheckActiveDeployments(org string, appName string) (bool, error) {

	selector := make(map[string]string, 3)
	selector[clusterConfig.OrgLabelName] = org
	selector[clusterConfig.AppLabelName] = appName
	selector["runtime"] = "shipyard"

	results, err := clusterConfig.Client.Deployments(api.NamespaceAll).List(v1.ListOptions{
		LabelSelector: labels.FormatLabels(selector),
	})

	if err != nil {
		return false, err
	}

	if len(results.Items) > 0 {
		return true, nil
	}

	return false, nil
}
