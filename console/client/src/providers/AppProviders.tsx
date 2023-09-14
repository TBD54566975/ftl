import { ThemeProvider } from '@mui/material'
import { App } from '../App'
import { createTheme } from '../theme'
import { DarkModeProvider } from './dark-mode-provider'
import { ModulesProvider } from './modules-provider'
import { NotificationsProvider } from './notifications-provider'
import { SchemaProvider } from './schema-provider'
import { SelectedModuleProvider } from './selected-module-provider'
import { SelectedTimelineEntryProvider } from './selected-timeline-entry-provider'
import { SidePanelProvider } from './side-panel-provider'

export const AppProviders = () => {
  const theme = createTheme()
  return (
    <ThemeProvider theme={theme}>
      <DarkModeProvider>
        <SchemaProvider>
          <ModulesProvider>
            <SelectedModuleProvider>
              <SelectedTimelineEntryProvider>
                <SidePanelProvider>
                  <NotificationsProvider>
                    <App />
                  </NotificationsProvider>
                </SidePanelProvider>
              </SelectedTimelineEntryProvider>
            </SelectedModuleProvider>
          </ModulesProvider>
        </SchemaProvider>
      </DarkModeProvider>
    </ThemeProvider>
  )
}
