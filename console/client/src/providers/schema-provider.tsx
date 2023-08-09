import { createContext, useEffect, useState } from 'react'
import { DeploymentChangeType, PullSchemaResponse } from '../protos/xyz/block/ftl/v1/ftl_pb'
import { useClient } from '../hooks/use-client'
import { ControllerService } from '../protos/xyz/block/ftl/v1/ftl_connect.ts'

export const schemaContext = createContext<PullSchemaResponse[]>([])

const SchemaProvider = props => {
  const client = useClient(ControllerService)
  const [ schema, setSchema ] = useState<PullSchemaResponse[]>([])

  useEffect(() => {
    async function fetchSchema() {
      const schemaMap = new Map<string, PullSchemaResponse>()
      for await (const response of client.pullSchema({})) {
        const moduleName = response.moduleName ?? ''
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
          setSchema(
            Array.from(schemaMap.values()).sort((a, b) => a.schema?.name?.localeCompare(b.schema?.name ?? '') ?? 0)
          )
        }
      }
    }
    fetchSchema()
  }, [ client ])

  return <schemaContext.Provider value={schema}>{props.children}</schemaContext.Provider>
}

export default SchemaProvider
