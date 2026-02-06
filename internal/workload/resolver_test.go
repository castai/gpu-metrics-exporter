package workload

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakedynamic "k8s.io/client-go/dynamic/fake"
)

var testGVRs = map[schema.GroupVersionResource]string{
	kindToGVR[KindPod]:         "PodList",
	kindToGVR[KindReplicaSet]:  "ReplicaSetList",
	kindToGVR[KindDeployment]:  "DeploymentList",
	kindToGVR[KindStatefulSet]: "StatefulSetList",
	kindToGVR[KindDaemonSet]:   "DaemonSetList",
	kindToGVR[KindJob]:         "JobList",
	kindToGVR[KindCronJob]:     "CronJobList",
	kindToGVR[KindRollout]:     "RolloutList",
}

func TestFindWorkloadForPod(t *testing.T) {
	ctx := context.Background()
	isController := true

	t.Run("pod with no owner returns pod itself", func(t *testing.T) {
		pod := newUnstructuredObj("v1", "Pod", "my-pod", "default", nil, nil)
		r := newTestResolver(t, nil, pod)

		w, err := r.FindWorkloadForPod(ctx, "my-pod", "default")
		require.NoError(t, err)
		require.Equal(t, &Workload{Name: "my-pod", Namespace: "default", Kind: KindPod}, w)
	})

	t.Run("pod owned by replicaset owned by deployment", func(t *testing.T) {
		pod := newUnstructuredObj("v1", "Pod", "my-deploy-abc-xyz", "default", nil,
			[]metav1.OwnerReference{{Kind: KindReplicaSet, Name: "my-deploy-abc", Controller: &isController}},
		)
		rs := newUnstructuredObj("apps/v1", "ReplicaSet", "my-deploy-abc", "default", nil,
			[]metav1.OwnerReference{{Kind: KindDeployment, Name: "my-deploy", Controller: &isController}},
		)
		r := newTestResolver(t, nil, pod, rs)

		w, err := r.FindWorkloadForPod(ctx, "my-deploy-abc-xyz", "default")
		require.NoError(t, err)
		require.Equal(t, &Workload{Name: "my-deploy", Namespace: "default", Kind: KindDeployment}, w)
	})

	t.Run("pod owned by replicaset with no deployment", func(t *testing.T) {
		pod := newUnstructuredObj("v1", "Pod", "my-rs-xyz", "default", nil,
			[]metav1.OwnerReference{{Kind: KindReplicaSet, Name: "my-rs", Controller: &isController}},
		)
		rs := newUnstructuredObj("apps/v1", "ReplicaSet", "my-rs", "default", nil, nil)
		r := newTestResolver(t, nil, pod, rs)

		w, err := r.FindWorkloadForPod(ctx, "my-rs-xyz", "default")
		require.NoError(t, err)
		require.Equal(t, &Workload{Name: "my-rs", Namespace: "default", Kind: KindReplicaSet}, w)
	})

	t.Run("pod owned by replicaset owned by rollout", func(t *testing.T) {
		pod := newUnstructuredObj("v1", "Pod", "my-rollout-abc-xyz", "default", nil,
			[]metav1.OwnerReference{{Kind: KindReplicaSet, Name: "my-rollout-abc", Controller: &isController}},
		)
		rs := newUnstructuredObj("apps/v1", "ReplicaSet", "my-rollout-abc", "default", nil,
			[]metav1.OwnerReference{{Kind: KindRollout, Name: "my-rollout", Controller: &isController}},
		)
		r := newTestResolver(t, nil, pod, rs)

		w, err := r.FindWorkloadForPod(ctx, "my-rollout-abc-xyz", "default")
		require.NoError(t, err)
		require.Equal(t, &Workload{Name: "my-rollout", Namespace: "default", Kind: KindRollout}, w)
	})

	t.Run("pod owned by statefulset", func(t *testing.T) {
		pod := newUnstructuredObj("v1", "Pod", "my-sts-0", "default", nil,
			[]metav1.OwnerReference{{Kind: KindStatefulSet, Name: "my-sts", Controller: &isController}},
		)
		r := newTestResolver(t, nil, pod)

		w, err := r.FindWorkloadForPod(ctx, "my-sts-0", "default")
		require.NoError(t, err)
		require.Equal(t, &Workload{Name: "my-sts", Namespace: "default", Kind: KindStatefulSet}, w)
	})

	t.Run("pod owned by daemonset", func(t *testing.T) {
		pod := newUnstructuredObj("v1", "Pod", "my-ds-abc", "default", nil,
			[]metav1.OwnerReference{{Kind: KindDaemonSet, Name: "my-ds", Controller: &isController}},
		)
		r := newTestResolver(t, nil, pod)

		w, err := r.FindWorkloadForPod(ctx, "my-ds-abc", "default")
		require.NoError(t, err)
		require.Equal(t, &Workload{Name: "my-ds", Namespace: "default", Kind: KindDaemonSet}, w)
	})

	t.Run("pod owned by job owned by cronjob", func(t *testing.T) {
		pod := newUnstructuredObj("v1", "Pod", "my-cron-job-abc-xyz", "default", nil,
			[]metav1.OwnerReference{{Kind: KindJob, Name: "my-cron-job-abc", Controller: &isController}},
		)
		job := newUnstructuredObj("batch/v1", "Job", "my-cron-job-abc", "default", nil,
			[]metav1.OwnerReference{{Kind: KindCronJob, Name: "my-cron", Controller: &isController}},
		)
		r := newTestResolver(t, nil, pod, job)

		w, err := r.FindWorkloadForPod(ctx, "my-cron-job-abc-xyz", "default")
		require.NoError(t, err)
		require.Equal(t, &Workload{Name: "my-cron", Namespace: "default", Kind: KindCronJob}, w)
	})

	t.Run("pod owned by job with no cronjob", func(t *testing.T) {
		pod := newUnstructuredObj("v1", "Pod", "my-job-xyz", "default", nil,
			[]metav1.OwnerReference{{Kind: KindJob, Name: "my-job", Controller: &isController}},
		)
		job := newUnstructuredObj("batch/v1", "Job", "my-job", "default", nil, nil)
		r := newTestResolver(t, nil, pod, job)

		w, err := r.FindWorkloadForPod(ctx, "my-job-xyz", "default")
		require.NoError(t, err)
		require.Equal(t, &Workload{Name: "my-job", Namespace: "default", Kind: KindJob}, w)
	})

	t.Run("label-based resolution uses label name and owner kind", func(t *testing.T) {
		pod := newUnstructuredObj("v1", "Pod", "my-deploy-abc-xyz", "default",
			map[string]string{"app.kubernetes.io/name": "my-app"},
			[]metav1.OwnerReference{{Kind: KindReplicaSet, Name: "my-deploy-abc", Controller: &isController}},
		)
		rs := newUnstructuredObj("apps/v1", "ReplicaSet", "my-deploy-abc", "default", nil,
			[]metav1.OwnerReference{{Kind: KindDeployment, Name: "my-deploy", Controller: &isController}},
		)
		r := newTestResolver(t, []string{"app.kubernetes.io/name"}, pod, rs)

		w, err := r.FindWorkloadForPod(ctx, "my-deploy-abc-xyz", "default")
		require.NoError(t, err)
		require.Equal(t, &Workload{Name: "my-app", Namespace: "default", Kind: KindDeployment}, w)
	})

	t.Run("label-based resolution with no owner returns pod kind", func(t *testing.T) {
		pod := newUnstructuredObj("v1", "Pod", "bare-pod", "default",
			map[string]string{"app": "my-app"},
			nil,
		)
		r := newTestResolver(t, []string{"app"}, pod)

		w, err := r.FindWorkloadForPod(ctx, "bare-pod", "default")
		require.NoError(t, err)
		require.Equal(t, &Workload{Name: "my-app", Namespace: "default", Kind: KindPod}, w)
	})

	t.Run("second call returns cached result", func(t *testing.T) {
		pod := newUnstructuredObj("v1", "Pod", "cached-pod", "default", nil, nil)
		r := newTestResolver(t, nil, pod)

		w1, err := r.FindWorkloadForPod(ctx, "cached-pod", "default")
		require.NoError(t, err)

		w2, err := r.FindWorkloadForPod(ctx, "cached-pod", "default")
		require.NoError(t, err)
		require.Same(t, w1, w2)
	})

	t.Run("pod not found returns error", func(t *testing.T) {
		r := newTestResolver(t, nil)

		_, err := r.FindWorkloadForPod(ctx, "missing-pod", "default")
		require.Error(t, err)
	})

	t.Run("replicaset fetch error falls back to pod", func(t *testing.T) {
		pod := newUnstructuredObj("v1", "Pod", "orphan-pod", "default", nil,
			[]metav1.OwnerReference{{Kind: KindReplicaSet, Name: "deleted-rs", Controller: &isController}},
		)
		// ReplicaSet not registered in fake client — will cause a 404
		r := newTestResolver(t, nil, pod)

		w, err := r.FindWorkloadForPod(ctx, "orphan-pod", "default")
		require.NoError(t, err)
		require.Equal(t, &Workload{Name: "orphan-pod", Namespace: "default", Kind: KindPod}, w)
	})
}

func newTestResolver(t *testing.T, labelKeys []string, objects ...*unstructured.Unstructured) Resolver {
	t.Helper()

	scheme := runtime.NewScheme()
	var runtimeObjects []runtime.Object
	for _, obj := range objects {
		runtimeObjects = append(runtimeObjects, obj)
	}

	dynClient := fakedynamic.NewSimpleDynamicClientWithCustomListKinds(scheme, testGVRs, runtimeObjects...)

	r, err := NewResolver(dynClient, Config{
		LabelKeys: labelKeys,
		CacheSize: 128,
	})
	require.NoError(t, err)
	return r
}

func newUnstructuredObj(apiVersion, kind, name, namespace string, labels map[string]string, ownerRefs []metav1.OwnerReference) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]any{
				"name":      name,
				"namespace": namespace,
			},
		},
	}

	if labels != nil {
		obj.SetLabels(labels)
	}

	if len(ownerRefs) > 0 {
		obj.SetOwnerReferences(ownerRefs)
	}

	return obj
}
