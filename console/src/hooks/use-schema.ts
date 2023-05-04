import { useEffect, useState } from 'react'
import { DevelService } from '../protos/xyz/block/ftl/v1/ftl_connect'
import { PullSchemaResponse } from '../protos/xyz/block/ftl/v1/ftl_pb'
import { useClient } from './use-client'

export function useSchema() {
  const client = useClient(DevelService)
  const [schema, setSchema] = useState<PullSchemaResponse[]>([])

  useEffect(() => {
    async function fetchSchema() {
      let schemaParts: PullSchemaResponse[] = []
      for await (const response of client.pullSchema({})) {
        const currentIndex = schemaParts.findIndex(
          res => res.schema?.name === response.schema?.name
        )
        if (currentIndex !== -1) {
          schemaParts[currentIndex] = response
        } else {
          schemaParts = [...schemaParts, response]
        }

        if (!response.more) {
          setSchema(
            schemaParts.sort(
              (a, b) => a.schema?.name?.localeCompare(b.schema?.name ?? '') ?? 0
            )
          )
        }
      }
    }
    fetchSchema()
  }, [client])

  return schema
}
