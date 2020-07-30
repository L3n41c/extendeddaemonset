// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.

package extendeddaemonset

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	datadoghqv1alpha1 "github.com/datadog/extendeddaemonset/pkg/apis/datadoghq/v1alpha1"
	test "github.com/datadog/extendeddaemonset/pkg/apis/datadoghq/v1alpha1/test"
	commontest "github.com/datadog/extendeddaemonset/pkg/controller/test"
	"github.com/datadog/extendeddaemonset/pkg/controller/utils/comparison"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func TestReconcileExtendedDaemonSet_selectNodes(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	log = logf.Log.WithName("TestReconcileExtendedDaemonSet_selectNodes")

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(datadoghqv1alpha1.SchemeGroupVersion, &datadoghqv1alpha1.ExtendedDaemonSet{})

	nodeOptions := &commontest.NewNodeOptions{
		Conditions: []corev1.NodeCondition{
			{
				Type:   corev1.NodeReady,
				Status: corev1.ConditionTrue,
			},
		},
	}
	node1 := commontest.NewNode("node1", nodeOptions)
	node2 := commontest.NewNode("node2", nodeOptions)
	node3 := commontest.NewNode("node3", nodeOptions)
	intString3 := intstr.FromInt(3)

	node2.Labels = map[string]string{
		"canary": "true",
	}

	options1 := &test.NewExtendedDaemonSetOptions{
		Canary: &datadoghqv1alpha1.ExtendedDaemonSetSpecStrategyCanary{
			Replicas: &intString3,
		},
		Status: &datadoghqv1alpha1.ExtendedDaemonSetStatus{
			ActiveReplicaSet: "foo-1",
			Canary: &datadoghqv1alpha1.ExtendedDaemonSetStatusCanary{
				ReplicaSet: "foo-2",
				Nodes:      []string{},
			},
		},
	}
	extendeddaemonset1 := test.NewExtendedDaemonSet("bar", "foo", options1)

	intString1 := intstr.FromInt(1)

	options2 := &test.NewExtendedDaemonSetOptions{
		Canary: &datadoghqv1alpha1.ExtendedDaemonSetSpecStrategyCanary{
			Replicas: &intString1,
			NodeSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"canary": "true",
				},
			},
		},
		Status: &datadoghqv1alpha1.ExtendedDaemonSetStatus{
			ActiveReplicaSet: "foo-1",
			Canary: &datadoghqv1alpha1.ExtendedDaemonSetStatusCanary{
				ReplicaSet: "foo-2",
				Nodes:      []string{},
			},
		},
	}
	extendeddaemonset2 := test.NewExtendedDaemonSet("bar", "foo", options2)

	type fields struct {
		client client.Client
		scheme *runtime.Scheme
	}
	type args struct {
		spec         *datadoghqv1alpha1.ExtendedDaemonSetSpec
		replicaset   *datadoghqv1alpha1.ExtendedDaemonSetReplicaSet
		canaryStatus *datadoghqv1alpha1.ExtendedDaemonSetStatusCanary
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		wantFunc func(*datadoghqv1alpha1.ExtendedDaemonSetStatusCanary) bool
	}{
		{
			name: "enough nodes",
			fields: fields{
				scheme: s,
				client: fake.NewFakeClient([]runtime.Object{node1, node2, node3}...),
			},
			args: args{
				spec:       &extendeddaemonset1.Spec,
				replicaset: &datadoghqv1alpha1.ExtendedDaemonSetReplicaSet{},
				canaryStatus: &datadoghqv1alpha1.ExtendedDaemonSetStatusCanary{
					ReplicaSet: "foo",
					Nodes:      []string{},
				},
			},
			wantErr: false,
		},
		{
			name: "missing nodes",
			fields: fields{
				scheme: s,
				client: fake.NewFakeClient([]runtime.Object{node1, node2}...),
			},
			args: args{
				spec:       &extendeddaemonset1.Spec,
				replicaset: &datadoghqv1alpha1.ExtendedDaemonSetReplicaSet{},
				canaryStatus: &datadoghqv1alpha1.ExtendedDaemonSetStatusCanary{
					ReplicaSet: "foo",
					Nodes:      []string{},
				},
			},
			wantErr: true,
		},
		{
			name: "enough nodes",
			fields: fields{
				scheme: s,
				client: fake.NewFakeClient([]runtime.Object{node1, node2, node3}...),
			},
			args: args{
				spec:       &extendeddaemonset1.Spec,
				replicaset: &datadoghqv1alpha1.ExtendedDaemonSetReplicaSet{},
				canaryStatus: &datadoghqv1alpha1.ExtendedDaemonSetStatusCanary{
					ReplicaSet: "foo",
					Nodes:      []string{node1.Name},
				},
			},
			wantErr: false,
		},
		{
			name: "dedicated canary nodes",
			fields: fields{
				scheme: s,
				client: fake.NewFakeClient([]runtime.Object{node1, node2, node3}...),
			},
			args: args{
				spec:       &extendeddaemonset2.Spec,
				replicaset: &datadoghqv1alpha1.ExtendedDaemonSetReplicaSet{},
				canaryStatus: &datadoghqv1alpha1.ExtendedDaemonSetStatusCanary{
					ReplicaSet: "foo",
					Nodes:      []string{},
				},
			},
			wantErr: false,
			wantFunc: func(canaryStatus *datadoghqv1alpha1.ExtendedDaemonSetStatusCanary) bool {
				return len(canaryStatus.Nodes) == 1 && canaryStatus.Nodes[0] == "node2"
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqLogger := log.WithValues("test:", tt.name)
			r := &ReconcileExtendedDaemonSet{
				client: tt.fields.client,
				scheme: tt.fields.scheme,
			}
			if err := r.selectNodes(reqLogger, tt.args.spec, tt.args.replicaset, tt.args.canaryStatus); (err != nil) != tt.wantErr {
				t.Errorf("ReconcileExtendedDaemonSet.selectNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantFunc != nil && !tt.wantFunc(tt.args.canaryStatus) {
				t.Errorf("ReconcileExtendedDaemonSet.selectNodes() didn’t pass the post-run checks")
			}
		})
	}
}

