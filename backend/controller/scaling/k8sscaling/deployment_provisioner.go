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

// DeploymentProvisioner reconciles a Foo object
type DeploymentProvisioner struct {
	Client           *kubernetes.Clientset
	MyDeploymentName string
	Namespace        string
	// Map of modules to known deployments
	// This needs to be re-done when we have proper rolling deployments
	KnownModules map[string]string
	FTLEndpoint  string
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
			logger.Errorf(err, "panic creating kube deployment")
		}
	}()
	err := r.handleSchemaChange(ctx, msg)
	if err != nil {
		logger.Errorf(err, "failed to handle schema change")
	}
	logger.Infof("handled schema change for %s", msg.ModuleName)
	return err
}
func (r *DeploymentProvisioner) handleSchemaChange(ctx context.Context, msg *ftlv1.PullSchemaResponse) error {
	if msg.DeploymentKey == "" {
		// Builtins don't have deployments
		return nil
	}
	logger := log.FromContext(ctx)
	logger.Infof("Handling schema change for %s", msg.DeploymentKey)
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
		r.KnownModules[msg.ModuleName] = msg.DeploymentKey
		if deploymentExists {
			logger.Infof("updating deployment %s", msg.DeploymentKey)
			return r.handleExistingDeployment(ctx, deployment, msg.Schema)
		} else {
			return r.handleNewDeployment(ctx, msg.Schema, msg.DeploymentKey)
		}
	case ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED:
		delete(r.KnownModules, msg.ModuleName)
		logger.Infof("removing deployment %s", msg.ModuleName)
		r.deleteMissingDeployments(ctx)
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
	deploymentClient := r.Client.AppsV1().Deployments(r.Namespace)
	logger := log.FromContext(ctx)
	logger.Infof("creating new kube deployment %s", name)
	thisImage, err := r.thisContainerImage(ctx)
	if err != nil {
		return fmt.Errorf("failed to get configMap %s: %w", configMapName, err)
	}
	cm, err := r.Client.CoreV1().ConfigMaps(r.Namespace).Get(ctx, configMapName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get configMap %s: %w", configMapName, err)
	}
	data := cm.Data[deploymentTemplate]
	deployment, err := decodeBytesToTDeployment([]byte(data))
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
	runnerImage := strings.ReplaceAll(ourImage, "controller", "runner")

	thisDeployment, err := deploymentClient.Get(ctx, thisDeploymentName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s: %w", thisDeploymentName, err)
	}
	deployment.Name = name
	deployment.OwnerReferences = []v1.OwnerReference{{APIVersion: "apps/v1", Kind: "deployment", Name: thisDeploymentName, UID: thisDeployment.UID}}
	deployment.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", runnerImage, ourVersion)
	deployment.Spec.Selector = &v1.LabelSelector{MatchLabels: map[string]string{"app": name}}
	if deployment.Spec.Template.ObjectMeta.Labels == nil {
		deployment.Spec.Template.ObjectMeta.Labels = map[string]string{}
	}
	deployment.Spec.Template.ObjectMeta.Labels["app"] = name
	changes, err := r.syncDeployment(ctx, thisImage, deployment, dep)
	if err != nil {
		return err
	}
	for _, change := range changes {
		change(deployment)
	}
	if deployment.Labels == nil {
		deployment.Labels = map[string]string{}
	}
	deployment.Labels[deploymentLabel] = name

	_, err = deploymentClient.Create(ctx, deployment, v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create deployment %s: %w", deployment.Name, err)
	}
	logger.Infof("created kube deployment %s", name)
	return nil

}
func decodeBytesToTDeployment(bytes []byte) (*kubeapps.Deployment, error) {
	deployment := kubeapps.Deployment{}
	decodingScheme := runtime.NewScheme()
	decoderCodecFactory := serializer.NewCodecFactory(decodingScheme)
	decoder := decoderCodecFactory.UniversalDecoder()
	err := runtime.DecodeInto(decoder, bytes, &deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to decode deployment: %w", err)
	}
	return &deployment, nil
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
			logger.Errorf(err, "could not determine base image for FTL deployment")
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
	deploymentClient := r.Client.AppsV1().Deployments(r.Namespace)

	list, err := deploymentClient.List(ctx, v1.ListOptions{LabelSelector: deploymentLabel})
	if err != nil {
		logger.Errorf(err, "failed to list deployments")
		return
	}
	knownDeployments := map[string]bool{}
	for _, deployment := range r.KnownModules {
		knownDeployments[deployment] = true
	}

	for _, deployment := range list.Items {
		if !knownDeployments[deployment.Name] {
			logger.Infof("deleting deployment %s", deployment.Name)
			err := deploymentClient.Delete(ctx, deployment.Name, v1.DeleteOptions{})
			if err != nil {
				logger.Errorf(err, "failed to delete deployment %s", deployment.Name)
			}

		}
	}
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
