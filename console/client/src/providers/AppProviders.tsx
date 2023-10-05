import { App } from '../App'
import { DarkModeProvider } from './dark-mode-provider'
import { ModulesProvider } from './modules-provider'
import { NotificationsProvider } from './notifications-provider'
import { SchemaProvider } from './schema-provider'

export const AppProviders = () => {
  return (
    <DarkModeProvider>
      <SchemaProvider>
        <ModulesProvider>
          <NotificationsProvider>
            <App />
          </NotificationsProvider>
        </ModulesProvider>
      </SchemaProvider>
    </DarkModeProvider>
  )
}
