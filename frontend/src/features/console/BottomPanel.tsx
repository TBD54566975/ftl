import { useMemo } from 'react'
import { Timeline } from '../timeline/Timeline'

const timeSettings = { isTailing: true, isPaused: false }

const BottomPanel = () => {
  const filters = useMemo(() => {
    return []
  }, [])
  return <Timeline timeSettings={timeSettings} filters={filters} />
}

export default BottomPanel
