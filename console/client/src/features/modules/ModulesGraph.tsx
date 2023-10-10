import { ArrowPathIcon, MinusCircleIcon, PlusCircleIcon } from '@heroicons/react/24/outline'
import React, { useContext, useEffect, useRef, useState } from 'react'
import { modulesContext } from '../../providers/modules-provider'
import { Panel } from './components'
import { dotToSVG, formatSVG, generateDot, svgZoom } from './graph'
import { VerbId, ZoomCallbacks } from './modules.constants'

export const ModulesGraph: React.FC<{
  className?: string
  zoomId?: string
  setSelectedVerbs: React.Dispatch<React.SetStateAction<VerbId[]>>
  selectedVerbs: VerbId[]
  setZoomCallbacks: React.Dispatch<React.SetStateAction<ZoomCallbacks | undefined>>
  zoomCallbacks?: ZoomCallbacks
}> = ({ className, setZoomCallbacks, zoomCallbacks }) => {
  const modules = useContext(modulesContext)
  const canvasRef = useRef<HTMLDivElement>(null)
  const [resizeCount, setResizeCount] = useState<number>(0)
  const previousDimensions = useRef({ width: 0, height: 0 }) // Store previous dimensions

  useEffect(() => {
    const canvasCur = canvasRef.current
    if (canvasCur) {
      const observer = new ResizeObserver((entries) => {
        for (const entry of entries) {
          const { width, height } = entry.contentRect
          // Check if dimensions have changed
          if (width !== previousDimensions.current.width || height !== previousDimensions.current.height) {
            setResizeCount((n) => n + 1)
            // Update previous dimensions
            previousDimensions.current = { width, height }
          }
        }
      })
      observer.observe(canvasCur)
      return () => {
        observer.disconnect()
      }
    }
  }, [canvasRef])

  useEffect(() => {
    const canvas = canvasRef.current
    const ready = canvas && Boolean(modules)
    let animationFrameId: number
    const renderSvg = async () => {
      const dot = generateDot(modules)
      const data = await dotToSVG(dot)
      if (data) {
        if (animationFrameId) {
          cancelAnimationFrame(animationFrameId)
        }
        animationFrameId = requestAnimationFrame(() => {
          const unformattedSVG = data
          const formattedSVG = formatSVG(unformattedSVG)
          canvas?.replaceChildren(formattedSVG)
          const zoom = svgZoom(formattedSVG, canvas?.clientWidth ?? 0, canvas?.clientHeight ?? 0)
          setZoomCallbacks(zoom)
        })
      }
    }
    ready && void renderSvg()
  }, [modules, resizeCount])

  return (
    <Panel className={className}>
      <Panel.Body className='overflow-hidden flex'>
        <div ref={canvasRef} className={'flex-1'} />
      </Panel.Body>
      <Panel.Header className='flex gap-0.5'>
        <button onClick={() => zoomCallbacks?.in()}>
          <span className='sr-only'>zoom in</span>
          <PlusCircleIcon className='w-6 h-6' />
        </button>
        <button onClick={() => zoomCallbacks?.out()}>
          <span className='sr-only'>zoom out</span>
          <MinusCircleIcon className='w-6 h-6' />
        </button>
        <button onClick={() => zoomCallbacks?.reset()}>
          <span className='sr-only'>reset</span>
          <ArrowPathIcon className='w-6 h-6' />
        </button>
      </Panel.Header>
    </Panel>
  )
}
