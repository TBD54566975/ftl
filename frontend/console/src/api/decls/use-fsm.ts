import { useClient } from '../../hooks/use-client'
import { ConsoleService } from '../../protos/xyz/block/ftl/v1/console/console_connect'
import { useDecl } from './use-decl'

export const useFsm = (moduleName: string, declName: string) => {
  const client = useClient(ConsoleService)
  return useDecl('fsm', moduleName, declName, client.getFSM)
}
