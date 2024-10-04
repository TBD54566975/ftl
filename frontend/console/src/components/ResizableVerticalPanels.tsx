import type React from 'react'
import { useEffect, useRef, useState } from 'react'

interface ResizableVerticalPanelsProps {
  topPanelContent: React.ReactNode
  bottomPanelContent: React.ReactNode
  initialTopPanelHeightPercent?: number
  minTopPanelHeight?: number
  minBottomPanelHeight?: number
}

export const ResizableVerticalPanels: React.FC<ResizableVerticalPanelsProps> = ({
  topPanelContent,
  bottomPanelContent,
  initialTopPanelHeightPercent = 50,
  minTopPanelHeight = 100,
  minBottomPanelHeight = 100,
}) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const [topPanelHeight, setTopPanelHeight] = useState<number>()
  const [isDragging, setIsDragging] = useState(false)

  const hasBottomPanel = !!bottomPanelContent

  useEffect(() => {
    const updateDimensions = () => {
      if (containerRef.current) {
        const parentHeight = containerRef.current.getBoundingClientRect().height
        const initialHeight = parentHeight * (initialTopPanelHeightPercent / 100)
        setTopPanelHeight(initialHeight)
      }
    }

    updateDimensions()
    window.addEventListener('resize', updateDimensions)
    return () => window.removeEventListener('resize', updateDimensions)
  }, [initialTopPanelHeightPercent, hasBottomPanel])

  const startDragging = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!hasBottomPanel) {
      return
    }
    e.preventDefault()
    setIsDragging(true)
  }

  const stopDragging = () => {
    setIsDragging(false)
  }

  const onDrag = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!isDragging || !containerRef.current || !hasBottomPanel) {
      return
    }
    const containerDims = containerRef.current.getBoundingClientRect()
    const newHeight = e.clientY - containerDims.top
    const maxHeight = containerDims.height - minBottomPanelHeight
    if (newHeight >= minTopPanelHeight && newHeight <= maxHeight) {
      setTopPanelHeight(newHeight)
    }
  }

  return (
    <div ref={containerRef} className='flex flex-col h-full w-full' onMouseMove={onDrag} onMouseUp={stopDragging} onMouseLeave={stopDragging}>
      <div style={{ height: hasBottomPanel ? `${topPanelHeight}px` : '100%' }} className='overflow-auto'>
        {' '}
        {topPanelContent}
      </div>
      {hasBottomPanel && (
        <>
          <div
            className='cursor-row-resize bg-gray-200 dark:bg-gray-700 hover:bg-indigo-600'
            onMouseDown={startDragging}
            style={{ height: '3px', cursor: 'row-resize' }}
          />
          <div className='flex-1 overflow-auto'>{bottomPanelContent}</div>
        </>
      )}
    </div>
  )
}
