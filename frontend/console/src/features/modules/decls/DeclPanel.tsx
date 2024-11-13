import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { useStreamModules } from '../../../api/modules/use-stream-modules'
import type { Config, Data, Database, Enum, Secret, Subscription, Topic, TypeAlias } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import { VerbPage } from '../../verbs/VerbPage'
import { declFromModules } from '../module.utils'
import { declSchemaFromModules } from '../schema/schema.utils'
import { ConfigPanel } from './ConfigPanel'
import { DataPanel } from './DataPanel'
import { DatabasePanel } from './DatabasePanel'
import { EnumPanel } from './EnumPanel'
import { SecretPanel } from './SecretPanel'
import { SubscriptionPanel } from './SubscriptionPanel'
import { TopicPanel } from './TopicPanel'
import { TypeAliasPanel } from './TypeAliasPanel'

export const DeclPanel = () => {
  const { moduleName, declCase, declName } = useParams()
  if (!moduleName || !declCase || !declName) {
    // Should be impossible, but validate anyway for type safety
    return
  }

  const modules = useStreamModules()
  const declSchema = useMemo(
    () => (moduleName && !!modules?.data ? declSchemaFromModules(moduleName, declName, modules?.data) : undefined),
    [moduleName, declName, modules?.data],
  )

  const decl = useMemo(() => declFromModules(moduleName, declCase, declName, modules?.data), [modules?.data, moduleName, declCase, declName])
  if (!declSchema || !decl) {
    return
  }

  const nameProps = { moduleName, declName }
  const commonProps = { moduleName, declName, schema: declSchema.schema }
  switch (declCase) {
    case 'config':
      return <ConfigPanel {...commonProps} value={decl as Config} />
    case 'data':
      return <DataPanel {...commonProps} value={decl as Data} />
    case 'database':
      return <DatabasePanel {...commonProps} value={decl as Database} />
    case 'enum':
      return <EnumPanel {...commonProps} value={decl as Enum} />
    case 'secret':
      return <SecretPanel {...commonProps} value={decl as Secret} />
    case 'subscription':
      return <SubscriptionPanel {...commonProps} value={decl as Subscription} />
    case 'topic':
      return <TopicPanel {...commonProps} value={decl as Topic} />
    case 'typealias':
      return <TypeAliasPanel {...commonProps} value={decl as TypeAlias} />
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
