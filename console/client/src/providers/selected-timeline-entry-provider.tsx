import { PropsWithChildren, createContext, useState } from 'react'
import { Event } from '../protos/xyz/block/ftl/v1/console/console_pb'

interface SelectedEventContextType {
  selectedEntry: Event | null
  setSelectedEntry: React.Dispatch<React.SetStateAction<Event | null>>
}

export const SelectedEventContext = createContext<SelectedEventContextType>({
  selectedEntry: null,
  setSelectedEntry: () => {},
})

export const SelectedEventProvider = (props: PropsWithChildren) => {
  const [selectedEntry, setSelectedEntry] = useState<Event | null>(null)

  return (
    <SelectedEventContext.Provider value={{ selectedEntry, setSelectedEntry }}>
      {props.children}
    </SelectedEventContext.Provider>
  )
}
