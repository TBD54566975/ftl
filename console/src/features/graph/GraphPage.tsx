import { useContext, useEffect } from 'react'
import ReactFlow, { Controls, MiniMap, useNodesState, useEdgesState } from 'reactflow'
import { schemaContext } from '../../providers/schema-provider'
import { GroupNode } from './GroupNode'
import { VerbNode } from './VerbNode'
import { layoutNodes } from './create-layout'
import 'reactflow/dist/style.css'

const nodeTypes = { groupNode: GroupNode, verbNode: VerbNode }

export default function GraphPage() {
  const schema = useContext(schemaContext)
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])

  useEffect(() => {
    const { nodes, edges } = layoutNodes(schema)
    setNodes(nodes)
    setEdges(edges)
  }, [schema, setEdges, setNodes])

  return (
    <div style={{ width: '100vw', height: '90vh' }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        fitView
      >
        <Controls />
        <MiniMap />
      </ReactFlow>
    </div>
  )
}
