import { useState } from 'react'
import { type NavigateFunction, useNavigate } from 'react-router-dom'
import { useModules } from '../../api/modules/use-modules'
import { Loader } from '../../components/Loader'
import { ResizablePanels } from '../../components/ResizablePanels'
import { Config, Module, Secret, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { type FTLNode, GraphPane } from '../graph/GraphPane'
import { configPanels } from '../modules/decls/config/ConfigRightPanels'
import { Timeline } from '../timeline/Timeline'
import type { ExpandablePanelProps } from './ExpandablePanel'
import { modulePanels } from './right-panel/ModulePanels'
import { headerForNode } from './right-panel/RightPanelHeader'
import { secretPanels } from './right-panel/SecretPanels'
import { verbPanels } from './right-panel/VerbPanels'

export const ConsolePage = () => {
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
  if (node instanceof Verb) {
    return verbPanels(node)
  }
  if (node instanceof Secret) {
    return secretPanels(node)
  }
  if (node instanceof Config) {
    return configPanels(node)
  }
  return [] as ExpandablePanelProps[]
}
