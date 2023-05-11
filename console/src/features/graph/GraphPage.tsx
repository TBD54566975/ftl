import { useContext } from 'react'
import ReactFlow, { Controls, MiniMap, Node, Edge } from 'reactflow'

import 'reactflow/dist/style.css'
import { schemaContext } from '../../providers/schema-provider'
import { MetadataCalls } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { GroupNode } from './GroupNode'
import { VerbNode } from './VerbNode'

const nodeTypes = { groupNode: GroupNode, verbNode: VerbNode }
const groupWidth = 200

export default function GraphPage() {
  const schema = useContext(schemaContext)

  const nodes: Node[] = []
  const edges: Edge[] = []
  let x = 0
  schema.forEach(module => {
    const verbs = module.schema?.decls.filter(decl => decl.value.case === 'verb')
    nodes.push({
      id: module.schema?.name ?? '',
      position: { x: x, y: 0 },
      data: { title: module.schema?.name },
      type: 'groupNode',
      style: {
        width: groupWidth,
        height: (verbs?.length ?? 1) * 50 + 50,
        zIndex: -1,
      },
    })
    let y = 40
    module.schema?.decls
      .filter(decl => decl.value.case === 'verb')
      .forEach(verb => {
        const calls = verb?.value.value?.metadata
          .filter(meta => meta.value.case === 'calls')
          .map(meta => meta.value.value as MetadataCalls)

        nodes.push({
          id: `${module.schema?.name}-${verb.value.value?.name}`,
          position: { x: x + 20, y: y },
          connectable: false,
          data: { title: verb.value.value?.name },
          type: 'verbNode',
          style: {
            width: groupWidth - 40,
            height: 40,
          },
        })

        calls?.map(call =>
          call.calls.forEach(call => {
            edges.push({
              id: `${module.schema?.name}-${verb.value.value?.name}-${call.module}-${call.name}`,
              source: `${module.schema?.name}-${verb.value.value?.name}`,
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

  return (
    <div style={{ width: '100vw', height: '90vh' }}>
      <ReactFlow nodes={nodes} edges={edges} nodeTypes={nodeTypes} fitView>
        <Controls />
        <MiniMap />
      </ReactFlow>
    </div>
  )
}
