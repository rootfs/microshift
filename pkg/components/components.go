package components

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	batchclientv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/microshift/pkg/config"
)

func StartComponents(cfg *config.MicroshiftConfig) error {
	if len(cfg.Components) == 0 {
		return nil
	}
	componentLoadNamespace := "component-loader-ns"
	restConfig, err := clientcmd.BuildConfigFromFlags("", cfg.DataDir+"/resources/kubeadmin/kubeconfig")
	if err != nil {
		return err
	}
	// create the namespace for component loader
	coreclient := coreclientv1.NewForConfigOrDie(rest.AddUserAgent(restConfig, "core-agent"))
	ns := corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: componentLoadNamespace,
		},
	}
	_, err = coreclient.Namespaces().Create(context.TODO(), &ns, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	batchclient := batchclientv1.NewForConfigOrDie(rest.AddUserAgent(restConfig, "batch-agent"))

	for _, component := range cfg.Components {
		image := component.Image
		// create a ConfigMap for loader to report readiness
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "microshift-component-loader-status-" + component.Name,
				Namespace: componentLoadNamespace,
			},
			Data: map[string]string{
				"Status": "Init",
			},
		}
		_, err = coreclient.ConfigMaps(componentLoadNamespace).Create(context.TODO(), cm, metav1.CreateOptions{})
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}

		args := []string{}
		volumes := []corev1.Volume{}
		volMounts := []corev1.VolumeMount{}
		if len(component.Parameters) > 0 {
			// if the component loader has parameters,
			// create a ConfigMap and volume-mount it
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "microshift-component-loader-" + component.Name,
					Namespace: componentLoadNamespace,
				},
				Data: component.Parameters,
			}
			_, err = coreclient.ConfigMaps(componentLoadNamespace).Create(context.TODO(), cm, metav1.CreateOptions{})
			if err != nil && !apierrors.IsAlreadyExists(err) {
				return err
			}
			mountPath := "/var/lib/loader.config"
			args = []string{"-c", mountPath}
			vol := corev1.VolumeMount{Name: "config", MountPath: mountPath}
			volMounts = append(volMounts, vol)
			volume := corev1.Volume{
				Name: "config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "microshift-component-loader-" + component.Name,
						},
					},
				},
			}
			volumes = append(volumes, volume)
		}
		container := corev1.Container{
			Name:         "microshift-component-loader",
			Image:        image,
			Command:      []string{"/loader"},
			Args:         args,
			VolumeMounts: volMounts,
			Env: []corev1.EnvVar{
				corev1.EnvVar{
					Name:  "MICROSHIFT_READINESS_STATUS_CONFIGMAP",
					Value: "microshift-component-loader-" + component.Name,
				},
			},
		}
		// create a job to start component loader
		job := &batchv1.Job{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Job",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "microshift-component-loader-" + component.Name,
				Namespace: componentLoadNamespace,
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							container,
						},
						Volumes: volumes,
					},
				},
			},
		}

		_, err = batchclient.Jobs(componentLoadNamespace).Create(context.TODO(), job, metav1.CreateOptions{})
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
		// watch job status
		// check status in configmap
	}
	return nil
}
