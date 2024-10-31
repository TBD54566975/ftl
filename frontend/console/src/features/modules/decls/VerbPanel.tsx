import type { Verb } from '../../../protos/xyz/block/ftl/v1/console/console_pb'
import { ResizablePanels } from '../../../components/ResizablePanels'
import { verbPanels } from '../../verbs/VerbRightPanel'
import { VerbPage } from '../../verbs/VerbPage'
import { VerbRequestForm } from '../../verbs/VerbRequestForm'
import { declTypeName } from '../module.utils'
import { Schema } from '../schema/Schema'
import { VerbRequestEditor } from './VerbRequestEditor'

const EmptyVerb = ({ value, schema, moduleName }: { value: Verb; schema: string; moduleName: string }) => {
  return (
    <div className='mt-6 text-sm'>
      <Schema schema={schema} moduleName={moduleName} />
      <button className='border rounded-md border-indigo-800 bg-indigo-700 dark:bg-indigo-800 text-white px-3 py-1 mt-6 ml-3.5'>Call</button>
      <span className='ml-1' title='An empty verb is one that takes Unit and returns Unit'>
        empty verb{' '}
      </span>
    </div>
  )
}

const MainContent = ({ value, schema, moduleName }: { value: Verb; schema: string; moduleName: string }) => {
  if (!value || !schema) {
    return
  }
  const decl = value.verb
  if (!decl) {
    return
  }
  const declType = declTypeName('verb', decl)
  //return <VerbRequestForm module={moduleName} verb={value} />
  switch (declType) {
    case 'cronjob':
      return (
        <div>
          <EmptyVerb value={value} schema={schema} moduleName={moduleName} />
        </div>
      )
    case 'ingress':
      return <VerbRequestEditor moduleName={moduleName} v={decl} />
    case 'subscriber':
      return <div>subscriber</div>
  }
  return <VerbRequestForm module={moduleName} verb={value} />
}

export const VerbPanel = ({ value, schema, moduleName, declName }: { value: Verb; schema: string; moduleName: string; declName: string }) => {
  if (!value || !schema) {
    return
  }
  const decl = value.verb
  if (!decl) {
    return
  }

  const nameProps = { moduleName, declName }
  //return <VerbPage {...nameProps} />

  return (
    <ResizablePanels
      mainContent={<MainContent value={value} schema={schema} moduleName={moduleName} />}
      rightPanelHeader={<div/>}
      rightPanelPanels={verbPanels(value, [] /*todo callers or replace this panel*/)}
      bottomPanelContent={<div/>}
    />
  )
}
