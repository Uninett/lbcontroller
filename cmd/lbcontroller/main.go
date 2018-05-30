/*
Copyright 2016 The Kubernetes Authors.
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

// https://engineering.bitnami.com/articles/kubewatch-an-example-of-kubernetes-custom-controller.html
package main

import (
	"flag"
	"fmt"
	"strings"

	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/uninett/lbcontroller"

	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var (
	// max retries for a queued object
	maxRetries  = 3
	lbendpoint  string
	client      *kubernetes.Clientset
	clusterName string
)

func main() {
	flag.Parse()

	//check is we have a load balancer endpoint
	lbendpoint = os.Getenv("lbcontroller_ENDPOINT")
	if len(lbendpoint) == 0 {

		log.Fatalln("No load balancer endpoint defined, lbcontroller_ENDPOINT not set in environment")
	}
	log.Printf("Load balancer endpoint: %s\n", lbendpoint)

	//check is we have a cluster name
	clusterName = os.Getenv("lbcontroller_CLUSTER")
	if len(clusterName) == 0 {

		log.Fatalln("No cluster name defined, lbcontroller_CLUSTER not set in environment")
	}
	log.Printf("Cluster name: %s\n", clusterName)

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Services(meta_v1.NamespaceAll).List(options)
				//return client.CoreV1().Services("default").List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Services(meta_v1.NamespaceAll).Watch(options)
				//return client.CoreV1().Services("default").Watch(options)
			},
		},
		&v1.Service{},
		0, //Skip resync
		cache.Indexers{},
	)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	})

	controller := &Controller{
		clientset: client,
		informer:  informer,
		queue:     queue,
	}

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(stop)

	// Wait forever
	select {}

}

// Controller object
type Controller struct {
	clientset kubernetes.Interface
	queue     workqueue.RateLimitingInterface
	informer  cache.SharedIndexInformer
	//eventHandler handlers.Handler
}

// Run will start the controller.
// StopCh channel is used to send interrupt signal to stop it.
func (c *Controller) Run(stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer utilruntime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer c.queue.ShutDown()

	log.Println("Starting l4lb controller")

	go c.informer.Run(stopCh)

	// wait for the caches to synchronize before starting the worker
	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	log.Println("l4lb controller synced and ready")

	// runWorker will loop until "something bad" happens.  The .Until will
	// then rekick the worker after one second
	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}
func (c *Controller) runWorker() {
	// processNextWorkItem will automatically wait until there's work available
	for c.processNextItem() {
		// continue looping
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (c *Controller) processNextItem() bool {
	// pull the next work item from queue.  It should be a key we use to lookup
	// something in a cache
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	// you always have to indicate to the queue that you've completed a piece of
	// work
	defer c.queue.Done(key)

	// do your work on the key.
	err := c.processItem(key.(string))

	if err == nil {
		// No error, tell the queue to stop tracking history
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < maxRetries {
		log.Printf("Error processing %s (will retry): %v", key, err)
		// requeue the item to work on later
		c.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		log.Printf("Error processing %s (giving up): %v", key, err)
		c.queue.Forget(key)
		utilruntime.HandleError(err)
	}

	return true
}

func (c *Controller) processItem(key string) error {
	log.Printf("Processing change to Service %s", key)

	obj, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return errors.Errorf("Error fetching object with key %s from store: %v", key, err)
	}
	log.Printf("Sync/Add/Update for Service %s\n", key)
	namespace, serviceName, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	//get the load balancer service name
	//the name is cluster-namespace-servicename-protocol
	//e.g. nird-ns9999k-mysql-tcp
	tcpKey := strings.Join([]string{clusterName, namespace, serviceName, "tcp"}, "-")
	udpKey := strings.Join([]string{clusterName, namespace, serviceName, "udp"}, "-")

	//delete the service(s), we have to delete all the
	//services (tcp/udp) because we don't know anymore
	//which protocols were configured.
	//The price ot pay to make this cleaner is to abbandon statelessness.
	if !exists {

		log.Printf("Deleting Service with key %s\n", tcpKey)
		err := lbcontroller.DeleteService(tcpKey, lbendpoint)
		if err != nil {
			return errors.Wrapf(err, "ERROR deleting service with key %s\n", tcpKey)
		}
		log.Printf("Deleted Service with key %s\n", tcpKey)
		log.Printf("Deleting Service with key %s\n", udpKey)
		err = lbcontroller.DeleteService(udpKey, lbendpoint)
		if err != nil {
			return errors.Wrapf(err, "ERROR deleting service with key %s\n", udpKey)
		}
		log.Printf("Deleted Service with key %s\n", udpKey)

		//TODO delete network policy as well

		return nil
	}

	svc := obj.(*v1.Service)

	//if not a loadBalancer service we dont care
	if svc.Spec.Type != v1.ServiceTypeLoadBalancer {
		return nil
	}

	//createor or update the service

	//is the service configured for TCP, UDP?
	var tcpProt bool
	for _, p := range svc.Spec.Ports {
		switch p.Protocol {
		case v1.ProtocolUDP:
			break
		case v1.ProtocolTCP:
			tcpProt = true
			break
		}
	}

	var (
		lbKey  string
		lbType lbcontroller.ServiceType
	)
	if tcpProt {
		lbKey = tcpKey
		lbType = lbcontroller.TCP

	} else {
		lbKey = udpKey
		lbType = lbcontroller.UDP
	}

	//check if the service exists on the loadbalncers
	lbSvc, found, err := lbcontroller.GetService(tcpKey, lbendpoint)
	if err != nil {
		log.Println("ERROR: ", err)
		return errors.Wrapf(err, "Error gettig load balancer service %s from endpoint", tcpKey)
	}

	if !found { //new service
		lbSvc.Type = lbType
		lbSvc.Metadata.Name = lbKey
		lbSvc.Config = lbcontroller.Config{
			Method: "least_conn",
			//Ports: svc.Spec.Ports
		}

		ingress, err := lbcontroller.NewService(*lbSvc, lbendpoint)
		if err != nil {
			//log.Println(err)
			return errors.Wrap(err, "Error creating a new load balancer service")
		}
		fmt.Printf("created load balancer for service %s ingress %v\n", lbKey, ingress)
		svc.Status.LoadBalancer.Ingress = ingress
		persistUpdate(namespace, svc)
	} else { //recofigure
		//do things here
	}

	fmt.Println("service status: ", svc.Status)

	return nil
}

func newlbcontrollerService(ks v1.Service, key, protocol string) lbcontroller.Service {
	svc := lbcontroller.Service{}
	svc.Type = lbcontroller.ServiceType(strings.ToLower(protocol))
	svc.Metadata.Name = key
	cfg := lbcontroller.Config{
		Method:           "least_conn",
		UpstreamMaxConns: 100,
	}
	cfg.Backends = backends
	if len(ks.Spec.LoadBalancerSourceRanges) != 0 {
		cfg.ACL = ks.Spec.LoadBalancerSourceRanges
	}
	if ks.Spec.HealthCheckNodePort != 0 {
		cfg.HealthCheck.Port = ks.Spec.HealthCheckNodePort
	} else if len(ks.Spec.Ports) > 0 {
		cfg.HealthCheck.Port = ks.Spec.Ports[0].NodePort
	}

	cfg.Ports = make(map[string]int32)
	for _, p := range ks.Spec.Ports {
		if string(p.Protocol) == protocol {
			port := fmt.Sprint(p.Port)
			cfg.Ports[port] = int32(p.NodePort)
		}
	}

	return svc
}

func persistUpdate(namespace string, service *v1.Service) {
	_, err := client.CoreV1().Services(service.Namespace).UpdateStatus(service)
	if err != nil {

		fmt.Println("error updating status:", err)
	}
}

//TODO put this in a config file
var backends = []lbcontroller.Backend{
	{
		Host:  "tos-spw01.nird.sigma2.no",
		Addrs: []string{"193.156.11.24", "2001:700:4a00:11::1024"},
	},
	{
		Host:  "tos-spw02.nird.sigma2.no",
		Addrs: []string{"193.156.11.25", "2001:700:4a00:11::1025"},
	},
	{
		Host:  "tos-spw03.nird.sigma2.no",
		Addrs: []string{"193.156.11.26", "2001:700:4a00:11::1026"},
	},
	{
		Host:  "tos-spw04.nird.sigma2.no",
		Addrs: []string{"193.156.11.27", "2001:700:4a00:11::1027"},
	},
	{
		Host:  "tos-spw05.nird.sigma2.no",
		Addrs: []string{"193.156.11.28", "2001:700:4a00:11::1028"},
	},
	{
		Host:  "tos-spw06.nird.sigma2.no",
		Addrs: []string{"193.156.11.29", "2001:700:4a00:11::1029"},
	},
	{
		Host:  "tos-spw07.nird.sigma2.no",
		Addrs: []string{"193.156.11.30", "2001:700:4a00:11::1030"},
	},
}
