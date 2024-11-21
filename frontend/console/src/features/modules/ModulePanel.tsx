import { useEffect, useMemo, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { useStreamModules } from '../../api/modules/use-stream-modules'
import { Schema } from './schema/Schema'

export const ModulePanel = () => {
  const { moduleName } = useParams()
  const modules = useStreamModules()
  const ref = useRef<HTMLDivElement>(null)

  const module = useMemo(() => {
    if (!modules?.data) {
      return
    }
    return modules.data.modules.find((module) => module.name === moduleName)
  }, [modules?.data, moduleName])

  useEffect(() => {
    ref?.current?.parentElement?.scrollTo({ top: 0, behavior: 'smooth' })
  }, [moduleName])

  if (!module) return

  return (
    <div ref={ref} className='mt-4 mx-4'>
      <Schema schema={module.schema} />
    </div>
  )
}
