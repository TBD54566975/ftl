# FTL Architecture

The actors in the diagrams are as follows:

| Actor      | Description                                                                                                |
| ---------- | ---------------------------------------------------------------------------------------------------------- |
| Backplane  | The coordination layer of FTL. This creates and manages Runner instances, routing, resource creation, etc. |
| Platform   | The platform FTL is running on, eg. Kubernetes, VMs, etc.                                                  |
| Runner     | The component of FTL that coordinates with the Backplane to spawn and route to user code.                  |
| Deployment | User code serving VerbService for a module written in a particular language.                               |

## System initialisation

```mermaid
sequenceDiagram
  participant B as Backplane
  participant P as Platform
  participant R as Runner

  B ->> P: CreateRunner(language)
  P ->> R: CreateInstance(language)
  R ->> B: RegisterRunner(language)
```

## Creating a deployment

```mermaid
sequenceDiagram
  participant C as Client
  participant B as Backplane
  box Module
    participant R as Runner
    participant M as Deployment
  end

  C ->> B: GetArtefactDiffs()
  B -->> C: missing_digests
  loop All artefacts
    C ->> B: UploadArtefact()
    B -->> C: digest
  end
  C ->> B: CreateDeployment()
  B -->> C: id
  B ->> R: Deploy(id)
  R -->> B: ack
  R ->> B: GetDeployment(id)
  B -->> R: schema
  R ->> B: GetDeploymentArtefacts(id)
  B -->> R: artefacts
  R ->> M: Start()
```

## Routing

This diagram shows a routing example of a client calling verb V0 which calls
verb V1.

```mermaid
sequenceDiagram
  %% autonumber

  participant C as Client
  participant B as Backplane
  box LightYellow Module0
    participant R0 as Runner0
    participant M0 as Deployment0
  end
  box LightGreen Module1
    participant R1 as Runner1
    participant M1 as Deployment1
  end

  C ->> B: Call(V0)
  B ->> R0: Call(V0)
  R0 ->> M0: Call(V0)
  M0 ->> B: Call(V1)
  B ->> R1: Call(V1)
  R1 ->> M1: Call(V1)
  M1 -->> R1: R1
  R1 -->> B: R1
  B -->> M0: R1
  M0 -->> R0: R0
  R0 -->> B: R0
  B -->> C: R0
```
