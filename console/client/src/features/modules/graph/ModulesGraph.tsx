import React from 'react'
import { PlusCircleIcon, MinusCircleIcon, ArrowPathIcon } from '@heroicons/react/24/outline'
import { modulesContext } from '../../../providers/modules-provider'
import { generateDot } from './generate-dot'
import { dotToSVG } from './dot-to-svg'
import { formatSVG } from './format-svg'
import { VerbId } from '../modules.constants'
import { Panel } from '../components'
import { svgZoom } from './svg-zoom'

import './graph.css'
export const ModulesGraph: React.FC<{
  className: string
  zoomId?: string
  setSelectedVerbs: React.Dispatch<React.SetStateAction<VerbId[]>>
  selectedVerbs: VerbId[]
}> = ({ className }) => {
  const modules = React.useContext(modulesContext)
  const canvasRef = React.useRef<HTMLDivElement>(null)
  const [canvas, setViewPort] = React.useState<HTMLDivElement>()
  const [svgRatioStyles, setSvgRatioStyles] = React.useState<React.CSSProperties>()
  const [zoomCallbacks, setZoomCallback] = React.useState<ReturnType<typeof svgZoom>>()
  React.useEffect(() => {
    const viewCur = canvasRef.current
    viewCur && setViewPort(viewCur)
  }, [])

  React.useEffect(() => {
    const renderSvg = async () => {
      const dot = generateDot(modules)
      const data = await dotToSVG(dot)
      if (data && canvas) {
        const [unformattedSVG, aspectRatio] = data
        const svgRatioStyles = aspectRatio >= 1 ? { width: '100%', aspectRatio } : { height: '100%', aspectRatio }
        setSvgRatioStyles(svgRatioStyles)
        const formattedSVG = formatSVG(unformattedSVG)
        const { width, height } = canvas.getBoundingClientRect()
        canvas?.replaceChildren(formattedSVG)
        const zoom = svgZoom([0, 0, width, height])
        setZoomCallback(zoom)
      }
    }
    canvas && void renderSvg()
  }, [modules, canvas])

  return (
    <Panel className={className}>
      <Panel.Header className='flex gap-2 justify-between'>
        Graph
        <div className='inline-flex gap-0.5'>
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
        </div>
      </Panel.Header>
      <Panel.Body>
        <div ref={canvasRef} className={`canvas flex-1 overflow-hidden h-full w-full`} />
      </Panel.Body>
    </Panel>
  )
}
