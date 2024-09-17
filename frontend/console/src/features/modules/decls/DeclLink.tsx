import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useSchema } from '../../../api/schema/use-schema'
import type { PullSchemaResponse } from '../../../protos/xyz/block/ftl/v1/ftl_pb.ts'
import type { Decl } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { DeclSnippet } from './DeclSnippet'

const SnippetContainer = ({ decl }: { decl: Decl }) => {
  return (
    <div className='absolute p-4 mt-4 -ml-1 rounded-md dark:bg-gray-700 dark:text-white text-xs'>
      <div className='absolute -mt-7 dark:text-gray-700'>
        <svg height='20' width='20'>
          <title>triangle</title>
          <polygon points='11,0 9,0 0,20 20,20' fill='currentColor' />
        </svg>
      </div>
      <DeclSnippet decl={decl} />
    </div>
  )
}

export const DeclLink = ({ moduleName, declName }: { moduleName?: string; declName: string }) => {
  const [isHovering, setIsHovering] = useState(false)
  const schema = useSchema()
  const decl = useMemo(() => {
    const modules = (schema?.data || []) as PullSchemaResponse[]
    const module = modules.find((m: PullSchemaResponse) => m.moduleName === moduleName)
    if (!module?.schema) {
      return
    }
    return module.schema.decls.find((d) => d.value.value?.name === declName)
  }, [moduleName, declName, schema?.data])

  const str = moduleName ? `${moduleName}.${declName}` : declName

  if (!decl) {
    return str
  }

  const navigate = useNavigate()
  return (
    <span
      className='inline-block rounded-md cursor-pointer text-indigo-600 dark:text-indigo-400 hover:bg-gray-100 hover:dark:bg-gray-700 p-1 -m-1'
      onClick={() => navigate(`/modules/${moduleName}/${decl.value.case}/${declName}`)}
      onMouseEnter={() => setIsHovering(true)}
      onMouseLeave={() => setIsHovering(false)}
    >
      {str}
      {isHovering && <SnippetContainer decl={decl} />}
    </span>
  )
}
