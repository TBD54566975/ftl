import { useClient } from '../../hooks/use-client'
import { ConsoleService } from '../../protos/xyz/block/ftl/v1/console/console_connect'
import { useDecl } from './use-decl'

export const useSubscription = (moduleName: string, declName: string) => {
  const client = useClient(ConsoleService)
  return useDecl('subscription', moduleName, declName, client.getSubscription)
}
