import type React from 'react'
import { useState } from 'react'
import type { ExpandablePanelProps } from '../features/console/ExpandablePanel'
import RightPanel from '../features/console/right-panel/RightPanel'
import useLocalStorage from '../hooks/use-local-storage'

interface ResizablePanelsProps {
  initialRightPanelWidth?: number
  initialBottomPanelHeight?: number
  minRightPanelWidth?: number
  minBottomPanelHeight?: number
  topBarHeight?: number
  rightPanelHeader: React.ReactNode
  rightPanelPanels: ExpandablePanelProps[]
  bottomPanelContent: React.ReactNode
  mainContent: React.ReactNode
}

export const ResizablePanels: React.FC<ResizablePanelsProps> = ({
  initialRightPanelWidth = 300,
  initialBottomPanelHeight = 200,
  minRightPanelWidth = 200,
  minBottomPanelHeight = 200,
  rightPanelHeader,
  rightPanelPanels,
  bottomPanelContent,
  mainContent,
}) => {
  const [rightPanelWidth, setRightPanelWidth] = useLocalStorage<number>('rightPanelWidth', initialRightPanelWidth)
  const [bottomPanelHeight, setBottomPanelHeight] = useLocalStorage<number>('bottomPanelHeight', initialBottomPanelHeight)

  const [isDraggingHorizontal, setIsDraggingHorizontal] = useState(false)
  const [isDraggingVertical, setIsDraggingVertical] = useState(false)

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
      const newWidth = Math.max(window.innerWidth - e.clientX, minRightPanelWidth)
      setRightPanelWidth(newWidth)
    }
  }

  const onDragVertical = (e: React.MouseEvent<HTMLDivElement>) => {
    if (isDraggingVertical) {
      const newHeight = Math.max(window.innerHeight - e.clientY, minBottomPanelHeight)
      setBottomPanelHeight(newHeight)
    }
  }

  return (
    <div
      className='flex h-full w-full flex-col'
      onMouseMove={(e) => {
        if (isDraggingHorizontal) onDragHorizontal(e)
        if (isDraggingVertical) onDragVertical(e)
      }}
      onMouseUp={stopDragging}
      onMouseLeave={stopDragging}
    >
      <div className='flex flex-1'>
        <div style={{ maxHeight: `calc(100vh - ${bottomPanelHeight + 46}px)` }} className='flex-1 flex-col min-h-64'>
          {mainContent}
        </div>
        <div
          className='cursor-col-resize bg-gray-200 dark:bg-gray-700 hover:bg-indigo-600'
          onMouseDown={startDraggingHorizontal}
          style={{ width: '3px', cursor: 'col-resize' }}
        />
        <div
          style={{ width: `${rightPanelWidth}px`, maxHeight: `calc(100vh - ${bottomPanelHeight + 46}px)` }}
          className='flex flex-col h-full overflow-y-scroll'
        >
          <RightPanel header={rightPanelHeader} panels={rightPanelPanels} />
        </div>
      </div>
      <div
        className='cursor-row-resize bg-gray-200 dark:bg-gray-700 hover:bg-indigo-600'
        onMouseDown={startDraggingVertical}
        style={{ height: '3px', cursor: 'row-resize' }}
      />
      <div style={{ height: `${bottomPanelHeight}px` }} className='overflow-auto'>
        {bottomPanelContent}
      </div>
    </div>
  )
}
