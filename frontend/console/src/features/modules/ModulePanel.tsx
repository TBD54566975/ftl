import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { useModules } from '../../api/modules/use-modules'
import { Schema } from './schema/Schema'

export const ModulePanel = () => {
  const { moduleName } = useParams()
  const modules = useModules()

  const module = useMemo(() => {
    if (!modules.isSuccess || modules.data.modules.length === 0) {
      return
    }
    return modules.data.modules.find((module) => module.name === moduleName)
  }, [modules?.data, moduleName])

  if (!module) return

  return <Schema schema={module.schema} />
}
