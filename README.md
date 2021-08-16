# Darwinia Node Liveness Probe

![](https://img.shields.io/github/workflow/status/darwinia-network/node-liveness-probe/Production)
![](https://img.shields.io/github/v/release/darwinia-network/node-liveness-probe)

The node-liveness-probe is a sidecar container that exposes an HTTP `/healthz` endpoint, which serves as kubelet's livenessProbe hook to monitor health of a Darwinia (or any Substrate-RPC-compatible) node.

It also experimentally provides a readiness probe endpoint `/readiness`, which reports if the node is ready to handle RPC requests, by determining if the syncing progress is done.

## Releases

Until `v1` released, node-liveness-probe is still under development, any API or CLI options may be changed in future versions.

The releases are under [GitHub's release page](https://github.com/darwinia-network/node-liveness-probe/releases). You can pull the image by using one of the versions, for example:

```bash
docker pull quay.io/darwinia-network/node-liveness-probe:v0.1.0
```

## HTTP Endpoints

- `/healthz` checks node health status by sending a sequence of JSON RPC requests. An HTTP `200` or `5xx` response should be returned, signifying node is healthy or not. See [How it Works](#how-it-works).

- `/healthz_block` also checks node health. Additionally, it checks if the latest and the finalized block of node are (potentially) stale. The threshold can be configured using the CLI option `--block-threshold-seconds`.

- `/readiness` checks if node is still in syncing. This can be used with ReadinessProbe, to make sure RPC requests to a Service will only be served by nodes that have followed up.

## Example

See the full manifests in Kubevali: <https://github.com/darwinia-network/kubevali/blob/master/deploy/manifests/statefulset.yaml>.

## Usage

```yaml
kind: Pod
spec:
  containers:
  ##
  # The node container
  ##
  - name: darwinia
    image: quay.io/darwinia-network/darwinia:NODE_VERSION
    # Defining port which will be used to GET plugin health status
    # 49944 is default, but can be changed.
    ports:
    - name: healthz
      containerPort: 49944
    # The liveness probe
    livenessProbe:
      httpGet:
        path: /healthz # Or /healthz_block
        port: healthz
      initialDelaySeconds: 60
      timeoutSeconds: 3
    # The experimental readiness probe
    readinessProbe:
      httpGet:
        path: /readiness
        port: healthz
    # ...

  ##
  # The liveness probe sidecar container
  ##
  - name: liveness-probe
    image: quay.io/darwinia-network/node-liveness-probe:VERSION
    # ...
```

Notice that the actual `livenessProbe` field is set on the node container. This way, Kubernetes restarts Darwinia node instead of node-liveness-probe when the probe fails. The liveness probe sidecar container only provides the HTTP endpoint for the probe and does not contain a `livenessProbe` section by itself.

It is recommended to increase the Pod spec `.containers.*.livenessProbe.timeoutSeconds` a bit (e.g. 3 seconds), if you have a heavy load on your node, since the probe process involves multiple RPC calls.

## Configuration

To get the full list of configurable options, please use `--help`:

```bash
docker run --rm -it quay.io/darwinia-network/node-liveness-probe:VERSION --help
```

## How it Works

When receives HTTP connections from `/healthz`, the node-liveness-probe tries to connect the node through WebSocket, then calls [several RPC methods](https://github.com/darwinia-network/node-liveness-probe/blob/master/probes/liveness_probe.go#L22) sequentially via the connection to check health of the node. If these requests all succeeded, it generates a `200` response. Otherwise, if there's any error including connection refused, RPC timed out, or JSON RPC error, it responds with HTTP `5xx`.

## Compatibility

The node liveness probe works with Darwinia nodes. It should be compatible with nodes of other Substrate-based chains too. We currently operates several chain node (e.g. Polkadot, Kusama, Kulupu, and Edgeware) as the infrastructure of [subscan.io](https://subscan.io) with node liveness probe. Please consider submitting an issue if you're experiencing any problems with these nodes to help us improve compatibility.

## Special Thanks

- <https://github.com/kubernetes-csi/livenessprobe>

## License

MIT
