/*
Copyright 2021 The Crossplane Authors.

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

package accesspoint

import (
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/s3control/v1alpha1"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupAccessPoint adds a controller that reconciles Stage.
func SetupAccessPoint(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.AccessPointGroupKind)

	opts := createOptions()
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.AccessPoint{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.AccessPointGroupVersionKind),
			managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
			managed.WithTypedExternalConnector(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func createOptions() []option {
	return []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.preDelete = preDelete
			hooks := newHooks(e.client)
			e.isUpToDate = hooks.isUpToDate
			e.update = hooks.update
		},
	}
}
