import React from 'react'
import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import { PageHeader } from '../../components/PageHeader'
import { modulesContext } from '../../providers/modules-provider'
import { ModuleNode } from './ModuleNode'
import { createLayoutDataStructure } from './create-layout-data-structure'
import ReactFlow, { useNodesState, useEdgesState, Controls, MiniMap } from 'reactflow'
import 'reactflow/dist/style.css'

const nodeTypes = {
  moduleNode: ModuleNode,
}

const defaultViewport = { x: 0, y: 0, zoom: 1.5 }
export const ModulesPage = () => {
  const modules = React.useContext(modulesContext)
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  console.log(nodes)
  React.useEffect(() => {
    const [nodes, edges] = createLayoutDataStructure(modules)
    setNodes(nodes)
    setEdges(edges)
  }, [modules, setEdges, setNodes])
  return (
    <div className='h-full w-full flex flex-col'>
      <PageHeader icon={<Square3Stack3DIcon />} title='Modules' />
      <div className='flex-1 relative p-8'>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          nodeTypes={nodeTypes}
          fitView
          attributionPosition='bottom-left'
          defaultViewport={defaultViewport}
        >
          <Controls />
          <MiniMap />
        </ReactFlow>
      </div>
    </div>
  )
}
