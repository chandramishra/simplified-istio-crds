import kopf
import logging
import json
from kubernetes import client, config

logging.basicConfig(level=logging.INFO)
logging.info("Harmless debug Message")

config.load_kube_config()

#------------------------------------------------------------------------------
def get_route(data):
    parent_route_data = data.get('spec',{}).get('conf',{}).get('split')
    child_route = []
    for each in parent_route_data:
      s =  {"destination":{"host": each['destination']['service'], "subset": each['destination']['version']}, "weight": each['weight']}
      child_route.append(s)
    return child_route

#------------------------------------------------------------------------------
def get_dr(data):
    parent_route_data = data.get('spec',{}).get('conf',{}).get('split')
    child_route = []
    for each in parent_route_data:
      s =  {"name":  each['destination']['version'], "labels": {"version":each['destination']['version']}}
      child_route.append(s)

    dr = {
        "apiVersion": "networking.istio.io/v1alpha3",
        "kind": "DestinationRule",
        "metadata": {
          "name": "canary-destination",
        },
        "spec": {
          "host": each.get('destination',{}).get('service'),
          "subsets": child_route
        }
    }
    return dr
	
#------------------------------------------------------------------------------
def get_vs_for_canary(data):
    vs = {
        "apiVersion": "networking.istio.io/v1alpha3",
        "kind": "VirtualService",
        "metadata": {
          "name": "canary-vs"
        },
        "spec": {
          "hosts": data.get("spec",{}).get("sources",[]),
          "http": [
            {
              "route": get_route(data)
            }  
          ]
        }
      } 
    
    return vs
#------------------------------------------------------------------------------
def get_vs_for_canary(data):
    vs = {
        "apiVersion": "networking.istio.io/v1alpha3",
        "kind": "VirtualService",
        "metadata": {
          "name": "canary-vs"
        },
        "spec": {
          "hosts": data.get("spec",{}).get("sources",[]),
          "http": [
            {
              "route": get_route(data)
            }  
          ]
        }
      } 
    
    return vs

#------------------------------------------------------------------------------
@kopf.on.create('istio.com','v1','pocs')
def create_fn(body, **kwargs):
    logging.info("create .........")
    vs = get_vs_for_canary(body)
    dr = get_dr(body)
    logging.info("**************VirtualService*********************")
    logging.info(vs)
    logging.info("**************DestinationRule*********************")
    myclient = client.CustomObjectsApi()
    api_response = myclient.create_namespaced_custom_object(namespace="test",group='networking.istio.io', version='v1alpha3', plural='virtualservices', body=vs)
    api_response = myclient.create_namespaced_custom_object(namespace="test",group='networking.istio.io', version='v1alpha3', plural='destinationrules', body=dr)
    
#------------------------------------------------------------------------------
@kopf.on.delete('istio.com','v1','pocs')
def delete_fn(body, **kwargs):
    logging.info("delete.........")
    vs = get_vs_for_canary(body)
    dr = get_dr(body)
    logging.info("**************VirtualService*********************")
    logging.info(vs)
    logging.info("**************DestinationRule*********************")
    logging.info(dr)
    myclient = client.CustomObjectsApi()
    api_response = myclient.delete_namespaced_custom_object(namespace="test",group='networking.istio.io', version='v1alpha3', plural='virtualservices', name=vs.get("metadata",{}).get("name"))
    api_response = myclient.delete_namespaced_custom_object(namespace="test",group='networking.istio.io', version='v1alpha3', plural='destinationrules', name=dr.get("metadata",{}).get("name"))
#------------------------------------------------------------------------------
