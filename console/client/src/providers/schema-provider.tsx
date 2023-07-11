import { PropsWithChildren, createContext, useEffect, useState } from 'react'
import { PullSchemaResponse } from '../protos/xyz/block/ftl/v1/ftl_pb'

// eslint-disable-next-line react-refresh/only-export-components
export const schemaContext = createContext<PullSchemaResponse[]>([])

const SchemaProvider = (props: PropsWithChildren) => {
  const [schema] = useState<PullSchemaResponse[]>([])

  useEffect(() => {
    async function fetchSchema() {
      return
    }
    fetchSchema()
  }, [])

  return <schemaContext.Provider value={schema}>{props.children}</schemaContext.Provider>
}

export default SchemaProvider
