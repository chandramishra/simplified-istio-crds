/*
Copyright 2022 The Crossplane Authors.

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

package canary

import (
	"context"
	"fmt"
	"log"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//"github.com/crossplane/provider-istio/internal/clients"

	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	"github.com/crossplane/provider-istio/apis/simplified/v1alpha1"
	apisv1alpha1 "github.com/crossplane/provider-istio/apis/v1alpha1"
	//"github.com/crossplane/provider-istio/internal/controller/features"
	"github.com/crossplane/provider-istio/internal/features"
)

const (
	errNotCanary    = "managed resource is not a Canary custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"

	errNewClient = "cannot create new Service"
)

type IstioService struct{
  istioObj *versionedclient.Clientset
}

var (
	newIstioService = func() (*IstioService, error) { 
		kubeconfig := "/root/.kube/config"
		restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatalf("Failed to create k8s rest client: %s", err)
		}

		ic, err := versionedclient.NewForConfig(restConfig)
		if err != nil {
			log.Fatalf("Failed to create istio client: %s", err)
		}
		return &IstioService{istioObj: ic}, nil 
		}
)

// Setup adds a controller that reconciles Canary managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.CanaryGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.CanaryGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: newIstioService}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.Canary{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func() (*IstioService, error)
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Canary)
	if !ok {
		return nil, errors.New(errNotCanary)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	//cd := pc.Spec.Credentials
	//data, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	//if err != nil {
	//	return nil, errors.Wrap(err, errGetCreds)
	//}
	//}

	svc, err := c.newServiceFn()
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &external{service: svc}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	service *IstioService
}
func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Canary)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCanary)
	}

	// These fmt statements should be removed in the real implementation.
	fmt.Println("******************Observing: Managed resource Manifest ************************************")
	//fmt.Printf("Observing: %+v", cr)
	//if true {
	fmt.Println(meta.GetExternalName(cr))
	if meta.GetExternalName(cr) != "helloworld" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		// Return false when the external resource does not exist. This lets
		// the managed resource reconciler know that it needs to call Create to
		// (re)create the resource, or that it has successfully been deleted.
		ResourceExists: true,

		// Return false when the external resource exists, but it not up to date
		// with the desired managed resource state. This lets the managed
		// resource reconciler know that it needs to call Update.
		ResourceUpToDate: true,

		// Return any details that may be required to connect to the external
		// resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Canary)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCanary)
	}

        fmt.Println("******************Managed resource Manifest ************************************")
	fmt.Printf("Creating:VS %+v", cr)
        fmt.Println("******************Started creating VS Resource ************************************")
	sources := cr.Spec.ForProvider.Sources[0]
	v1_w := cr.Spec.ForProvider.Split[0].Weight
	v2_w := cr.Spec.ForProvider.Split[1].Weight
	vs := string(fmt.Sprintf("{\"apiVersion\": \"networking.istio.io/v1alpha3\",\"kind\": \"VirtualService\",\"metadata\": {\"name\": \"helloworld\"},\"spec\": {\"hosts\": [\"%s\"],\"http\": [{\"route\": [{\"destination\": {\"host\": \"%s\",\"subset\": \"v1\"},\"weight\": %d},{\"destination\": {\"host\": \"%s\",\"subset\": \"v2\"},\"weight\": %d}]}]}}",sources,sources,v1_w,sources,v2_w))
	bytes := []byte(vs)
	virtualService := &v1alpha3.VirtualService{}
	json.Unmarshal(bytes, &virtualService)
	c.service.istioObj.NetworkingV1alpha3().VirtualServices("test").Create(context.TODO(), virtualService, metav1.CreateOptions{})
        fmt.Println("******************Started creating DR Resource ************************************")
	dr := "{\"apiVersion\": \"networking.istio.io/v1alpha3\",\"kind\": \"DestinationRule\",\"metadata\": {\"name\": \"helloworld\"},\"spec\": {\"host\": \"helloworld\",\"subsets\": [{\"name\": \"v1\",\"labels\": {\"version\": \"v1\"}},{\"name\": \"v2\",\"labels\": {\"version\": \"v2\"}}]}}"
	fmt.Printf("Creating DR: %+v", dr)
	dr_bytes := []byte(dr)
	destinationRule := &v1alpha3.DestinationRule{}
	json.Unmarshal(dr_bytes, &destinationRule)
	c.service.istioObj.NetworkingV1alpha3().DestinationRules("test").Create(context.TODO(), destinationRule, metav1.CreateOptions{})
	meta.SetExternalName(cr, "helloworld")

	//cr.Status.AtProvider.State = true
	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Canary)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCanary)
	}

	fmt.Printf("Updating: %+v", cr)

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Canary)
	if !ok {
		return errors.New(errNotCanary)
	}
        fmt.Println("******************Deleting ************************************\n")
	c.service.istioObj.NetworkingV1alpha3().VirtualServices("test").Delete(context.TODO(), "helloworld", metav1.DeleteOptions{})
	c.service.istioObj.NetworkingV1alpha3().DestinationRules("test").Delete(context.TODO(), "helloworld", metav1.DeleteOptions{})
	meta.RemoveFinalizer(cr, "finalizer.managedresource.crossplane.io")
	fmt.Printf("Updating: %+v", cr)
        //return c.kube.client.Delete(ctx, "helloworld")
	return nil
}
