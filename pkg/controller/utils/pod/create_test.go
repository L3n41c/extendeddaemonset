// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.

package pod

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	datadoghqv1alpha1 "github.com/datadog/extendeddaemonset/pkg/apis/datadoghq/v1alpha1"
	ctrltest "github.com/datadog/extendeddaemonset/pkg/controller/test"
	"github.com/google/go-cmp/cmp"
)

func Test_overwriteResourcesFromNode(t *testing.T) {
	type args struct {
		template     *corev1.PodTemplateSpec
		edsNamespace string
		edsName      string
		node         *corev1.Node
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		wantTemplate *corev1.PodTemplateSpec
	}{
		{
			name: "nil node",
			args: args{
				template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "container1"},
						},
					},
				},
				edsNamespace: "bar",
				edsName:      "foo",
			},
			wantErr: false,
		},
		{
			name: "no annotation",
			args: args{
				template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "container1"},
						},
					},
				},
				edsNamespace: "bar",
				edsName:      "foo",
				node:         ctrltest.NewNode("node1", nil),
			},
			wantErr: false,
		},
		{
			name: "annotation requests.cpu",
			args: args{
				template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "container1"},
						},
					},
				},
				edsNamespace: "bar",
				edsName:      "foo",
				node: ctrltest.NewNode("node1", &ctrltest.NewNodeOptions{
					Annotations: map[string]string{
						fmt.Sprintf(datadoghqv1alpha1.ExtendedDaemonSetRessourceNodeAnnotationKey, "bar", "foo", "container1"): `{"Requests": {"cpu": "1.5"}}`,
					}}),
			},
			wantErr: false,
			wantTemplate: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "container1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("1.5"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "annotation requests.cpu",
			args: args{
				template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "container1"},
						},
					},
				},
				edsNamespace: "bar",
				edsName:      "foo",
				node: ctrltest.NewNode("node1", &ctrltest.NewNodeOptions{
					Annotations: map[string]string{
						fmt.Sprintf(datadoghqv1alpha1.ExtendedDaemonSetRessourceNodeAnnotationKey, "bar", "foo", "container1"): `{"Requests": invalid {"cpu": "1.5"}}`,
					}}),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := overwriteResourcesFromNode(tt.args.template, tt.args.edsNamespace, tt.args.edsName, tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("overwriteResourcesFromNode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantTemplate != nil {
				if diff := cmp.Diff(tt.wantTemplate, tt.args.template); diff != "" {
					t.Errorf("template mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
