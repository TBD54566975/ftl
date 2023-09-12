import { App } from '../App'
import { DarkModeProvider } from './dark-mode-provider'
import { ModulesProvider } from './modules-provider'
import { NotificationsProvider } from './notifications-provider'
import { SchemaProvider } from './schema-provider'
import { SelectedModuleProvider } from './selected-module-provider'
import { SelectedTimelineEntryProvider } from './selected-timeline-entry-provider'
import { SidePanelProvider } from './side-panel-provider'
import { TabsProvider } from './tabs-provider'

export const AppProviders = () => {
  return (
    <DarkModeProvider>
      <SchemaProvider>
        <ModulesProvider>
          <SelectedModuleProvider>
            <SelectedTimelineEntryProvider>
              <TabsProvider>
                <SidePanelProvider>
                  <NotificationsProvider>
                    <App />
                  </NotificationsProvider>
                </SidePanelProvider>
              </TabsProvider>
            </SelectedTimelineEntryProvider>
          </SelectedModuleProvider>
        </ModulesProvider>
      </SchemaProvider>
    </DarkModeProvider>
  )
}
