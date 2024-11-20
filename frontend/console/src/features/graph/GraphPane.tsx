import { useCallback, useEffect } from 'react'
import ReactFlow, { Background, Controls, useEdgesState, useNodesState, useReactFlow, ReactFlowProvider } from 'reactflow'
import 'reactflow/dist/style.css'
import React from 'react'
import { useModules } from '../../api/modules/use-modules'
import type { Config, Module, Secret, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { ConfigNode } from './ConfigNode'
import { GroupNode } from './GroupNode'
import { SecretNode } from './SecretNode'
import { VerbNode } from './VerbNode'
import { layoutNodes } from './create-layout'
const nodeTypes = { groupNode: GroupNode, verbNode: VerbNode, secretNode: SecretNode, configNode: ConfigNode }

export type FTLNode = Module | Verb | Secret | Config

interface GraphPaneProps {
  onTapped?: (item: FTLNode | null) => void
}

const GraphPaneInner: React.FC<GraphPaneProps> = ({ onTapped }) => {
  const modules = useModules()
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const [selectedNode, setSelectedNode] = React.useState<FTLNode | null>(null)
  const [selectedModule, setSelectedModule] = React.useState('')
  const { setCenter } = useReactFlow()

  useEffect(() => {
    if (!modules.isSuccess) return
    const { nodes: newNodes, edges: newEdges } = layoutNodes(modules.data.modules, modules.data.topology)

    // Need to update after render loop for ReactFlow to pick up the changes
    setTimeout(() => {
      setNodes(newNodes)
      setEdges(newEdges)
    }, 0)
  }, [modules.data?.modules])

  useEffect(() => {
    const currentNodes = nodes.map((node) => {
      return { ...node, data: { ...node.data, selected: node.data.item === selectedNode } }
    })
    setNodes(currentNodes)
  }, [selectedNode])

  const moduleNodes = nodes.filter((node) => node.type === 'groupNode')

  const handleModuleSelect = useCallback(
    (moduleId: string) => {
      if (!moduleId) return

      const matchingNode = nodes.find((node) => node.type === 'groupNode' && node.id === moduleId)

      if (matchingNode) {
        setCenter(matchingNode.position.x, matchingNode.position.y, { zoom: 1.5, duration: 800 })
        setSelectedNode(matchingNode.data.item)
        onTapped?.(matchingNode.data.item)
        setSelectedModule(moduleId)
      }
    },
    [nodes, setCenter, onTapped],
  )

  return (
    <div style={{ width: '100%', height: '100%', position: 'relative' }}>
      <div
        style={{
          position: 'absolute',
          top: 10,
          left: 10,
          zIndex: 4,
          background: 'white',
          padding: '8px',
          borderRadius: '4px',
          boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
        }}
      >
        <select value={selectedModule} onChange={(e) => handleModuleSelect(e.target.value)} style={{ padding: '4px 8px', minWidth: '200px' }}>
          <option value=''>Select a module...</option>
          {moduleNodes.map((node) => (
            <option key={node.id} value={node.id}>
              {node.data.title}
            </option>
          ))}
        </select>
      </div>

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
    </div>
  )
}

export const GraphPane: React.FC<GraphPaneProps> = (props) => {
  return (
    <ReactFlowProvider>
      <GraphPaneInner {...props} />
    </ReactFlowProvider>
  )
}
