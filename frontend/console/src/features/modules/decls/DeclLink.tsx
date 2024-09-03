import { useMemo } from 'react'
import { Link } from 'react-router-dom'
import { useSchema } from '../../../api/schema/use-schema'
import type { PullSchemaResponse } from '../../../protos/xyz/block/ftl/v1/ftl_pb.ts'

export const DeclLink = ({ moduleName, declName }: { moduleName?: string; declName: string }) => {
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

  return !decl ? (
    str
  ) : (
    <Link className='rounded-md cursor-pointer hover:bg-gray-100 hover:dark:bg-gray-700 p-1 -m-1' to={`/modules/${moduleName}/${decl.value.case}/${declName}`}>
      {str}
    </Link>
  )
}
