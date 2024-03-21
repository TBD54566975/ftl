import React from 'react'
import { ExpandablePanel, ExpandablePanelProps } from './ExpandablePanel'

interface RightPanelProps {
  width: number
  panels: ExpandablePanelProps[]
}

const RightPanel: React.FC<RightPanelProps> = ({ width, panels }) => {
  return (
    <div style={{ width: `${width}px` }} className='overflow-auto bg-slate-800'>
      {panels.map((panel, index) => (
        <ExpandablePanel key={index} title={panel.title} expanded={panel.expanded}>
          {panel.children}
        </ExpandablePanel>
      ))}
    </div>
  )
}

export default RightPanel
