import { useContext, useState } from 'react'
import BottomPanel from './BottomPanel'
import { Config, Module, Secret, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { ExpandablePanelProps } from './ExpandablePanel'
import { FTLNode, GraphPane } from '../graph/GraphPane'
import { Page } from '../../layout'
import { CubeTransparentIcon } from '@heroicons/react/24/outline'
import { verbPanels } from './right-panel/VerbPanels'
import { modulesContext } from '../../providers/modules-provider'
import { NavigateFunction, useNavigate } from 'react-router-dom'
import { headerForNode } from './right-panel/RightPanelHeader'
import { modulePanels } from './right-panel/ModulePanels'
import { secretPanels } from './right-panel/SecretPanels'
import { configPanels } from './right-panel/ConfigPanels'
import { ResizablePanels } from '../../components/ResizablePanels'

export const ConsolePage = () => {
  const modules = useContext(modulesContext)
  const navigate = useNavigate()
  const [selectedNode, setSelectedNode] = useState<FTLNode | null>(null)

  return (
    <Page>
      <Page.Header icon={<CubeTransparentIcon />} title='Console' />
      <Page.Body className='flex h-full'>
        <ResizablePanels
          mainContent={<GraphPane onTapped={setSelectedNode} />}
          rightPanelHeader={headerForNode(selectedNode)}
          rightPanelPanels={panelsForNode(modules.modules, selectedNode, navigate)}
          bottomPanelContent={<BottomPanel />}
        />
      </Page.Body>
    </Page>
  )
}

const panelsForNode = (modules: Module[], node: FTLNode | null, navigate: NavigateFunction) => {
  if (node instanceof Module) {
    return modulePanels(modules, node, navigate)
  } else if (node instanceof Verb) {
    return verbPanels(node)
  } else if (node instanceof Secret) {
    return secretPanels(node)
  } else if (node instanceof Config) {
    return configPanels(node)
  } else {
    return [] as ExpandablePanelProps[]
  }
}
