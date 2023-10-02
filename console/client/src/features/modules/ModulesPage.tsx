import React from 'react'
import { ModulesSidebar } from './ModulesSidebar'
import { PageHeader } from '../../components'
import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import { ModulesGraph } from './graph/ModulesGraph'
import { ModulesRequests } from './ModulesRequests'
import { ModulesTimeline } from './ModulesTimeline'
import { ModulesSchema } from './ModulesSchema'
import { ModulesTestCalls } from './ModulesTestCalls'
import { modulesContext } from '../../providers/modules-provider'
import { classNames } from '../../utils'
import { VerbId } from './modules.constants'

import styles from  './ModulesPage.module.css'

export const ModulesPage = () => {
  const { modules } = React.useContext(modulesContext)
  const [zoomId, setZoomId] = React.useState<string>()
  const [selectedVerbs, setSelectedVerbs] = React.useState<VerbId[]>([])
  const [selectedModules, setSelectedModules] = React.useState<string[]>([])
  const [hoveredEdge, setHoveredEdge] = React.useState<string>()

  return (
    <div className={classNames(
      styles.page,
      styles.template,
      selectedVerbs.length && styles.templateSelectedVerb,
    )}>
      <PageHeader className={styles.header} icon={<Square3Stack3DIcon />} title='Modules'/>
      <ModulesSidebar
        className={styles.sidebar}
        modules={modules}
        setSelectedVerbs={setSelectedVerbs}
        selectedVerbs={selectedVerbs}
        setSelectedModules={setSelectedModules}
        selectedModules={selectedModules}
        setZoomId={setZoomId}
      />
      <ModulesGraph className={styles.graph}/>
      {selectedVerbs && <ModulesSchema
        className={styles.schema}
        modules={modules}
        selectedVerbs={selectedVerbs}
        />}
      {selectedVerbs && <ModulesRequests
        className={styles.schema}
        modules={modules}
        selectedVerbs={selectedVerbs}
        />}
      {selectedVerbs && <ModulesTestCalls
        className={styles.call}
        modules={modules}
        selectedVerbs={selectedVerbs}
        />}
      {selectedVerbs && <ModulesTimeline className={styles.timeline} />}
    </div>
  )
}
