import React, { useState } from 'react'
import RightPanel from './RightPanel'
import BottomPanel from './BottomPanel'
import TopBar from './TopBar'
import { GraphPage } from '../graph/GraphPage'
import { Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { ExpandablePanelProps } from './ExpandablePanel'

const MIN_RIGHT_PANEL_WIDTH = 200
const MIN_BOTTOM_PANEL_HEIGHT = 100

const ConsolePage = () => {
  const [rightPanelWidth, setRightPanelWidth] = useState(200)
  const [bottomPanelHeight, setBottomPanelHeight] = useState(150)
  const [isDraggingHorizontal, setIsDraggingHorizontal] = useState(false)
  const [isDraggingVertical, setIsDraggingVertical] = useState(false)
  const [selectedNode, setSelectedNode] = useState<Module | Verb | null>(null)

  const startDraggingHorizontal = (e: React.MouseEvent<HTMLDivElement>) => {
    e.preventDefault()
    setIsDraggingHorizontal(true)
  }

  const startDraggingVertical = (e: React.MouseEvent<HTMLDivElement>) => {
    e.preventDefault()
    setIsDraggingVertical(true)
  }

  const stopDragging = () => {
    setIsDraggingHorizontal(false)
    setIsDraggingVertical(false)
  }

  const onDragHorizontal = (e: React.MouseEvent<HTMLDivElement>) => {
    if (isDraggingHorizontal) {
      const newWidth = Math.max(window.innerWidth - e.clientX, MIN_RIGHT_PANEL_WIDTH)
      setRightPanelWidth(newWidth > 0 ? newWidth : 0)
    }
  }

  const onDragVertical = (e: React.MouseEvent<HTMLDivElement>) => {
    if (isDraggingVertical) {
      const newHeight = Math.max(window.innerHeight - e.clientY, MIN_BOTTOM_PANEL_HEIGHT)
      setBottomPanelHeight(newHeight > 0 ? newHeight : 0)
    }
  }

  return (
    <div
      className='flex flex-col h-screen'
      onMouseMove={(e) => {
        if (isDraggingHorizontal) onDragHorizontal(e)
        if (isDraggingVertical) onDragVertical(e)
      }}
      onMouseUp={stopDragging}
      onMouseLeave={stopDragging}
    >
      <TopBar />
      <div className='flex flex-1'>
        <div className='flex-1 bg-gray-700 text-white'>
          <GraphPage onTapped={setSelectedNode} />
        </div>
        <div
          className='cursor-col-resize bg-gray-600 hover:bg-indigo-600'
          onMouseDown={startDraggingHorizontal}
          style={{ width: '3px', cursor: 'col-resize' }}
        />
        <RightPanel width={rightPanelWidth} panels={panelsForNode(selectedNode)} />
      </div>
      <div
        className='cursor-row-resize bg-gray-600 hover:bg-indigo-600'
        onMouseDown={startDraggingVertical}
        style={{ height: '3px', cursor: 'row-resize' }}
      />
      <BottomPanel height={bottomPanelHeight} />
    </div>
  )
}

const panelsForNode = (node: Module | Verb | null) => {
  if (node instanceof Module) {
    return [
      { title: node.name, expanded: true, children: 'Module Content' },
      { title: 'Verbs', expanded: true, children: `${node.verbs.map((v) => v.verb?.name).join(', ')}` },
    ] as ExpandablePanelProps[]
  } else if (node instanceof Verb) {
    return [
      { title: node.verb?.name, expanded: true, children: 'Verb Content' },
      { title: 'Verb 2', expanded: true, children: 'Verb 2 Content' },
    ] as ExpandablePanelProps[]
  } else {
    return [] as ExpandablePanelProps[]
  }
}

export default ConsolePage
