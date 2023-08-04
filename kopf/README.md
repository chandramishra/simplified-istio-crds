# istio-simplified-crd
The pupose of this repo is to provide a generic simplified istio crd for non kubernet users.


## Pre-Requiste:
  - kubernet cluster
  - install istio on a kubernet cluster
  - python3.7

## Install istio
- curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.18.0 TARGET_ARCH=x86_64 sh -
- cd istio-1.18.0/bin
- istioctl install --set profile=demo -y

## Install python kopf framework
- pip install kopf

## Deploy Simplified CRDs:

 - git clone https://github.com/chandramishra/istio-simplified-crd.git
 - cd kopf

 Deploy crds:
```
  kubectl apply -f canary-crd.yml  
```
 Create istio policies using simplified crds:

 - create a namespace
```
   kubectl create ns test
```
 - run locally controller(or in a pod) 
```
  python3.7 -m kopf run controller.py --verbose
```
 - Create actual simplified canary 
```
   kubectl apply -f simplified-canary-cr.yml -n test
```
