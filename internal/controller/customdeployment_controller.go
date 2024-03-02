/*
Copyright 2024.

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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mycustomalphav1 "github.com/deployemtn.aas/api/alphav1"
)

// CustomDeploymentReconciler reconciles a CustomDeployment object
type CustomDeploymentReconciler struct {
        client.Client
        Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mycustom.deployment.aas,resources=customdeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mycustom.deployment.aas,resources=customdeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mycustom.deployment.aas,resources=customdeployments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CustomDeployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *CustomDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
        log := log.FromContext(ctx)
        appCr := &mycustomalphav1.CustomDeployment{}

        err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, appCr)
        if err != nil {
                log.Info(fmt.Sprintf("faied to get %s/%s", appCr.Name, appCr.Namespace))
                return ctrl.Result{}, err
        }

        var replicas int32
        currentTime := time.Now()
        currentHour := currentTime.Hour() + 5
        currentMin := currentTime.Minute()

        log.Info(fmt.Sprintf("\nP!! %v \n %v \n %v", appCr.Spec.PickHourStart, appCr.Spec.PickHourEnd, appCr.Spec.PickMinsEnd))
        log.Info(fmt.Sprintf("\nP22 %v \n %v \n %v\n", currentHour, currentMin,appCr.Status.IsFirstTime))
        log.Info(fmt.Sprintf("\n======================== \nHERE WE GO\n====================\n %v",currentHour))

        if currentHour >= int(appCr.Spec.PickHourStart) && currentMin >= int(appCr.Spec.PickMinsEnd) && currentHour <= int(appCr.Spec.PickHourEnd) {
                replicas = 10
        } else {
                replicas = 2
        }

        podLabel := map[string]string{
                "tata": "mata",
        }

        deployment := &appsv1.Deployment{
                ObjectMeta: metav1.ObjectMeta{
                        Name:      appCr.Name,
                        Namespace: appCr.Namespace,
                        //Labels:    podLabel,
                },
                Spec: appsv1.DeploymentSpec{
                        Replicas: &replicas,
                        Selector: &metav1.LabelSelector{
                                MatchLabels: podLabel,
                        },
                        Template: corev1.PodTemplateSpec{
                                ObjectMeta: metav1.ObjectMeta{
                                        // Name:      appCr.Name,
                                        // Namespace: appCr.Namespace,
                                        Labels: podLabel,
                                },
                                Spec: corev1.PodSpec{
                                        Containers: []corev1.Container{
                                                {
                                                        Name:  "my-nginx",
                                                        Image: appCr.Spec.Image,

                                                        Ports: []corev1.ContainerPort{
                                                                {
                                                                        Name:          "http",
                                                                        Protocol:      corev1.ProtocolTCP,
                                                                        ContainerPort: appCr.Spec.Port,
                                                                },
                                                        },
                                                },
                                        },
                                },
                        },
                },
        }


        if appCr.Status.IsFirstTime {
        // If it's not the first time, update the existing Deployment
        err = r.Update(ctx, deployment)
        if err != nil {
            log.Error(err, "Failed to update Deployment")
            return ctrl.Result{}, err
        }
    } else {
        // If it's the first time, create the Deployment
        err = r.Create(ctx, deployment)
        if err != nil {
            log.Error(err, "Failed to create Deployment")
            return ctrl.Result{}, err
        }
        // Update status to mark it's not the first time
        appCr.Status.IsFirstTime = true
        err = r.Status().Update(ctx, appCr)
        if err != nil {
            log.Error(err, "Failed to update CustomDeployment status")
            return ctrl.Result{}, err
        }
    }

        // TODO(user): your logic here

        return ctrl.Result{RequeueAfter: time.Duration(30 * time.Second)}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
        return ctrl.NewControllerManagedBy(mgr).
                For(&mycustomalphav1.CustomDeployment{}).
                Complete(r)
}
                      