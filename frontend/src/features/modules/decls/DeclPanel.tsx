import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { useSchema } from '../../../api/schema/use-schema'
import { DataPanel } from './DataPanel'
import { VerbPanel } from './VerbPanel'
import { declFromSchema } from '../module.utils'

export const DeclPanel = () => {
  const { moduleName, declCase, declName } = useParams()
  if (!moduleName || !declName) {
    // Should be impossible, but validate anyway for type safety
    return []
  }

  const schema = useSchema()
  const decl = useMemo(() => declFromSchema(moduleName, declName, schema?.data || []), [schema?.data])
  if (!decl) {
    return []
  }

  const nameProps = {moduleName, declName}
  switch (decl.value.case) {
      case 'data': return <DataPanel value={decl.value.value} {...nameProps} />
      case 'verb': return <VerbPanel value={decl.value.value} {...nameProps} />
  }
  return (
    <div className='flex-1 py-2 px-4'>
      <p>{declCase} declaration: {moduleName}.{declName}</p>
    </div>
  )
}
