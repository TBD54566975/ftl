// https://codesandbox.io/p/sandbox/elkjs-layout-subflows-9og9hl?file=%2FApp.js%3A6%2C1
// Imports slightly refactored:
//  - older flow package to xyflow
//  - elk import from bundled js

import { useEffect, useState } from 'react'
import {
  ReactFlow,
  ReactFlowProvider,
  useNodesState,
  useEdgesState,
  useReactFlow,
  Controls,
  Background,
  Handle,
  Position,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'

import ELK from 'elkjs/lib/elk.bundled.js'

const CustomNodeA = ({ data }) => {
  return (
    <div className='h-full w-full border border-gray-600 rounded-sm bg-indigo-100 truncate px-4 py-2 text-center'>
      <Handle
        type='target'
        position={Position.Top}
        isConnectable={true}
      />
      <div>{data?.label}</div>
      <Handle
        type='source'
        position={Position.Bottom}
        isConnectable={true}
      />
    </div>
  )
}

const nodeTypes = {
  customNodeA: CustomNodeA,
}

const initialNodes = [
  {
    id: 'A',
    group: '1'
  },
  {
    id: 'B',
    group: '1'
  },
  {
    id: 'C',
    group: '1'
  },
  {
    id: 'D',
    group: '2'
  },
  {
    id: 'E',
    group: '2'
  },
  {
    id: 'F',
    group: '3'
  },
  {
    id: 'G',
    group: '3'
  },
  {
    id: 'H',
    group: '1'
  },
  {
    id: 'I',
    group: '1'
  }
]

const initialGroups = [
  {
    id: '1',
    width: 100,
    height: 100
  },
  {
    id: '2',
    width: 100,
    height: 100
  },
  {
    id: '3',
    width: 100,
    height: 100
  }
]

const initialEdges = [
  { id: '1', source: '1', target: '2' },
  { id: '2', source: '2', target: '3' },
  { id: '3', source: 'A', target: 'B' },
  { id: '4', source: 'A', target: 'I' },
  { id: '5', source: 'B', target: 'C' },
  { id: '6', source: 'B', target: 'H' },
]

const elk = new ELK()

const graph = {
  id: 'root',
  layoutOptions: {
    'elk.algorithm': 'mrtree',
    'elk.direction': 'DOWN'
  },
  children: initialGroups.map((group) => ({
    id: group.id,
    width: group.width,
    height: group.height,
    layoutOptions: {
      'elk.direction': 'DOWN'
    },
    children: initialNodes
      .filter((node) => node.group === group.id)
      .map((node) => ({
        id: node.id,
        width: 100,
        height: 50,
        layoutOptions: {
          'elk.direction': 'DOWN'
        }
      }))
  })),
  edges: initialEdges.map((edge) => ({
    id: edge.id,
    sources: [edge.source],
    targets: [edge.target]
  }))
}

export default async function createLayout() {
  const layout = await elk.layout(graph)
  const nodes = layout.children.reduce((result, current) => {
    result.push({
      id: current.id,
      position: { x: current.x, y: current.y },
      style: { width: current.width, height: current.height }
    })

    current.children.forEach((child) =>
      result.push({
        id: child.id,
        position: { x: child.x, y: child.y },
        type: 'customNodeA',
        data: { label: child.id, title: 'title' },
        style: { width: child.width, height: child.height },
        parentId: current.id,
      })
    )

    return result
  }, [])

  return {
    nodes,
    edges: initialEdges
  }
}

function Flow() {
  const [graph, setGraph] = useState(null)

  useEffect(() => {
    (async () => {
      const { nodes, edges } = await createLayout()
      setGraph({ nodes, edges })
    })()
  }, [])

  return (
    <div style={{ height: '100%' }}>
      {graph && (
        <ReactFlow
          defaultNodes={graph.nodes}
          defaultEdges={graph.edges}
          nodeTypes={nodeTypes}
          fitView
          defaultEdgeOptions={{
            //type: 'step',
            zIndex: 100,
            pathOptions: { offset: 1 }
          }}
        >
          <Background />
          <Controls />
        </ReactFlow>
      )}
    </div>
  )
}

export const GraphPane = () => {
  return (
    <ReactFlowProvider>
      <Flow />
    </ReactFlowProvider>
  )
}
