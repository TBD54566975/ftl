import type React from 'react'
import { useMemo, useState } from 'react'
import { useSchema } from '../../api/schema/use-schema'
import { ResizableHorizontalPanels } from '../../components/ResizableHorizontalPanels'
import { DeploymentsPage } from '../deployments/DeploymentsPage'
import { ModulesTree } from './ModulesTree'
import { moduleTreeFromSchema } from './module.utils'

const treeWidthStorageKey = 'tree_w'

export const ModulesPanel = () => {
  return <DeploymentsPage />
}

export const ModulesPage = ({ body }: { body: React.ReactNode }) => {
  const schema = useSchema()
  const tree = useMemo(() => moduleTreeFromSchema(schema?.data || []), [schema?.data])
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
