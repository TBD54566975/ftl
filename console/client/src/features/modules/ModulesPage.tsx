import React from 'react'
import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import { PageHeader } from '../../components/PageHeader'
import { modulesContext } from '../../providers/modules-provider'
import { generateMermaidMarkdown } from './generate-mermaid-diagram'

const defaultViewport = { x: 0, y: 0, zoom: 1.5 }
export const ModulesPage = () => {
  const modules = React.useContext(modulesContext)
  console.log(generateMermaidMarkdown(modules))
  return (
    <div className='h-full w-full flex flex-col'>
      <PageHeader icon={<Square3Stack3DIcon />} title='Modules' />
      <div className='flex-1 relative p-8'></div>
    </div>
  )
}
