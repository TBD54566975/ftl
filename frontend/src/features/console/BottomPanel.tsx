import React, { useMemo } from 'react'
import { Timeline } from '../timeline/Timeline'

interface BottomPanelProps {
  height: number
}
const timeSettings = { isTailing: true, isPaused: false }
const BottomPanel: React.FC<BottomPanelProps> = ({ height }) => {
  const filters = useMemo(() => {
    return []
  }, [])
  return (
    <div style={{ height: `${height}px` }} className='overflow-auto '>
      {<Timeline timeSettings={timeSettings} filters={filters} />}
    </div>
  )
}

export default BottomPanel
