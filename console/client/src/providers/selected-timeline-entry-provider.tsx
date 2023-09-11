import {PropsWithChildren, createContext, useState} from 'react'
import {TimelineEvent} from '../protos/xyz/block/ftl/v1/console/console_pb'

type SelectedTimelineEntryContextType = {
  selectedEntry: TimelineEvent | null
  setSelectedEntry: React.Dispatch<React.SetStateAction<TimelineEvent | null>>
}

export const SelectedTimelineEntryContext =
  createContext<SelectedTimelineEntryContextType>({
    selectedEntry: null,
    setSelectedEntry: () => {},
  })

export const SelectedTimelineEntryProvider = (props: PropsWithChildren) => {
  const [selectedEntry, setSelectedEntry] = useState<TimelineEvent | null>(null)

  return (
    <SelectedTimelineEntryContext.Provider
      value={{selectedEntry, setSelectedEntry}}
    >
      {props.children}
    </SelectedTimelineEntryContext.Provider>
  )
}
