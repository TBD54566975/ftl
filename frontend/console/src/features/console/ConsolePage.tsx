import { useState } from 'react'
import { type NavigateFunction, useNavigate } from 'react-router-dom'
import { useModules } from '../../api/modules/use-modules'
import { Loader } from '../../components/Loader'
import { ResizablePanels } from '../../components/ResizablePanels'
import { Config, Data, Database, Enum, Module, Secret, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import type { FTLNode } from '../graph/GraphPane'
import { GraphPane } from '../graph/GraphPane'
import { NewGraphPane } from '../graph/NewGraphPane'
import { configPanels } from '../modules/decls/config/ConfigRightPanels'
import { dataPanels } from '../modules/decls/data/DataRightPanels'
import { databasePanels } from '../modules/decls/database/DatabaseRightPanels'
import { enumPanels } from '../modules/decls/enum/EnumRightPanels'
import { secretPanels } from '../modules/decls/secret/SecretRightPanels'
import { verbPanels } from '../modules/decls/verb/VerbRightPanel'
import { Timeline } from '../timeline/Timeline'
import type { ExpandablePanelProps } from './ExpandablePanel'
import { modulePanels } from './ModulePanels'
import { headerForNode } from './RightPanelHeader'

export const ConsolePage = () => {
  const modules = useModules()
  const navigate = useNavigate()
  const [selectedNode, setSelectedNode] = useState<FTLNode | null>(null)
  const [graphType, setGraphType] = useState<'new' | 'legacy'>('new')

  if (!modules.isSuccess) {
    return (
      <div className='flex justify-center items-center h-full'>
        <Loader />
      </div>
    )
  }

  const renderGraph = () => {
    return (
      <div className='flex flex-col h-full'>
        <div className='py-2 px-4 border-b border-gray-200 dark:border-gray-700 flex justify-end'>
          <label className='flex items-center gap-2 cursor-pointer'>
            <span className='text-sm text-gray-600 dark:text-gray-400'>Use new graph</span>
            <input type='checkbox' checked={graphType === 'new'} onChange={(e) => setGraphType(e.target.checked ? 'new' : 'legacy')} className='sr-only peer' />
            <div className="relative w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600" />
          </label>
        </div>
        <div className='flex-1'>{graphType === 'new' ? <NewGraphPane onTapped={setSelectedNode} /> : <GraphPane onTapped={setSelectedNode} />}</div>
      </div>
    )
  }

  return (
    <div className='flex h-full'>
      <ResizablePanels
        mainContent={renderGraph()}
        rightPanelHeader={headerForNode(selectedNode)}
        rightPanelPanels={panelsForNode(modules.data.modules, selectedNode, navigate)}
        bottomPanelContent={<Timeline timeSettings={{ isTailing: true, isPaused: false }} filters={[]} />}
      />
    </div>
  )
}

const panelsForNode = (modules: Module[], node: FTLNode | null, navigate: NavigateFunction) => {
  if (node instanceof Module) {
    return modulePanels(modules, node, navigate)
  }

  if (node instanceof Config) {
    return configPanels(node)
  }
  if (node instanceof Secret) {
    return secretPanels(node)
  }
  if (node instanceof Database) {
    return databasePanels(node)
  }
  if (node instanceof Enum) {
    return enumPanels(node)
  }
  if (node instanceof Data) {
    return dataPanels(node)
  }
  if (node instanceof Verb) {
    return verbPanels(node)
  }
  return [] as ExpandablePanelProps[]
}
