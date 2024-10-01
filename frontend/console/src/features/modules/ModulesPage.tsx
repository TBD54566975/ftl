import type React from 'react'
import { useMemo, useState } from 'react'
import { useStreamModules } from '../../api/modules/use-stream-modules'
import { ResizableHorizontalPanels } from '../../components/ResizableHorizontalPanels'
import { ModulesTree } from './ModulesTree'
import { moduleTreeFromStream } from './module.utils'

const treeWidthStorageKey = 'tree_w'

export const ModulesPage = ({ body }: { body: React.ReactNode }) => {
  const modules = useStreamModules()
  const tree = useMemo(() => moduleTreeFromStream(modules?.data || []), [modules?.data])
  const [treeWidth, setTreeWidth] = useState(Number(localStorage.getItem(treeWidthStorageKey)) || 300)

  function setTreeWidthWithLS(newWidth: number) {
    localStorage.setItem(treeWidthStorageKey, `${newWidth}`)
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
