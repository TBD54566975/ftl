# Deploy ftl-controller to k3d

Create a k3d cluster:
```
k3d cluster create
```

Run `pg-init.sh` to create a postgres database.


## Debugging

To exec into the k3d node:
```
docker exec -it k3d-k3s-default-server-0 sh
```

Create a one-shot pod:

```
kubectl run -it --rm --restart=Never --image ubuntu:22.04 tempshell -- sh
```

List all the things:

```
kubectl get pod,statefulset,svc,configmap,pv,pvc -o wide
```