import { App } from '../App'
import { DarkModeProvider } from './dark-mode-provider'
import { ModulesProvider } from './modules-provider'
import { NotificationsProvider } from './notifications-provider'

export const AppProviders = () => {
  return (
    <DarkModeProvider>
      <ModulesProvider>
        <NotificationsProvider>
          <App />
        </NotificationsProvider>
      </ModulesProvider>
    </DarkModeProvider>
  )
}
