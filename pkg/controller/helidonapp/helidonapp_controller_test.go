// Copyright (c) 2020, Oracle Corporation and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package helidonapp

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/stretchr/testify/assert"
	vz "github.com/verrazzano/verrazzano-helidon-app-operator/pkg/apis/verrazzano/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

func TestNewService(t *testing.T) {
	appName := "myHelidonApp"
	appNs := "myns"
	app := vz.HelidonApp{}
	app.Spec.Name = appName
	app.Spec.Namespace = appNs
	svc := newService(&app)
	assert.Equal(t, corev1.ServiceTypeClusterIP, svc.Spec.Type, "Expected ServiceTypeClusterIP")
	assert.Equal(t, 1, len(svc.Spec.Ports), "Expected 1 svc.Spec.Port")
	port := svc.Spec.Ports[0]
	assert.Equal(t, 8080, int(port.Port), "Expected default port")
	assert.Equal(t, intstr.FromInt(8080), port.TargetPort, "Expected default targetPort")
	expectedPort := int32(8010)
	expectedTargetPort := int32(8011)
	app.Spec.Port = expectedPort
	app.Spec.TargetPort = expectedTargetPort
	svc = newService(&app)
	t.Log("Generated Service", svc)
	assert.Equal(t, appName, svc.Name, "Expected Name")
	assert.Equal(t, appNs, svc.Namespace, "Expected Namespace")
	assert.Equal(t, corev1.ServiceTypeClusterIP, svc.Spec.Type, "Expected ServiceTypeClusterIP")
	assert.Equal(t, 1, len(svc.Spec.Ports), "Expected 1 svc.Spec.Port")
	port = svc.Spec.Ports[0]
	assert.Equal(t, expectedPort, port.Port, "Expected default port")
	assert.Equal(t, intstr.FromInt(int(expectedTargetPort)), port.TargetPort, "Expected default targetPort")

	// Verify that target port is set to a non-default port if the target port is not specified
	expectedPort = int32(8079)
	expectedTargetPort = int32(8079)
	app.Spec.Port = expectedPort
	app.Spec.TargetPort = 0
	svc = newService(&app)
	assert.Equal(t, 1, len(svc.Spec.Ports), "Expected 1 svc.Spec.Port")
	port = svc.Spec.Ports[0]
	assert.Equal(t, expectedPort, port.Port, "Expected default port")
	assert.Equal(t, intstr.FromInt(int(expectedTargetPort)), port.TargetPort, "Expected default targetPort")

}

// Test Helidon CR that specified volumes
func TestNewDeploymentWithVolumes(t *testing.T) {
	appName := "myHelidonApp"
	appNs := "myns"
	app := vz.HelidonApp{}
	app.Spec.Name = appName
	app.Spec.Namespace = appNs
	deploy := newDeployment(&app)
	assert.Equal(t, 0, len(deploy.Spec.Template.Spec.Volumes), "Expected 0 volumes for deployment")

	app.Spec.Volumes = createVolumes()
	deploy = newDeployment(&app)
	assert.Equal(t, 1, len(deploy.Spec.Template.Spec.Volumes), "Expected 1 volume for deployment")
	name := "varlog"
	assert.Equal(t, name, deploy.Spec.Template.Spec.Volumes[0].Name, fmt.Sprintf("Expected volume name to be %s", name))
	name = "/var/log"
	assert.Equal(t, name, deploy.Spec.Template.Spec.Volumes[0].VolumeSource.HostPath.Path, fmt.Sprintf("Expected volume hostpath to be %s", name))
}

// Test Helidon CR that specified sidecar containers
func TestNewDeploymentWithContainers(t *testing.T) {
	appName := "myHelidonApp"
	appNs := "myns"
	appImage := "myImage"
	app := vz.HelidonApp{}
	app.Spec.Name = appName
	app.Spec.Namespace = appNs
	app.Spec.Image = appImage
	deploy := newDeployment(&app)
	assert.Equal(t, 1, len(deploy.Spec.Template.Spec.Containers), "Expected 1 container for deployment")
	assert.Equal(t, appName, deploy.Spec.Template.Spec.Containers[0].Name, fmt.Sprintf("Expected name to be %s", appName))
	assert.Equal(t, appImage, deploy.Spec.Template.Spec.Containers[0].Image, fmt.Sprintf("Expected image to be %s", appImage))

	app.Spec.Containers = createContainers()
	deploy = newDeployment(&app)
	assert.Equal(t, 2, len(deploy.Spec.Template.Spec.Containers), "Expected 2 containers for deployment")
	assert.Equal(t, appName, deploy.Spec.Template.Spec.Containers[0].Name, fmt.Sprintf("Expected name to be %s", appName))
	assert.Equal(t, appImage, deploy.Spec.Template.Spec.Containers[0].Image, fmt.Sprintf("Expected image to be %s", appImage))
	assert.Equal(t, "sidecar-name", deploy.Spec.Template.Spec.Containers[1].Name, "Expected name to be sidecar-name")
	assert.Equal(t, "sidecar-image", deploy.Spec.Template.Spec.Containers[1].Image, "Expected name to be sidecar-image")
}

func createVolumes() []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "varlog",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/var/log",
				},
			},
		},
	}
}

func createContainers() []corev1.Container {
	return []corev1.Container{
		{
			Name:  "sidecar-name",
			Image: "sidecar-image",
		},
	}
}
