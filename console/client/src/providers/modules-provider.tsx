import { Code, ConnectError } from '@bufbuild/connect'
import { PropsWithChildren, createContext, useContext, useEffect, useState } from 'react'
import { useClient } from '../hooks/use-client'
import { ConsoleService } from '../protos/xyz/block/ftl/v1/console/console_connect'
import { GetModulesResponse } from '../protos/xyz/block/ftl/v1/console/console_pb'
import { schemaContext } from './schema-provider'

export const modulesContext = createContext<GetModulesResponse>(new GetModulesResponse())

export const ModulesProvider = (props: PropsWithChildren) => {
  const schema = useContext(schemaContext)
  const client = useClient(ConsoleService)
  const [modules, setModules] = useState<GetModulesResponse>(new GetModulesResponse())

  useEffect(() => {
    const abortController = new AbortController()
    const fetchModules = async () => {
      try {
        const modules = await client.getModules({}, { signal: abortController.signal })
        setModules(modules ?? [])
      } catch (error) {
        if (error instanceof ConnectError) {
          if (error.code !== Code.Canceled) {
            console.error('ModulesProvider - Connect error:', error)
          }
        } else {
          console.error('ModulesProvider:', error)
        }
      }

      return
    }

    fetchModules()
    return () => {
      abortController.abort()
    }
  }, [client, schema])

  return <modulesContext.Provider value={modules}>{props.children}</modulesContext.Provider>
}
