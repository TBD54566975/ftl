import { useClient } from '../../hooks/use-client'
import { ConsoleService } from '../../protos/xyz/block/ftl/v1/console/console_connect'
import { useDecl } from './use-decl'

export const useVerb = (moduleName: string, declName: string) => {
  const client = useClient(ConsoleService)
  return useDecl('verb', moduleName, declName, client.getVerb)
}
