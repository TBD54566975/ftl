import { useState } from 'react'

export const Layout = () => {
  const [leftPaneWidth, setLeftPaneWidth] = useState(200)
  const [rightPaneWidth, setRightPaneWidth] = useState(200)
  const [bottomPaneHeight, setBottomPaneHeight] = useState(100)

  const handleMouseDown = (e: React.MouseEvent, direction: string) => {
    const startX = e.clientX
    const startY = e.clientY
    const startLeftPaneWidth = leftPaneWidth
    const startRightPaneWidth = rightPaneWidth
    const startBottomPaneHeight = bottomPaneHeight

    const handleMouseMove = (e: MouseEvent) => {
      if (direction === 'left') {
        const newLeftPaneWidth = startLeftPaneWidth + e.clientX - startX
        setLeftPaneWidth(newLeftPaneWidth > 100 ? newLeftPaneWidth : 100)
      } else if (direction === 'right') {
        const newRightPaneWidth = startRightPaneWidth - (e.clientX - startX)
        setRightPaneWidth(newRightPaneWidth > 100 ? newRightPaneWidth : 100)
      } else if (direction === 'bottom') {
        const newBottomPaneHeight = startBottomPaneHeight - (e.clientY - startY)
        setBottomPaneHeight(newBottomPaneHeight > 50 ? newBottomPaneHeight : 50)
      }
    }

    const handleMouseUp = () => {
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
    }

    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseup', handleMouseUp)
  }

  return (
    <div className='h-screen flex flex-col'>
      <div className='flex flex-1'>
        <div className='w-16 bg-gray-800 text-gray-200'>
          {/* Activity Bar */}
        </div>
        <div className='flex flex-1 relative'>
          <div className='bg-gray-900 text-gray-200 relative' style={{ width: leftPaneWidth }}>
            {/* Left Pane */}
            <div
              className='absolute top-0 right-0 h-full w-1 cursor-ew-resize bg-gray-600'
              onMouseDown={(e) => handleMouseDown(e, 'left')}
            />
          </div>
          <div className='flex flex-1 flex-col relative'>
            <div className='flex flex-1'>
              <div className='flex flex-1 bg-gray-700 text-gray-200'>
                {/* Main Editor Area */}
              </div>
              <div className='bg-gray-900 text-gray-200 relative' style={{ width: rightPaneWidth }}>
                {/* Right Pane */}
                <div
                  className='absolute top-0 left-0 h-full w-1 cursor-ew-resize bg-gray-600'
                  onMouseDown={(e) => handleMouseDown(e, 'right')}
                />
              </div>
            </div>
            <div className='bg-gray-800 text-gray-200 relative' style={{ height: bottomPaneHeight }}>
              {/* Bottom Pane */}
              <div
                className='absolute left-0 right-0 top-0 h-1 cursor-ns-resize bg-gray-600'
                onMouseDown={(e) => handleMouseDown(e, 'bottom')}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
