import React from 'react'
import { ExpandablePanel, ExpandablePanelProps } from '../ExpandablePanel'

interface RightPanelProps {
  width: number
  header: React.ReactNode
  panels: ExpandablePanelProps[]
}

const RightPanel: React.FC<RightPanelProps> = ({ width, header, panels }) => {
  return (
    <div style={{ width: `${width}px` }} className='overflow-y-auto flex flex-col'>
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
