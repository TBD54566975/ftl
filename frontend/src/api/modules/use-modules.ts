import { ConsoleService } from '../../protos/xyz/block/ftl/v1/console/console_connect'
import { GetModulesResponse } from '../../protos/xyz/block/ftl/v1/console/console_pb'

import { Code, ConnectError } from '@connectrpc/connect'
import { useClient } from '../../hooks/use-client'

const fetchModules = async (client: ConsoleService, isVisible: boolean): Promise<GetModulesResponse> => {
  if (!isVisible) {
    throw new Error('Component is not visible')
  }

  const abortController = new AbortController()

  try {
    const modules = await client.getModules({}, { signal: abortController.signal })
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
  } finally {
    abortController.abort()
  }
}

export const useModules = () => {
  const client = useClient(ConsoleService)
  const isVisible = useVisibility()
  const schema = useSchema()

  return useQuery<GetModulesResponse>(
    ['modules', schema, isVisible], // The query key, include schema and isVisible as dependencies
    () => fetchModules(client, isVisible),
    {
      enabled: isVisible, // Only run the query when the component is visible
      refetchOnWindowFocus: false, // Optional: Disable refetching on window focus
      staleTime: 1000 * 60 * 5, // Optional: Cache data for 5 minutes
    }
  )
}
