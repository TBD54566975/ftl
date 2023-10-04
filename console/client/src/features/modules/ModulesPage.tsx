import React from 'react'
import { ModulesSidebar } from './ModulesSidebar'
import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import { ModulesGraph } from './ModulesGraph'
import { ModulesRequests } from './ModulesRequests'
import { ModulesSelectedVerbs } from './ModulesSelectedVerbs'
import { modulesContext } from '../../providers/modules-provider'
import { classNames } from '../../utils'
import { VerbId } from './modules.constants'
import { Page } from '../../layout'
import type { ZoomCallbacks } from './modules.constants'

import styles from './ModulesPage.module.css'

export const ModulesPage = () => {
  const { modules } = React.useContext(modulesContext)
  const [selectedVerbs, setSelectedVerbs] = React.useState<VerbId[]>([])
  const hasVerbs = Boolean(selectedVerbs.length)
  const [zoomCallbacks, setZoomCallbacks] = React.useState<ZoomCallbacks>()

  return (
    <Page>
      <Page.Header icon={<Square3Stack3DIcon />} title='Modules' />
      <Page.Body className='gap-2 py-2 flex'>
        <ModulesSidebar
          className={`flex-none w-72`}
          modules={modules}
          setSelectedVerbs={setSelectedVerbs}
          selectedVerbs={selectedVerbs}
          zoomCallbacks={zoomCallbacks}
        />
        <div className={classNames('flex-1', styles.page, styles.template, hasVerbs && styles.templateSelectedVerb)}>
          <ModulesGraph
            className={classNames(
              styles.graph,
              styles.panel,
              hasVerbs && 'border border-gray-300 dark:border-slate-700',
            )}
            setSelectedVerbs={setSelectedVerbs}
            selectedVerbs={selectedVerbs}
            setZoomCallbacks={setZoomCallbacks}
            zoomCallbacks={zoomCallbacks}
          />
          {/* {selectedVerbs && <ModulesRequests
            className={styles.requests}
            modules={modules}
            selectedVerbs={selectedVerbs}
            />} */}
          {hasVerbs && (
            <ModulesSelectedVerbs
              className={classNames(
                styles.verbs,
                styles.panel,
                hasVerbs && 'border border-gray-300 dark:border-slate-700',
              )}
              modules={modules}
              selectedVerbs={selectedVerbs}
            />
          )}
        </div>
      </Page.Body>
    </Page>
  )
}
