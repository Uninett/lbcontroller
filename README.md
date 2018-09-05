
# lbcontroller (WIP)

lbcontroller uses Kubernetes [metacontroller](https://metacontroller.app/) to automatically configure the Uninett loadbalancers.
If a Service of `type: LoadBalancer` is added to the cluster the metacontroller should call the sync hook of cmd/lbmetacontroller, which in turin will communicate to the Uninett API to configure the loadbalancers.

A biref panoramic of the content:

- lbcontrollercli is probably broken and was used for testing.
- lbcontrollertest is a mock of the Uninett loadbalancer's API.
- lbmetacontroller is the web hook that the metacontroller should call to sync the state of the services with the loadbalancers.

## Development/test with minikube on Mac

Install go 1.11.

Install docker and minikube.

Run minikube.

Set yuor shell to use the docker daemon running in minikube, so we don't need a Docker registry, with `eval $(minikube docker-env)`.

Export `GOOS=linux` in your shell, we are compiling for linux.

Clone this repo if you haven't yet.

Install the metacontroller namespace, service account, and role/binding, etc.. as explained [here](https://metacontroller.app/guide/install/), note that in minikube you are already cluster admin so you can skip the first step.

Build both lbmetacontroller and lbcontrollertest by moving in their dirs and run `go bulid`.

Build both docker images with `docker build -t lb-api-test .` for cmd/lbcontrollertest and `docker build -t lb-hook .` for cmd/lbmetacontroller.

Run `kubectl apply -f lb-api-test.yaml` from cmd/lbcontrollertest.

Run `kubectl apply -f lb-hook.yaml` from cmd/lbmetacontroller.

Run the metacontroller that actually listens for services in the cluster `kubectl apply -f metacontroller.yaml` from cmd/lbmetacontroller.

Now all should be set up, to test that it also works run `kubectl apply -f test-service.yaml`. This create a service of `type: LoadBalancer` and triggers the metacontroller in calling the sync hook (lbmetacontroller) which in turn should call the test API (lbcontrollertest).

The result shuold be visible as new networkpolicy called `nginx-lb` and an annotation on the service `nginx`, that is the test service created in the previous step.

If something is wrong check the logs and open an issue.

## Development/test with minikube on Linux

*I have not tried this on Linux so send a pull request if you find something not working.*

Pretty much the same as before just do not export `GOOS=linux`, or do it is already in your environment. Also if you run minikube as `minkube staert --vm-driver=none` you probably dont have to reuse the docker daemon in minkube.

## Notes

The yaml files in cmd/{lbmetaontroller, lbcontrollertest} have `imagePullPolicy: Never`, this is because if reuse the Docker daemon in minikube without a registry teh images are already there and cannot be pulled. If you want to use a different setup you might want to changhe the pull policy.

# TODOs

- Add clenup functionality as mentioned [here](https://github.com/GoogleCloudPlatform/metacontroller/issues/60)
- Better names for *lbmetacontroller* and *lbcontrollertest*?
- Automate some test?
- ...
