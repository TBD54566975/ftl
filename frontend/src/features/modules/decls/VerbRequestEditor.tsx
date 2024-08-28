import { useEffect, useState } from 'react'
import { useModules } from '../../../api/modules/use-modules'
import { CodeEditor } from '../../../components/CodeEditorV2'
import { defaultRequest, simpleJsonSchema } from '../../verbs/verb.utils'
import { deploymentKeyModuleName } from '../module.utils'
import type { Verb as SchemaVerb } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import type { /*Module, */Verb } from '../../../protos/xyz/block/ftl/v1/console/console_pb'

export const VerbRequestEditor = ({ moduleName, v }: { moduleName: string, v: SchemaVerb }) => {
  const modules = useModules()
  const [verb, setVerb] = useState<Verb | undefined>()
  const [editorText, setEditorText] = useState<string | undefined>()

  useEffect(() => {
    if (!modules.isSuccess) return
    if (modules.data.modules.length === 0) return
    const module = modules.data.modules.find((module) => module.name === moduleName)
    const verb = module?.verbs.find((verb) => verb.verb?.name.toLocaleLowerCase() === v.name.toLocaleLowerCase())
    setVerb(verb)
    setEditorText(defaultRequest(verb))
  }, [modules.data])

  if (!verb) {
    // Editor text has not yet been populated with the default request
    return (
      <div className='h-full'>
        <CodeEditor readOnly={true} content='' />
      </div>
    )
  }

  const schemaString = verb ? JSON.stringify(simpleJsonSchema(verb)) : '{}'
  return (
    <div className='h-full relative'>
      <div className='absolute z-10 top-2 text-sm px-2 py-0.5 right-4 cursor-pointer shadow-lg rounded-md bg-gray-200 hover:bg-gray-300 dark:bg-gray-900 hover:dark:bg-gray-700' onClick={() => setEditorText(defaultRequest(verb))}>Reset</div>
      <CodeEditor content={editorText} schema={schemaString} onChange={setEditorText} />
    </div>
  )
}
