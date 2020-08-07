<p align="center">
<img src="https://raw.githubusercontent.com/mosn/community/master/icons/png/mosn-labeled-horizontal.png" width="350" title="MOSN Logo" alt="MOSN logo">
</p>

[![Build Status](https://travis-ci.com/mosn/mosn.svg?branch=master)](https://travis-ci.com/mosn/mosn)
[![codecov](https://codecov.io/gh/mosn/mosn/branch/master/graph/badge.svg)](https://codecov.io/gh/mosn/mosn)
[![Go Report Card](https://goreportcard.com/badge/github.com/mosn/mosn)](https://goreportcard.com/report/github.com/mosn/mosn)
![license](https://img.shields.io/badge/license-Apache--2.0-green.svg)

[中文](README_ZH.md)

MOSN is a network proxy written in Golang. It can be used as a cloud-native network data plane, providing services with the following proxy functions: multi-protocol, modular, intelligent, and secure. MOSN is the short name of Modular Open Smart Network-proxy. MOSN can be integrated with any Service Mesh which support xDS API. It also can be used as an independent Layer 4 or Layer 7 load balancer, API Gateway, cloud-native Ingress, etc.

## Features

As an open source network proxy, MOSN has the following core functions:

- Support full dynamic resource configuration through xDS API integrated with Service Mesh.
- Support proxy with TCP, HTTP, and RPC protocols.
- Support rich routing features.
- Support reliable upstream management and load balancing capabilities.
- Support network and protocol layer observability.
- Support mTLS and protocols on TLS.
- Support rich extension mechanism to provide highly customizable expansion capabilities.
- Support process smooth upgrade.

## Download&Install

Use `go get -u mosn.io/mosn`, or you can git clone the repository to `$GOPATH/src/mosn.io/mosn`.

**Notice**

- If you need to use code before 0.8.1, you may needs to run the script `transfer_path.sh` to fix the import path.
- If you are in Linux, you should modify the `SED_CMD` in `transfer_path.sh`, see the comment in the script file.

## Documentation

- [Website](https://mosn.io)
- [Changelog](CHANGELOG.md)

## Contributing

See our [contributor guide](CONTRIBUTING.md).

## Community

See our community materials on <https://github.com/mosn/community>.

Scan the QR code below with [DingTalk(钉钉)](https://www.dingtalk.com) to join the MOSN user group.

<p align="center">
<img src="https://gw.alipayobjects.com/mdn/rms_91f3e6/afts/img/A*NyEzRp3Xq28AAAAAAAAAAABkARQnAQ" width="150">
</p>

## Landscapes

<p align="center">
<img src="https://landscape.cncf.io/images/left-logo.svg" width="150"/>&nbsp;&nbsp;<img src="https://landscape.cncf.io/images/right-logo.svg" width="200"/>
<br/><br/>
MOSN enriches the <a href="https://landscape.cncf.io/landscape=observability-and-analysis&license=apache-license-2-0">CNCF CLOUD NATIVE Landscape.</a>
</p>

## istio configmap modify

### download istioctl

```shell
$ curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.6.7 sh -
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   107  100   107    0     0     58      0  0:00:01  0:00:01 --:--:--    58
100  3896  100  3896    0     0   1683      0  0:00:02  0:00:02 --:--:-- 3804k
Downloading istio-1.6.7 from https://github.com/istio/istio/releases/download/1.6.7/istio-1.6.7-osx.tar.gz ...
Istio 1.6.7 Download Complete!

Istio has been successfully downloaded into the istio-1.6.7 folder on your system.

Next Steps:
See https://istio.io/docs/setup/kubernetes/install/ to add Istio to your Kubernetes cluster.

To configure the istioctl client tool for your workstation,
add the /tmp/istio-1.6.7/bin directory to your environment path variable with:
	 export PATH="$PATH:/tmp/istio-1.6.7/bin"

Begin the Istio pre-installation verification check by running:
	 istioctl verify-install

Need more information? Visit https://istio.io/docs/setup/kubernetes/install/
```

### install istio

```shell
$ ./istio-1.6.7/bin/istioctl manifest apply --set profile=minimal --set values.global.jwtPolicy=first-party-jwt --set addonComponents.grafana.enabled=false --set addonComponents.istiocoredns.enabled=false --set addonComponents.kiali.enabled=true --set addonComponents.prometheus.enabled=false --set addonComponents.tracing.enabled=false --set components.pilot.hub=docker.io/istio --set meshConfig.defaultConfig.binaryPath=/usr/local/bin/mosn --set meshConfig.defaultConfig.customConfigFile=/etc/istio/mosn/mosn_config_dubbo_xds.json --set meshConfig.defaultConfig.statusPort=15021 --set values.sidecarInjectorWebhook.rewriteAppHTTPProbe=false --set values.global.proxy.logLevel=info
Detected that your cluster does not support third party JWT authentication. Falling back to less secure first party JWT. See https://istio.io/docs/ops/best-practices/security/#configure-third-party-service-account-tokens for details.
! global.mtls.auto is deprecated; use meshConfig.enableAutoMtls instead
✔ Istio core installed
✔ Istiod installed
✔ Addons installed
✔ Installation complete

```

if already install istioctl

> istioctl manifest apply --set profile=minimal --set values.global.jwtPolicy=first-party-jwt --set addonComponents.grafana.enabled=false --set addonComponents.istiocoredns.enabled=false --set addonComponents.kiali.enabled=true --set addonComponents.prometheus.enabled=false --set addonComponents.tracing.enabled=false --set components.pilot.hub=docker.io/istio --set meshConfig.defaultConfig.binaryPath=/usr/local/bin/mosn --set meshConfig.defaultConfig.customConfigFile=/etc/istio/mosn/mosn_config_dubbo_xds.json --set meshConfig.defaultConfig.statusPort=15021 --set values.sidecarInjectorWebhook.rewriteAppHTTPProbe=false --set values.global.proxy.logLevel=info

### modify configmap

```shell
kubectl edit configmap -n istio-system istio-sidecar-injector
```

- modify `data.config.policy`: disabled

- delete `data.config.template.initContainers`

- add `data.config.template.containers.env`, such as: MOSN_ZK_ADDRESS

**finally result**

```yaml
apiVersion: v1
data:
  config: |-
    # modify
    policy: disabled
    alwaysInjectSelector:
      []
    neverInjectSelector:
      []
    injectedAnnotations:

    template: |
      rewriteAppHTTPProbe: {{ valueOrDefault .Values.sidecarInjectorWebhook.rewriteAppHTTPProbe false }}
      # initContainers delete
      containers:
      - name: istio-proxy
      {{- if contains "/" (annotation .ObjectMeta `sidecar.istio.io/proxyImage` .Values.global.proxy.image) }}
        image: "{{ annotation .ObjectMeta `sidecar.istio.io/proxyImage` .Values.global.proxy.image }}"
      {{- else }}
        image: "{{ .Values.global.hub }}/{{ .Values.global.proxy.image }}:{{ .Values.global.tag }}"
      {{- end }}
        ports:
        - containerPort: 15090
          protocol: TCP
          name: http-envoy-prom
        args:
        - proxy
        - sidecar
        - --domain
        - $(POD_NAMESPACE).svc.{{ .Values.global.proxy.clusterDomain }}
        - --serviceCluster
        {{ if ne "" (index .ObjectMeta.Labels "app") -}}
        - "{{ index .ObjectMeta.Labels `app` }}.$(POD_NAMESPACE)"
        {{ else -}}
        - "{{ valueOrDefault .DeploymentMeta.Name `istio-proxy` }}.{{ valueOrDefault .DeploymentMeta.Namespace `default` }}"
        {{ end -}}
        - --proxyLogLevel={{ annotation .ObjectMeta `sidecar.istio.io/logLevel` .Values.global.proxy.logLevel}}
        - --proxyComponentLogLevel={{ annotation .ObjectMeta `sidecar.istio.io/componentLogLevel` .Values.global.proxy.componentLogLevel}}
      {{- if .Values.global.sts.servicePort }}
        - --stsPort={{ .Values.global.sts.servicePort }}
      {{- end }}
      {{- if .Values.global.trustDomain }}
        - --trust-domain={{ .Values.global.trustDomain }}
      {{- end }}
      {{- if .Values.global.logAsJson }}
        - --log_as_json
      {{- end }}
      {{- if gt .ProxyConfig.Concurrency 0 }}
        - --concurrency
        - "{{ .ProxyConfig.Concurrency }}"
      {{- end -}}
      {{- if .Values.global.proxy.lifecycle }}
        lifecycle:
          {{ toYaml .Values.global.proxy.lifecycle | indent 4 }}
        {{- end }}
        env:
        # add env MOSN_ZK_ADDRESS
        - name: MOSN_ZK_ADDRESS
          value: 127.0.0.1:2181
        - name: JWT_POLICY
          value: {{ .Values.global.jwtPolicy }}
......
```
