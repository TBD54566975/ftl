import type React from 'react'
import { useRef, useState } from 'react'

interface ResizableHorizontalPanelsProps {
  leftPanelContent: React.ReactNode
  rightPanelContent: React.ReactNode
  minLeftPanelWidth?: number
  minRightPanelWidth?: number
  leftPanelWidth: number
  setLeftPanelWidth: (n: number) => void
}

export const ResizableHorizontalPanels: React.FC<ResizableHorizontalPanelsProps> = ({
  leftPanelContent,
  rightPanelContent,
  minLeftPanelWidth = 100,
  minRightPanelWidth = 100,
  leftPanelWidth,
  setLeftPanelWidth,
}) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const [isDragging, setIsDragging] = useState(false)

  const startDragging = (e: React.MouseEvent<HTMLDivElement>) => {
    e.preventDefault()
    setIsDragging(true)
  }

  const stopDragging = () => setIsDragging(false)

  const onDrag = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!isDragging || !containerRef.current) {
      return
    }
    const containerDims = containerRef.current.getBoundingClientRect()
    const newWidth = e.clientX - containerDims.x
    const maxWidth = containerDims.width - minRightPanelWidth
    if (newWidth >= minLeftPanelWidth && newWidth <= maxWidth) {
      setLeftPanelWidth(newWidth)
    }
  }

  return (
    <div ref={containerRef} className='flex flex-row h-full w-full' onMouseMove={onDrag} onMouseUp={stopDragging} onMouseLeave={stopDragging}>
      <div style={{ width: `${leftPanelWidth}px` }} className='overflow-auto'>
        {leftPanelContent}
      </div>
      <div
        className='cursor-col-resize hover:bg-indigo-600 hover:dark:bg-indigo-600'
        onMouseDown={startDragging}
        style={{ width: '3px', cursor: 'col-resize' }}
      />
      <div className='flex-1 overflow-auto'>{rightPanelContent}</div>
    </div>
  )
}
