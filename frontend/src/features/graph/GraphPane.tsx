import { useContext, useEffect } from 'react'
import ReactFlow, { Background, Controls, useEdgesState, useNodesState } from 'reactflow'
import 'reactflow/dist/style.css'
import { modulesContext } from '../../providers/modules-provider'
import { GroupNode } from './GroupNode'
import { VerbNode } from './VerbNode'
import { layoutNodes } from './create-layout'
import { Config, Module, Secret, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import React from 'react'
import { SecretNode } from './SecretNode'
import { ConfigNode } from './ConfigNode'
const nodeTypes = { groupNode: GroupNode, verbNode: VerbNode, secretNode: SecretNode, configNode: ConfigNode }

export type FTLNode = Module | Verb | Secret | Config

interface GraphPaneProps {
  onTapped?: (item: FTLNode | null) => void
}

export const GraphPane: React.FC<GraphPaneProps> = ({ onTapped }) => {
  const modules = useContext(modulesContext)
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const [selectedNode, setSelectedNode] = React.useState<FTLNode | null>(null)

  useEffect(() => {
    const { nodes: newNodes, edges: newEdges } = layoutNodes(modules.modules, modules.topology)

    // Need to update after render loop for ReactFlow to pick up the changes
    setTimeout(() => {
      setNodes(newNodes)
      setEdges(newEdges)
    }, 0)
  }, [modules.modules])

  useEffect(() => {
    const currentNodes = nodes.map((node) => {
      return { ...node, data: { ...node.data, selected: node.data.item === selectedNode } }
    })
    setNodes(currentNodes)
  }, [selectedNode])

  return (
    <ReactFlow
      key={`${nodes.length}-${edges.length}`}
      proOptions={{ hideAttribution: true }}
      nodes={nodes}
      edges={edges}
      nodeTypes={nodeTypes}
      onNodesChange={onNodesChange}
      onEdgesChange={onEdgesChange}
      maxZoom={2}
      minZoom={0.1}
      nodeDragThreshold={2}
      onNodeClick={(_, node) => {
        setSelectedNode(node.data.item)
        onTapped?.(node.data.item)
      }}
      onNodeDragStart={(_, node) => {
        setSelectedNode(node.data.item)
        onTapped?.(node.data.item)
      }}
      onPaneClick={() => {
        setSelectedNode(null)
        onTapped?.(null)
      }}
      fitView
    >
      <Controls />
      <Background color='#888' gap={16} size={1} />
    </ReactFlow>
  )
}
