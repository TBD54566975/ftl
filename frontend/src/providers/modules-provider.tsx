import { Code, ConnectError } from '@connectrpc/connect'
import { type PropsWithChildren, createContext, useEffect, useState } from 'react'
import { useClient } from '../hooks/use-client'
import { useSchema } from '../hooks/use-schema'
import { useVisibility } from '../hooks/use-visibility'
import { ConsoleService } from '../protos/xyz/block/ftl/v1/console/console_connect'
import { GetModulesResponse } from '../protos/xyz/block/ftl/v1/console/console_pb'

export const modulesContext = createContext<GetModulesResponse>(new GetModulesResponse())

export const ModulesProvider = ({ children }: PropsWithChildren) => {
  const schema = useSchema()
  const client = useClient(ConsoleService)
  const [modules, setModules] = useState<GetModulesResponse>(new GetModulesResponse())
  const isVisible = useVisibility()

  useEffect(() => {
    const abortController = new AbortController()

    const fetchModules = async () => {
      if (!isVisible) {
        abortController.abort()
        return
      }

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
  }, [client, schema, isVisible])

  return <modulesContext.Provider value={modules}>{children}</modulesContext.Provider>
}
