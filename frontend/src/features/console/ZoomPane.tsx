import React, { useState } from 'react'

interface ZoomPaneProps {
  children: React.ReactNode // This allows any React content to be passed in
}

const ZoomPane: React.FC<ZoomPaneProps> = ({ children }) => {
  const [viewBox, setViewBox] = useState({ x: 0, y: 0, width: 800, height: 600 })
  const [isPanning, setIsPanning] = useState(false)
  const [startPan, setStartPan] = useState({ x: 0, y: 0 })

  const handleWheel = (event: React.WheelEvent) => {
    event.preventDefault()
    const scale = event.deltaY < 0 ? 1.1 : 0.9 // Adjust these values for zoom speed
    const factor = event.deltaY < 0 ? scale : 1 / scale

    // Compute the new viewBox dimensions
    const newWidth = Math.max(100, Math.min(800, viewBox.width * factor)) // Min and max width limits
    const newHeight = Math.max(100, Math.min(600, viewBox.height * factor)) // Min and max height limits
    const newX = viewBox.x - (newWidth - viewBox.width) / 2
    const newY = viewBox.y - (newHeight - viewBox.height) / 2

    setViewBox({ x: newX, y: newY, width: newWidth, height: newHeight })
  }

  const handleMouseDown = (event: React.MouseEvent) => {
    setIsPanning(true)
    setStartPan({ x: event.clientX, y: event.clientY })
  }

  const handleMouseMove = (event: React.MouseEvent) => {
    if (isPanning) {
      const dx = ((event.clientX - startPan.x) * viewBox.width) / window.innerWidth
      const dy = ((event.clientY - startPan.y) * viewBox.height) / window.innerHeight

      setViewBox((prevViewBox) => ({
        ...prevViewBox,
        x: prevViewBox.x - dx,
        y: prevViewBox.y - dy,
      }))

      setStartPan({ x: event.clientX, y: event.clientY })
    }
  }

  const handleMouseUp = () => {
    setIsPanning(false)
  }

  return (
    <div
      onMouseDown={handleMouseDown}
      onMouseMove={handleMouseMove}
      onMouseUp={handleMouseUp}
      onMouseLeave={handleMouseUp}
      onWheel={handleWheel}
      style={{ cursor: isPanning ? 'grabbing' : 'grab', width: '100%', height: '100%', overflow: 'hidden' }}
    >
      <svg
        width='100%'
        height='100%'
        viewBox={`${viewBox.x} ${viewBox.y} ${viewBox.width} ${viewBox.height}`}
        xmlns='http://www.w3.org/2000/svg'
        style={{ display: 'block' }}
      >
        {children}
      </svg>
    </div>
  )
}

export default ZoomPane