func Test_newReplicaSetFromInstance(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	tests := []struct {
		name      string
		daemonset *datadoghqv1alpha1.ExtendedDaemonSet
		want      *datadoghqv1alpha1.ExtendedDaemonSetReplicaSet
		wantErr   bool
	}{
		{
			name:      "default test",
			daemonset: test.NewExtendedDaemonSet("bar", "foo", nil),
			wantErr:   false,
			want: &datadoghqv1alpha1.ExtendedDaemonSetReplicaSet{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    "bar",
					GenerateName: "foo-",
					Labels:       map[string]string{"extendeddaemonset.datadoghq.com/name": "foo"},
					Annotations:  map[string]string{"extendeddaemonset.datadoghq.com/templatehash": "a2bb34618483323482d9a56ae2515eed"},
				},
				Spec: datadoghqv1alpha1.ExtendedDaemonSetReplicaSetSpec{
					TemplateGeneration: "a2bb34618483323482d9a56ae2515eed",
				},
			},
		},
		{
			name:      "with label",
			daemonset: test.NewExtendedDaemonSet("bar", "foo", &test.NewExtendedDaemonSetOptions{Labels: map[string]string{"foo-key": "bar-value"}}),
			wantErr:   false,
			want: &datadoghqv1alpha1.ExtendedDaemonSetReplicaSet{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    "bar",
					GenerateName: "foo-",
					Labels:       map[string]string{"foo-key": "bar-value", "extendeddaemonset.datadoghq.com/name": "foo"},
					Annotations:  map[string]string{"extendeddaemonset.datadoghq.com/templatehash": "a2bb34618483323482d9a56ae2515eed"},
				},
				Spec: datadoghqv1alpha1.ExtendedDaemonSetReplicaSetSpec{
					TemplateGeneration: "a2bb34618483323482d9a56ae2515eed",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newReplicaSetFromInstance(tt.daemonset)
			if (err != nil) != tt.wantErr {
				t.Errorf("newReplicaSetFromInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !apiequality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("newReplicaSetFromInstance() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_selectCurrentReplicaSet(t *testing.T) {
	now := time.Now()
	t.Logf("now: %v", now)
	creationTimeDaemonset := now.Add(-10 * time.Minute)
	creationTimeRSDone := now.Add(-6 * time.Minute)

	replicassetUpToDate := test.NewExtendedDaemonSetReplicaSet("bar", "foo-1", &test.NewExtendedDaemonSetReplicaSetOptions{
		CreationTime: &now,
		Labels:       map[string]string{"foo-key": "bar-value"}})

	replicassetUpToDateDone := test.NewExtendedDaemonSetReplicaSet("bar", "foo-1", &test.NewExtendedDaemonSetReplicaSetOptions{
		CreationTime: &creationTimeRSDone,
		Labels:       map[string]string{"foo-key": "bar-value"}})
	replicassetOld := test.NewExtendedDaemonSetReplicaSet("bar", "foo-old", &test.NewExtendedDaemonSetReplicaSetOptions{
		CreationTime: &creationTimeDaemonset,
		Labels:       map[string]string{"foo-key": "old-value"}})

	daemonset := test.NewExtendedDaemonSet("bar", "foo", &test.NewExtendedDaemonSetOptions{Labels: map[string]string{"foo-key": "bar-value"}})
	intString1 := intstr.FromInt(1)
	daemonsetWithCanary := test.NewExtendedDaemonSet("bar", "foo", &test.NewExtendedDaemonSetOptions{
		CreationTime: &creationTimeDaemonset,
		Labels:       map[string]string{"foo-key": "bar-value"},
		Canary: &datadoghqv1alpha1.ExtendedDaemonSetSpecStrategyCanary{
			Replicas: &intString1,
			Duration: &metav1.Duration{Duration: 5 * time.Minute},
		},
		Status: &datadoghqv1alpha1.ExtendedDaemonSetStatus{
			ActiveReplicaSet: replicassetOld.Name,
		},
	})
	daemonsetWithCanaryValid := test.NewExtendedDaemonSet("bar", "foo", &test.NewExtendedDaemonSetOptions{
		CreationTime: &creationTimeDaemonset,
		Labels:       map[string]string{"foo-key": "bar-value"},
		Annotations:  map[string]string{datadoghqv1alpha1.ExtendedDaemonSetCanaryValidAnnotationKey: "foo-1"},
		Canary: &datadoghqv1alpha1.ExtendedDaemonSetSpecStrategyCanary{
			Replicas: &intString1,
			Duration: &metav1.Duration{Duration: 5 * time.Minute},
		},
		Status: &datadoghqv1alpha1.ExtendedDaemonSetStatus{
			ActiveReplicaSet: replicassetOld.Name,
		},
	})

	type args struct {
		daemonset  *datadoghqv1alpha1.ExtendedDaemonSet
		upToDateRS *datadoghqv1alpha1.ExtendedDaemonSetReplicaSet
		activeRS   *datadoghqv1alpha1.ExtendedDaemonSetReplicaSet
		now        time.Time
	}
	tests := []struct {
		name  string
		args  args
		want  *datadoghqv1alpha1.ExtendedDaemonSetReplicaSet
		want1 time.Duration
	}{
		{
			name: "one RS, update to date",
			args: args{
				daemonset:  daemonset,
				upToDateRS: replicassetUpToDate,
				activeRS:   replicassetUpToDate,
				now:        now,
			},
			want:  replicassetUpToDate,
			want1: 0,
		},
		{
			name: "two RS, update to date, canary not set",
			args: args{
				daemonset:  daemonset,
				upToDateRS: replicassetUpToDate,
				activeRS:   replicassetOld,
				now:        now,
			},
			want:  replicassetUpToDate,
			want1: 0,
		},
		{
			name: "two RS, update to date, canary set not done",
			args: args{
				daemonset:  daemonsetWithCanary,
				upToDateRS: replicassetUpToDate,
				activeRS:   replicassetOld,
				now:        now,
			},

			want:  replicassetOld,
			want1: 5 * time.Minute,
		},
		{
			name: "two RS, update to date, canary set and done",
			args: args{
				daemonset:  daemonsetWithCanary,
				upToDateRS: replicassetUpToDateDone,
				activeRS:   replicassetOld,
				now:        now,
			},
			want:  replicassetUpToDateDone,
			want1: -time.Minute,
		},
		{
			name: "two RS, update to date, canary set, canary duration not done, canary valid",
			args: args{
				daemonset:  daemonsetWithCanaryValid,
				upToDateRS: replicassetUpToDate,
				activeRS:   replicassetOld,
				now:        now,
			},
			want:  replicassetUpToDate,
			want1: 5 * time.Minute,
		},
		{
			name: "two RS, update to date, canary set, canary duration done, canary valid",
			args: args{
				daemonset:  daemonsetWithCanaryValid,
				upToDateRS: replicassetUpToDateDone,
				activeRS:   replicassetOld,
				now:        now,
			},
			want:  replicassetUpToDateDone,
			want1: -time.Minute,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("daemonset: %v", tt.args.daemonset)
			got, got1 := selectCurrentReplicaSet(tt.args.daemonset, tt.args.activeRS, tt.args.upToDateRS, tt.args.now)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("selectCurrentReplicaSet() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("selectCurrentReplicaSet() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestReconcileExtendedDaemonSet_cleanupReplicaSet(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	log = logf.Log.WithName("TestReconcileExtendedDaemonSet_cleanupReplicaSet")

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(t.Logf)

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(datadoghqv1alpha1.SchemeGroupVersion, &datadoghqv1alpha1.ExtendedDaemonSetReplicaSet{})
	s.AddKnownTypes(datadoghqv1alpha1.SchemeGroupVersion, &datadoghqv1alpha1.ExtendedDaemonSet{})

	replicassetUpToDate := test.NewExtendedDaemonSetReplicaSet("bar", "foo-1", &test.NewExtendedDaemonSetReplicaSetOptions{
		Labels: map[string]string{"foo-key": "bar-value"}})
	replicassetCurrent := test.NewExtendedDaemonSetReplicaSet("bar", "current", &test.NewExtendedDaemonSetReplicaSetOptions{
		Labels: map[string]string{"foo-key": "bar-value"}})

	replicassetOld := test.NewExtendedDaemonSetReplicaSet("bar", "old", &test.NewExtendedDaemonSetReplicaSetOptions{
		Labels: map[string]string{"foo-key": "bar-value"}})

	type fields struct {
		client client.Client
		scheme *runtime.Scheme
	}
	type args struct {
		rsList       *datadoghqv1alpha1.ExtendedDaemonSetReplicaSetList
		current      *datadoghqv1alpha1.ExtendedDaemonSetReplicaSet
		updatetodate *datadoghqv1alpha1.ExtendedDaemonSetReplicaSet
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "nothing to delete",
			fields: fields{
				client: fake.NewFakeClient(),
				scheme: s,
			},
			args: args{
				rsList: &datadoghqv1alpha1.ExtendedDaemonSetReplicaSetList{
					Items: []datadoghqv1alpha1.ExtendedDaemonSetReplicaSet{},
				},
			},
			wantErr: false,
		},
		{
			name: "on RS to delete",
			fields: fields{
				client: fake.NewFakeClient(replicassetOld, replicassetUpToDate, replicassetCurrent),
				scheme: s,
			},
			args: args{
				rsList: &datadoghqv1alpha1.ExtendedDaemonSetReplicaSetList{
					Items: []datadoghqv1alpha1.ExtendedDaemonSetReplicaSet{*replicassetOld, *replicassetUpToDate, *replicassetCurrent},
				},
				updatetodate: replicassetUpToDate,
				current:      replicassetCurrent,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqLogger := log.WithValues("test:", tt.name)
			r := &ReconcileExtendedDaemonSet{
				client:   tt.fields.client,
				scheme:   tt.fields.scheme,
				recorder: eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: tt.name}),
			}
			if err := r.cleanupReplicaSet(reqLogger, tt.args.rsList, tt.args.current, tt.args.updatetodate); (err != nil) != tt.wantErr {
				t.Errorf("ReconcileExtendedDaemonSet.cleanupReplicaSet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReconcileExtendedDaemonSet_createNewReplicaSet(t *testing.T) {
	eventBroadcaster := record.NewBroadcaster()

	logf.SetLogger(logf.ZapLogger(true))
	log = logf.Log.WithName("TestReconcileExtendedDaemonSet_createNewReplicaSet")

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(datadoghqv1alpha1.SchemeGroupVersion, &datadoghqv1alpha1.ExtendedDaemonSetReplicaSet{})
	s.AddKnownTypes(datadoghqv1alpha1.SchemeGroupVersion, &datadoghqv1alpha1.ExtendedDaemonSet{})

	type fields struct {
		client client.Client
		scheme *runtime.Scheme
	}
	type args struct {
		logger    logr.Logger
		daemonset *datadoghqv1alpha1.ExtendedDaemonSet
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    reconcile.Result
		wantErr bool
	}{
		{
			name: "create new RS",
			fields: fields{
				client: fake.NewFakeClient(),
				scheme: s,
			},
			args: args{
				logger:    log,
				daemonset: test.NewExtendedDaemonSet("bar", "foo", &test.NewExtendedDaemonSetOptions{Labels: map[string]string{"foo-key": "bar-value"}}),
			},
			want:    reconcile.Result{Requeue: true},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReconcileExtendedDaemonSet{
				client:   tt.fields.client,
				scheme:   tt.fields.scheme,
				recorder: eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "TestReconcileExtendedDaemonSet_cleanupReplicaSet"}),
			}
			got, err := r.createNewReplicaSet(tt.args.logger, tt.args.daemonset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReconcileExtendedDaemonSet.createNewReplicaSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReconcileExtendedDaemonSet.createNewReplicaSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReconcileExtendedDaemonSet_updateInstanceWithCurrentRS(t *testing.T) {
	eventBroadcaster := record.NewBroadcaster()

	logf.SetLogger(logf.ZapLogger(true))
	log = logf.Log.WithName("TestReconcileExtendedDaemonSet_updateStatusWithNewRS")

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(datadoghqv1alpha1.SchemeGroupVersion, &datadoghqv1alpha1.ExtendedDaemonSetReplicaSet{})
	s.AddKnownTypes(datadoghqv1alpha1.SchemeGroupVersion, &datadoghqv1alpha1.ExtendedDaemonSet{})

	daemonset := test.NewExtendedDaemonSet("bar", "foo", &test.NewExtendedDaemonSetOptions{Labels: map[string]string{"foo-key": "bar-value"}})
	replicassetUpToDate := test.NewExtendedDaemonSetReplicaSet("bar", "foo-1", &test.NewExtendedDaemonSetReplicaSetOptions{
		Labels: map[string]string{"foo-key": "bar-value"}})
	replicassetCurrent := test.NewExtendedDaemonSetReplicaSet("bar", "current", &test.NewExtendedDaemonSetReplicaSetOptions{
		Labels: map[string]string{"foo-key": "current-value"},
		Status: &datadoghqv1alpha1.ExtendedDaemonSetReplicaSetStatus{
			Desired:   3,
			Available: 3,
			Ready:     2,
		}})

	daemonsetWithStatus := daemonset.DeepCopy()
	daemonsetWithStatus.ResourceVersion = "2"
	daemonsetWithStatus.Status = datadoghqv1alpha1.ExtendedDaemonSetStatus{
		ActiveReplicaSet: "current",
		Desired:          3,
		Current:          3,
		Available:        3,
		Ready:            2,
		UpToDate:         3,
		State:            "Running",
	}
	intString1 := intstr.FromInt(1)
	daemonsetWithCanaryWithStatus := daemonsetWithStatus.DeepCopy()
	{
		daemonsetWithCanaryWithStatus.Status.State = datadoghqv1alpha1.ExtendedDaemonSetStatusStateCanary
		daemonsetWithCanaryWithStatus.Spec.Strategy.Canary = &datadoghqv1alpha1.ExtendedDaemonSetSpecStrategyCanary{
			Replicas: &intString1,
			Duration: &metav1.Duration{Duration: 10 * time.Minute},
		}
		daemonsetWithCanaryWithStatus.Status.Canary = &datadoghqv1alpha1.ExtendedDaemonSetStatusCanary{
			Nodes:      []string{"node1"},
			ReplicaSet: "foo-1",
		}
	}
	daemonsetWithCanaryFailedOldStatus := daemonsetWithCanaryWithStatus.DeepCopy()
	{
		daemonsetWithCanaryFailedOldStatus.Annotations[datadoghqv1alpha1.ExtendedDaemonSetCanaryFailedAnnotationKey] = "true"
		daemonsetWithCanaryFailedOldStatus.Status.Canary = &datadoghqv1alpha1.ExtendedDaemonSetStatusCanary{
			Nodes:      []string{"node1"},
			ReplicaSet: "foo-1",
		}
	}
	daemonsetWithCanaryFailedNewStatus := daemonsetWithCanaryFailedOldStatus.DeepCopy()
	{
		daemonsetWithCanaryFailedNewStatus.ResourceVersion = "3"
		daemonsetWithCanaryFailedNewStatus.Status.State = datadoghqv1alpha1.ExtendedDaemonSetStatusStateCanaryFailed
		daemonsetWithCanaryFailedNewStatus.Status.Canary = nil
	}

	type fields struct {
		client client.Client
		scheme *runtime.Scheme
	}
	type args struct {
		logger      logr.Logger
		daemonset   *datadoghqv1alpha1.ExtendedDaemonSet
		current     *datadoghqv1alpha1.ExtendedDaemonSetReplicaSet
		upToDate    *datadoghqv1alpha1.ExtendedDaemonSetReplicaSet
		podsCounter podsCounterType
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       *datadoghqv1alpha1.ExtendedDaemonSet
		wantResult reconcile.Result
		wantErr    bool
	}{
		{
			name: "no replicaset == no update",
			fields: fields{
				client: fake.NewFakeClient(daemonset),
				scheme: s,
			},
			args: args{
				logger:    log,
				daemonset: daemonset,
				current:   nil,
				upToDate:  nil,
			},
			want:       daemonset,
			wantResult: reconcile.Result{Requeue: false},
			wantErr:    false,
		},
		{
			name: "current == upToDate; status empty => update",
			fields: fields{
				client: fake.NewFakeClient(daemonset, replicassetCurrent, replicassetUpToDate),
				scheme: s,
			},
			args: args{
				logger:    log,
				daemonset: daemonset,
				current:   replicassetCurrent,
				upToDate:  replicassetCurrent,
				podsCounter: podsCounterType{
					Current: 3,
					Ready:   2,
				},
			},
			want:       daemonsetWithStatus,
			wantResult: reconcile.Result{Requeue: false},
			wantErr:    false,
		},
		{
			name: "current != upToDate; canary active => update",
			fields: fields{
				client: fake.NewFakeClient(daemonset, replicassetCurrent, replicassetUpToDate),
				scheme: s,
			},
			args: args{
				logger:    log,
				daemonset: daemonsetWithCanaryWithStatus,
				current:   replicassetCurrent,
				upToDate:  replicassetUpToDate,
				podsCounter: podsCounterType{
					Current: 3,
					Ready:   2,
				},
			},
			want:       daemonsetWithCanaryWithStatus,
			wantResult: reconcile.Result{Requeue: false},
			wantErr:    false,
		},
		{
			name: "canary failed => update",
			fields: fields{
				client: fake.NewFakeClient(daemonset, replicassetCurrent, replicassetUpToDate),
				scheme: s,
			},
			args: args{
				logger:    log,
				daemonset: daemonsetWithCanaryFailedOldStatus,
				current:   replicassetCurrent,
				upToDate:  replicassetUpToDate,
				podsCounter: podsCounterType{
					Current: 3,
					Ready:   2,
				},
			},
			want:       daemonsetWithCanaryFailedNewStatus,
			wantResult: reconcile.Result{Requeue: false},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReconcileExtendedDaemonSet{
				client:   tt.fields.client,
				scheme:   tt.fields.scheme,
				recorder: eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "TestReconcileExtendedDaemonSet_cleanupReplicaSet"}),
			}
			got, got1, err := r.updateInstanceWithCurrentRS(tt.args.logger, tt.args.daemonset, tt.args.current, tt.args.upToDate, tt.args.podsCounter)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReconcileExtendedDaemonSet.updateInstanceWithCurrentRS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !apiequality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("ReconcileExtendedDaemonSet.updateInstanceWithCurrentRS() got = %#v, \n want %#v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.wantResult) {
				t.Errorf("ReconcileExtendedDaemonSet.updateInstanceWithCurrentRS() gotResult = %v, \n wantResult %v", got1, tt.wantResult)
			}
		})
	}
}

func TestReconcileExtendedDaemonSet_Reconcile(t *testing.T) {
	eventBroadcaster := record.NewBroadcaster()
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "TestReconcileExtendedDaemonSet_Reconcile"})
	logf.SetLogger(logf.ZapLogger(true))

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(datadoghqv1alpha1.SchemeGroupVersion, &datadoghqv1alpha1.ExtendedDaemonSetReplicaSet{})
	s.AddKnownTypes(datadoghqv1alpha1.SchemeGroupVersion, &datadoghqv1alpha1.ExtendedDaemonSetReplicaSetList{})
	s.AddKnownTypes(datadoghqv1alpha1.SchemeGroupVersion, &datadoghqv1alpha1.ExtendedDaemonSet{})
	s.AddKnownTypes(datadoghqv1alpha1.SchemeGroupVersion, &datadoghqv1alpha1.ExtendedDaemonSetList{})

	type fields struct {
		client   client.Client
		scheme   *runtime.Scheme
		recorder record.EventRecorder
	}
	type args struct {
		request  reconcile.Request
		loadFunc func(c client.Client)
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		want     reconcile.Result
		wantErr  bool
		wantFunc func(c client.Client) error
	}{
		{
			name: "ExtendedDaemonset not found",
			fields: fields{
				client:   fake.NewFakeClient(),
				scheme:   s,
				recorder: recorder,
			},
			args: args{
				request: newRequest("but", "faa"),
			},
			want:    reconcile.Result{},
			wantErr: false,
		},
		{
			name: "ExtendedDaemonset found, but not defaulted",
			fields: fields{
				client:   fake.NewFakeClient(),
				scheme:   s,
				recorder: recorder,
			},
			args: args{
				request: newRequest("bar", "foo"),
				loadFunc: func(c client.Client) {
					_ = c.Create(context.TODO(), test.NewExtendedDaemonSet("bar", "foo", &test.NewExtendedDaemonSetOptions{Labels: map[string]string{"foo-key": "bar-value"}}))
				},
			},
			want:    reconcile.Result{Requeue: true},
			wantErr: false,
		},
		{
			name: "ExtendedDaemonset found and defaulted => create the replicaset",
			fields: fields{
				client:   fake.NewFakeClient(),
				scheme:   s,
				recorder: recorder,
			},
			args: args{
				request: newRequest("bar", "foo"),
				loadFunc: func(c client.Client) {
					dd := test.NewExtendedDaemonSet("bar", "foo", &test.NewExtendedDaemonSetOptions{Labels: map[string]string{"foo-key": "bar-value"}})
					dd = datadoghqv1alpha1.DefaultExtendedDaemonSet(dd)
					_ = c.Create(context.TODO(), dd)
				},
			},
			want:    reconcile.Result{Requeue: true},
			wantErr: false,
			wantFunc: func(c client.Client) error {
				replicasetList := &datadoghqv1alpha1.ExtendedDaemonSetReplicaSetList{}
				listOptions := []client.ListOption{
					client.InNamespace("bar"),
				}
				if err := c.List(context.TODO(), replicasetList, listOptions...); err != nil {
					return err
				}
				if len(replicasetList.Items) != 1 {
					return fmt.Errorf("len(replicasetList.Items) is not equal to 1")
				}
				if replicasetList.Items[0].GenerateName != "foo-" {
					return fmt.Errorf("replicasetList.Items[0] bad generated name, should be: 'foo-', current: %s", replicasetList.Items[0].GenerateName)
				}

				return nil
			},
		},
		{
			name: "ExtendedDaemonset found and defaulted, replicaset already exist",
			fields: fields{
				client:   fake.NewFakeClient(),
				scheme:   s,
				recorder: recorder,
			},
			args: args{
				request: newRequest("bar", "foo"),
				loadFunc: func(c client.Client) {
					dd := test.NewExtendedDaemonSet("bar", "foo", &test.NewExtendedDaemonSetOptions{Labels: map[string]string{"foo-key": "bar-value"}})
					dd = datadoghqv1alpha1.DefaultExtendedDaemonSet(dd)

					hash, _ := comparison.GenerateMD5PodTemplateSpec(&dd.Spec.Template)
					rsOptions := &test.NewExtendedDaemonSetReplicaSetOptions{
						GenerateName: "foo-",
						Labels:       map[string]string{"foo-key": "bar-value", datadoghqv1alpha1.ExtendedDaemonSetNameLabelKey: "foo"},
						Annotations:  map[string]string{string(datadoghqv1alpha1.MD5ExtendedDaemonSetAnnotationKey): hash},
					}
					rs := test.NewExtendedDaemonSetReplicaSet("bar", "", rsOptions)

					_ = c.Create(context.TODO(), dd)
					_ = c.Create(context.TODO(), rs)
				},
			},
			want:    reconcile.Result{Requeue: false},
			wantErr: false,
			wantFunc: func(c client.Client) error {
				replicasetList := &datadoghqv1alpha1.ExtendedDaemonSetReplicaSetList{}
				listOptions := []client.ListOption{
					client.InNamespace("bar"),
				}
				if err := c.List(context.TODO(), replicasetList, listOptions...); err != nil {
					return err
				}
				if len(replicasetList.Items) != 1 {
					return fmt.Errorf("len(replicasetList.Items) is not equal to 1")
				}
				if replicasetList.Items[0].GenerateName != "foo-" {
					return fmt.Errorf("replicasetList.Items[0] bad generated name, should be: 'foo-', current: %s", replicasetList.Items[0].GenerateName)
				}

				return nil
			},
		},

		{
			name: "ExtendedDaemonset found and defaulted, replicaset already but not uptodate",
			fields: fields{
				client:   fake.NewFakeClient(),
				scheme:   s,
				recorder: recorder,
			},
			args: args{
				request: newRequest("bar", "foo"),
				loadFunc: func(c client.Client) {
					dd := test.NewExtendedDaemonSet("bar", "foo", &test.NewExtendedDaemonSetOptions{Labels: map[string]string{"foo-key": "bar-value"}})
					dd = datadoghqv1alpha1.DefaultExtendedDaemonSet(dd)

					rsOptions := &test.NewExtendedDaemonSetReplicaSetOptions{
						GenerateName: "foo-",
						Labels:       map[string]string{"foo-key": "old-value"},
						Annotations:  map[string]string{string(datadoghqv1alpha1.MD5ExtendedDaemonSetAnnotationKey): "oldhash"},
					}
					rs := test.NewExtendedDaemonSetReplicaSet("bar", "foo-old", rsOptions)

					_ = c.Create(context.TODO(), dd)
					_ = c.Create(context.TODO(), rs)
				},
			},
			want:    reconcile.Result{Requeue: true},
			wantErr: false,
			wantFunc: func(c client.Client) error {
				replicasetList := &datadoghqv1alpha1.ExtendedDaemonSetReplicaSetList{}
				listOptions := []client.ListOption{
					client.InNamespace("bar"),
				}
				if err := c.List(context.TODO(), replicasetList, listOptions...); err != nil {
					return err
				}
				if len(replicasetList.Items) != 2 {
					return fmt.Errorf("len(replicasetList.Items) is not equal to 1")
				}

				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReconcileExtendedDaemonSet{
				client:   tt.fields.client,
				scheme:   tt.fields.scheme,
				recorder: tt.fields.recorder,
			}
			if tt.args.loadFunc != nil {
				tt.args.loadFunc(r.client)
			}
			got, err := r.Reconcile(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReconcileExtendedDaemonSet.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReconcileExtendedDaemonSet.Reconcile() = %v, want %v", got, tt.want)
			}
			if tt.wantFunc != nil {
				if err := tt.wantFunc(r.client); err != nil {
					t.Errorf("ReconcileExtendedDaemonSet.Reconcile() wantFunc validation error: %v", err)
				}
			}
		})
	}
}

func newRequest(ns, name string) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: ns,
			Name:      name,
		},
	}
}

func Test_getAntiAffinityKeysValue(t *testing.T) {
	node := commontest.NewNode("node", &commontest.NewNodeOptions{
		Labels: map[string]string{
			"app":     "foo",
			"service": "bar",
			"unused":  "baz",
		},
	})

	tests := []struct {
		name          string
		node          corev1.Node
		daemonsetSpec datadoghqv1alpha1.ExtendedDaemonSetSpec
		want          string
	}{
		{
			name: "basic",
			node: *node,
			daemonsetSpec: datadoghqv1alpha1.ExtendedDaemonSetSpec{
				Strategy: datadoghqv1alpha1.ExtendedDaemonSetSpecStrategy{
					Canary: &datadoghqv1alpha1.ExtendedDaemonSetSpecStrategyCanary{
						NodeAntiAffinityKeys: []string{
							"app",
							"missing",
							"service",
						},
					},
				},
			},
			want: "foo$$bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getAntiAffinityKeysValue(&tt.node, &tt.daemonsetSpec)
			if got != tt.want {
				t.Errorf("getAntiAffinityKeysValue(%#v, %#v) = %s, want %s", tt.node, tt.daemonsetSpec, got, tt.want)
			}
		})
	}
}
