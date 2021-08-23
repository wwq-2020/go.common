package rpc

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/wwq-2020/go.common/app"
	corev1 "k8s.io/api/core/v1"
	informerscorev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// ResolverFactory ResolverFactory
type ResolverFactory func(string) Resolver

// Resolver Resolver
type Resolver interface {
	Start()
	OnAdd(func(string))
	OnDel(func(string))
}

type k8sResolver struct {
	addr      string
	namespace string
	onAdds    []func(string)
	onDels    []func(string)
	m         sync.Mutex
}

// NewK8SResolver NewK8SResolver
func NewK8SResolver(addr string) Resolver {
	r := &k8sResolver{
		namespace: os.Getenv("KUBE_NAMESPACE"),
		addr:      addr,
	}
	return r
}

func (r *k8sResolver) OnAdd(onAdd func(string)) {
	r.m.Lock()
	defer r.m.Unlock()
	r.onAdds = append(r.onAdds, onAdd)
}

func (r *k8sResolver) OnDel(onDel func(string)) {
	r.m.Lock()
	defer r.m.Unlock()
	r.onDels = append(r.onDels, onDel)
}

func (r *k8sResolver) Start() {
	namespace := r.namespace
	if namespace == "" {
		r.onAdd(r.addr)
		return
	}
	masterURL, kubeconfigPath := os.Getenv("KUBE_MASTER_URL"), os.Getenv("KUBE_CONFIG_PATH")
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		panic(err)
	}
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}
	informer := informerscorev1.NewEndpointsInformer(clientSet, namespace, time.Minute, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			endpoints := obj.(*corev1.Endpoints)
			if endpoints.Name == r.addr {
				r.addEndpoints(endpoints)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldEndpoints := oldObj.(*corev1.Endpoints)
			newEndpoints := newObj.(*corev1.Endpoints)
			if oldEndpoints.Name == r.addr {
				r.delEndpoints(oldEndpoints)
			}
			if newEndpoints.Name == r.addr {
				r.addEndpoints(newEndpoints)
			}
		},
		DeleteFunc: func(obj interface{}) {
			endpoints := obj.(*corev1.Endpoints)
			if endpoints.Name == r.addr {
				r.delEndpoints(endpoints)
			}
		},
	})
	go informer.Run(app.Done())
	cache.WaitForCacheSync(app.Done(), informer.HasSynced)
}

func (r *k8sResolver) onAdd(endpoint string) {
	for _, onAdd := range r.onAdds {
		onAdd(endpoint)
	}
}

func (r *k8sResolver) onDel(endpoint string) {
	for _, onDel := range r.onDels {
		onDel(endpoint)
	}
}

func (r *k8sResolver) addEndpoints(endpoints *corev1.Endpoints) {
	r.m.Lock()
	defer r.m.Unlock()
	for _, subset := range endpoints.Subsets {
		if len(subset.Ports) == 0 {
			continue
		}
		for _, address := range subset.Addresses {
			endpoint := fmt.Sprintf("%s:%d", address.IP, subset.Ports[0].Port)
			r.onAdd(endpoint)
		}
	}
}

func (r *k8sResolver) delEndpoints(endpoints *corev1.Endpoints) {
	r.m.Lock()
	defer r.m.Unlock()
	for _, subset := range endpoints.Subsets {
		if len(subset.Ports) == 0 {
			continue
		}
		for _, address := range subset.Addresses {
			endpoint := fmt.Sprintf("%s:%d", address.IP, subset.Ports[0].Port)
			r.onDel(endpoint)
		}
	}
}
