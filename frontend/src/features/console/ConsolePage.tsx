import React, { useContext, useState } from 'react'
import RightPanel from './right-panel/RightPanel'
import BottomPanel from './BottomPanel'
import { FTLNode } from '../graph/GraphPage'
import { Config, Module, Secret, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { ExpandablePanelProps } from './ExpandablePanel'
import { headerForNode } from './right-panel/RightPanelHeader'
import { modulePanels } from './right-panel/ModulePanels'
import { GraphPane } from '../graph/GraphPane'
import { Page } from '../../layout'
import { CubeTransparentIcon } from '@heroicons/react/24/outline'
import { verbPanels } from './right-panel/VerbPanels'
import { secretPanels } from './right-panel/SecretPanels'
import { configPanels } from './right-panel/ConfigPanels'
import { modulesContext } from '../../providers/modules-provider'
import { NavigateFunction, useNavigate } from 'react-router-dom'

const MIN_RIGHT_PANEL_WIDTH = 200
const MIN_BOTTOM_PANEL_HEIGHT = 200

const ConsolePage = () => {
  const modules = useContext(modulesContext)
  const navigate = useNavigate()
  const [rightPanelWidth, setRightPanelWidth] = useState(300)
  const [bottomPanelHeight, setBottomPanelHeight] = useState(250)
  const [isDraggingHorizontal, setIsDraggingHorizontal] = useState(false)
  const [isDraggingVertical, setIsDraggingVertical] = useState(false)
  const [selectedNode, setSelectedNode] = useState<FTLNode | null>(null)

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
    <Page>
      <Page.Header icon={<CubeTransparentIcon />} title='Graph' />
      <Page.Body className='flex h-full'>
        <div
          className='flex flex-col h-screen'
          onMouseMove={(e) => {
            if (isDraggingHorizontal) onDragHorizontal(e)
            if (isDraggingVertical) onDragVertical(e)
          }}
          onMouseUp={stopDragging}
          onMouseLeave={stopDragging}
        >
          <div className='flex flex-1'>
            <div className='flex-1 bg-gray-800 text-white'>
              <GraphPane onTapped={setSelectedNode} />
            </div>
            <div
              className='cursor-col-resize bg-gray-200 dark:bg-gray-700 hover:bg-indigo-600'
              onMouseDown={startDraggingHorizontal}
              style={{ width: '3px', cursor: 'col-resize' }}
            />
            <RightPanel
              width={rightPanelWidth}
              header={headerForNode(selectedNode)}
              panels={panelsForNode(modules.modules, selectedNode, navigate)}
            />
          </div>
          <div
            className='cursor-row-resize bg-gray-200 dark:bg-gray-700 hover:bg-indigo-600'
            onMouseDown={startDraggingVertical}
            style={{ height: '3px', cursor: 'row-resize' }}
          />
          <BottomPanel height={bottomPanelHeight} />
        </div>
      </Page.Body>
    </Page>
  )
}

const panelsForNode = (modules: Module[], node: FTLNode | null, navigate: NavigateFunction) => {
  if (node instanceof Module) {
    return modulePanels(modules, node, navigate)
  } else if (node instanceof Verb) {
    return verbPanels(node)
  } else if (node instanceof Secret) {
    return secretPanels(node)
  } else if (node instanceof Config) {
    return configPanels(node)
  } else {
    return [] as ExpandablePanelProps[]
  }
}

export default ConsolePage
