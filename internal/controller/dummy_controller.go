/*
Copyright 2025.

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
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	interviewv1alpha1 "github.com/mquhuy/dummy-controller/api/v1alpha1"
)

// DummyReconciler reconciles a Dummy object
type DummyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var logger logr.Logger

const dummyFinalizer = "interview.com/finalizer"

// +kubebuilder:rbac:groups=interview.com,resources=dummies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=interview.com,resources=dummies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=interview.com,resources=dummies/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;delete

func (r *DummyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger = log.FromContext(ctx)

	dummyObj := &interviewv1alpha1.Dummy{}
	err := r.Get(ctx, req.NamespacedName, dummyObj)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get dummy", "dummyName", req.Name)
		return ctrl.Result{}, err
	}

	nginxPodName := fmt.Sprintf("%s-nginx", req.Name)

	// Check if the Dummy resource is being deleted, in that case we delete the nginx pod
	// before removing the finalizer.
	if !dummyObj.DeletionTimestamp.IsZero() {
		// Delete the Nginx pod if it exists
		pod := &corev1.Pod{}
		err := r.Get(ctx, types.NamespacedName{Name: nginxPodName, Namespace: req.Namespace}, pod)
		if err == nil {
			// If the pod is already being deleted, wait for another 1s
			if !pod.DeletionTimestamp.IsZero() {
				return ctrl.Result{RequeueAfter: 1}, nil
			}
			if deleteErr := r.Delete(ctx, pod); deleteErr != nil {
				logger.Error(deleteErr, "Failed to delete nginx pod", "podName", nginxPodName)
				return ctrl.Result{}, deleteErr
			}
			logger.Info("Nginx pod was deleted", "podName", nginxPodName)
			// Wait for 5s before requeuing, to make sure the pod is deleted.
			return ctrl.Result{}, nil
		}
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error checking nginx pod", "podName", nginxPodName)
			return ctrl.Result{}, err
		}
		// Remove finalizer
		finalizers := dummyObj.GetFinalizers()
		// Generation check to prevent duplicate updates
		if dummyObj.Generation == dummyObj.Status.ObservedGeneration {
			logger.Info("Skipping finalizer removal - already processed this generation")
			return ctrl.Result{}, nil
		}
		if containsString(finalizers, dummyFinalizer) {
			dummyObj.SetFinalizers(removeString(finalizers, dummyFinalizer))
			dummyObj.Status.ObservedGeneration = dummyObj.Generation
			if err := r.Update(ctx, dummyObj); err != nil {
				logger.Error(err, "Failed to remove finalizer from Dummy", "dummyName", req.Name)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	}

	// Add a finalizer to the Dummy resource if it doesn't already have one
	if !containsString(dummyObj.GetFinalizers(), dummyFinalizer) {
		// Generation check to prevent duplicate updates
		if dummyObj.Generation == dummyObj.Status.ObservedGeneration {
			logger.Info("Skipping finalizer addition - already processed this generation")
			return ctrl.Result{}, nil
		}
		logger.Info("Adding finalizer to Dummy", "dummyName", req.Name)
		dummyObj.SetFinalizers(append(dummyObj.GetFinalizers(), dummyFinalizer))
		dummyObj.Status.ObservedGeneration = dummyObj.Generation
		err := r.Update(ctx, dummyObj)
		if err != nil {
			logger.Error(err, "Failed to add finalizer to Dummy", "dummyName", req.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Check if an Nginx pod exists
	pod := &corev1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: nginxPodName, Namespace: req.Namespace}, pod)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Pod does not exist, create it
			logger.Info("Nginx pod not found. Creating.", "podName", nginxPodName)
			pod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      nginxPodName,
					Namespace: req.Namespace,
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			}
			// set dummyObj as the owner of the nginx pod
			if err := controllerutil.SetControllerReference(dummyObj, pod, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
			if err := r.Create(ctx, pod); err != nil {
				logger.Error(err, "Failed to create Nginx pod", "podName", nginxPodName)
				return ctrl.Result{}, err
			}
			logger.Info("Nginx pod created successfully", "podName", nginxPodName)
		} else {
			logger.Error(err, "Failed to get nginx pod", "podName", nginxPodName)
			return ctrl.Result{}, err
		}
	}

	// Update Dummy status with SpecEcho if a message exists
	if dummyObj.Spec.Message != "" {
		logger.Info("", "Dummy Name", dummyObj.Name, "Namespace", dummyObj.Namespace, "Message", dummyObj.Spec.Message)
		dummyObj.Status.SpecEcho = dummyObj.Spec.Message
	}

	// Check the status of the Nginx pod
	if pod.Status.Phase != "" {
		dummyObj.Status.PodStatus = pod.Status.Phase
	}

	if err := r.Status().Update(ctx, dummyObj); err != nil {
		logger.Error(err, "Failed to update Dummy status", "dummyName", req.Name)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// Helper function to check if a finalizer exists
func containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// Helper function to remove a finalizer from a slice
func removeString(slice []string, str string) []string {
	for i, item := range slice {
		if item == str {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// SetupWithManager sets up the controller with the Manager.
func (r *DummyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&interviewv1alpha1.Dummy{}).
		Owns(&corev1.Pod{},
			builder.WithPredicates(predicate.Funcs{
				// Filter only events for Pods with status.phase in transitioning
				UpdateFunc: func(e event.UpdateEvent) bool {
					oldPod := e.ObjectOld.(*corev1.Pod)
					newPod := e.ObjectNew.(*corev1.Pod)
					// Reconcile only if the Pod status phase transitions.
					return oldPod.Status.Phase != newPod.Status.Phase
				},
				// Ignore create events
				CreateFunc: func(e event.CreateEvent) bool {
					return false
				},
				// Ignore delete events
				DeleteFunc: func(e event.DeleteEvent) bool {
					return false
				},
				// Ignore generic events
				GenericFunc: func(e event.GenericEvent) bool {
					return false
				},
			}),
		).
		Complete(r)
}
