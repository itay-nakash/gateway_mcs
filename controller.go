package multicluster_gw

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	mcsv1a "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"
)

// ServiceImportReconciler reconciles a ServiceImport object
type ServiceImportReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var SIset Set

//+kubebuilder:rbac:groups=app.my.domain,resources=serviceimports,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.my.domain,resources=serviceimports/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.my.domain,resources=serviceimports/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

// TODO: check about the 'cntx' (should I get it in Reconcile or not?)
func (r *ServiceImportReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("serviceimport", req.NamespacedName)
	log.Info("Enter Reconcile", "req", req)

	si := &mcsv1a.ServiceImport{}
	siNameNs := types.NamespacedName{Name: req.Name, Namespace: req.Namespace}
	ctx := context.Background()
	err := r.Get(ctx, siNameNs, si)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("ServiceImport resource not found. Assume the corresponding SI was deleted")
			// deleting the service name and ns:
			SIset.Delete(GenerateNameAsString(siNameNs))

			return ctrl.Result{}, nil
		}
		// Error reading the object (other than not found...) - requeue the request
		log.Error(err, "Failed to get ServiceImport")
		return ctrl.Result{}, err
	}

	// if it got here, the serviceImport is existing, so its a new ServiceImport:
	// how to know if its a new one, or changing in an existing one?
	// #TODO what should I do if the

	// update them in the data structure:
	SIset.Add(GenerateNameAsString(siNameNs))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceImportReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&mcsv1a.ServiceImport{}).
		Complete(r)
}

func GenerateNameAsString(siNameNs types.NamespacedName) string {
	return siNameNs.Name + "." + siNameNs.Namespace
}
