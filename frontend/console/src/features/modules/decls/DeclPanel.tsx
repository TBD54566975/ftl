import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { useSchema } from '../../../api/schema/use-schema'
import { VerbPage } from '../../verbs/VerbPage'
import { declFromSchema } from '../module.utils'
import { ConfigPanel } from './ConfigPanel'
import { DataPanel } from './DataPanel'
import { DatabasePanel } from './DatabasePanel'
import { EnumPanel } from './EnumPanel'
import { SecretPanel } from './SecretPanel'
import { TypeAliasPanel } from './TypeAliasPanel'

export const DeclPanel = () => {
  const { moduleName, declCase, declName } = useParams()
  if (!moduleName || !declName) {
    // Should be impossible, but validate anyway for type safety
    return
  }

  const schema = useSchema()
  const decl = useMemo(() => declFromSchema(moduleName, declName, schema?.data || []), [schema?.data, moduleName, declCase, declName])
  if (!decl) {
    return
  }

  const nameProps = { moduleName, declName }
  switch (decl.value.case) {
    case 'config':
      return <ConfigPanel value={decl.value.value} {...nameProps} />
    case 'data':
      return <DataPanel value={decl.value.value} {...nameProps} />
    case 'database':
      return <DatabasePanel value={decl.value.value} {...nameProps} />
    case 'enum':
      return <EnumPanel value={decl.value.value} {...nameProps} />
    case 'secret':
      return <SecretPanel value={decl.value.value} {...nameProps} />
    case 'typeAlias':
      return <TypeAliasPanel value={decl.value.value} {...nameProps} />
    case 'verb':
      return <VerbPage {...nameProps} />
  }
  return (
    <div className='flex-1 py-2 px-4'>
      <p>
        {declCase} declaration: {moduleName}.{declName}
      </p>
    </div>
  )
}
