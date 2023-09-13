import { Schema } from '@mui/icons-material'
import { useContext, useEffect } from 'react'
import ReactFlow, { Controls, MiniMap, useEdgesState, useNodesState } from 'reactflow'
import 'reactflow/dist/style.css'
import { PageHeader } from '../../components/PageHeader'
import { modulesContext } from '../../providers/modules-provider'
import { GroupNode } from './GroupNode'
import { VerbNode } from './VerbNode'
import { layoutNodes } from './create-layout'

const nodeTypes = { groupNode: GroupNode, verbNode: VerbNode }

export const GraphPage = () => {
  const modules = useContext(modulesContext)
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])

  useEffect(() => {
    const { nodes, edges } = layoutNodes(modules.modules)
    setNodes(nodes)
    setEdges(edges)
  }, [modules, setEdges, setNodes])

  return (
    <>
      <PageHeader icon={<Schema />} title='Graph' />
      <div style={{ width: '100vw', height: '100vh' }}>
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
    </>
  )
}
