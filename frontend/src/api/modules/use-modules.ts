import { Code, ConnectError } from '@connectrpc/connect'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useClient } from '../../hooks/use-client'
import { ConsoleService } from '../../protos/xyz/block/ftl/v1/console/console_connect'
import { useSchema } from '../schema/use-schema'

const useModulesKey = 'modules'

export const useModules = () => {
  const client = useClient(ConsoleService)
  const queryClient = useQueryClient()
  const { data: streamingData } = useSchema()

  useEffect(() => {
    if (streamingData) {
      queryClient.invalidateQueries({
        queryKey: [useModulesKey],
      })
    }
  }, [streamingData, queryClient])

  const fetchModules = async (signal: AbortSignal) => {
    try {
      console.debug('fetching modules from FTL')
      const modules = await client.getModules({}, { signal })
      return modules ?? []
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code !== Code.Canceled) {
          console.error('fetchModules - Connect error:', error)
        }
      } else {
        console.error('fetchModules:', error)
      }
      throw error
    }
  }

  return useQuery({
    queryKey: [useModulesKey],
    queryFn: async ({ signal }) => fetchModules(signal),
    enabled: !!streamingData,
  })
}
