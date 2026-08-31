/*
Copyright 2026.

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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "github.com/minhngo248/go-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WebAppReconciler reconciles a WebApp object
type WebAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=webapp.kodekloud.com,resources=webapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=webapp.kodekloud.com,resources=webapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=webapp.kodekloud.com,resources=webapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the WebApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *WebAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Create an empty WebApp object that will be populated by the client.
	var webapp webappv1.WebApp

	// Fetch the WebApp instance from the API server.
	if err := r.Get(ctx, req.NamespacedName, &webapp); err != nil {
		// Ignore not-found errors (deleted after reconcile request). Return other errors.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("fetched WebApp", "image", webapp.Spec.Image, "replicas", webapp.Spec.Replicas)
	labels := map[string]string{"app": webapp.Name}
	selectLabels := map[string]string{"app": webapp.Name}

	desiredCM := configMapFor(&webapp, labels)
	if err := controllerutil.SetControllerReference(&webapp, desiredCM, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	foundCM := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: desiredCM.Name, Namespace: desiredCM.Namespace}, foundCM)
	if apierrors.IsNotFound(err) {
		log.Info("creating ConfigMap", "name", desiredCM.Name)
		if err := r.Create(ctx, desiredCM); err != nil {
			return ctrl.Result{}, err
		}
	} else if err != nil {
		return ctrl.Result{}, err
	}

	desiredDeployment := deploymentFor(&webapp, labels)
	if err := controllerutil.SetControllerReference(&webapp, desiredDeployment, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: desiredDeployment.Name, Namespace: desiredDeployment.Namespace}, found)

	if apierrors.IsNotFound(err) {
		log.Info("creating Deployment", "name", desiredDeployment.Name)
		if err := r.Create(ctx, desiredDeployment); err != nil {
			return ctrl.Result{}, err
		}
	} else if err != nil {
		return ctrl.Result{}, err
	}

	desiredSvc := serviceFor(&webapp, labels, selectLabels)
	if err := controllerutil.SetControllerReference(&webapp, desiredSvc, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	foundSvc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: desiredSvc.Name, Namespace: desiredSvc.Namespace}, foundSvc)
	if apierrors.IsNotFound(err) {
		log.Info("creating Service", "name", desiredSvc.Name)
		if err := r.Create(ctx, desiredSvc); err != nil {
			return ctrl.Result{}, err
		}
	} else if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func deploymentFor(w *webappv1.WebApp, labels map[string]string) *appsv1.Deployment {
	replicas := w.Spec.Replicas
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.Name,
			Namespace: w.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "web",
						Image: w.Spec.Image,
						Ports: []corev1.ContainerPort{{ContainerPort: 80}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "config-vol",
							MountPath: "/usr/share/nginx/html",
						}},
					}},
					Volumes: []corev1.Volume{
						{
							Name: "config-vol",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: w.Name + "-config",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "welcome.html",
											Path: "index.html",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func configMapFor(w *webappv1.WebApp, labels map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.Name + "-config",
			Namespace: w.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"welcome.html": "<h1>Hello from " + w.Name + "</h1>",
		},
	}
}

func serviceFor(w *webappv1.WebApp, labels, selectLabels map[string]string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.Name + "-service",
			Namespace: w.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: selectLabels,
			Ports: []corev1.ServicePort{
				{
					Port:     80,
					Protocol: "TCP",
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *WebAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.WebApp{}).
		Named("webapp").
		Complete(r)
}
