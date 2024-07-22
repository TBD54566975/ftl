# Deploy ftl-controller to k3d

This is a kubernetes environment that runs under k3d for testing purposes, and can also be used as a basis for deployments.

The [Justfile](Justfile) contains the commands to spin up a cluster and set up everything to have a running FTL system.

## Setup

`just setup` will create a k3s cluster and registry under Docker. You should only run this once, or after running `just teardown`.

`just install` will then set up FTL in the Kubernetes cluster:
- Build and install the FTL controller and runner
- Run postgres DB
- Run localstack for AWS Secrets Manager
- Run FTL migrations

You can use these commands from the root of the project with `just k8s install`, or within the `deployment` directory with `just install`.

`just teardown` will remove the cluster and all resources, including the registry and database.

Show all processes:

`just ps`

There will be a lot of retries as the migration and controller waits
for the database to be ready, but it should eventually reconcile to a working state.

e.g. For the first 2-5 minutes of the cluster starting up this is normal:

```
pod/ftl-controller-7f8b5f5785-wnj74   0/1     CrashLoopBackOff
```

## Redeploying FTL

When creating changes to the kubernetes resources, or want to re-deploy resources that are deleted, you can `just apply`.

However if you changed FTL and want to deploy, use `just install` which will build the docker images and reapply kubernetes resources.

## FTL

The web console should be available at `http://localhost:8892`.

You can connect to the FTL controller using the `ftl` CLI that you have on your machine.

By default, the endpoint should be pointing to `http://localhost:8892`, so the `--endpoint` doesn't need to be specified.

```
ftl status

{
  "controllers":  [
    {
      "key":  "ctr-10.42.0.7-8892-55zkljc9cl7zl0p1",
      "endpoint":  "http://10.42.0.7:8892",
      "version":  "0.296.3-6-gd09c1cab"
    },
    {
      "key":  "ctr-10.42.2.5-8892-5g41ueriqti7905j",
      "endpoint":  "http://10.42.2.5:8892",
      "version":  "0.296.3-6-gd09c1cab"
    }
  ]
}
```

Deploy some modules:

```
ftl deploy ./examples/go
```

## Debugging

After viewing `just ps`, e.g.:

```
just ps

kubectl get deployment,pod,statefulset,svc,configmap,pv,pvc,ingress -o wide
NAME                             READY   UP-TO-DATE   AVAILABLE   AGE     CONTAINERS   IMAGES                           SELECTOR
deployment.apps/ftl-runner       10/10   10           10          2m19s   app          ftl:5000/ftl-runner:latest       app=ftl-runner
deployment.apps/localstack       1/1     1            1           2m19s   localstack   localstack/localstack            app=localstack
deployment.apps/ftl-controller   2/2     2            2           2m19s   app          ftl:5000/ftl-controller:latest   app=ftl-controller

NAME                                  READY   STATUS      RESTARTS      AGE     IP           NODE               NOMINATED NODE   READINESS GATES
pod/ftl-runner-79b546fb4d-bfhnr       1/1     Running     0             2m19s   10.42.2.7    k3d-ftl-server-0   <none>           <none>
pod/ftl-runner-79b546fb4d-jb242       1/1     Running     0             2m19s   10.42.2.8    k3d-ftl-server-0   <none>           <none>
pod/ftl-runner-79b546fb4d-96fk9       1/1     Running     0             2m18s   10.42.2.9    k3d-ftl-server-0   <none>           <none>
pod/ftl-runner-79b546fb4d-h85ws       1/1     Running     0             2m19s   10.42.1.5    k3d-ftl-agent-1    <none>           <none>
pod/ftl-runner-79b546fb4d-hb4zq       1/1     Running     0             2m19s   10.42.1.7    k3d-ftl-agent-1    <none>           <none>
pod/ftl-runner-79b546fb4d-l852m       1/1     Running     0             2m18s   10.42.1.8    k3d-ftl-agent-1    <none>           <none>
pod/ftl-runner-79b546fb4d-9qb7h       1/1     Running     0             2m19s   10.42.0.9    k3d-ftl-agent-0    <none>           <none>
pod/ftl-runner-79b546fb4d-rtzw9       1/1     Running     0             2m19s   10.42.0.8    k3d-ftl-agent-0    <none>           <none>
pod/ftl-runner-79b546fb4d-xjsm9       1/1     Running     0             2m18s   10.42.0.10   k3d-ftl-agent-0    <none>           <none>
pod/ftl-runner-79b546fb4d-gr5h4       1/1     Running     0             2m19s   10.42.0.6    k3d-ftl-agent-0    <none>           <none>
pod/localstack-57b975d597-lj6vl       1/1     Running     0             2m19s   10.42.1.6    k3d-ftl-agent-1    <none>           <none>
pod/ftl-pg-cluster-1-0                1/1     Running     0             111s    10.42.2.11   k3d-ftl-server-0   <none>           <none>
pod/ftl-db-migrate-n8h2f              0/1     Completed   3             2m19s   10.42.2.6    k3d-ftl-server-0   <none>           <none>
pod/ftl-controller-7f8b5f5785-xvxlm   1/1     Running     4 (84s ago)   2m19s   10.42.0.7    k3d-ftl-agent-0    <none>           <none>
pod/ftl-controller-7f8b5f5785-wnj74   1/1     Running     4 (92s ago)   2m19s   10.42.2.5    k3d-ftl-server-0   <none>           <none>

NAME                                READY   AGE    CONTAINERS         IMAGES
statefulset.apps/ftl-pg-cluster-1   1/1     111s   ftl-pg-cluster-1   postgres:14.1

NAME                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE     SELECTOR
service/kubernetes       ClusterIP   10.43.0.1       <none>        443/TCP             7m17s   <none>
service/ftl-controller   ClusterIP   10.43.49.8      <none>        8891/TCP,8892/TCP   2m19s   app=ftl-controller
service/localstack       ClusterIP   10.43.231.229   <none>        4566/TCP            2m19s   app=localstack
service/ftl-pg-cluster   ClusterIP   None            <none>        5432/TCP            79s     app=ftl-pg-cluster,replicationRole=primary

NAME                                         DATA   AGE
configmap/kube-root-ca.crt                   1      7m2s
configmap/ftl-pg-cluster-conf                1      2m19s
configmap/ftl-db-migrate-config-h4fmggb56d   2      2m19s
configmap/base-kubegres-config               7      112s

NAME                                                        CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                    STORAGECLASS   REASON   AGE    VOLUMEMODE
persistentvolume/pvc-f26addf6-c4d3-487b-a289-acdc97b73a32   200Mi      RWO            Delete           Bound    default/postgres-db-ftl-pg-cluster-1-0   local-path              101s   Filesystem

NAME                                                   STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE    VOLUMEMODE
persistentvolumeclaim/postgres-db-ftl-pg-cluster-1-0   Bound    pvc-f26addf6-c4d3-487b-a289-acdc97b73a32   200Mi      RWO            local-path     111s   Filesystem

NAME                                       CLASS     HOSTS   ADDRESS                                     PORTS   AGE
ingress.networking.k8s.io/ftl-controller   traefik   *       192.168.247.3,192.168.247.4,192.168.247.5   80      2m17s
```

