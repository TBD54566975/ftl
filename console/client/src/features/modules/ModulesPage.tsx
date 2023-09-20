import React from 'react'
import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import { PageHeader } from '../../components/PageHeader'
import { modulesContext } from '../../providers/modules-provider'
import { generateDot } from './generate-dot'
import { dotToSVG } from './dot-to-svg'
import { formatSVG } from './format-svg'
export const ModulesPage = () => {
  const modules = React.useContext(modulesContext)
  const dot = generateDot(modules)
  const ref = React.useRef<HTMLDivElement>(null)
  const [viewport, setViewPort] = React.useState<HTMLDivElement>()
  React.useEffect(() => {
    const cur = ref.current
    cur && setViewPort(cur)
  }, [])
  React.useEffect(() => {
    const renderSvg = async () => {
      const svg = await dotToSVG(dot)
      svg && viewport?.replaceChildren(formatSVG(svg))
    }
    viewport && void renderSvg()
  }, [dot, viewport])
  // console.log(generateDotFile(modules))
  return (
    <div className='h-full w-full flex flex-col'>
      <PageHeader icon={<Square3Stack3DIcon />} title='Modules' />
      <div ref={ref} className='viewport' />
    </div>
  )
}
