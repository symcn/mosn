# istio install

## install

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

## modify configmap

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
