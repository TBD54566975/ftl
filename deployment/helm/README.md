# Apply Helm to K3D

Create a K3D cluster:
```
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

As kubegres does not have Helm Charts, for now the CRDs need to be added manually.

```bash
kubectl apply -f https://raw.githubusercontent.com/reactive-tech/kubegres/v1.16/kubegres.yaml
```

Once that is done the Helm Charts can be installed
```bash
helm install ftl .
```