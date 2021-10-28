/*
Copyright 2021.

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

package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/example-inc/lib-go-plugin-operator/api/cache.my.domain/v1alpha1"
	clientV1alpha1 "github.com/example-inc/lib-go-plugin-operator/api/generated/clientset/versioned/typed/cache.my.domain/v1alpha1"
	opInformer "github.com/example-inc/lib-go-plugin-operator/api/generated/informers/externalversions/cache.my.domain/v1alpha1"
	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsinformersv1 "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
)

//+kubebuilder:rbac:groups=cache.example.com,resources=memcacheds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.example.com,resources=memcacheds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.example.com,resources=memcacheds/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

type MemcachedController struct {
	// TODO: fill in the controller you would like to configure
	// for the resource.
	name           string
	operatorClient *clientV1alpha1.CacheV1alpha1Client
	kubeclient     kubernetes.Interface
	deployInformer appsinformersv1.DeploymentInformer
	recorder       events.Recorder
	namespace      string
}

func NewMemcachedController(name string,
	operatorClient *clientV1alpha1.CacheV1alpha1Client,
	kubeclient kubernetes.Interface,
	deployInformer appsinformersv1.DeploymentInformer,
	recorder events.Recorder,
	operatorInformer opInformer.MemcachedInformer,
	ns string) factory.Controller {
	// TODO: use this to create a new instance of your controller.

	c := &MemcachedController{
		name:           name,
		operatorClient: operatorClient,
		kubeclient:     kubeclient,
		deployInformer: deployInformer,
		recorder:       recorder,
		namespace:      ns,
	}

	return factory.New().WithInformers(deployInformer.Informer(), operatorInformer.Informer()).WithSync(c.sync).ResyncEvery(time.Minute).ToController(c.name, recorder.WithComponentSuffix(strings.ToLower(name)+"-deployment-controller-"))
}

// sync contains the logic of the reconciler
func (c *MemcachedController) sync(ctx context.Context, syncContext factory.SyncContext) error {

	// TODO: implement your reconciler logic here.
	log.Info("*********** reconciling **************")
	memcached, err := c.operatorClient.Memcacheds(c.namespace).Get(ctx, c.name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Error("memcached resource not found. Ignoring and reconciling again since object maybe deleted")
			return nil
		}
		log.Error("failed to get memcached")
		return nil
	}

	// if a deployment for memcached is not found create a new one.
	found, err := c.kubeclient.AppsV1().Deployments(c.namespace).Get(ctx, c.name, v1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		dep := c.deploymentForProject(memcached)
		log.Info("creating new deployment")
		_, err := c.kubeclient.AppsV1().Deployments(c.namespace).Create(ctx, dep, metav1.CreateOptions{})
		if err != nil {
			fmt.Println(err)
			return nil
		}
		return nil
	} else if err != nil {
		fmt.Println(err)
	}

	// if the number of replicas are not same the size specified in the spec, reconcile accordingly.
	size := memcached.Spec.Size
	if *found.Spec.Replicas != size {
		log.Info("Difference in number of replicas, reconciling again")
		found.Spec.Replicas = &size
		_, err := c.kubeclient.AppsV1().Deployments(c.namespace).Update(ctx, found, v1.UpdateOptions{})
		if err != nil {
			fmt.Println(err)
		}
		return nil
	}

	return nil
}

// Return deployment template to create for memcached.
func (r *MemcachedController) deploymentForProject(m *v1alpha1.Memcached) *appsv1.Deployment {
	ls := labelsForMemcached(m.Name)
	replica := m.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   "memcached:1.4.36-alpine",
						Name:    "projects",
						Command: []string{"memcached", "-m=64", "-o", "modern", "-v"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 11211,
							Name:          "projects",
						}},
					}},
				},
			},
		},
	}
	return dep
}

// labelsForMemcached returns the labels for selecting the resources
// belonging to the given memcached CR name.
func labelsForMemcached(name string) map[string]string {
	return map[string]string{"app": "project", "project_cr": name}
}
