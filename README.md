# Darwinia Node Liveness Probe

![](https://img.shields.io/github/workflow/status/darwinia-network/node-liveness-probe/Production)
![](https://img.shields.io/github/v/release/darwinia-network/node-liveness-probe)

The node liveness probe is a sidecar container that exposes an HTTP `/healthz` endpoint, which calls several RPC methods via a WebSocket connection to check and serves as kubelet's livenessProbe hook to monitor health of a Darwinia node.

It also experimentally provides a readiness probe endpoint `/readiness`, which reports if the node is ready to handle RPC requests by determining if the syncing progress is done.

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
    # The liveness probe
    livenessProbe:
      httpGet:
        path: /healthz
        port: healthz
      initialDelaySeconds: 60
      timeoutSeconds: 3
    # The experimental readiness probe
    readinessProbe:
      httpGet:
        path: /readiness
        port: healthz
    # ...

  # The liveness probe sidecar container
  - name: liveness-probe
    image: quay.io/darwinia-network/node-liveness-probe:VERSION
    args:
      - --timeout=3
    # ...
```

Notice that the actual `livenessProbe` field is set on the node container. This way, Kubernetes restarts darwinia node instead of node-liveness-probe when the probe fails. The liveness probe sidecar container only provides the HTTP endpoint for the probe and does not contain livenessProbe section by itself.

It is recommended to increase the option `--timeout` and Pod spec `.containers.*.livenessProbe.timeoutSeconds` a bit (e.g. 3 seconds), if you have a heavy load on your node, since the probe process involves multiple RPC calls.

## Configuration

To get the full list of configurable options, please use `--help`:

```bash
docker run --rm -it quay.io/darwinia-network/node-liveness-probe:VERSION --help
```

## Compatibility

The node liveness probe should be compatible with nodes of other Substrate-based chains, such as Polkadot and Kusama, although that hasn't been well tested. Please consider submitting an issue if you're experiencing any problem with these nodes to help us improve the compatibilities.

## Special Thanks

- <https://github.com/kubernetes-csi/livenessprobe>

## License

MIT
