# istio-simplified-crd
The pupose of this repo is to provide a generic simplified istio crd for non kubernet users.


## Pre-Requiste:
  - kubernet cluster
  - install meta-controller on a kubernet cluster. For more details please refer this link - https://github.com/metacontroller/metacontroller
  - install istio on a kubernet cluster

## Install istio
- curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.18.0 TARGET_ARCH=x86_64 sh -
- cd istio-1.18.0/bin
- istioctl install --set profile=demo -y

## Install meta-controller
- kubectl apply -k https://github.com/metacontroller/metacontroller/manifests/production

## Deploy Simplified CRDs:

 - git clone https://github.com/chandramishra/istio-simplified-crd.git
 - cd istio-simplified-crd/metaController

 Deploy crds:
```
  kubectl apply -f kube_crds/canary-crd.yml  
  kubectl apply -f kube_crds/mirror-crd.yml  
```
 Deploy controller for each crd:
 ```
  kubectl apply -f controllers//canary-controller.yml  
  kubectl apply -f controllers/mirror-controller.yml  
```
 Create istio policies using simplified crds:

 - create a namespace
```
   kubectl create ns test
```
 - Create configMap
```
   kubectl -n test create configmap istio-webhook --from-file=webhook/sync.py
```
 - Deploy web hook
```
   kubectl apply -f webhook/web-hook.yml -n test
```
 - Create actual simplified canary & mirror policy
```
   kubectl apply -f simplified-canary-cr.yml -n test
   kubectl apply -f simplified-mirror-cr.yml -n test
```
