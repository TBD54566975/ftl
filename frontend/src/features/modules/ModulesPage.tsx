import type React from 'react'
import { useMemo, useState } from 'react'
import { useSchema } from '../../api/schema/use-schema'
import { ResizableHorizontalPanels } from '../../components/ResizableHorizontalPanels'
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
  const [treeWidth, setTreeWidth] = useState(Number(localStorage.getItem('tree_w')) || 300)

  function setTreeWidthWithLS(newWidth: number) {
    localStorage.setItem('tree_w', `${newWidth}`)
    setTreeWidth(newWidth)
  }

  return (
    <ResizableHorizontalPanels
      leftPanelContent={<ModulesTree modules={tree} />}
      rightPanelContent={body}
      leftPanelWidth={treeWidth}
      setLeftPanelWidth={setTreeWidthWithLS}
    />
  )
}
