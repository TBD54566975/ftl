import type React from 'react'
import { ExpandablePanel, type ExpandablePanelProps } from '../features/graph/ExpandablePanel'

interface RightPanelProps {
  header: React.ReactNode
  panels: ExpandablePanelProps[]
}

const RightPanel: React.FC<RightPanelProps> = ({ header, panels }) => {
  return (
    <div>
      {header}
      {panels.map((panel, index) => (
        <ExpandablePanel key={`panel-${index}`} {...panel}>
          {panel.children}
        </ExpandablePanel>
      ))}
    </div>
  )
}

export default RightPanel
