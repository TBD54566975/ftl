import { CubeTransparentIcon } from '@heroicons/react/24/outline'
import { useContext, useEffect } from 'react'
import ReactFlow, { Background, Controls, useEdgesState, useNodesState } from 'reactflow'
import 'reactflow/dist/style.css'
import { modulesContext } from '../../providers/modules-provider'
import { GroupNode } from './GroupNode'
import { VerbNode } from './VerbNode'
import { layoutNodes } from './create-layout'
import { Page } from '../../layout'
import { Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import React from 'react'
const nodeTypes = { groupNode: GroupNode, verbNode: VerbNode }

interface GraphPageProps {
  onTapped?: (item: Module | Verb) => void
}

export const GraphPage: React.FC<GraphPageProps> = ({ onTapped }) => {
  const modules = useContext(modulesContext)
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const [selectedNode, setSelectedNode] = React.useState<Module | Verb | null>(null)

  useEffect(() => {
    const { nodes, edges } = layoutNodes(modules.modules)
    setNodes(nodes)
    setEdges(edges)
  }, [modules])

  useEffect(() => {
    if (!selectedNode) return

    const currentNodes = nodes.map((node) => {
      return { ...node, data: { ...node.data, selected: node.data.item === selectedNode } }
    })
    setNodes(currentNodes)
  }, [selectedNode])

  return (
    <Page>
      <Page.Header icon={<CubeTransparentIcon />} title='Graph' />
      <Page.Body className='flex h-full bg-slate-800'>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          nodeTypes={nodeTypes}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          maxZoom={2}
          minZoom={0.1}
          onNodeClick={(_, node) => {
            setSelectedNode(node.data.item)
            onTapped?.(node.data.item)
          }}
          onNodeDragStart={(_, node) => {
            setSelectedNode(node.data.item)
            onTapped?.(node.data.item)
          }}
          fitView
        >
          <Controls />
          <Background
            color='#888' // Color of the grid lines
            gap={16} // Distance between grid lines
            size={1} // Thickness of the grid lines
          />
        </ReactFlow>
      </Page.Body>
    </Page>
  )
}
