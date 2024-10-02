package k8sscaling

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/alecthomas/types/optional"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/controller/scaling"
	"github.com/TBD54566975/ftl/internal/log"
)

var _ scaling.RunnerScaling = &k8sScaling{}

type k8sScaling struct {
	disableIstio bool
}

func (k k8sScaling) Start(ctx context.Context, controller url.URL, leaser leases.Leaser) error {

	logger := log.FromContext(ctx).Scope("K8sScaling")
	ctx = log.ContextWithLogger(ctx, logger)
	clientset, err := CreateClientSet()
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	namespace, err := GetCurrentNamespace()
	if err != nil {
		// Nothing we can do here, if we don't have a namespace we have no runners
		return fmt.Errorf("failed to get current namespace: %w", err)
	}

	var sec *istioclient.Clientset
	if !k.disableIstio {
		groups, err := clientset.Discovery().ServerGroups()
		if err != nil {
			return fmt.Errorf("failed to get server groups: %w", err)
		}
		// If istio is present and not explicitly disabled we create the client
		for _, group := range groups.Groups {
			if group.Name == "security.istio.io" {
				sec, err = CreateIstioClientSet()
				if err != nil {
					return fmt.Errorf("failed to create istio clientset: %w", err)
				}
				break
			}
		}
	}

	logger.Debugf("Using namespace %s", namespace)
	deploymentReconciler := &DeploymentProvisioner{
		Client:           clientset,
		Namespace:        namespace,
		KnownDeployments: map[string]bool{},
		FTLEndpoint:      controller.String(),
		IstioSecurity:    optional.Ptr(sec),
	}
	scaling.BeginGrpcScaling(ctx, controller, leaser, deploymentReconciler.HandleSchemaChange)
	return nil
}

func CreateClientSet() (*kubernetes.Clientset, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client set: %w", err)
	}
	return clientset, nil
}

func CreateIstioClientSet() (*istioclient.Clientset, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := istioclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client set: %w", err)
	}
	return clientset, nil
}

func getKubeConfig() (*rest.Config, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		// if we're not in a cluster, use the kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
		}
	}
	return config, nil
}

func (k k8sScaling) GetEndpointForDeployment(ctx context.Context, module string, deployment string) (optional.Option[url.URL], error) {
	// TODO: hard coded port? It's hard to deal with as we might not have the lease
	// I think requiring this port is fine for now
	return optional.Some(url.URL{Scheme: "http",
		Host: fmt.Sprintf("%s:8893", deployment),
	}), nil
}

func NewK8sScaling(disableIstio bool) scaling.RunnerScaling {
	return &k8sScaling{disableIstio: disableIstio}
}

func GetCurrentNamespace() (string, error) {
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
