# multicluster_gw

## Name

*multicluster_gw* - implementation of plugin for multicluster services networking, based on gateway-service.

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

* `kubeconfig` **KUBECONFIG [CONTEXT]** authenticates the connection to a remote k8s cluster using a kubeconfig file. **[CONTEXT]** is optional, if not set, then the current context specified in kubeconfig will be used. It supports TLS, username and password, or token-based authentication. This option is ignored if connecting in-cluster (i.e., the endpoint is not specified).
* `fallthrough` **[ZONES...]** If a query for a record in the zones for which the plugin is authoritative results in NXDOMAIN, normally that is what the response will be. However, if you specify this option, the query will instead be passed on down the plugin chain, which can include another plugin to handle the query. If **[ZONES...]** is omitted, then fallthrough happens for all zones for which the plugin is authoritative. If specific zones are listed (for example `in-addr.arpa` and `ip6.arpa`), then only queries for those zones will be subject to fallthrough.
* `gateway_ip` **GATEWAY_IP** The wanted ip for our gateway service


## Config example

Handle all queries in the `clusterset.local` zone, and refer them to the service in the ip `6.6.6.6`. Connect to Kubernetes in-cluster.

```
.:53 {
    `multicluster_gw coredns.local .svc.clusterset.local. {
    gateway_ip 6.6.6.6
}`
}
```

## Installation

See CoreDNS documentation about [Compile Time Enabling or Disabling Plugins](https://coredns.io/2017/07/25/compile-time-enabling-or-disabling-plugins/).

### Recompile coredns

Add the plugin to  `plugins.cfg` file. The [ordering of plugins matters](https://coredns.io/2017/06/08/how-queries-are-processed-in-coredns/),
add it just below `kubernetes` plugin that has very similar functionality:

```
...
kubernetes:kubernetes
multicluster_gw:github.com/itay-nakash/multicluster_gw
...
```

Follow the [coredns README](https://github.com/coredns/coredns#readme) file to build it.
