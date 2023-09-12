import { Edge, Node } from 'reactflow'
import { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { MetadataCalls } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'

const groupWidth = 200

export const layoutNodes = (modules: Module[]) => {
  let x = 0
  const nodes: Node[] = []
  const edges: Edge[] = []
  modules.forEach((module) => {
    const verbs = module.verbs
    nodes.push({
      id: module.name ?? '',
      position: { x: x, y: 0 },
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

      calls?.map((call) =>
        call.calls.forEach((call) => {
          edges.push({
            id: `${module.name}-${verb.verb?.name}-${call.module}-${call.name}`,
            source: `${module.name}-${verb.verb?.name}`,
            target: `${call.module}-${call.name}`,
            style: { stroke: 'rgb(251 113 133)' },
            animated: true,
          })
          call.name
          call.module
        }),
      )

      y += 50
    })
    x += 300
  })
  return { nodes, edges }
}
