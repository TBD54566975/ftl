import { CubeTransparentIcon } from '@heroicons/react/24/outline'
import { useContext, useState } from 'react'
import { type NavigateFunction, useNavigate } from 'react-router-dom'
import { ResizablePanels } from '../../components/ResizablePanels'
import { Page } from '../../layout'
import { Config, Module, Secret, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { modulesContext } from '../../providers/modules-provider'
import { type FTLNode, GraphPane } from '../graph/GraphPane'
import BottomPanel from './BottomPanel'
import type { ExpandablePanelProps } from './ExpandablePanel'
import { configPanels } from './right-panel/ConfigPanels'
import { modulePanels } from './right-panel/ModulePanels'
import { headerForNode } from './right-panel/RightPanelHeader'
import { secretPanels } from './right-panel/SecretPanels'
import { verbPanels } from './right-panel/VerbPanels'

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
