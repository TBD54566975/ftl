import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { useSchema } from '../../../api/schema/use-schema'
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

  switch (decl.value.case) {
      case 'verb': return <VerbPanel v={decl.value.value} moduleName={moduleName} verbName={declName} />
  }
  return (
    <div className='flex-1 py-2 px-4'>
      <p>{declCase} declaration: {moduleName}.{declName}</p>
    </div>
  )
}
