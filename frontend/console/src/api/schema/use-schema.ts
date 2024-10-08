import { Code, ConnectError } from '@connectrpc/connect'
import { type UseQueryResult, useQuery, useQueryClient } from '@tanstack/react-query'
import { useClient } from '../../hooks/use-client.ts'
import { useVisibility } from '../../hooks/use-visibility.ts'
import { ControllerService } from '../../protos/xyz/block/ftl/v1/ftl_connect.ts'
import { DeploymentChangeType, type PullSchemaResponse } from '../../protos/xyz/block/ftl/v1/ftl_pb.ts'

const streamingSchemaKey = 'streamingSchema'
const currentDeployments: Record<string, string> = {}
const schemaMap: Record<string, PullSchemaResponse> = {}

export const useSchema = (): UseQueryResult<PullSchemaResponse[], Error> => {
  const client = useClient(ControllerService)
  const queryClient = useQueryClient()
  const isVisible = useVisibility()

  const streamSchema = async (signal: AbortSignal) => {
    try {
      for await (const response of client.pullSchema({}, { signal })) {
        const moduleName = response.moduleName ?? ''
        const deploymentKey = response.deploymentKey ?? ''
        console.log(`schema changed: ${DeploymentChangeType[response.changeType]} ${deploymentKey}`)
        switch (response.changeType) {
          case DeploymentChangeType.DEPLOYMENT_ADDED:
          case DeploymentChangeType.DEPLOYMENT_CHANGED: {
            const previousDeploymentKey = currentDeployments[moduleName]

            currentDeployments[moduleName] = deploymentKey

            if (previousDeploymentKey && previousDeploymentKey !== deploymentKey) {
              delete schemaMap[previousDeploymentKey]
            }

            schemaMap[deploymentKey] = response
            break
          }

          case DeploymentChangeType.DEPLOYMENT_REMOVED:
            if (currentDeployments[moduleName] === deploymentKey) {
              delete schemaMap[deploymentKey]
              delete currentDeployments[moduleName]
            }
            break
        }

        if (!response.more) {
          const schema = Object.values(schemaMap).sort((a, b) => a.schema?.name?.localeCompare(b.schema?.name ?? '') ?? 0)
          queryClient.setQueryData([streamingSchemaKey], schema ?? [])
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
