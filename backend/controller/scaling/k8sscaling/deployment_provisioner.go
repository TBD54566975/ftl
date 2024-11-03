/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package k8sscaling

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/puzpuzpuz/xsync/v3"
	istiosecmodel "istio.io/api/security/v1"
	"istio.io/api/type/v1beta1"
	istiosec "istio.io/client-go/pkg/apis/security/v1"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	kubeapps "k8s.io/api/apps/v1"
	kubecore "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/internal/log"
)

const thisDeploymentName = "ftl-controller"
const deploymentLabel = "ftl-deployment"
const configMapName = "ftl-controller-deployment-config"
const deploymentTemplate = "deploymentTemplate"
const serviceTemplate = "serviceTemplate"
const serviceAccountTemplate = "serviceAccountTemplate"

// DeploymentProvisioner reconciles a Foo object
type DeploymentProvisioner struct {
	Client           *kubernetes.Clientset
	MyDeploymentName string
	Namespace        string
	// Map of known deployments
	KnownDeployments *xsync.Map
	FTLEndpoint      string
	IstioSecurity    optional.Option[istioclient.Clientset]
}

func (r *DeploymentProvisioner) updateDeployment(ctx context.Context, name string, mod func(deployment *kubeapps.Deployment)) error {
	deploymentClient := r.Client.AppsV1().Deployments(r.Namespace)
	for range 10 {

		get, err := deploymentClient.Get(ctx, name, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get deployment %s: %w", name, err)
		}
		mod(get)
		_, err = deploymentClient.Update(ctx, get, v1.UpdateOptions{})
		if err != nil {
			if errors.IsConflict(err) {
				time.Sleep(time.Second)
				continue
			}
			return fmt.Errorf("failed to update deployment %s: %w", name, err)
		}
		return nil
	}
	return fmt.Errorf("failed to update deployment %s, 10 clonflicts in a row", name)
}

func (r *DeploymentProvisioner) HandleSchemaChange(ctx context.Context, msg *ftlv1.PullSchemaResponse) error {
	logger := log.FromContext(ctx).Scope("DeploymentProvisioner")
	ctx = log.ContextWithLogger(ctx, logger)
	defer func() {
		if r := recover(); r != nil {
			var err error
			if rerr, ok := r.(error); ok {
				err = rerr
			} else {
				err = fmt.Errorf("%v", r)
			}
			logger.Errorf(err, "Panic creating kube deployment")
		}
	}()
	err := r.handleSchemaChange(ctx, msg)
	if err != nil {
		logger.Errorf(err, "Failed to handle schema change")
	}
	logger.Debugf("Handled schema change for %s", msg.ModuleName)
	return err
}
func (r *DeploymentProvisioner) handleSchemaChange(ctx context.Context, msg *ftlv1.PullSchemaResponse) error {
	if !msg.More {
		defer r.deleteMissingDeployments(ctx)
	}
	if msg.DeploymentKey == "" {
		// Builtins don't have deployments
		return nil
	}
	logger := log.FromContext(ctx)
	logger = logger.Module(msg.ModuleName)
	ctx = log.ContextWithLogger(ctx, logger)
	logger.Debugf("Handling schema change for %s", msg.DeploymentKey)
	deploymentClient := r.Client.AppsV1().Deployments(r.Namespace)
	deployment, err := deploymentClient.Get(ctx, msg.DeploymentKey, v1.GetOptions{})
	deploymentExists := true
	if err != nil {
		if errors.IsNotFound(err) {
			deploymentExists = false
		} else {
			return fmt.Errorf("failed to get deployment %s: %w", msg.DeploymentKey, err)
		}
	}

	switch msg.ChangeType {
	case ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED, ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED:

		// Note that a change is now currently usually and add and a delete
		// As it should really be called a module changed, not a deployment changed
		// This will need to be fixed as part of the support for rolling deployments
		r.KnownDeployments.Store(msg.DeploymentKey, true)
		if deploymentExists {
			logger.Debugf("Updating deployment %s", msg.DeploymentKey)
			return r.handleExistingDeployment(ctx, deployment, msg.Schema)
		} else {
			return r.handleNewDeployment(ctx, msg.Schema, msg.DeploymentKey)
		}
	case ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED:
		if deploymentExists {
			go func() {

				// Nasty hack, we want all the controllers to have updated their route tables before we kill the runner
				// so we add a slight delay here
				time.Sleep(time.Second * 10)
				r.KnownDeployments.Delete(msg.DeploymentKey)
				logger.Debugf("Deleting service %s", msg.ModuleName)
				err = r.Client.CoreV1().Services(r.Namespace).Delete(ctx, msg.DeploymentKey, v1.DeleteOptions{})
				if err != nil {
					if !errors.IsNotFound(err) {
						logger.Errorf(err, "Failed to delete service %s", msg.ModuleName)
					}
				}
			}()
		}
	}
	return nil
}

