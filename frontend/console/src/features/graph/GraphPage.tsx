import { useState } from 'react'
import { type NavigateFunction, useNavigate } from 'react-router-dom'
import { useModules } from '../../api/modules/use-modules'
import { Loader } from '../../components/Loader'
import { ResizablePanels } from '../../components/ResizablePanels'
import { Config, Data, Database, Enum, Module, Secret, Topic, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { configPanels } from '../modules/decls/config/ConfigRightPanels'
import { dataPanels } from '../modules/decls/data/DataRightPanels'
import { databasePanels } from '../modules/decls/database/DatabaseRightPanels'
import { enumPanels } from '../modules/decls/enum/EnumRightPanels'
import { secretPanels } from '../modules/decls/secret/SecretRightPanels'
import { topicPanels } from '../modules/decls/topic/TopicRightPanels'
import { verbPanels } from '../modules/decls/verb/VerbRightPanel'
import { Timeline } from '../timeline/Timeline'
import type { ExpandablePanelProps } from './ExpandablePanel'
import { GraphPane } from './GraphPane'
import { modulePanels } from './ModulePanels'
import { headerForNode } from './RightPanelHeader'
import type { FTLNode } from './graph-utils'

export const GraphPage = () => {
  const modules = useModules()
  const navigate = useNavigate()
  const [selectedNode, setSelectedNode] = useState<FTLNode | null>(null)

  if (!modules.isSuccess) {
    return (
      <div className='flex justify-center items-center h-full'>
        <Loader />
      </div>
    )
  }

  return (
    <div className='flex h-full'>
      <ResizablePanels
        mainContent={<GraphPane onTapped={setSelectedNode} />}
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
  if (node instanceof Topic) {
    return topicPanels(node)
  }
  if (node instanceof Verb) {
    return verbPanels(node)
  }
  return [] as ExpandablePanelProps[]
}
