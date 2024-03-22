import { useContext, useEffect, useRef, useState } from 'react'
import ELK, { ElkExtendedEdge, ElkNode } from 'elkjs/lib/elk.bundled.js'
import { modulesContext } from '../../providers/modules-provider'
import React from 'react'
import { layoutNodes } from './graph'
import svgPanZoom from 'svg-pan-zoom'

const elk = new ELK()

const ConsolePane = () => {
  const modules = useContext(modulesContext)
  const [layout, setLayout] = useState<ElkNode | null>(null)
  const [nodes, setNodes] = useState<ElkNode[]>([])
  const [edges, setEdges] = useState<ElkExtendedEdge[]>([])
  const svgRef = useRef(null)

  useEffect(() => {
    let panZoomInstance: SvgPanZoom.Instance | null = null

    if (svgRef.current) {
      panZoomInstance = svgPanZoom(svgRef.current, {
        zoomEnabled: true,
        controlIconsEnabled: true,
        fit: true,
        center: true,
      })
    }

    return () => panZoomInstance?.destroy()
  }, [])

  useEffect(() => {
    const { nodes, edges } = layoutNodes(modules.modules)
    setNodes(nodes)
    setEdges(edges)
  }, [modules])

  useEffect(() => {
    const graph = {
      id: 'root',
      layoutOptions: {
        'elk.algorithm': 'layered',
      },
      children: nodes.map((node) => ({
        ...node,
        layoutOptions: {
          'elk.nodeKind': 'group',
          'elk.padding': '[top=20,left=20,bottom=20,right=20]', // Group padding
        },
      })),
      edges: edges,
    }

    elk
      .layout(graph)
      .then((result) => {
        setLayout(result)
      })
      .catch((error) => console.error('ELK Layout calculation failed', error))
  }, [nodes, edges])

  if (!layout) {
    return <div>Loading graph...</div>
  }

  return (
    <div className='flex-1 bg-gray-700 text-white'>
      <svg ref={svgRef} width='100%' height='100%'>
        <defs>
          <marker id='arrowhead' markerWidth='10' markerHeight='7' refX='9' refY='3.5' orient='auto'>
            <polygon points='0 0, 10 3.5, 0 7' fill='white' />
          </marker>
        </defs>

        {layout.children?.map((moduleNode) => (
          <React.Fragment key={moduleNode.id}>
            <rect
              x={moduleNode.x}
              y={moduleNode.y}
              width={moduleNode.width}
              height={moduleNode.height}
              fill='lightblue'
            />

            {moduleNode.labels?.map((label, labelIndex) => (
              <text
                key={`${moduleNode.id}-label-${labelIndex}`}
                x={(moduleNode.x ?? 0) + 5}
                y={(moduleNode.y ?? 0) + 20}
                fill='black'
              >
                {label.text}
              </text>
            ))}
            {moduleNode.children?.map((verbNode) => (
              <React.Fragment key={verbNode.id}>
                <rect x={verbNode.x} y={verbNode.y} width={verbNode.width} height={verbNode.height} fill='lightgreen' />{' '}
                {verbNode.labels?.map((label, labelIndex) => (
                  <text
                    key={`${verbNode.id}-label-${labelIndex}`}
                    x={(verbNode.x ?? 0) + 5}
                    y={(verbNode.y ?? 0) + 20}
                    fill='black'
                  >
                    {label.text}
                  </text>
                ))}
              </React.Fragment>
            ))}
          </React.Fragment>
        ))}

        {/* Edges */}
        {layout.edges?.map((edge, index) => {
          const x1 = edge.sections?.[0]?.startPoint.x ?? 0
          const y1 = edge.sections?.[0]?.startPoint.y ?? 0
          const x2 = edge.sections?.[0]?.endPoint.x ?? 0
          const y2 = edge.sections?.[0]?.endPoint.y ?? 0

          return (
            <line
              key={edge.id || index}
              x1={x1}
              y1={y1}
              x2={x2}
              y2={y2}
              stroke='white'
              strokeWidth={1}
              markerEnd='url(#arrowhead)'
            />
          )
        })}
      </svg>
    </div>
  )
}

export default ConsolePane
