import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { useStreamModules } from '../../api/modules/use-stream-modules'
import { Schema } from './schema/Schema'

export const ModulePanel = () => {
  const { moduleName } = useParams()
  const modules = useStreamModules()

  const module = useMemo(() => {
    if (!modules?.data) {
      return
    }
    return modules.data.find((module) => module.name === moduleName)
  }, [modules?.data, moduleName])

  if (!module) return

  return (
    <div className='mt-4 mx-4'>
      <Schema schema={module.schema} />
    </div>
  )
}
