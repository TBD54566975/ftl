# Deploy ftl-controller to k3d

## Create a k3d cluster with a local Docker registry

```
#
k3d registry create registry.localhost --port 5000
k3d cluster create --api-port 6550 -p "8892:80@loadbalancer" --agents 2 \
    --registry-use k3d-registry.localhost:5000 \
    --registry-config <(cat <<EOF
mirrors:
  "localhost:5000":
    endpoint:
      - http://k3d-registry.localhost:5000
EOF
)
```

## Deploy ftl-controller and all dependencies

```
kubectl kustomize --load-restrictor=LoadRestrictionsNone | kubectl apply -f -
```

### Monitor the rollout

```
kubectl get events -w
```

There will be a lot of retries as the migration and controller waits
for the database to be ready, but it should eventually reconcile to a working state.

## Check the ftl-controller is up

```
ftl status
```

If the controller is not up, check the logs:

```
$ kubectl logs -f deployment/ftl-controller
info: Starting FTL controller
info: Listening on http://0.0.0.0:8892
info: Starting DB listener
```

## To deploy a local ftl-controller or ftl-runner image

Build the image locally:

```
make docker-controller
```

Tag the image for the local registry:

```
docker tag ftl0/ftl-controller:latest localhost:5000/ftl-controller
```

Push the image to the local registry:

```
docker push localhost:5000/ftl-controller
```

## Debugging

To exec into the k3d node:

```
docker exec -it k3d-k3s-default-server-0 sh
```

Exec into the PG cluster:

```
kubectl exec -it ftl-pg-cluster-1-0 -- /bin/bash
```

Create a one-shot shell pod:

```
kubectl run -it --rm --restart=Never --image ubuntu:22.04 tempshell -- bash
```

List all the things:

```
kubectl get deployment,pod,statefulset,svc,configmap,pv,pvc,ingress -o wide
```
