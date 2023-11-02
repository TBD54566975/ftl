import { CubeTransparentIcon } from '@heroicons/react/24/outline'
import { useContext, useEffect } from 'react'
import ReactFlow, { Controls, MiniMap, useEdgesState, useNodesState } from 'reactflow'
import 'reactflow/dist/style.css'
import { modulesContext } from '../../providers/modules-provider'
import { GroupNode } from './GroupNode'
import { VerbNode } from './VerbNode'
import { layoutNodes } from './create-layout'
import { Page } from '../../layout'
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
    <Page>
      <Page.Header icon={<CubeTransparentIcon />} title='Graph' />
      <Page.Body className='flex h-full'>
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
      </Page.Body>
    </Page>
  )
}
