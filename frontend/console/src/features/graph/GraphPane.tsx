import { useCallback, useEffect } from 'react'
import ReactFlow, { Background, Controls, useEdgesState, useNodesState, useReactFlow, ReactFlowProvider } from 'reactflow'
import 'reactflow/dist/style.css'
import React from 'react'
import { useStreamModules } from '../../api/modules/use-stream-modules'
import type { Config, Data, Database, Enum, Module, Secret, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { DeclNode } from './DeclNode'
import { GroupNode } from './GroupNode'
import { VerbNode } from './VerbNode'
import { layoutNodes } from './create-layout'
const nodeTypes = { groupNode: GroupNode, verbNode: VerbNode, declNode: DeclNode }

export type FTLNode = Module | Verb | Secret | Config | Data | Database | Enum

interface GraphPaneProps {
  onTapped?: (item: FTLNode | null) => void
}

const GraphPaneInner: React.FC<GraphPaneProps> = ({ onTapped }) => {
  const modules = useStreamModules()
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
      <div className='absolute bottom-3.5 left-12 z-10'>
        <select
          value={selectedModule}
          onChange={(e) => handleModuleSelect(e.target.value)}
          className='w-48 px-2 py-1 text-sm dark:bg-gray-800 dark:text-gray-100 border border-gray-200 dark:border-gray-700 rounded-md shadow-sm
            dark:hover:bg-gray-700 hover:bg-gray-100 focus:outline-none focus:ring-0
            backdrop-blur-sm'
        >
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
