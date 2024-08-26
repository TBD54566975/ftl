import type React from 'react'
import { useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'
import { ResizableHorizontalPanels } from '../../components/ResizableHorizontalPanels'
import { useSchema } from '../../api/schema/use-schema'
import { ModulesTree } from './ModulesTree'
import { moduleTreeFromSchema } from './module.utils'

export const ModulesPanel = () => {
  return (
    <div className='flex-1 py-2 px-4'>
      <p>Content</p>
    </div>
  )
}

export const ModulesPage = ({ body }: { body: React.ReactNode }) => {
  const schema = useSchema()
  const tree = useMemo(() => moduleTreeFromSchema(schema?.data || []), [schema?.data])
  const [searchParams, setSearchParams] = useSearchParams()

  function setTreeWidthWithParams(newWidth: number) {
    searchParams.set('tree_w', `${newWidth}`)
    setSearchParams(searchParams)
  }

  return (
    <ResizableHorizontalPanels
      leftPanelContent={<ModulesTree modules={tree} />}
      rightPanelContent={body}
      leftPanelWidth={Number(searchParams.get('tree_w')) || 300}
      setLeftPanelWidth={setTreeWidthWithParams}
    />
  )
}
