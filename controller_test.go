package multicluster_gw

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcsv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"
)

const (
	serviceName = "svc"
	serviceNS   = "svc-ns"
	cluster1    = "c1"
)

var (
	timeout int32 = 10

	serviceImport = &mcsv1a1.ServiceImport{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: serviceNS,
			Name:      serviceName,
		},
		//TODO: Don't think I need it, meanwhile kept in comment:
		//		Spec: mcsv1a1.ServiceImportSpec{
		//			Type:  es.Service.Type,
		//			Ports: es.Service.Ports,
		//		},
		//		Status: mcsv1a1.ServiceImportStatus{
		//			Clusters: []mcsv1a1.ClusterStatus{{Cluster: lhutil.GetOriginalObjectCluster(se.ObjectMeta)}},
		//		},
	}
)

func TestController(t *testing.T) {
	tests := []struct {
		shouldAddToSet   bool
		shouldErr        bool // True if test case is expected to produce an error.
		preloadedObjects []runtime.Object
	}{
		// positive
		{
			true,
			false,
			[]runtime.Object{serviceImport},
		},
		// try to add SI that dosent exists:

		{
			false,
			false,
			[]runtime.Object{},
		},
	}

	assert := require.New(t)

	for _, test := range tests {

		ser := ServiceImportReconciler{
			Client: getClient(test.preloadedObjects),
			Log:    logr.Logger{},
			Scheme: getScheme(),
		}
		Mcgw.SISet = *NewSiSet()
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      serviceImport.GetName(),
				Namespace: serviceImport.GetNamespace(),
			}}

		result, err := ser.Reconcile(context.TODO(), req)
		if !test.shouldErr {
			assert.Nil(err)
			assert.False(result.Requeue, "unexpected requeue")
		}
		isContains := Mcgw.SISet.Contains(GenerateNameAsString(serviceImport.GetName(), serviceImport.GetNamespace()))
		assert.Equal(test.shouldAddToSet, isContains)
	}
}

// generate a fake client with preloaded objects
func getClient(objs []runtime.Object) client.Client {
	return fake.NewClientBuilder().WithScheme(getScheme()).WithRuntimeObjects(objs...).Build()
}

// satisfy the logr.Logger interface with a nil logger
type logger struct {
	enabled bool
	t       *testing.T
	name    string
	kv      map[string]interface{}
}

// Tocomplete:
func newLogger(t *testing.T, enabled bool) *logger {
	return &logger{
		enabled: enabled,
		t:       t,
		kv:      make(map[string]interface{}),
	}
}

// return a scheme
func getScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	utilruntime.Must(mcsv1a1.AddToScheme(scheme))
	return scheme
}
