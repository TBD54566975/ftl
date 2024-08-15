import { ModulesProvider } from './modules-provider'
import { NotificationsProvider } from './notifications-provider'
import { ReactQueryProvider } from './react-query-provider'
import { RoutingProvider } from './routing-provider'

export const AppProvider = () => {
  return (
    <ReactQueryProvider>
      <ModulesProvider>
        <NotificationsProvider>
          <RoutingProvider />
        </NotificationsProvider>
      </ModulesProvider>
    </ReactQueryProvider>
  )
}
