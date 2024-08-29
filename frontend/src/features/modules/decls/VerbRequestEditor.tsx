import { useEffect, useMemo, useState } from 'react'
import { useModules } from '../../../api/modules/use-modules'
import { CodeEditor } from '../../../components/CodeEditorV2'
import type { Verb } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import type { Verb as SchemaVerb } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { defaultRequest, simpleJsonSchema } from '../../verbs/verb.utils'

export const VerbRequestEditor = ({ moduleName, v }: { moduleName: string; v: SchemaVerb }) => {
  const modules = useModules()

  const [verb, setVerb] = useState<Verb | undefined>()
  const defaultContent = useMemo(() => defaultRequest(verb), [verb])

  const [editorText, setEditorText] = useState(defaultContent)
  const schemaString = useMemo(() => (verb ? JSON.stringify(simpleJsonSchema(verb)) : '{}'), [verb])

  useEffect(() => {
    if (!modules.isSuccess) return
    if (modules.data.modules.length === 0) return
    const module = modules.data.modules.find((module) => module.name === moduleName)
    const verb = module?.verbs.find((verb) => verb.verb?.name.toLocaleLowerCase() === v.name.toLocaleLowerCase())
    setVerb(verb)
    setEditorText(defaultContent)
  }, [modules.data])

  if (!verb) {
    // Editor text has not yet been populated with the default request
    return <CodeEditor readOnly={true} content='' />
  }

  return <CodeEditor content={editorText} schema={schemaString} onChange={setEditorText} defaultContent={defaultContent} />
}
