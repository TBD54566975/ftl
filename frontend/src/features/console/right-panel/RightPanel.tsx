import React from 'react'
import { ExpandablePanel, ExpandablePanelProps } from '../ExpandablePanel'

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
