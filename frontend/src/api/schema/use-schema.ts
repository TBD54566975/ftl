import { Code, ConnectError } from '@connectrpc/connect'
import { useEffect, useState } from 'react'
import { useClient } from '../../hooks/use-client.ts'
import { useVisibility } from '../../hooks/use-visibility.ts'
import { ControllerService } from '../../protos/xyz/block/ftl/v1/ftl_connect.ts'
import { DeploymentChangeType, type PullSchemaResponse } from '../../protos/xyz/block/ftl/v1/ftl_pb.ts'

export const useSchema = () => {
  const client = useClient(ControllerService)
  const [schema, setSchema] = useState<PullSchemaResponse[]>([])
  const isVisible = useVisibility()

  useEffect(() => {
    const abortController = new AbortController()

    const fetchSchema = async () => {
      try {
        if (!isVisible) {
          abortController.abort()
          return
        }

        const schemaMap = new Map<string, PullSchemaResponse>()
        for await (const response of client.pullSchema(
          {},
          {
            signal: abortController.signal,
          },
        )) {
          const moduleName = response.moduleName ?? ''
          console.log(`${response.changeType} ${moduleName}`)
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
            setSchema(Array.from(schemaMap.values()).sort((a, b) => a.schema?.name?.localeCompare(b.schema?.name ?? '') ?? 0))
          }
        }
      } catch (error) {
        if (error instanceof ConnectError) {
          if (error.code !== Code.Canceled) {
            console.error('Console service - streamEvents - Connect error:', error)
          }
        } else {
          console.error('Console service - streamEvents:', error)
        }
      }
    }

    fetchSchema()
    return () => {
      abortController.abort()
    }
  }, [client, isVisible])

  return schema
}
