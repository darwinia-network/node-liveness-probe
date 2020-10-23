# Darwinia Node Liveness Probe

![](https://img.shields.io/github/workflow/status/darwinia-network/node-liveness-probe/Production)
![](https://img.shields.io/github/v/release/darwinia-network/node-liveness-probe)

The node liveness probe is a sidecar container that exposes an HTTP `/healthz` endpoint, which serves as kubelet's livenessProbe hook to monitor health of a Darwinia node.

## Releases

The releases are under [GitHub's release page](https://github.com/darwinia-network/node-liveness-probe/releases). You can pull the image by using one of the versions, for example:

```bash
docker pull quay.io/darwinia-network/node-liveness-probe:v0.1.0
```

## Usage

```yaml
kind: Pod
spec:
  containers:

  # The node container
  - name: darwinia
    image: darwinianetwork/darwinia:NODE_VERSION
    # Defining port which will be used to GET plugin health status
    # 49944 is default, but can be changed.
    ports:
    - name: healthz
      containerPort: 49944
    # The probe
    readinessProbe: &probe
      httpGet:
        path: /healthz
        port: healthz
      timeoutSeconds: 3
    livenessProbe:
      <<: *probe
      initialDelaySeconds: 60
    # ...

  # The liveness probe sidecar container
  - name: liveness-probe
    image: quay.io/darwinia-network/node-liveness-probe:VERSION
    args:
      - --timeout=3
    # ...
```

Notice that the actual `livenessProbe` field is set on the node container. This way, Kubernetes restarts darwinia node instead of node-liveness-probe when the probe fails. The liveness probe sidecar container only provides the HTTP endpoint for the probe and does not contain livenessProbe section by itself.

To get the full list of configurable options, please use `--help`:

```bash
docker run --rm -it quay.io/darwinia-network/node-liveness-probe:VERSION --help
```

## Special Thanks

- <https://github.com/kubernetes-csi/livenessprobe>

## License

MIT
