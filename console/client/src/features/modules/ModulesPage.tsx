import React from 'react'
import { ModulesSidebar } from './ModulesSidebar'
import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import { ModulesGraph } from './graph/ModulesGraph'
import { ModulesRequests } from './ModulesRequests'
import { ModulesSchema } from './ModulesSchema'
import { ModulesTestCalls } from './ModulesTestCalls'
import { modulesContext } from '../../providers/modules-provider'
import { classNames } from '../../utils'
import { VerbId } from './modules.constants'
import { Page } from '../../layout'
import styles from  './ModulesPage.module.css'


export const ModulesPage = () => {
  const { modules } = React.useContext(modulesContext)
  const [zoomId, setZoomId] = React.useState<string>()
  const [selectedVerbs, setSelectedVerbs] = React.useState<VerbId[]>([])
  const hasVerbs = Boolean(selectedVerbs.length)
  return (
    <Page>
      <Page.Header icon={<Square3Stack3DIcon />} title='Modules'/>
      <Page.Body className='gap-2 py-2 flex'>
        <ModulesSidebar
          className={`flex-none w-72`}
          modules={modules}
          setSelectedVerbs={setSelectedVerbs}
          selectedVerbs={selectedVerbs}
          setZoomId={setZoomId}
        />
        <div className={classNames(
          'flex-1',
          styles.page,
          styles.template,
          hasVerbs && styles.templateSelectedVerb,
          )}>
          <ModulesGraph
            className={classNames(styles.graph, styles.panel, hasVerbs && 'border border-gray-300 dark:border-slate-700')}
            zoomId={zoomId}
            setSelectedVerbs={setSelectedVerbs}
            selectedVerbs={selectedVerbs}
          />
          {hasVerbs && <ModulesSchema
            className={classNames(styles.schema, styles.panel, hasVerbs && 'border border-gray-300 dark:border-slate-700')}
            modules={modules}
            selectedVerbs={selectedVerbs}
            />}
          {/* {selectedVerbs && <ModulesRequests
            className={styles.requests}
            modules={modules}
            selectedVerbs={selectedVerbs}
            />} */}
          {hasVerbs && <ModulesTestCalls
            className={classNames(styles.call, styles.panel, hasVerbs && 'border border-gray-300 dark:border-slate-700')}
            modules={modules}
            selectedVerbs={selectedVerbs}
            />}
        </div>
      </Page.Body>
    </Page>
  )
}
