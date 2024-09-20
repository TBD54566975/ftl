import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useSchema } from '../../../api/schema/use-schema'
import type { PullSchemaResponse } from '../../../protos/xyz/block/ftl/v1/ftl_pb.ts'
import type { Decl } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { classNames } from '../../../utils'
import { DeclSnippet } from './DeclSnippet'

const SnippetContainer = ({ decl }: { decl: Decl }) => {
  return (
    <div className='absolute p-4 mt-4 -ml-1 rounded-md bg-gray-200 dark:bg-gray-900 text-gray-700 dark:text-white text-xs font-normal z-10 drop-shadow-xl'>
      <div className='-mt-7 mb-2 text-gray-200 dark:text-gray-900'>
        <svg height='20' width='20'>
          <title>triangle</title>
          <polygon points='11,0 9,0 0,20 20,20' fill='currentColor' />
        </svg>
      </div>
      <DeclSnippet decl={decl} />
    </div>
  )
}

// When `slim` is true, print only the decl name, not the module name, and show nothing on hover.
export const DeclLink = ({
  moduleName,
  declName,
  slim,
  textColors = 'text-indigo-600 dark:text-indigo-400',
}: { moduleName?: string; declName: string; slim?: boolean; textColors?: string }) => {
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

  const str = moduleName && slim !== true ? `${moduleName}.${declName}` : declName

  if (!decl) {
    return str
  }

  const navigate = useNavigate()
  return (
    <span
      className={classNames(textColors, 'inline-block rounded-md cursor-pointer hover:bg-gray-100 hover:dark:bg-gray-700 p-1 -m-1 relative')}
      onClick={() => navigate(`/modules/${moduleName}/${decl.value.case}/${declName}`)}
      onMouseEnter={() => setIsHovering(true)}
      onMouseLeave={() => setIsHovering(false)}
    >
      {str}
      {!slim && isHovering && <SnippetContainer decl={decl} />}
    </span>
  )
}
