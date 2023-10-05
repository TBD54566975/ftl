import { Square3Stack3DIcon } from '@heroicons/react/24/outline'
import React from 'react'
import { Page } from '../../layout'
import { modulesContext } from '../../providers/modules-provider'
import { SidePanelProvider } from '../../providers/side-panel-provider'
import { classNames } from '../../utils'
import { ModulesGraph } from './ModulesGraph'
import { ModulesRequests } from './ModulesRequests'
import { ModulesSelectedVerbs } from './ModulesSelectedVerbs'
import { ModulesSidebar } from './ModulesSidebar'
import type { ZoomCallbacks } from './modules.constants'
import { VerbId } from './modules.constants'

export const ModulesPage = () => {
  const { modules } = React.useContext(modulesContext)
  const [selectedVerbs, setSelectedVerbs] = React.useState<VerbId[]>([])
  const hasVerbs = Boolean(selectedVerbs.length)
  const [zoomCallbacks, setZoomCallbacks] = React.useState<ZoomCallbacks>()

  return (
    <SidePanelProvider>
      <Page>
        <Page.Header icon={<Square3Stack3DIcon />} title='Modules' />
        <Page.Body className='gap-2 p-2 flex'>
          <ModulesSidebar
            className={`flex-none w-72`}
            modules={modules}
            setSelectedVerbs={setSelectedVerbs}
            selectedVerbs={selectedVerbs}
            zoomCallbacks={zoomCallbacks}
          />
          <div
            className={classNames(
              'flex-grow grid gap-2',
              !hasVerbs && 'grid grid-rows-2 grid-cols-fr',
              hasVerbs && 'grid-cols-2 grid-rows-2',
            )}
          >
            <ModulesGraph
              setSelectedVerbs={setSelectedVerbs}
              selectedVerbs={selectedVerbs}
              setZoomCallbacks={setZoomCallbacks}
              zoomCallbacks={zoomCallbacks}
              className={classNames(hasVerbs && 'row-start-1 row-span-1 col-start-1 col-span-1')}
            />
            <ModulesRequests
              modules={modules}
              className={classNames(hasVerbs && 'row-start-2 row-span-1 col-start-1 col-span-1')}
            />
            {hasVerbs && (
              <ModulesSelectedVerbs
                modules={modules}
                selectedVerbs={selectedVerbs}
                className='row-start-1 row-span-2 col-start-2 col-span-1'
              />
            )}
          </div>
        </Page.Body>
      </Page>
    </SidePanelProvider>
  )
}
