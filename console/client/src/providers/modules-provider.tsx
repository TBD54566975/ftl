import {
  PropsWithChildren,
  createContext,
  useEffect,
  useState,
  useContext,
} from 'react'
import {useClient} from '../hooks/use-client'
import {ConsoleService} from '../protos/xyz/block/ftl/v1/console/console_connect'
import {GetModulesResponse} from '../protos/xyz/block/ftl/v1/console/console_pb'
import {schemaContext} from './schema-provider'

export const modulesContext = createContext<GetModulesResponse>(
  new GetModulesResponse()
)

const ModulesProvider = (props: PropsWithChildren) => {
  const schema = useContext(schemaContext)
  const client = useClient(ConsoleService)
  const [modules, setModules] = useState<GetModulesResponse>(
    new GetModulesResponse()
  )

  useEffect(() => {
    async function fetchModules() {
      const modules = await client.getModules({})
      setModules(modules ?? [])

      return
    }
    fetchModules()
  }, [client, schema])

  return (
    <modulesContext.Provider value={modules}>
      {props.children}
    </modulesContext.Provider>
  )
}

export default ModulesProvider
