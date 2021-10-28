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

	"github.com/example-inc/lib-go-plugin-operator/api/cache.my.domain/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	ctx := context.TODO()

	cfg := ctrl.GetConfigOrDie()

	// Register custom resource to the scheme.
	v1alpha1.AddToScheme(scheme.Scheme)

	// Start the informers to make sure their caches are in sync and are updated periodically.
	for _, informer := range []interface {
		Start(stopCh <-chan struct{})
	}{
		// TODO: If there are any informers for your controller, make sure to
		// add them here to start the informer.
	} {
		informer.Start(ctx.Done())
	}

	// Start and run the controller
	for _, controllerint := range []interface {
		Run(ctx context.Context, workers int)
	}{
		// TODO: Add the name of controllers which have been instantiated previosuly for the
		// operator.
	} {
		go controllerint.Run(ctx, 1)
	}

	<-ctx.Done()
	return

}
