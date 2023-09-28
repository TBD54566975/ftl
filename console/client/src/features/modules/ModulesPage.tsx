import React from 'react'
import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import { PageHeader } from '../../components/PageHeader'
import { modulesContext } from '../../providers/modules-provider'
import { generateDot } from './generate-dot'
import { dotToSVG } from './dot-to-svg'
import { formatSVG } from './format-svg'
import { svgZoom } from './svg-zoom'
import { createControls } from './create-controls'
import './Modules.css'

export const ModulesPage = () => {
  const modules = React.useContext(modulesContext)
  const viewportRef = React.useRef<HTMLDivElement>(null)
  const controlRef = React.useRef<HTMLDivElement>(null)
  const [viewport, setViewPort] = React.useState<HTMLDivElement>()
  const [controls, setControls] = React.useState<HTMLDivElement>()
  const [svg, setSVG] = React.useState<SVGSVGElement>()

  React.useEffect(() => {
    const viewCur = viewportRef.current
    viewCur && setViewPort(viewCur)

    const ctlCur = controlRef.current
    ctlCur && setControls(ctlCur)
  }, [])

  React.useEffect(() => {
    const renderSvg = async () => {
      const dot = generateDot(modules)
      const unformattedSVG = await dotToSVG(dot)
      if (unformattedSVG) {
        const formattedSVG = formatSVG(unformattedSVG)
        viewport?.replaceChildren(formattedSVG)
        setSVG(formattedSVG)
      }
    }
    viewport && void renderSvg()
  }, [modules, viewport])

  React.useEffect(() => {
    if (controls && svg) {
      const zoom = svgZoom()
      const [buttons, removeListeners] = createControls(zoom)
      controls.replaceChildren(...buttons.values())
      return () => {
        removeListeners()
      }
    }
  }, [controls, svg])
  return (
    <div className='h-full w-full flex flex-col'>
      <PageHeader icon={<Square3Stack3DIcon />} title='Modules' />
      <div ref={controlRef} className='zoom-pan-controls'></div>
      <div ref={viewportRef} className='viewport flex-1 overflow-hidden' />
    </div>
  )
}