func (r *DeploymentProvisioner) thisContainerImage(ctx context.Context) (string, error) {
	deploymentClient := r.Client.AppsV1().Deployments(r.Namespace)
	thisDeployment, err := deploymentClient.Get(ctx, thisDeploymentName, v1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get deployment %s: %w", thisDeploymentName, err)
	}
	return thisDeployment.Spec.Template.Spec.Containers[0].Image, nil

}

func (r *DeploymentProvisioner) handleNewDeployment(ctx context.Context, dep *schemapb.Module, name string) error {
	if dep.Runtime == nil {
		return nil
	}
	logger := log.FromContext(ctx)

	cm, err := r.Client.CoreV1().ConfigMaps(r.Namespace).Get(ctx, configMapName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get configMap %s: %w", configMapName, err)
	}
	deploymentClient := r.Client.AppsV1().Deployments(r.Namespace)
	thisDeployment, err := deploymentClient.Get(ctx, thisDeploymentName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s: %w", thisDeploymentName, err)
	}

	// First create a Service, this will be the root owner of all the other resources
	// Only create if it does not exist already
	servicesClient := r.Client.CoreV1().Services(r.Namespace)
	service, err := servicesClient.Get(ctx, name, v1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get service %s: %w", name, err)
		}
		logger.Debugf("Creating new kube service %s", name)
		err = decodeBytesToObject([]byte(cm.Data[serviceTemplate]), service)
		if err != nil {
			return fmt.Errorf("failed to decode service from configMap %s: %w", configMapName, err)
		}
		service.Name = name
		service.Labels = addLabel(service.Labels, "app", name)
		service.Labels = addLabel(service.Labels, deploymentLabel, name)
		service.OwnerReferences = []v1.OwnerReference{{APIVersion: "apps/v1", Kind: "deployment", Name: thisDeploymentName, UID: thisDeployment.UID}}
		service.Spec.Selector = map[string]string{"app": name}
		service, err = servicesClient.Create(ctx, service, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create service %s: %w", name, err)
		}
		logger.Debugf("Created kube service %s", name)
	} else {
		logger.Debugf("Service %s already exists", name)
	}

	// Now create a ServiceAccount, we mostly need this for Istio but we create it for all deployments
	// To keep things consistent
	serviceAccountClient := r.Client.CoreV1().ServiceAccounts(r.Namespace)
	serviceAccount, err := serviceAccountClient.Get(ctx, name, v1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get service account %s: %w", name, err)
		}
		logger.Debugf("Creating new kube service account %s", name)
		err = decodeBytesToObject([]byte(cm.Data[serviceAccountTemplate]), serviceAccount)
		if err != nil {
			return fmt.Errorf("failed to decode service account from configMap %s: %w", configMapName, err)
		}
		serviceAccount.Name = name
		serviceAccount.Labels = addLabel(serviceAccount.Labels, "app", name)
		serviceAccount.OwnerReferences = []v1.OwnerReference{{APIVersion: "v1", Kind: "service", Name: name, UID: service.UID}}
		_, err = serviceAccountClient.Create(ctx, serviceAccount, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create service account%s: %w", name, err)
		}
		logger.Debugf("Created kube service  account%s", name)
	} else {
		logger.Debugf("Service account %s already exists", name)
	}

	// Sync the istio policy if applicable
	if sec, ok := r.IstioSecurity.Get(); ok {
		err = r.syncIstioPolicy(ctx, sec, name, service, thisDeployment)
		if err != nil {
			return err
		}
	}

	// Now create the deployment

	logger.Debugf("Creating new kube deployment %s", name)
	thisImage, err := r.thisContainerImage(ctx)
	if err != nil {
		return fmt.Errorf("failed to get container image: %w", err)
	}
	data := cm.Data[deploymentTemplate]
	deployment := &kubeapps.Deployment{}
	err = decodeBytesToObject([]byte(data), deployment)
	if err != nil {
		return fmt.Errorf("failed to decode deployment from configMap %s: %w", configMapName, err)
	}
	ourVersion, err := extractTag(thisImage)
	if err != nil {
		return err
	}
	ourImage, err := extractBase(thisImage)
	if err != nil {
		return err
	}

	// runner images use the same tag as the controller
	var runnerImage string
	if dep.Runtime.Image != "" {
		runnerImage = strings.ReplaceAll(ourImage, "ftl0/ftl-controller", dep.Runtime.Image)
	} else {
		runnerImage = strings.ReplaceAll(ourImage, "ftl0/ftl-controller", "ftl0/ftl-runner")
	}

	deployment.Name = name
	deployment.OwnerReferences = []v1.OwnerReference{{APIVersion: "v1", Kind: "service", Name: name, UID: service.UID}}
	deployment.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", runnerImage, ourVersion)
	deployment.Spec.Selector = &v1.LabelSelector{MatchLabels: map[string]string{"app": name}}
	if deployment.Spec.Template.ObjectMeta.Labels == nil {
		deployment.Spec.Template.ObjectMeta.Labels = map[string]string{}
	}

	deployment.Spec.Template.ObjectMeta.Labels = addLabel(deployment.Spec.Template.ObjectMeta.Labels, "app", name)
	deployment.Spec.Template.Spec.ServiceAccountName = name
	changes, err := r.syncDeployment(ctx, thisImage, deployment, dep)

	if err != nil {
		return err
	}
	for _, change := range changes {
		change(deployment)
	}
	deployment.Labels = addLabel(deployment.Labels, deploymentLabel, name)
	deployment.Labels = addLabel(deployment.Labels, "app", name)

	deployment, err = deploymentClient.Create(ctx, deployment, v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create deployment %s: %w", deployment.Name, err)
	}
	logger.Debugf("Created kube deployment %s", name)

	return nil
}

