import cytoscape from 'cytoscape'
import type { FcoseLayoutOptions } from 'cytoscape-fcose'
import fcose from 'cytoscape-fcose'

import { useEffect, useRef, useState } from 'react'
import type React from 'react'
import { useStreamModules } from '../../api/modules/use-stream-modules'
import { useUserPreferences } from '../../providers/user-preferences-provider'
import { createGraphStyles } from './graph-styles'
import { type FTLNode, getGraphData } from './graph-utils'

cytoscape.use(fcose)

interface GraphPaneProps {
  onTapped?: (item: FTLNode | null) => void
}

const ZOOM_THRESHOLD = 0.6

export const GraphPane: React.FC<GraphPaneProps> = ({ onTapped }) => {
  const modules = useStreamModules()
  const { isDarkMode } = useUserPreferences()

  const cyRef = useRef<HTMLDivElement>(null)
  const cyInstance = useRef<cytoscape.Core | null>(null)
  const [nodePositions, setNodePositions] = useState<Record<string, { x: number; y: number }>>({})
  const resizeObserverRef = useRef<ResizeObserver | null>(null)

  // Initialize Cytoscape and ResizeObserver
  useEffect(() => {
    if (!cyRef.current) return

    cyInstance.current = cytoscape({
      container: cyRef.current,
      userZoomingEnabled: true,
      userPanningEnabled: true,
      boxSelectionEnabled: false,
      autoungrabify: true,
    })

    // Create ResizeObserver
    resizeObserverRef.current = new ResizeObserver(() => {
      if (!cyInstance.current) return

      cyInstance.current?.resize()
    })

    // Start observing the container
    resizeObserverRef.current.observe(cyRef.current)

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
      if (resizeObserverRef.current) {
        resizeObserverRef.current.disconnect()
      }
      cyInstance.current?.destroy()
    }
  }, [onTapped])

  // Modify the data loading effect
  useEffect(() => {
    if (!cyInstance.current) return

    const elements = getGraphData(modules.data, isDarkMode, nodePositions)
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
      const layoutOptions = {
        name: 'fcose',
        animate: false,
        quality: 'default',
        nodeSeparation: 50,
        idealEdgeLength: 50,
        nodeRepulsion: 4500,
        padding: 20,
        randomize: false,
        // Make the layout more deterministic
        tile: true,
        tilingPaddingVertical: 20,
        tilingPaddingHorizontal: 20,
      } as FcoseLayoutOptions

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
  }, [nodePositions, modules.data, isDarkMode])

  useEffect(() => {
    // Update your cytoscape instance with new styles when dark mode changes
    cyInstance.current?.style(createGraphStyles(isDarkMode))
  }, [isDarkMode])

  return (
    <div style={{ width: '100%', height: '100%', position: 'relative', minWidth: 0 }}>
      <div
        ref={cyRef}
        style={{
          width: '100%',
          height: '100%',
          position: 'absolute',
          zIndex: 0, // Ensure graph stays below other elements
        }}
      />
    </div>
  )
}
