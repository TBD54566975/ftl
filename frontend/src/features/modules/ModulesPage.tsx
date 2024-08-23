import { useMemo } from 'react'
import { useSchema } from '../../api/schema/use-schema'
import { ModulesTree } from './ModulesTree'
import { moduleTreeFromSchema } from './module.utils'

export const ModulesPage = () => {
  const schema = useSchema()
  const tree = useMemo(() => moduleTreeFromSchema(schema?.data || []), [schema?.data])

  return (
    <div className='flex h-full'>
      <div className='w-64 h-full'>
        <ModulesTree modules={tree} />
      </div>

      <div className='flex-1 py-2 px-4'>
        <p>Content</p>
      </div>
    </div>
  )
}
