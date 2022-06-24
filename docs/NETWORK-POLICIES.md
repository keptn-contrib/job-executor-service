## Job executor service network policies

In release
[0.2.1](https://github.com/keptn-contrib/job-executor-service/releases/tag/0.2.1)
opt-in network policies have been added to job-executor-service helm chart to
limit connection both in ingress and egress direction. This document will
explain a little more in detail the reasoning behind such network policy
definitions.
For more information about kubernetes network policies please refer to the relevant 
[kubernetes official documentation](https://kubernetes.io/docs/concepts/services-networking/network-policies/).

### Ingress network policy

Job-executor-service does not need to accept connections from the outside, so
the ingress network policy is defined as not allowing *any* incoming connection
for the job-executor-service pod:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: jes-ingress-network-policy
spec:
  podSelector:
    matchLabels:
      {{- include "job-executor-service.selectorLabels" . | nindent 6 }}
  policyTypes:
    - Ingress
```
Such network policy definition will prevent any incoming connection to the
job-executor-service from anywhere (equivalent to the 
[deny all example](https://kubernetes.io/docs/concepts/services-networking/network-policies/#default-deny-all-ingress-traffic)
from the [kubernetes documentation](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
but applied only to job-executor-service pod through the `podSelector`)

#### Enabling ingress network policy
To enable the ingress network policy definition, set the helm
value `networkPolicy.ingress.enabled` to `true` during job-executor-service
installation/upgrade.

### Egress network policy

Job-executor-service needs to connect to a few services for receiving keptn
cloud events and spawning jobs to run tasks: this means that the egress network
policy must allow communication to such services not to break job-executor
functionality.

The services and the reason why job-executor-service needs access to each are:

- Kubernetes [cluster DNS](https://kubernetes.io/docs/concepts/overview/components/#dns):
  this is used to translate 
  [kubernetes services network names](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/)
- Kubernetes [apiserver](https://kubernetes.io/docs/concepts/overview/components/#kube-apiserver):
  job-executor-service needs to reach Kubernetes API to create/get/watch jobs
  and related pods
- Keptn control-plane: when job-executor-service is installed as a remote
  execution plane (that is in a separate cluster/namespace from where Keptn control plane is
  running), it needs to connect to the Keptn API to fetch the pending triggered cloudevents
- OAuth provider: if OAuth authentication is activated for connecting to Keptn
  control plane job-executor-service needs access to the OAuth provider to
  retrieve confguration and get OAuth tokens to be able to authenticate
  correctly on Keptn API

Each of the service job-executor-service needs to access will have a dedicated
rule in the network policy definition so let's have a look at the whole network
policy definition before getting into details of each rule:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: jes-egress-network-policy
spec:
  podSelector:
    matchLabels:
      {{- include "job-executor-service.selectorLabels" . | nindent 6 }}
  policyTypes:
    - Egress
  egress:
    # Add egress to the keptn api gateway POD
    - to:
        - namespaceSelector: {}
          podSelector:
            matchLabels:
              app.kubernetes.io/name: api-gateway-nginx
              app.kubernetes.io/instance: keptn

    # Add egress to apiserver endpoints
    {{ if .Values.networkPolicy.egress.k8sMasterCIDR }}

    # We have a value set for the master CIDR
    - to:
        - ipBlock:
            cidr: {{ .Values.networkPolicy.egress.k8sMasterCIDR }}
      {{ if .Values.networkPolicy.egress.k8sMasterPort }}
      ports:
        - port: {{ .Values.networkPolicy.egress.k8sMasterPort }}
      {{ end }}

    {{else}}

    # Attempt at autodetect if the k8s master CIDR is not set
    {{ $endpoint := (lookup "v1" "Endpoints" "default" "kubernetes") }}
    {{ $https_port := 0 }}
    {{ range $_, $subset := $endpoint.subsets }}
    {{ range $_, $port := $subset.ports }}
    {{ if eq "https" $port.name }}
    {{ range $_, $address := $subset.addresses }}
    - to:
        - ipBlock:
            cidr: {{ printf "%s/32" $address.ip }}
      ports:
        - port: {{ $port.port }}
    {{ end }}
    {{ end }}
    {{ end }}
    {{ end }}
    {{ end }}

    {{ if eq .Values.remoteControlPlane.api.authMode "oauth" }}
    # Allow traffic to OAuth endpoint
    {{ $oauthHost := getHostByName (urlParse .Values.remoteControlPlane.api.oauth.clientDiscovery).host }}
    - to:
        - ipBlock:
            cidr: {{ printf "%s/32" ($oauthHost) }}
    {{ end }}

    # Allow traffic to External Keptn control plane
    {{ if .Values.remoteControlPlane.api.hostname }}
    - to:
        - ipBlock:
            cidr: {{ printf "%s/32" (getHostByName .Values.remoteControlPlane.api.hostname) }}
    {{ end }}

    # Allow DNS traffic to kube-dns pod
    - to:
        - namespaceSelector: {}
          podSelector:
            matchLabels:
              k8s-app: kube-dns
      ports:
        - protocol: UDP
          port: 53
```
There's a bit of Helm function calls going on to be able to autodetect service
endpoints as much as possible using DNS and kubernetes service lookups but be
aware that *external IPs are retrieved only when installing/upgrading
job-executor-service*: if those IPs are subjected to change the
job-executor-service may stop working without warning.

### Kubernetes cluster DNS rule

The following rule
```yaml
    # Allow DNS traffic to kube-dns pod
    - to:
        - namespaceSelector: {}
          podSelector:
            matchLabels:
              k8s-app: kube-dns
      ports:
        - protocol: UDP
          port: 53
```
will allow egress traffic from the job-executor-service to pods in *any*
namespace matching the label `k8s-app: kube-dns` on UDP port 53 (DNS).
This implies that access to kubernetes dns cluster (it usually is a coreDNS pod
in kube-system namespace that has the matching label) is allowed by the rule
definition above.

### Kubernetes Apiserver

As already stated above job-executor-service needs to access the kubernetes API
to spawn jobs so connectivity to the kubernetes master(s) must be allowed. The
apiserver is usually running as a regular process (not a pod) on the master
nodes of a kubernetes cluster (some local dev setup may differ, most notably
minikube) so we have to define the rule in terms of IP addresses rather than
pod/namespaces selectors:

```yaml
    # Add egress to apiserver endpoints
    {{ if .Values.networkPolicy.egress.k8sMasterCIDR }}

    # We have a value set for the master CIDR
    - to:
        - ipBlock:
            cidr: {{ .Values.networkPolicy.egress.k8sMasterCIDR }}
      {{ if .Values.networkPolicy.egress.k8sMasterPort }}
      ports:
        - port: {{ .Values.networkPolicy.egress.k8sMasterPort }}
      {{ end }}

    {{else}}

    # Attempt at autodetect if the k8s master CIDR is not set
    {{ $endpoint := (lookup "v1" "Endpoints" "default" "kubernetes") }}
    {{ $https_port := 0 }}
    {{ range $_, $subset := $endpoint.subsets }}
    {{ range $_, $port := $subset.ports }}
    {{ if eq "https" $port.name }}
    {{ range $_, $address := $subset.addresses }}
    - to:
        - ipBlock:
            cidr: {{ printf "%s/32" $address.ip }}
      ports:
        - port: {{ $port.port }}
    {{ end }}
    {{ end }}
    {{ end }}
    {{ end }}
    {{ end }}

```
The first part of the rule (top level if branch) defines a network policy rule
where the subnet of the master nodes and the port on which the apiserver is
available are provided as helm values. In such a case the configuration is
pretty straightforward: allow egress to the specified subnet (optionally
restricting to a specific TCP port where the apiserver is listening).


If master node subnet is not provided (else branch) then we attempt to
autodetect the endpoints for service `kubernetes` in the `default` namespace
and define a rule for each endpoint address and https port pair. If master
nodes IP address are stable (also though kubernetes updates) this is a good
enough solution for kubernetes apiserver automatic detection. If the IPs of the
master nodes are not stable then passing the master CIDR is preferable as
explained above.

### Keptn control plane

Access to the Keptn control plane is allowed by two separate rules. 

The first rule deals with the case where the Keptn control plane is running as
a POD in the same cluster as the job-executor-service:

```yaml
    - to:
        - namespaceSelector: {}
          podSelector:
            matchLabels:
              app.kubernetes.io/name: api-gateway-nginx
              app.kubernetes.io/instance: keptn
```
The definition above allows egress to the api-gateway-nginx POD selected by the
labels in *any* namespace.

The second rule deals with the case where there is an address configured using 
`.Values.remoteControlPlane.api.hostname` helm value:
```yaml
    # Allow traffic to External Keptn control plane
    {{ if .Values.remoteControlPlane.api.hostname }}
    - to:
        - ipBlock:
            cidr: {{ printf "%s/32" (getHostByName .Values.remoteControlPlane.api.hostname) }}
    {{ end }}
```
In such a case we define an ipBlock using the host we get from a DNS lookup of
the `remoteControlPlane.api.hostname` string (if it's already an IP address it
will be returned as-is).

The union of the 2 rules should guarantee access to the keptn control plane
(with the usual caveat that the control plane not running on the same
kubernetes cluster must have a stable IP address, if that's not the case it's
advisable to define additional network policies).

### OAuth endpoint

If the OAuth client discovery URL is defined in `.Values.remoteControlPlane.api.oauth.clientDiscovery`
the following rule of the network policy is defined:
```yaml
    {{ if eq .Values.remoteControlPlane.api.authMode "oauth" }}
    # Allow traffic to OAuth endpoint
    {{ $oauthHost := getHostByName (urlParse .Values.remoteControlPlane.api.oauth.clientDiscovery).host }}
    - to:
        - ipBlock:
            cidr: {{ printf "%s/32" ($oauthHost) }}
    {{ end }}
```

Similarly to what we do for keptn remote control plane api endpoint, we perform
a DNS lookup on the host part of the OAuth client config URL and add the
resulting address as an allowed ipBlock. The same caveat as for the other DNS
lookups apply here: if the IP address of the OAuth endpoint is not stable or it
doesn't match with the host there client configuration is retrieved from some
additional network policies mey be needed.

#### Enabling egress network policy
To enable the egress network policy definition, set the helm
value `networkPolicy.egress.enabled` to `true` during job-executor-service
installation/upgrade.
