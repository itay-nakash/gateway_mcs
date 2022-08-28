# multicluster_gw

Implementation of plugin for multicluster services networking, based on gateway-service.

## Description

This plugin direct requsets with the configure zone to a gateway service,
which will refer them to the wanted service in an outside cluster.

## Syntax

```
multicluster [ZONES...] {
    kubeconfig KUBECONFIG [CONTEXT]
    fallthrough [ZONES...]
    gateway_ip GATEWAY_IP
}
```

* `fallthrough` **[ZONES...]** If a query for a record in the zones for which the plugin is authoritative results in NXDOMAIN, normally that is what the response will be. However, if you specify this option, the query will instead be passed on down the plugin chain, which can include another plugin to handle the query. If **[ZONES...]** is omitted, then fallthrough happens for all zones for which the plugin is authoritative. If specific zones are listed (for example `in-addr.arpa` and `ip6.arpa`), then only queries for those zones will be subject to fallthrough.
* `gateway_ip` **GATEWAY_IP** The wanted ip for our gateway service


## Config example

Example for a core-config file for k8s cluster.
Handle all queries in the `clusterset.local` zone, and refer them to the service in the ip `6.6.6.6`. Connect to Kubernetes in-cluster.

```
.:53 {
        errors
        health {
           lameduck 5s
        }
        ready
        multicluster_gw svc.clusterset.local {
                    gateway_ip 6.6.6.6
        }
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           fallthrough in-addr.arpa ip6.arpa
           ttl 30
        }
        prometheus :9153
        forward . /etc/resolv.conf {
           max_concurrent 1000
        }
        cache 30
        loop
        reload
        loadbalance
    }
```

## How to use the plugin

Installation, and plugin setup steps:

1. Clone core-dns repo
2. Add the plugin to  `plugins.cfg` file. The [ordering of plugins matters](https://coredns.io/2017/06/08/how-queries-are-processed-in-coredns/),
   add it just below `kubernetes` plugin that has very similar functionality.
   You should add to plugin.cfg the line: 'multicluster_gw:github.com/itay-nakash/multicluster_gw' 
3. Recompile corends (using their makefile)
4. make sure that the compiled coredns incluedes 'multicluster_gw' plugin by running `./corends --plugins`
5. Build docker-image for your new dns server
6. Replace the image in the core-dns deployment in your cluster to your custom image (that inculdes the plugin)
7. Change the corefile to configure it to include the plugin (for example, as the example deatiled above)
8. Terminate the current coredns pod (to let it come back with the new core-config settings).
for convenience you may want to edit the deployment to include only one replica of the coredns.
9. Enjoy your brand-new coredns server :))

