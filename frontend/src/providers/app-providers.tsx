import { NotificationsProvider } from './notifications-provider'
import { ReactQueryProvider } from './react-query-provider'
import { RoutingProvider } from './routing-provider'
import { UserPreferencesProvider } from './user-preferences-provider'

export const AppProvider = () => {
  return (
    <ReactQueryProvider>
      <UserPreferencesProvider>
        <NotificationsProvider>
          <RoutingProvider />
        </NotificationsProvider>
      </UserPreferencesProvider>
    </ReactQueryProvider>
  )
}
