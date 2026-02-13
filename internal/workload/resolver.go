package workload

import (
	"context"
	"fmt"

	"github.com/hashicorp/golang-lru/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	KindPod         = "Pod"
	KindJob         = "Job"
	KindCronJob     = "CronJob"
	KindRollout     = "Rollout"
	KindDaemonSet   = "DaemonSet"
	KindDeployment  = "Deployment"
	KindStatefulSet = "StatefulSet"
	KindReplicaSet  = "ReplicaSet"
)

type Resolver interface {
	FindWorkloadForPod(ctx context.Context, name, namespace string) (*Workload, error)
}

var kindToGVR = map[string]schema.GroupVersionResource{
	KindPod:         {Group: "", Version: "v1", Resource: "pods"},
	KindReplicaSet:  {Group: "apps", Version: "v1", Resource: "replicasets"},
	KindDeployment:  {Group: "apps", Version: "v1", Resource: "deployments"},
	KindStatefulSet: {Group: "apps", Version: "v1", Resource: "statefulsets"},
	KindDaemonSet:   {Group: "apps", Version: "v1", Resource: "daemonsets"},
	KindJob:         {Group: "batch", Version: "v1", Resource: "jobs"},
	KindCronJob:     {Group: "batch", Version: "v1", Resource: "cronjobs"},
	KindRollout:     {Group: "argoproj.io", Version: "v1alpha1", Resource: "rollouts"},
}

type resolver struct {
	dynamic   dynamic.Interface
	lru       *lru.Cache[cacheKey, *Workload]
	labelKeys []string
}

type cacheKey struct {
	namespace string
	name      string
}

type Config struct {
	LabelKeys []string
	CacheSize int
}

func NewResolver(dynClient dynamic.Interface, cfg Config) (Resolver, error) {
	cache, err := lru.New[cacheKey, *Workload](cfg.CacheSize)
	if err != nil {
		return nil, err
	}

	return &resolver{
		dynamic:   dynClient,
		lru:       cache,
		labelKeys: cfg.LabelKeys,
	}, nil
}

func (m *resolver) FindWorkloadForPod(ctx context.Context, name, namespace string) (*Workload, error) {
	key := cacheKey{
		namespace: namespace,
		name:      name,
	}
	if w, ok := m.lru.Get(key); ok {
		return w, nil
	}

	pod, err := m.getPod(ctx, name, namespace)
	if err != nil {
		return nil, err
	}

	if workloadName, ok := m.findWorkloadNameFromLabels(pod.GetLabels()); ok {
		kind := m.findTopControllerKind(ctx, pod)
		w := &Workload{
			Name:      workloadName,
			Namespace: pod.GetNamespace(),
			Kind:      kind,
		}
		m.lru.Add(key, w)
		return w, nil
	}

	w, err := m.findPodOwner(ctx, pod)
	if err != nil {
		w = &Workload{
			Name:      pod.GetName(),
			Namespace: pod.GetNamespace(),
			Kind:      KindPod,
		}
	}

	m.lru.Add(key, w)

	return w, nil
}

func (m *resolver) getPod(ctx context.Context, name, namespace string) (metav1.Object, error) {
	return m.dynamic.Resource(kindToGVR[KindPod]).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (m *resolver) findWorkloadNameFromLabels(labels map[string]string) (string, bool) {
	if len(m.labelKeys) == 0 {
		return "", false
	}

	for _, key := range m.labelKeys {
		if val, ok := labels[key]; ok {
			return val, true
		}
	}

	return "", false
}

func (m *resolver) findTopControllerKind(ctx context.Context, obj metav1.Object) string {
	ownerRef := metav1.GetControllerOfNoCopy(obj)
	if ownerRef == nil {
		return KindPod
	}

	kind := ownerRef.Kind
	name := ownerRef.Name
	namespace := obj.GetNamespace()

	for {
		gvr, ok := kindToGVR[kind]
		if !ok {
			return kind
		}

		next, err := m.dynamic.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return kind
		}

		nextOwner := metav1.GetControllerOfNoCopy(next)
		if nextOwner == nil {
			return kind
		}

		kind = nextOwner.Kind
		name = nextOwner.Name
	}
}

func (m *resolver) findPodOwner(ctx context.Context, pod metav1.Object) (*Workload, error) {
	ownerRef := metav1.GetControllerOfNoCopy(pod)
	if ownerRef == nil {
		return &Workload{
			Name:      pod.GetName(),
			Namespace: pod.GetNamespace(),
			Kind:      KindPod,
		}, nil
	}

	namespace := pod.GetNamespace()

	switch ownerRef.Kind {
	case KindReplicaSet:
		rs, err := m.dynamic.Resource(kindToGVR[KindReplicaSet]).Namespace(namespace).Get(ctx, ownerRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("getting replicaset %s/%s: %w", namespace, ownerRef.Name, err)
		}

		if rsOwner := metav1.GetControllerOfNoCopy(rs); rsOwner != nil && (rsOwner.Kind == KindDeployment || rsOwner.Kind == KindRollout) {
			return &Workload{
				Name:      rsOwner.Name,
				Namespace: namespace,
				Kind:      rsOwner.Kind,
			}, nil
		}

		return &Workload{
			Name:      ownerRef.Name,
			Namespace: namespace,
			Kind:      KindReplicaSet,
		}, nil

	case KindJob:
		job, err := m.dynamic.Resource(kindToGVR[KindJob]).Namespace(namespace).Get(ctx, ownerRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("getting job %s/%s: %w", namespace, ownerRef.Name, err)
		}
		if jobOwner := metav1.GetControllerOfNoCopy(job); jobOwner != nil && jobOwner.Kind == KindCronJob {
			return &Workload{
				Name:      jobOwner.Name,
				Namespace: namespace,
				Kind:      KindCronJob,
			}, nil
		}
		return &Workload{
			Name:      ownerRef.Name,
			Namespace: namespace,
			Kind:      KindJob,
		}, nil

	default:
		return &Workload{
			Name:      ownerRef.Name,
			Namespace: namespace,
			Kind:      ownerRef.Kind,
		}, nil
	}
}
