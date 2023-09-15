import { App } from '../App'
import { DarkModeProvider } from './dark-mode-provider'
import { ModulesProvider } from './modules-provider'
import { NotificationsProvider } from './notifications-provider'
import { SchemaProvider } from './schema-provider'
import { SelectedModuleProvider } from './selected-module-provider'
import { SelectedEventProvider } from './selected-timeline-entry-provider'
import { SidePanelProvider } from './side-panel-provider'

export const AppProviders = () => {
  return (
    <DarkModeProvider>
      <SchemaProvider>
        <ModulesProvider>
          <SelectedModuleProvider>
            <SelectedEventProvider>
              <SidePanelProvider>
                <NotificationsProvider>
                  <App />
                </NotificationsProvider>
              </SidePanelProvider>
            </SelectedEventProvider>
          </SelectedModuleProvider>
        </ModulesProvider>
      </SchemaProvider>
    </DarkModeProvider>
  )
}
