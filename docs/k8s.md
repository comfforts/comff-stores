# comff-stores on local Kind

- Setup (Infra)[https://github.com/comfforts/comff-infra]

## Install App stack
### Setup env, config, policy & cred files

### Build container images used by manifests
Update docker image references in `docker-compose-stores.yaml` & run `make build-stores` and/or `make build-stores-debug`

### Load images into cluster registry
```bash
make load-stores-imgs-k

docker exec -it comff-control-plane crictl images
docker exec -it comff-control-plane crictl rmi <IMAGE ID>
```

### Setup stores certs, config, policies & secrets
- `stores/kustomization.yaml`

### Stores service
```bash
make set-stores-k

make rm-stores-k

kubectl -n comff get cm,secret
kubectl -n comff get pvc,pods,svc | grep stores-server
kubectl -n comff describe certificate stores-grpc-server
kubectl -n comff rollout status deploy/stores-server --timeout=300s
kubectl -n comff get svc stores-server -o jsonpath='{.metadata.name}.{.metadata.namespace}.svc.cluster.local{"\n"}'
```
#### get pod and inspect init container status/logs:
```bash
kubectl -n comff get pod -l app=stores-server
kubectl -n comff describe pod <POD_NAME>
kubectl -n comff logs <POD_NAME> --all-containers

kubectl -n comff debug -it pod/<POD_NAME> --target=stores-server --image=busybox:1.36
```

#### check debug pod:
```bash
kubectl -n comff exec -it <POD_NAME> -c stores-server -- /bin/sh
```

#### watch rollout if pod is re-created:
```bash
kubectl -n comff rollout status deploy/stores-server
kubectl -n comff get pods -w

kubectl apply -f k8s/stores/deploy-stores.yaml
kubectl -n comff rollout restart deploy/stores-server
```

#### client certs
```bash
make cp-stores-cl-certs-k
```
#### port-forward stores-server
```bash
kubectl -n comff port-forward svc/stores-server 62151:62151
```

```bash
kubectl -n comff get deploy stores-server -o yaml

# Check node container names
docker ps --format '{{.Names}}' | grep '^comff-'

# Update RAM & CPU caps (example)
docker update --memory 4g --memory-swap 4g --cpus 2.0 comff-control-plane
docker update --memory 8g --memory-swap 8g --cpus 3.0 comff-worker
docker update --memory 8g --memory-swap 8g --cpus 3.0 comff-worker2

# Check RAM & CPU caps
docker inspect -f '{{.Name}} mem={{.HostConfig.Memory}} cpus={{.HostConfig.NanoCpus}}' \
  comff-control-plane comff-worker comff-worker2

```
