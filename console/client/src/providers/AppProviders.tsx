import App from '../App'
import ModulesProvider from './modules-provider'
import SchemaProvider from './schema-provider'
import { SelectedModuleProvider } from './selected-module-provider'
import { SelectedTimelineEntryProvider } from './selected-timeline-entry-provider'
import { SidePanelProvider } from './side-panel-provider'
import { TabsProvider } from './tabs-provider'

export function AppProviders() {
  return (
    <SchemaProvider>
      <ModulesProvider>
        <SelectedModuleProvider>
          <SelectedTimelineEntryProvider>
            <TabsProvider>
              <SidePanelProvider>
                <App />
              </SidePanelProvider>
            </TabsProvider>
          </SelectedTimelineEntryProvider>
        </SelectedModuleProvider>
      </ModulesProvider>
    </SchemaProvider>
  )
}
