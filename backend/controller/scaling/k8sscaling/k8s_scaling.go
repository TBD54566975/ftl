package k8sscaling

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/controller/scaling"
	"github.com/TBD54566975/ftl/internal/log"
)

var _ = (scaling.RunnerScaling)(NewK8sScaling)

func NewK8sScaling(ctx context.Context, controller url.URL, leaser leases.Leaser) error {

	logger := log.FromContext(ctx).Scope("K8sScaling")
	ctx = log.ContextWithLogger(ctx, logger)
	// creates the in-cluster config
	config, err := rest.InClusterConfig()

	if err != nil {
		// if we're not in a cluster, use the kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			return fmt.Errorf("failed to get kubeconfig: %w", err)
		}
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	namespace, err := getCurrentNamespace()
	if err != nil {
		// Nothing we can do here, if we don't have a namespace we have no runners
		return fmt.Errorf("failed to get current namespace: %w", err)
	}
	logger.Infof("using namespace %s", namespace)
	deploymentReconciler := &DeploymentProvisioner{
		Client:       clientset,
		Namespace:    namespace,
		KnownModules: map[string]string{},
		FTLEndpoint:  controller.String(),
	}
	scaling.BeginGrpcScaling(ctx, controller, leaser, deploymentReconciler.HandleSchemaChange)
	return nil
}

func getCurrentNamespace() (string, error) {
	namespaceFile := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	namespace, err := os.ReadFile(namespaceFile)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to read namespace file: %w", err)
	} else if err == nil {
		return string(namespace), nil
	}

	// If not running in a cluster, get the namespace from the kubeconfig
	configAccess := clientcmd.NewDefaultPathOptions()
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	currentContext := config.CurrentContext
	if currentContext == "" {
		return "", fmt.Errorf("no current context found in kubeconfig")
	}

	context, exists := config.Contexts[currentContext]
	if !exists {
		return "", fmt.Errorf("context %s not found in kubeconfig", currentContext)
	}

	return context.Namespace, nil
}
