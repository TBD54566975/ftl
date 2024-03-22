import { ElkExtendedEdge, ElkNode } from 'elkjs'
import { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { MetadataCalls } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'

export const layoutNodes = (modules: Module[]) => {
  const nodes: ElkNode[] = []
  const edges: ElkExtendedEdge[] = []

  modules.forEach((module) => {
    const verbs = module.verbs

    verbs.forEach((verb) => {
      const calls = verb?.verb?.metadata
        .filter((meta) => meta.value.case === 'calls')
        .map((meta) => meta.value.value as MetadataCalls)

      nodes.push({
        id: `${module.name}-${verb.verb?.name}`,
        width: 200,
        height: 50,
        labels: [{ text: `${module.name}.${verb.verb?.name}` }],
      })

      const uniqueEdgeIds = new Set<string>()
      calls?.forEach((call) =>
        call.calls.forEach((call) => {
          const sourceNode = `${module.name}-${verb.verb?.name}`
          const targetNode = `${call.module}-${call.name}`
          const edgeId = `${sourceNode}-${targetNode}`
          if (!uniqueEdgeIds.has(edgeId)) {
            uniqueEdgeIds.add(edgeId)
            edges.push({ id: edgeId, sources: [sourceNode], targets: [targetNode] })
          }
        }),
      )
    })
  })

  return { nodes, edges }
}
