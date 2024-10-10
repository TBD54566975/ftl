import type React from 'react'
import { useMemo, useState } from 'react'
import { useStreamModules } from '../../api/modules/use-stream-modules'
import { ResizableHorizontalPanels } from '../../components/ResizableHorizontalPanels'
import { ModulesTree } from './ModulesTree'
import { getTreeWidthFromLS, moduleTreeFromStream, setTreeWidthInLS } from './module.utils'

export const ModulesPage = ({ body }: { body: React.ReactNode }) => {
  const modules = useStreamModules()
  const tree = useMemo(() => moduleTreeFromStream(modules?.data || []), [modules?.data])
  const [treeWidth, setTreeWidth] = useState(getTreeWidthFromLS())

  function setTreeWidthWithLS(newWidth: number) {
    setTreeWidthInLS(newWidth)
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
