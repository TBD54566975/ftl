import { PropsWithChildren, createContext, useState } from 'react'
import { StreamTimelineResponse } from '../protos/xyz/block/ftl/v1/console/console_pb'

type SelectedTimelineEntryContextType = {
  selectedEntry: StreamTimelineResponse | null;
  setSelectedEntry: React.Dispatch<React.SetStateAction<StreamTimelineResponse | null>>;
}

export const SelectedTimelineEntryContext = createContext<SelectedTimelineEntryContextType>({ selectedEntry: null, setSelectedEntry: () => {} })

export const SelectedTimelineEntryProvider = (props: PropsWithChildren) => {
  const [ selectedEntry, setSelectedEntry ] = useState<StreamTimelineResponse | null>(null)

  return <SelectedTimelineEntryContext.Provider value={{ selectedEntry, setSelectedEntry }}>{props.children}</SelectedTimelineEntryContext.Provider>
}
