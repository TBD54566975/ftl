import { Edge, Node } from 'reactflow'
import { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { MetadataCalls } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'

const groupWidth = 200

const calculateModuleDepths = (modules: Module[]): Record<string, number> => {
  const depths: Record<string, number> = {}
  const adjList: Record<string, string[]> = {}

  // Initialize adjacency list
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
    depths[node] = Math.max(depths[node] ?? 0, depth)
    adjList[node].forEach((neighbor) => dfs(neighbor, depth + 1))
  }

  // Initialize DFS from each node
  Object.keys(adjList).forEach((node) => dfs(node, 0))

  return depths
}

export const layoutNodes = (modules: Module[]) => {
  const nodes: Node[] = []
  const edges: Edge[] = []
  const xCounters: Record<number, number> = {}

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
      data: { title: module.name },
      type: 'groupNode',
      style: {
        width: groupWidth,
        height: (verbs?.length ?? 1) * 50 + 50,
        zIndex: -1,
      },
    })
    let y = 40
    verbs.forEach((verb) => {
      const calls = verb?.verb?.metadata
        .filter((meta) => meta.value.case === 'calls')
        .map((meta) => meta.value.value as MetadataCalls)

      nodes.push({
        id: `${module.name}-${verb.verb?.name}`,
        position: { x: 20, y: y },
        connectable: false,
        data: { title: verb.verb?.name },
        type: 'verbNode',
        parentNode: module.name,
        style: {
          width: groupWidth - 40,
          height: 40,
        },
      })

      const uniqueEdgeIds = new Set<string>()
      calls?.map((call) =>
        call.calls.forEach((call) => {
          const edgeId = `${module.name}-${verb.verb?.name}-${call.module}-${call.name}`
          if (!uniqueEdgeIds.has(edgeId)) {
            uniqueEdgeIds.add(edgeId)
            edges.push({
              id: edgeId,
              source: `${module.name}-${verb.verb?.name}`,
              target: `${call.module}-${call.name}`,
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
