package endpointgroup

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/globalaccelerator"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/globalaccelerator/v1alpha1"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupEndpointGroup adds a controller that reconciles an EndpointGroup.
func SetupEndpointGroup(mgr ctrl.Manager, o controller.Options) error {
	fmt.Println("Setup endpointgroup")
	name := managed.ControllerName(svcapitypes.EndpointGroupGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preUpdate = preUpdate
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.preDelete = preDelete
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.EndpointGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.EndpointGroupGroupVersionKind),
			managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
			managed.WithTypedExternalConnector(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(ctx context.Context, cr *svcapitypes.EndpointGroup, obj *svcsdk.DescribeEndpointGroupInput) error {
	obj.EndpointGroupArn = aws.String(meta.GetExternalName(cr))
	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.EndpointGroup, obj *svcsdk.CreateEndpointGroupInput) error {
	obj.ListenerArn = aws.String(ptr.Deref(cr.Spec.ForProvider.CustomEndpointGroupParameters.ListenerARN, ""))
	obj.IdempotencyToken = aws.String(string(cr.UID))
	return nil
}

func preUpdate(_ context.Context, cr *svcapitypes.EndpointGroup, obj *svcsdk.UpdateEndpointGroupInput) error {
	obj.EndpointGroupArn = aws.String(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.EndpointGroup, obj *svcsdk.DeleteEndpointGroupInput) (bool, error) {
	obj.EndpointGroupArn = aws.String(meta.GetExternalName(cr))
	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.EndpointGroup, resp *svcsdk.DescribeEndpointGroupOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.EndpointGroup, resp *svcsdk.CreateEndpointGroupOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	meta.SetExternalName(cr, aws.StringValue(resp.EndpointGroup.EndpointGroupArn))

	return cre, err
}