func addLabel(labels map[string]string, key string, value string) map[string]string {
	if labels == nil {
		labels = map[string]string{}
	}
	labels[key] = value
	return labels
}

func decodeBytesToObject(bytes []byte, deployment runtime.Object) error {
	decodingScheme := runtime.NewScheme()
	decoderCodecFactory := serializer.NewCodecFactory(decodingScheme)
	decoder := decoderCodecFactory.UniversalDecoder()
	err := runtime.DecodeInto(decoder, bytes, deployment)
	if err != nil {
		return fmt.Errorf("failed to decode deployment: %w", err)
	}
	return nil
}

func (r *DeploymentProvisioner) handleExistingDeployment(ctx context.Context, deployment *kubeapps.Deployment, ftlDeployment *schemapb.Module) error {

	thisContainerImage, err := r.thisContainerImage(ctx)
	if err != nil {
		return err
	}
	changes, err := r.syncDeployment(ctx, thisContainerImage, deployment, ftlDeployment)
	if err != nil {
		return err
	}

	// If we have queued changes we apply them here. Changes can fail and need to be retried
	// Which is why they are supplied as a list of functions
	if len(changes) > 0 {
		err = r.updateDeployment(ctx, deployment.Name, func(deployment *kubeapps.Deployment) {
			for _, change := range changes {
				change(deployment)
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *DeploymentProvisioner) syncDeployment(ctx context.Context, thisImage string, deployment *kubeapps.Deployment, ftlDeployment *schemapb.Module) ([]func(*kubeapps.Deployment), error) {
	logger := log.FromContext(ctx)
	changes := []func(*kubeapps.Deployment){}
	ourVersion, err := extractTag(thisImage)
	if err != nil {
		return nil, err
	}
	deploymentVersion, err := extractTag(deployment.Spec.Template.Spec.Containers[0].Image)
	if err != nil {
		return nil, err
	}
	if ourVersion != deploymentVersion {
		// This means there has been an FTL upgrade
		// We are assuming the runner and provisioner run the same version
		// If they are different it means the provisioner has been upgraded and we need
		// to upgrade the deployments

		base, err := extractBase(deployment.Spec.Template.Spec.Containers[0].Image)
		if err != nil {
			logger.Errorf(err, "Could not determine base image for FTL deployment")
		} else {
			changes = append(changes, func(deployment *kubeapps.Deployment) {
				deployment.Spec.Template.Spec.Containers[0].Image = base + ":" + ourVersion
			})
			if err != nil {
				return nil, err
			}
		}
	}

	// For now we just make sure the number of replicas match
	if deployment.Spec.Replicas == nil || *deployment.Spec.Replicas != ftlDeployment.Runtime.MinReplicas {
		changes = append(changes, func(deployment *kubeapps.Deployment) {
			deployment.Spec.Replicas = &ftlDeployment.Runtime.MinReplicas
		})
	}
	changes = r.updateEnvVar(deployment, "FTL_DEPLOYMENT", deployment.Name, changes)
	changes = r.updateEnvVar(deployment, "FTL_ENDPOINT", r.FTLEndpoint, changes)
	return changes, nil
}

func (r *DeploymentProvisioner) updateEnvVar(deployment *kubeapps.Deployment, envVerName string, envVarValue string, changes []func(*kubeapps.Deployment)) []func(*kubeapps.Deployment) {
	found := false
	for pos, env := range deployment.Spec.Template.Spec.Containers[0].Env {
		if env.Name == envVerName {
			found = true
			if env.Value != envVarValue {
				changes = append(changes, func(deployment *kubeapps.Deployment) {
					deployment.Spec.Template.Spec.Containers[0].Env[pos] = kubecore.EnvVar{
						Name:  envVerName,
						Value: envVarValue,
					}
				})
			}
			break
		}
	}
	if !found {
		changes = append(changes, func(deployment *kubeapps.Deployment) {
			deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, kubecore.EnvVar{
				Name:  envVerName,
				Value: envVarValue,
			})
		})
	}
	return changes
}

func (r *DeploymentProvisioner) deleteMissingDeployments(ctx context.Context) {
	logger := log.FromContext(ctx)
	serviceClient := r.Client.CoreV1().Services(r.Namespace)

	list, err := serviceClient.List(ctx, v1.ListOptions{LabelSelector: deploymentLabel})
	if err != nil {
		logger.Errorf(err, "Failed to list services")
		return
	}

	for _, service := range list.Items {
		if _, ok := r.KnownDeployments.Load(service.Name); !ok {
			// With owner references the deployments should be deleted automatically
			// However this is in transition so delete both
			logger.Debugf("Deleting service %s as it is not a known module", service.Name)
			err := serviceClient.Delete(ctx, service.Name, v1.DeleteOptions{})
			if err != nil {
				logger.Errorf(err, "Failed to delete service %s", service.Name)
			}
		}
	}
}

func (r *DeploymentProvisioner) syncIstioPolicy(ctx context.Context, sec istioclient.Clientset, name string, service *kubecore.Service, thisDeployment *kubeapps.Deployment) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Creating new istio policy for %s", name)
	var update func(policy *istiosec.AuthorizationPolicy) error

	policiesClient := sec.SecurityV1().AuthorizationPolicies(r.Namespace)
	policy, err := policiesClient.Get(ctx, name, v1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get istio policy %s: %w", name, err)
		}
		policy = &istiosec.AuthorizationPolicy{}
		policy.Name = name
		policy.Namespace = r.Namespace
		update = func(policy *istiosec.AuthorizationPolicy) error {
			_, err := policiesClient.Create(ctx, policy, v1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create istio policy %s: %w", name, err)
			}
			return nil
		}
	} else {
		update = func(policy *istiosec.AuthorizationPolicy) error {
			_, err := policiesClient.Update(ctx, policy, v1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to update istio policy %s: %w", name, err)
			}
			return nil
		}
	}
	policy.OwnerReferences = []v1.OwnerReference{{APIVersion: "v1", Kind: "service", Name: name, UID: service.UID}}
	// At present we only allow ingress from the controller
	policy.Spec.Selector = &v1beta1.WorkloadSelector{MatchLabels: map[string]string{"app": name}}
	policy.Spec.Action = istiosecmodel.AuthorizationPolicy_ALLOW
	policy.Spec.Rules = []*istiosecmodel.Rule{
		{
			From: []*istiosecmodel.Rule_From{
				{
					Source: &istiosecmodel.Source{
						Principals: []string{"cluster.local/ns/" + r.Namespace + "/sa/" + thisDeployment.Spec.Template.Spec.ServiceAccountName},
					},
				},
			},
		},
	}

	return update(policy)
}

func extractTag(image string) (string, error) {
	idx := strings.LastIndex(image, ":")
	if idx == -1 {
		return "", fmt.Errorf("no tag found in image %s", image)
	}
	ret := image[idx+1:]
	at := strings.LastIndex(ret, "@")
	if at != -1 {
		ret = image[:at]
	}
	return ret, nil
}

func extractBase(image string) (string, error) {
	idx := strings.LastIndex(image, ":")
	if idx == -1 {
		return "", fmt.Errorf("no tag found in image %s", image)
	}
	return image[:idx], nil
}
