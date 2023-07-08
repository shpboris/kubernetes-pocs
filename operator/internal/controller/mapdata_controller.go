/*
Copyright 2023.

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

package controller

import (
	"context"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "shpboris/operator/api/v1"
)

// MapDataReconciler reconciles a MapData object
type MapDataReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=infra.shpboris,resources=mapdata,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infra.shpboris,resources=mapdata/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infra.shpboris,resources=mapdata/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MapData object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *MapDataReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logrus.Print("Started Reconcile")
	var mapData infrav1.MapData
	if err := r.Get(ctx, req.NamespacedName, &mapData); err != nil {
		logrus.Error(err, "Unable to fetch MapData")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logrus.Info("Received new MapData object named: " + mapData.Spec.Mapname)
	configMapList := &v1.ConfigMapList{}
	err := r.Client.List(ctx, configMapList)
	if err != nil {
		logrus.Error(err, "Failed to retrieve config maps list")
		return ctrl.Result{}, err
	}
	found := false
	for _, configMap := range configMapList.Items {
		if configMap.Name == mapData.Spec.Mapname {
			logrus.Info("Found existing config map named: " + mapData.Spec.Mapname)
			found = true
		}
	}
	if !found {
		logrus.Info("Creating config map named: " + mapData.Spec.Mapname)
		configMap := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      mapData.Spec.Mapname,
				Namespace: "default",
			},
			Data: map[string]string{
				"key1": mapData.Spec.Key1,
				"key2": mapData.Spec.Key2,
			},
		}
		if err = ctrl.SetControllerReference(&mapData, configMap, r.Scheme); err != nil {
			logrus.Error(err, "Failed to set controller reference")
			return ctrl.Result{}, err
		}
		err = r.Client.Create(ctx, configMap)
		if err != nil {
			logrus.Error(err, "Failed to create config map: "+mapData.Spec.Mapname)
			return ctrl.Result{}, err
		}
	}
	logrus.Print("Completed Reconcile")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MapDataReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.MapData{}).
		Complete(r)
}
