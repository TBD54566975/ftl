import cytoscape from 'cytoscape'
import type { FcoseLayoutOptions } from 'cytoscape-fcose'
import fcose from 'cytoscape-fcose'

import { useEffect, useRef, useState } from 'react'
import type React from 'react'
import { useStreamModules } from '../../api/modules/use-stream-modules'
import type { FTLNode } from './GraphPane'
import { getGraphData } from './graph-utils'

cytoscape.use(fcose)

interface NewGraphPaneProps {
  onTapped?: (item: FTLNode | null) => void
}

const ZOOM_THRESHOLD = 0.6

export const NewGraphPane: React.FC<NewGraphPaneProps> = ({ onTapped }) => {
  const modules = useStreamModules()

  const cyRef = useRef<HTMLDivElement>(null)
  const cyInstance = useRef<cytoscape.Core | null>(null)
  const [nodePositions, setNodePositions] = useState<Record<string, { x: number; y: number }>>({})

  // Initialize Cytoscape
  useEffect(() => {
    if (!cyRef.current) return

    cyInstance.current = cytoscape({
      container: cyRef.current,
      userZoomingEnabled: true,
      userPanningEnabled: true,
      boxSelectionEnabled: false,
      autoungrabify: true,
      style: [
        {
          selector: 'node',
          style: {
            'background-color': '#64748b',
            label: 'data(label)',
            'text-valign': 'center',
            'text-halign': 'center',
            shape: 'round-rectangle',
            width: '120px',
            height: '40px',
            'text-wrap': 'wrap',
            'text-max-width': '100px',
            'text-overflow-wrap': 'anywhere',
            'font-size': '12px',
          },
        },
        {
          selector: 'edge',
          style: {
            width: 2,
            'line-color': '#6366f1',
            'curve-style': 'bezier',
            'target-arrow-shape': 'triangle',
            'target-arrow-color': '#6366f1',
            'arrow-scale': 1,
          },
        },
        {
          selector: '$node > node',
          style: {
            'padding-top': '10px',
            'padding-left': '10px',
            'padding-bottom': '10px',
            'padding-right': '10px',
            'text-valign': 'top',
            'text-halign': 'center',
            'background-color': '#94a3b8',
          },
        },
        {
          selector: 'node[type="groupNode"]',
          style: {
            'background-color': '#6366f1',
            'background-opacity': 0.8,
            shape: 'round-rectangle',
            width: '180px',
            height: '120px',
            'text-valign': 'top',
            'text-halign': 'center',
            'text-wrap': 'wrap',
            'text-max-width': '120px',
            'text-overflow-wrap': 'anywhere',
            'font-size': '14px',
          },
        },
        {
          selector: ':parent',
          style: {
            'text-valign': 'top',
            'text-halign': 'center',
            'background-opacity': 0.3,
          },
        },
        {
          selector: '.selected',
          style: {
            'background-color': '#3b82f6',
            'border-width': 2,
            'border-color': '#60a5fa',
          },
        },
        {
          selector: 'node[type="node"]',
          style: {
            'background-color': 'data(backgroundColor)',
            shape: 'round-rectangle',
            width: '100px',
            height: '30px',
            'border-width': '1px',
            'border-color': '#475569',
            'text-wrap': 'wrap',
            'text-max-width': '80px',
            'text-overflow-wrap': 'anywhere',
            'font-size': '11px',
          },
        },
      ],
    })

    // Add click handlers
    cyInstance.current.on('tap', 'node', (evt) => {
      const node = evt.target
      const nodeType = node.data('type')
      const item = node.data('item')
      const zoom = evt.cy.zoom()

      if (zoom < ZOOM_THRESHOLD) {
        if (nodeType === 'node') {
          const parent = node.parent()
          if (parent.length) {
            onTapped?.(parent.data('item'))
            return
          }
        }
      }

      if (nodeType === 'groupNode' || (nodeType === 'node' && zoom >= ZOOM_THRESHOLD)) {
        onTapped?.(item)
      }
    })

    cyInstance.current.on('tap', (evt) => {
      if (evt.target === cyInstance.current) {
        onTapped?.(null)
      }
    })

    // Update zoom level event handler
    cyInstance.current.on('zoom', (evt) => {
      const zoom = evt.target.zoom()
      const elements = evt.target.elements()

      if (zoom < ZOOM_THRESHOLD) {
        // Hide child nodes
        elements.nodes('node[type != "groupNode"]').style('opacity', 0)

        // Show only module-level edges (type="moduleConnection")
        elements.edges('[type = "moduleConnection"]').style('opacity', 1)
        elements.edges('[type = "childConnection"]').style('opacity', 0)

        // Updated text settings for zoomed out view
        elements.nodes('node[type = "groupNode"]').style({
          'text-valign': 'center',
          'text-halign': 'center',
          'font-size': '18px',
          'text-max-width': '160px',
          width: '180px',
        })
      } else {
        // Show all nodes
        elements.nodes().style('opacity', 1)

        // Show only verb-level edges (type="childConnection")
        elements.edges('[type = "moduleConnection"]').style('opacity', 0)
        elements.edges('[type = "childConnection"]').style('opacity', 1)

        // Move text to top when zoomed in
        elements.nodes('node[type = "groupNode"]').style({
          'text-valign': 'top',
          'text-halign': 'center',
          'font-size': '14px',
          'text-max-width': '160px',
          width: '180px',
        })
      }
    })

    return () => {
      cyInstance.current?.destroy()
    }
  }, [onTapped])

  // Modify the data loading effect
  useEffect(() => {
    if (!cyInstance.current) return

    const elements = getGraphData(modules.data, nodePositions)
    const cy = cyInstance.current

    // Update existing elements and add new ones
    for (const element of elements) {
      const id = element.data?.id
      if (!id) continue // Skip elements without an id

      const existingElement = cy.getElementById(id)

      if (existingElement.length) {
        // Update existing element data
        existingElement.data(element.data)

        // If it's a node and doesn't have saved position, update position
        if (element.group === 'nodes' && !nodePositions[id]) {
          existingElement.position(element.position || { x: 0, y: 0 })
        }
      } else {
        // Add new element
        cy.add(element)
      }
    }

    // Remove elements that no longer exist in the data
    for (const element of cy.elements()) {
      const elementId = element.data('id')
      const stillExists = elements.some((e) => e.data?.id === elementId)
      if (!stillExists) {
        element.remove()
      }
    }

    // Only run layout for new nodes without positions
    const hasNewNodesWithoutPositions = cy.nodes().some((node) => {
      const nodeId = node.data('id')
      return node.data('type') === 'groupNode' && !nodePositions[nodeId]
    })

    if (hasNewNodesWithoutPositions) {
      const layoutOptions: FcoseLayoutOptions = {
        name: 'fcose',
        animate: false,
        quality: 'proof',
        nodeSeparation: 75,
        idealEdgeLength: 50,
        nodeRepulsion: 4500,
        padding: 30,
        randomize: false,
        // Make the layout more deterministic
        tile: true,
        tilingPaddingVertical: 20,
        tilingPaddingHorizontal: 20,
      }

      const layout = cy.layout(layoutOptions)
      layout.run()

      layout.on('layoutstop', () => {
        const newPositions = { ...nodePositions }
        for (const node of cy.nodes()) {
          const nodeId = node.data('id')
          newPositions[nodeId] = node.position()
        }
        setNodePositions(newPositions)
      })
    }

    cy.fit()
  }, [nodePositions, modules.data])

  return (
    <div style={{ width: '100%', height: '100%', position: 'relative' }}>
      <div ref={cyRef} style={{ width: '100%', height: '100%' }} />
    </div>
  )
}
