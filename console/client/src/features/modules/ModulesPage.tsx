import React from 'react'
import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import { PageHeader } from '../../components/PageHeader'
import { ModulesSidebar } from './sidebar/ModulesSidebar'
import styles from  './ModulesPage.module.css'
import { ModulesUI } from './ui'
import { ModulesGraph } from './graph/ModulesGraph'

export const ModulesPage = () => {
  const [zoomID, setZoomID] = React.useState<`#${string}`>()
  const [selectedVerb, setSelectedVerb] = React.useState<`${string}.${string}`>()
  const [selectedModule, setSelectedModule] = React.useState<string>()
  const [selectedEdges, setSelectedEdges] = React.useState<`#${string}`[]>()
  return (
    <div className={styles.page}>
      <PageHeader icon={<Square3Stack3DIcon />} title='Modules' className={styles.header}/>
      <ModulesSidebar
        className={styles.sidebar}
        setZoomID={setZoomID}
        setSelectedEdges={setSelectedEdges}
        setSelectedVerb={setSelectedVerb}
        setSelectedModule={setSelectedModule}
      />
      <ModulesGraph className={styles.graph}/>
      <ModulesUI
        className={styles.ui}
        withSidebarCls={styles.uiWithSidebar}
        withoutSidebarCls={styles.uiWithoutSidebar}
        controlBarCls={styles.controlBar}
        sidebarCls={styles.uiSidebar}
      />
    </div>
  )
}
