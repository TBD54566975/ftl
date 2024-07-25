import React, { useEffect, useRef, useState } from 'react'

interface ResizableVerticalPanelsProps {
  topPanelContent: React.ReactNode;
  bottomPanelContent: React.ReactNode;
  initialTopPanelHeightPercent?: number;
  minTopPanelHeight?: number;
  minBottomPanelHeight?: number;
}

export const ResizableVerticalPanels: React.FC<ResizableVerticalPanelsProps> = ({
  topPanelContent,
  bottomPanelContent,
  initialTopPanelHeightPercent = 50,
  minTopPanelHeight = 100,
  minBottomPanelHeight = 100,
}) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const [topPanelHeight, setTopPanelHeight] =  useState<number>()
  const [isDragging, setIsDragging] = useState(false)

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
  },[initialTopPanelHeightPercent])

  const startDragging = (e: React.MouseEvent<HTMLDivElement>) => {
    e.preventDefault()
    setIsDragging(true)
  }

  const stopDragging = () => {
    setIsDragging(false)
  }

  const onDrag = (e: React.MouseEvent<HTMLDivElement>) => {
    if (isDragging) {
      const newHeight = e.clientY
      const maxHeight = window.innerHeight - minBottomPanelHeight
      if (newHeight >= minTopPanelHeight && newHeight <= maxHeight) {
        setTopPanelHeight(newHeight - 44)
      }
    }
  }

  return (
    <div
      ref={containerRef}
      className='flex flex-col h-full w-full'
      onMouseMove={onDrag}
      onMouseUp={stopDragging}
      onMouseLeave={stopDragging}
    >
      <div style={{ height: `${topPanelHeight}px` }} className='overflow-auto'>
        {topPanelContent}
      </div>
      <div
        className='cursor-row-resize bg-gray-200 dark:bg-gray-700 hover:bg-indigo-600'
        onMouseDown={startDragging}
        style={{ height: '3px', cursor: 'row-resize' }}
      />
      <div className='flex-1 overflow-auto'>
        {bottomPanelContent}
      </div>
    </div>
  )
}
