import { PropsWithChildren, createContext, useEffect, useState } from 'react'
import { useClient } from '../hooks/use-client'
import { ConsoleService } from '../protos/xyz/block/ftl/v1/console/console_connect'
import { GetModulesResponse } from '../protos/xyz/block/ftl/v1/console/console_pb'

// eslint-disable-next-line react-refresh/only-export-components
export const modulesContext = createContext<GetModulesResponse>(new GetModulesResponse())

const ModulesProvider = (props: PropsWithChildren) => {
  const client = useClient(ConsoleService)
  const [modules, setModules] = useState<GetModulesResponse>(new GetModulesResponse())

  useEffect(() => {
    async function fetchModules() {
      const modules = await client.getModules({})
      setModules(modules ?? [])

      return
    }
    fetchModules()
  }, [client])

  return <modulesContext.Provider value={modules}>{props.children}</modulesContext.Provider>
}

export default ModulesProvider
