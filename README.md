
# lbcontroller (WIP)

lbcontroller uses Kubernetes [metacontroller](https://metacontroller.app/) to automatically configure the Uninett loadbalancers.
If a Service of `type: LoadBalancer` is added to the cluster the metacontroller should call the sync hook of cmd/lbmetacontroller, which in turin will communicate to the Uninett API to configure the loadbalancers.

A biref panoramic of the content:

- lbcontrollercli is probably broken and was used for testing.
- lbcontrollertest is a mock of the Uninett loadbalancer's API.
- lbmetacontroller is the web hook that the metacontroller should call to sync the state of the services with the loadbalancers.

## Development/test with minikube on Mac

Install go 1.11.

**This project is developed with go 1.11 and [modules](https://github.com/golang/go/wiki/Modules), to compile clone this repo outside the GOPATH or set GO111MODULE=on.**

Install docker and minikube.

Run minikube.

Set yuor shell to use the docker daemon running in minikube, so we don't need a Docker registry, with `eval $(minikube docker-env)`.

Export `GOOS=linux` in your shell, we are compiling for linux.

Clone this repo if you haven't yet.

Install the metacontroller namespace, service account, and role/binding, etc.. as explained [here](https://metacontroller.app/guide/install/), note that in minikube you are already cluster admin so you can skip the first step.

Build both lbmetacontroller by moving in the repo root dir and run `go bulid`.

Build the docker image with `docker build -t lb-hook .`, this will push the image to the docker daemon runnin in the minikube cluster.

Run the web hook of the metacontroller on the minikube cluster `kubectl apply -f lb-hook.yaml`, this is the service that will get invokec by the metacontroller when a service is added to the cluster. It will in turn communicate/sync the new service to the load balancer API.

Run the metacontroller that actually listens for services in the cluster `kubectl apply -f metacontroller.yaml`, this is the controller that will invoke the web hook that we just ran.

Now all should be set up, to test that it also works run `kubectl apply -f test-service.yaml`. This create a service of `type: LoadBalancer` and triggers the metacontroller in calling the sync hook (lb-hook) which in turn should call the load balancer API.

The result shuold be visible as new networkpolicy called `nginx-lb` and an annotation on the service `nginx`, that is the test service created in the previous step.

If something is wrong check the logs and open an issue.
If you followed the instructions you should be able to see the logs of the lb-hook with `kubectl logs lb-hook` and the logs of the metacontroller with `kubectl -n metacontroller logs pod/metacontroller-0 -f`.

## Development/test with minikube on Linux

*I have not tried this on Linux so send a pull request if you find something not working.*

Pretty much the same as for Mac OS just do not export `GOOS=linux` (it should be already in your environment).

Also if you run minikube as `minkube staert --vm-driver=none` you probably dont have to reuse the docker daemon in minkube.

# Configuration

There are four envroment variables used to configure the behavior of the controller.
`LBC_CLUSTER_NAME` is the name of the cluster. This varible is not mandatory  and will default to *nird* the other two must be defined.
`LBC_ENDPOINT` is the API endpoint of the load balancer. This varible is mandatory.
`LBC_TOKEN` is the token to use to authenticate the API. This varible is mandatory.
`LBC_PEERS` is load babalancers IPs, comma separated in CIDR form. This varible is mandatory.

## Notes

The yaml file lb-hook.yaml have `imagePullPolicy: Never`, this is because if reuse the Docker daemon in minikube without a registry the image is already available and does not need to be pulled. If you want to use a different setup, maybe with a registry, you might want to changhe the pull policy.



# TODOs

- Add clenup functionality as mentioned [here](https://github.com/GoogleCloudPlatform/metacontroller/issues/60)
- Automate some test?
- ...