View logs with `just logs <pod>`:

```
just logs pod/ftl-controller-7f8b5f5785-xvxlm

kubectl logs -f pod/ftl-controller-7f8b5f5785-xvxlm
debug: Starting FTL controller
info: Web console available at: http://10.42.0.7:8892
debug: Listening on http://10.42.0.7:8892
debug: Advertising as http://10.42.0.7:8892
info: HTTP ingress server listening on: http://10.42.0.7:8891
debug: new leader for /system/asm: http://10.42.0.7:8892
debug:lease:/system/asm: Acquired lease
debug: Seeded 0 deployments
debug:lease:/system/scheduledtask/reconcileRunners: Acquired lease
debug:lease:/system/scheduledtask/reconcileDeployments: Acquired lease
debug:lease:/system/scheduledtask/reapStaleRunners: Acquired lease
debug:lease:/system/scheduledtask/releaseExpiredReservations: Acquired lease
```

You can access the DB with `just psql`:

```
just psql

just enter statefulset.apps/ftl-pg-cluster-1 env PGPASSWORD=secret psql -U postgres ftl
kubectl exec -it statefulset.apps/ftl-pg-cluster-1 -- env PGPASSWORD=secret psql -U postgres ftl
psql (14.1 (Debian 14.1-1.pgdg110+1))
Type "help" for help.

ftl=# \d
                     List of relations
 Schema |            Name             |   Type   |  Owner
--------+-----------------------------+----------+----------
 public | artefacts                   | table    | postgres
 public | artefacts_id_seq            | sequence | postgres
 public | async_calls                 | table    | postgres

...
```

Or shell into a pod with `just enter <pod>`:

```
just enter pod/ftl-controller-7f8b5f5785-xvxlm

kubectl exec -it pod/ftl-controller-7f8b5f5785-xvxlm -- bash
root@ftl-controller-7f8b5f5785-xvxlm:~# ps aux
USER         PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root           1  108  0.6 1686216 51600 ?       Ssl  00:51   4:18 [rosetta] /root/ftl-controller /root/ftl-controller
root          30  3.8  0.0 418324  7424 pts/1    Ss   00:55   0:00 [rosetta] /usr/bin/bash bash
root          38  100  0.0 421360  6272 pts/1    R+   00:55   0:00 ps aux
root@ftl-controller-7f8b5f5785-xvxlm:~#
```

Create a one-shot shell pod:

```
kubectl run -it --rm --restart=Never --image ubuntu:22.04 tempshell -- bash
```
