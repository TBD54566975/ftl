import { Edge, Node } from 'reactflow'
import { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { MetadataCalls } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'

const groupWidth = 200

const calculateModuleDepths = (modules: Module[]): Record<string, number> => {
  const depths: Record<string, number> = {}
  const adjList: Record<string, string[]> = {}
  const visitedInCurrentPath: Set<string> = new Set() // For cycle detection

  modules.forEach((module) => {
    adjList[module.name ?? ''] = []
    ;(module.verbs ?? []).forEach((verb) => {
      const calls = verb?.verb?.metadata
        .filter((meta) => meta.value.case === 'calls')
        .map((meta) => meta.value.value as MetadataCalls)

      calls?.forEach((call) => {
        call.calls.forEach((c) => {
          adjList[module.name ?? ''].push(c.module)
        })
      })
    })
  })

  // Depth-first search to calculate depths
  const dfs = (node: string, depth: number) => {
    if (visitedInCurrentPath.has(node)) {
      // Detected a cycle
      return
    }
    visitedInCurrentPath.add(node)

    depths[node] = Math.max(depths[node] ?? 0, depth)
    adjList[node].forEach((neighbor) => {
      dfs(neighbor, depth + 1)
    })

    visitedInCurrentPath.delete(node) // Remove the node from the current path after exploring all neighbors
  }

  // Initialize DFS from each node
  Object.keys(adjList).forEach((node) => {
    visitedInCurrentPath.clear() // Clear the path before starting a new DFS
    dfs(node, 0)
  })

  return depths
}

const nodeGraph = (modules: Module[]) => {
  const out: Record<string, string[]> = {}

  for (const module of modules) {
    out[module.name ?? ''] = []
    const verbs = module.verbs
    verbs.forEach((verb) => {
      const calls = verb?.verb?.metadata
        .filter((meta) => meta.value.case === 'calls')
        .map((meta) => meta.value.value as MetadataCalls)

      calls?.map((call) =>
        call.calls.forEach((call) => {
          if (out[module.name ?? ''] === undefined) {
            out[module.name ?? ''] = []
          }
          out[module.name ?? ''].push(call.module)
        }),
      )
    })
  }

  return out
}

export const layoutNodes = (modules: Module[]) => {
  const nodes: Node[] = []
  const edges: Edge[] = []
  const xCounters: Record<number, number> = {}

  const ng = nodeGraph(modules)
  console.log(ng)

  const moduleDepths = calculateModuleDepths(modules)

  modules.forEach((module) => {
    const depth = moduleDepths[module.name ?? ''] ?? 0
    if (xCounters[depth] === undefined) {
      xCounters[depth] = 0
    }

    const x = xCounters[depth]
    const yOffset = depth * 300

    const verbs = module.verbs

    nodes.push({
      id: module.name ?? '',
      position: { x: x, y: yOffset },
      data: { title: module.name, item: module },
      type: 'groupNode',
      draggable: true,
      style: {
        width: groupWidth,
        height: (verbs?.length ?? 1) * 50 + 50,
        zIndex: 1,
      },
    })
    let y = 40
    verbs.forEach((verb) => {
      const calls = verb?.verb?.metadata
        .filter((meta) => meta.value.case === 'calls')
        .map((meta) => meta.value.value as MetadataCalls)

      nodes.push({
        id: `${module.name}.${verb.verb?.name}`,
        position: { x: 20, y: y },
        connectable: false,
        data: { title: verb.verb?.name, item: verb },
        type: 'verbNode',
        parentNode: module.name,
        style: {
          width: groupWidth - 40,
          height: 40,
        },
        draggable: false,
        zIndex: 2,
      })

      const uniqueEdgeIds = new Set<string>()
      calls?.map((call) =>
        call.calls.forEach((call) => {
          const edgeId = `${module.name}.${verb.verb?.name}-${call.module}.${call.name}`
          if (!uniqueEdgeIds.has(edgeId)) {
            uniqueEdgeIds.add(edgeId)
            edges.push({
              id: edgeId,
              source: `${module.name}.${verb.verb?.name}`,
              target: `${call.module}.${call.name}`,
              style: { stroke: 'rgb(251 113 133)' },
              animated: true,
            })
            call.name
            call.module
          }
        }),
      )

      y += 50
    })
    xCounters[depth] += 300
  })
  return { nodes, edges }
}
