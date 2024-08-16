import { Code, ConnectError } from '@connectrpc/connect'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useClient } from '../../hooks/use-client.ts'
import { useVisibility } from '../../hooks/use-visibility.ts'
import { ControllerService } from '../../protos/xyz/block/ftl/v1/ftl_connect.ts'
import { DeploymentChangeType, type PullSchemaResponse } from '../../protos/xyz/block/ftl/v1/ftl_pb.ts'

const streamingSchemaKey = 'streamingSchema'

export const useSchema = () => {
  const client = useClient(ControllerService)
  const queryClient = useQueryClient()
  const isVisible = useVisibility()

  const streamSchema = async (signal: AbortSignal) => {
    try {
      const schemaMap = new Map<string, PullSchemaResponse>()
      for await (const response of client.pullSchema({}, { signal })) {
        const moduleName = response.moduleName ?? ''
        console.log(`schema changed: ${DeploymentChangeType[response.changeType]} ${moduleName}`)
        switch (response.changeType) {
          case DeploymentChangeType.DEPLOYMENT_ADDED:
            schemaMap.set(moduleName, response)
            break
          case DeploymentChangeType.DEPLOYMENT_CHANGED:
            schemaMap.set(moduleName, response)
            break
          case DeploymentChangeType.DEPLOYMENT_REMOVED:
            schemaMap.delete(moduleName)
        }

        if (!response.more) {
          const schema = Array.from(schemaMap.values()).sort((a, b) => a.schema?.name?.localeCompare(b.schema?.name ?? '') ?? 0)
          queryClient.setQueryData([streamingSchemaKey], schema)
        }
      }
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code !== Code.Canceled) {
          console.error('useSchema - streamSchema - Connect error:', error)
        }
      } else {
        console.error('useSchema - streamSchema:', error)
      }
    }
  }

  return useQuery({
    queryKey: [streamingSchemaKey],
    queryFn: async ({ signal }) => streamSchema(signal),
    enabled: isVisible,
  })
}
