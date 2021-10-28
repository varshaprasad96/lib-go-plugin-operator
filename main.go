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
package main

import (
	"context"
	"time"

	"github.com/example-inc/lib-go-plugin-operator/api/cache.my.domain/v1alpha1"
	operatorversionedclient "github.com/example-inc/lib-go-plugin-operator/api/generated/clientset/versioned"
	clientV1alpha1 "github.com/example-inc/lib-go-plugin-operator/api/generated/clientset/versioned/typed/cache.my.domain/v1alpha1"
	"github.com/example-inc/lib-go-plugin-operator/controllers"
	"github.com/openshift/library-go/pkg/operator/events"
	coreinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"

	opInformer "github.com/example-inc/lib-go-plugin-operator/api/generated/informers/externalversions"
)

func main() {
	ctx := context.TODO()

	cfg := ctrl.GetConfigOrDie()
	var err error

	// Register custom resource to the scheme.
	v1alpha1.AddToScheme(scheme.Scheme)

	// Create a config to work with custom resources
	clientSet, err := clientV1alpha1.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}

	// create a versioned client which will be used to create informers in turn.
	operatorConfigClient, err := operatorversionedclient.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}

	// create kubeclient to handle other resources like deployment.
	kubeclient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}

	// create an informer for memcached resource.
	memcachedInformer := opInformer.NewSharedInformerFactoryWithOptions(
		operatorConfigClient,
		time.Minute,
	)

	// use coreInformer to set up an informer for the deployment object.
	coreInformerFactory := coreinformers.NewSharedInformerFactory(kubeclient, 0)

	memcachedController := controllers.NewMemcachedController("memcached-sample", clientSet, kubeclient, coreInformerFactory.Apps().V1().Deployments(), events.NewInMemoryRecorder("memcached"), memcachedInformer.Cache().V1alpha1().Memcacheds(), "default")

	// Start the informers to make sure their caches are in sync and are updated periodically.
	for _, informer := range []interface {
		Start(stopCh <-chan struct{})
	}{
		// TODO: If there are any informers for your controller, make sure to
		// add them here to start the informer.
		coreInformerFactory,
		memcachedInformer,
	} {
		informer.Start(ctx.Done())
	}

	// Start and run the controller
	for _, controllerint := range []interface {
		Run(ctx context.Context, workers int)
	}{
		// TODO: Add the name of controllers which have been instantiated previosuly for the
		// operator.
		memcachedController,
	} {
		go controllerint.Run(ctx, 1)
	}

	<-ctx.Done()
	return

}
