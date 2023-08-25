import React from 'react'

export const TabType = {
  Timeline: 'timeline',
  Verb: 'verb',
} as const

export interface Tab {
  id: string
  label: string
  type: (typeof TabType)[keyof typeof TabType]
}

export const TabSearchParams = {
  active: 'active-tab',
} as const

export const timelineTab = {
  id: 'timeline',
  label: 'Timeline',
  type: TabType.Timeline,
}

interface TabsContextType {
  tabs: Tab[]
  activeTab?: number
  setTabs: React.Dispatch<React.SetStateAction<Tab[]>>
  setActiveTab: React.Dispatch<React.SetStateAction<number>>
}

export const TabsContext = React.createContext<TabsContextType>({
  tabs: [],
  activeTab: 0,
  setTabs: () => {},
  setActiveTab: () => {},
})

export const TabsProvider = (props: React.PropsWithChildren) => {
  const [tabs, setTabs] = React.useState<Tab[]>([timelineTab])
  const [activeTab, setActiveTab] = React.useState<number>(0)

  return (
    <TabsContext.Provider value={{ tabs, setTabs, activeTab, setActiveTab }}>{props.children}</TabsContext.Provider>
  )
}
