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
	"k8s.io/apimachinery/pkg/util/intstr"

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

// config mmap function
func (r *CustomDeploymentReconciler) CreateOrUpdateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) error {
	existingConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKey{Name: configMap.Name, Namespace: configMap.Namespace}, existingConfigMap)
	if err != nil {
		// ConfigMap doesn't exist, create it
		if err := r.Create(ctx, configMap); err != nil {
			return err
		}
		return nil
	}

	existingConfigMap.Data = configMap.Data

	if err := r.Update(ctx, existingConfigMap); err != nil {
		return err
	}
	return nil
}

func (r *CustomDeploymentReconciler) CreateOrUpdateSecret(ctx context.Context, secret *corev1.Secret) error {
	existSecret := &corev1.Secret{}

	err := r.Get(ctx, client.ObjectKey{Name: secret.Name, Namespace: secret.Namespace}, existSecret)

	if err != nil {
		if err := r.Create(ctx, secret); err != nil {
			return err
		}
		return nil
	}

	existSecret.Data = secret.Data
	if err := r.Update(ctx, secret); err != nil {
		return err
	}
	return nil
}

func (r *CustomDeploymentReconciler) GetSecret(ctx context.Context, name string, nameSpace string) (*corev1.Secret, error) {
	getSecret := &corev1.Secret{}

	err := r.Get(ctx, client.ObjectKey{Name: name, Namespace: nameSpace}, getSecret)

	if err != nil {
		return nil, err
	}

	return getSecret, nil

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
	customResource := &mycustomalphav1.CustomDeployment{}

	err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, customResource)
	if err != nil {
		log.Info(fmt.Sprintf("\nx HHHHHHHHHHH -====- %v -====- HHHHHHHHHHH\n", err))
		return ctrl.Result{}, err
	}

	deployment := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: customResource.Name, Namespace: customResource.Namespace}, deployment)
	if err == nil {
		log.Info(fmt.Sprintf("\nx HHHHHHHHHHH -====- %v -====- HHHHHHHHHHH\n", err))
		deployment.Spec.Replicas = &customResource.Spec.Replicas
		er := r.Update(ctx, deployment)
		if er != nil {
			log.Error(er, "Error duing update")
			return ctrl.Result{}, nil
		}
		log.Info("====== Update success ===========\n")
		return ctrl.Result{}, nil
	}

	podLabel := map[string]string{
		"app": customResource.Spec.Selector,
	}

	// configMap := corev1.ConfigMap{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      "Config name",
	// 		Namespace: "Default",
	// 		Labels:    podLabel,
	// 	},
	// 	Data: map[string]string{
	// 		"PORT":     "3000",
	// 		"BASE_API": "https://abc.com",
	// 	},
	// }

	// secret := corev1.Secret{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      "Secret name",
	// 		Namespace: "my namespace",
	// 		Labels:    podLabel,
	// 	},
	// }

	// r.CreateOrUpdateConfigMap(ctx, &configMap)
	// r.CreateOrUpdateSecret(ctx, &secret)

	getSecret, err := r.GetSecret(ctx, "my-secret", "secret-vault")

	if err != nil {
		log.Info("Something went wrong")
	}

	log.Info(fmt.Sprintf("\nHERE IS YOUR SECRET: %v  %v\n ", string(getSecret.Data["username"]), string(getSecret.Data["password"])))

	// metav1.TypeMeta{
	// 	Kind:       "adsfb",
	// 	APIVersion: "v1",
	// }

	service := &corev1.Service{}

	err = r.Get(ctx, types.NamespacedName{Name: customResource.Name, Namespace: customResource.Namespace}, service)

	if err != nil {
		log.Info("================= I am here =j=============")
		service = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      customResource.Name,
				Namespace: customResource.Namespace,
				Labels:    podLabel,
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": customResource.Spec.Selector,
				},
				Ports: []corev1.ServicePort{
					{
						Name:       "http",
						Protocol:   corev1.ProtocolTCP,
						Port:       customResource.Spec.ContainerPort,
						TargetPort: intstr.FromInt(customResource.Spec.TargetPort),
					},
				},
				Type: corev1.ServiceTypeNodePort,
			},
			// Spec: appsv1,
		}

		if err := r.Create(ctx, service); err != nil {
			log.Error(err, "FAILED TO CREATE SERVICE")
			return ctrl.Result{}, err
		}
	}

	deployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      customResource.Name,
			Namespace: customResource.Namespace,
			Labels:    podLabel,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &customResource.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabel,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					// Name:      customResource.Name,
					// Namespace: customResource.Namespace,
					Labels: podLabel,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "my-nginx",
							Image: customResource.Spec.Image,

							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: customResource.Spec.ContainerPort,
								},
							},
						},
					},
				},
			},
		},
	}

	err = r.Create(ctx, deployment)
	if err != nil {
		log.Error(err, "Failed to create Deployment :(")
	}

	// TODO(user): your logic here

	return ctrl.Result{RequeueAfter: time.Duration(5 * time.Second)}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mycustomalphav1.CustomDeployment{}).
		Complete(r)
}
