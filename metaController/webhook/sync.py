from http.server import BaseHTTPRequestHandler, HTTPServer
import json
import logging
logging.basicConfig(level=logging.INFO)
logging.info("Start debug Message")

class Controller(BaseHTTPRequestHandler):
  #------------------------------------------------------------------------
  def get_mirror(self, data):
    mirror_data = data.get('spec',{}).get('conf',{}).get('mirror')
    s =  {"host":mirror_data.get("destination",{}).get("service",""),"subset": mirror_data.get('destination',{}).get('version',"")}
    return s

  #------------------------------------------------------------------------
  def get_mirror_perc(self, data):
    mirror_data = data.get('spec',{}).get('conf',{}).get('mirrorPercentage',0)
    return mirror_data
  #------------------------------------------------------------------------
  def get_route(self, data):
    parent_route_data = data.get('spec',{}).get('conf',{}).get('split')
    child_route = []
    for each in parent_route_data:
      s =  {"destination":{"host": each['destination']['service'], "subset": each['destination']['version']}, "weight": each['weight']}
      child_route.append(s)
    return child_route

  #------------------------------------------------------------------------
  def get_dr(self, data):
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
	
  def get_vs_for_canary(self, data):
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
              "route": self.get_route(data)
            }  
          ]
        }
      } 
    
    return vs
  def get_vs_for_mirror(self, data):
    vs = {
        "apiVersion": "networking.istio.io/v1alpha3",
        "kind": "VirtualService",
        "metadata": {
          "name": "mirror-vs"
        },
        "spec": {
          "hosts": data.get("spec",{}).get("sources",[]),
          "http": [
            {
              "route": self.get_route(data),
              "mirror": self.get_mirror(data),
              "mirror_percent": self.get_mirror_perc(data)
            }  
          ]
        }
      } 
    
    return vs
 
  def sync(self, parent, children):
    # Compute status based on observed state.
    desired_status = {
      "virtualservices": 1 #len(children["VirtualService.networking.istio.io/v1alpha3"])
    }
    logging.info("********* parent policy*********")
    logging.info(parent)
    if parent.get("kind","") == "Canary":
      desired = [self.get_vs_for_canary(parent)]
    elif parent.get("kind","") == "Mirror":
      desired = [self.get_vs_for_mirror(parent)]
    else:
      desired = []
    desired.append(self.get_dr(parent))
    logging.info("*********desired child policy*********")
    logging.info(desired)
    return {"status": desired_status, "children": desired}

  def do_POST(self):
    # Serve the sync() function as a JSON webhook.
    observed = json.loads(self.rfile.read(int(self.headers.get("content-length"))))
    desired = self.sync(observed["parent"], observed["children"])

    self.send_response(200)
    self.send_header("Content-type", "application/json")
    self.end_headers()
    self.wfile.write(json.dumps(desired).encode())

HTTPServer(("", 80), Controller).serve_forever()
