import { useCallback, useEffect } from 'react'
import type { Edge, Node } from '@xyflow/react'
import {
  Background,
  Controls,
  ReactFlow,
  ReactFlowProvider,
  Panel,
  useNodesState,
  useEdgesState,
  useReactFlow,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'

import ELK from 'elkjs/lib/elk.bundled.js'

import React from 'react'
import { useStreamModules } from '../../api/modules/use-stream-modules'
import type { Config, Module, Secret, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import type { Ref } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { DeclNode } from './DeclNode'
import { ModuleNode } from './ModuleNode'

const moduleWidth = 200
const declWidth = 160
const declHeight = 28

function countDecls(m: Module) {
  return (
    m.configs.length +
    m.data.length +
    m.databases.length +
    m.enums.length +
    m.secrets.length +
    m.subscriptions.length +
    m.topics.length +
    m.typealiases.length +
    m.verbs.length
  )
}

function moduleHeight(m: Module) {
  return 32 * (countDecls(m) + 1)
}

function modulesToGraph(modules: Module[]) {
  const moduleNames = modules.map((m) => m.name)
  let nodes: Node[] = []
  const edges: Edge[] = []
  const existingEdges: { [key: string]: boolean } = {}
  const addRef = (r: Ref, m: Module, name?: string) => {
    if (!name) {
      return
    }
    if (r.module === m.name) {
      return
    }
    if (!moduleNames.includes(r.module)) {
      return
    }
    edges.push({
      id: `${m.name}.${name}-${r.module}.${r.name}`,
      source: `${m.name}.${name}`,
      target: `${r.module}.${r.name}`,
      style: { stroke: 'rgb(151 13 33)' },
      animated: true,
    })
    if (!existingEdges[`${m.name}-${r.module}`]) {
      existingEdges[`${m.name}-${r.module}`] = true
      edges.push({
        id: `${m.name}-${r.module}`,
        source: `${m.name}`,
        target: `${r.module}`,
        style: { stroke: 'rgb(251 113 133)' },
        animated: true,
      })
    }
  }
  modules.forEach((m, i) => {
    const children: Node[] = []
    const addChild = (d, name?: string) => {
      if (!name) {
        return
      }
      children.push({
        id: `${m.name}.${name}`,
        parentId: `${m.name}`,
        extent: 'parent',
        position: { x: 0, y: 0 },
        data: { title: name, item: d },
        type: 'declNode',
        draggable: true,
        style: {
          width: declWidth, //groupWidth,
          height: declHeight,
          zIndex: 0,
        },
        layoutOptions: {
          'elk.direction': 'DOWN',
          'elk.spacing.nodeNode': 90,
        },
      })
    }
    for (const d of m.configs) {
      addChild(d, d.config?.name)
      for (const r of d.references) {
        addRef(r, m, d.config?.name)
      }
    }
    for (const d of m.data) {
      addChild(d, d.data?.name)
      for (const r of d.references) {
        addRef(r, m, d.data?.name)
      }
    }
    for (const d of m.databases) {
      addChild(d, d.database?.name)
      for (const r of d.references) {
        addRef(r, m, d.database?.name)
      }
    }
    for (const d of m.enums) {
      addChild(d, d.enum?.name)
      for (const r of d.references) {
        addRef(r, m, d.enum?.name)
      }
    }
    for (const d of m.secrets) {
      addChild(d, d.secret?.name)
      for (const r of d.references) {
        addRef(r, m, d.secret?.name)
      }
    }
    for (const d of m.subscriptions) {
      addChild(d, d.subscription?.name)
      for (const r of d.references) {
        addRef(r, m, d.subscription?.name)
      }
    }
    for (const d of m.topics) {
      addChild(d, d.topic?.name)
      for (const r of d.references) {
        addRef(r, m, d.topic?.name)
      }
    }
    for (const d of m.typealiases) {
      addChild(d, d.typealias?.name)
      for (const r of d.references) {
        addRef(r, m, d.typealias?.name)
      }
    }
    for (const d of m.verbs) {
      addChild(d, d.verb?.name)
      for (const r of d.references) {
        addRef(r, m, d.verb?.name)
      }
    }
    nodes.push({
      id: m.name ?? '',
      position: { x: i * 210, y: 0 },
      children,
      data: { title: m.name, item: m },
      type: 'moduleNode',
      draggable: true,
      style: {
        width: moduleWidth, //groupWidth,
        height: moduleHeight(m),
        zIndex: -1,
      },
    })
  })
  return { nodes, edges }
}

const elk = new ELK()

const useLayoutedElements = () => {
  //const { getNodes, setNodes, getEdges, fitView } = useReactFlow();
  const getNodes = () => []
  const setNodes = () => {}
  const getEdges = () => []
  const fitView = () => {}
  const defaultOptions = {
    'elk.algorithm': 'layered',
    'elk.layered.spacing.nodeNodeBetweenLayers': 100,
    'elk.spacing.nodeNode': 80,
  };

  const getLayoutedElements = useCallback((options) => {
    const layoutOptions = { ...defaultOptions, ...options };
    const graph = {
      id: 'root',
      layoutOptions: layoutOptions,
      children: getNodes().map((node) => ({
        ...node,
        width: node.measured.width,
        height: node.measured.height,
      })),
      edges: getEdges(),
    };

    elk.layout(graph).then(({ children }) => {
      // By mutating the children in-place we saves ourselves from creating a
      // needless copy of the nodes array.
      children.forEach((node) => {
        node.position = { x: node.x, y: node.y };
      });

      setNodes(children);
      window.requestAnimationFrame(() => {
        fitView();
      });
    });
  }, []);

  return { getLayoutedElements };
}

export type FTLNode = Module | Verb | Secret | Config

interface GraphPaneProps {
  onTapped?: (item: FTLNode | null) => void
}

export const GraphPane: React.FC<GraphPaneProps> = ({ onTapped }) => {
  //const modules = useModules()
  const streamed = useStreamModules()
  if (!streamed?.data) {
    return null
  }

  return (
    <ReactFlowProvider>
      <ActualGraph modules={streamed.data} />
    </ReactFlowProvider>
  )
}

const nodeTypes = { 'declNode': DeclNode, 'moduleNode': ModuleNode }

const ActualGraph = ({ modules }: {modules: Module[]}) => {
  const { nodes: initialNodes, edges: initialEdges } = modulesToGraph(modules)
  const [nodes,  , onNodesChange] = useNodesState(initialNodes)
  const [edges,  , onEdgesChange] = useEdgesState(initialEdges)
  const { getLayoutedElements } = useLayoutedElements()
  //const nodes = []
  //const onNodesChange = () => {}

  let flatNodes = nodes.map((n) => ({...n, children: undefined}))
  for (const node of nodes) {
    if (node.children) {
      flatNodes = flatNodes.concat(node.children)
    }
  }
  //console.log(flatNodes)

  return (
    <ReactFlow
      nodes={flatNodes}
      edges={edges}
      nodeTypes={nodeTypes}
      onNodesChange={onNodesChange}
      onEdgesChange={onEdgesChange}
      fitView
    >
      <Controls />
      <Background color='#888' gap={16} size={1} />
      <Panel position='top-right'>
         <button
          onClick={() =>
            getLayoutedElements({
              'elk.algorithm': 'org.eclipse.elk.force',
            })
          }
         >
           force layout
         </button>
      </Panel>
    </ReactFlow>
  )
}
