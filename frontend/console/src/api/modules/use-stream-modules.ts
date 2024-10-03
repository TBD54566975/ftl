import { Code, ConnectError } from '@connectrpc/connect'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useClient } from '../../hooks/use-client'
import { useVisibility } from '../../hooks/use-visibility'
import { ConsoleService } from '../../protos/xyz/block/ftl/v1/console/console_connect'
import type { Module } from '../../protos/xyz/block/ftl/v1/console/console_pb'

const streamModulesKey = 'streamModules'

export const useStreamModules = () => {
  const client = useClient(ConsoleService)
  const queryClient = useQueryClient()
  const isVisible = useVisibility()

  const queryKey = [streamModulesKey]

  const streamModules = async ({ signal }: { signal: AbortSignal }) => {
    try {
      console.debug('streaming modules')
      let hasModules = false
      for await (const response of client.streamModules({}, { signal })) {
        console.debug('stream-modules-response:', response)
        if (response.modules) {
          hasModules = true
          const newModuleNames = response.modules.map((m) => m.name)
          queryClient.setQueryData<Module[]>(queryKey, (prev = []) => {
            return [...response.modules, ...prev.filter((m) => !newModuleNames.includes(m.name))]
          })
        }
      }
      return hasModules ? queryClient.getQueryData(queryKey) : []
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code !== Code.Canceled) {
          console.error('Console service - streamModules - Connect error:', error)
        }
      } else {
        console.error('Console service - streamModules:', error)
      }
      return []
    }
  }

  return useQuery({
    queryKey: queryKey,
    queryFn: async ({ signal }) => streamModules({ signal }),
    enabled: isVisible,
  })
}
