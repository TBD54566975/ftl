import React from 'react'
import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import { PageHeader } from '../../components/PageHeader'
import { modulesContext } from '../../providers/modules-provider'
import { generateDot } from './generate-dot'
import { dot2Svg } from './create-svg'

export const ModulesPage = () => {
  const modules = React.useContext(modulesContext)
  const dot = generateDot(modules)
  React.useEffect(() => {
    const renderSvg = async () => {
      const svg = await dot2Svg(dot)
    }
    void renderSvg()
  }, [dot])
  // console.log(generateDotFile(modules))
  return (
    <div className='h-full w-full flex flex-col'>
      <PageHeader icon={<Square3Stack3DIcon />} title='Modules' />
      <div className='flex-1 relative p-8'></div>
    </div>
  )
}
