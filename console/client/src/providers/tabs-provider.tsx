import React from 'react'

export const TabType  = {
  Timeline: 'timeline',
  Verb: 'verb',
} as const

export type Tab = {
  id: string;
  label: string;
  type: typeof TabType[keyof typeof TabType];
}

export const TabSearchParams = {
  id: 'tab-id',
  type: 'tab-type',
} as const

export const timelineTab = { id: 'timeline', label: 'Timeline', type: TabType.Timeline }

export type ActiveTab = {
  id: string;
  type: string;
} | undefined

type TabsContextType = {
  tabs: Tab[];
  activeTab?: ActiveTab;
  setTabs: React.Dispatch<React.SetStateAction<Tab[]>>;
  setActiveTab: React.Dispatch<React.SetStateAction<ActiveTab>>;
}

export const TabsContext = React.createContext<TabsContextType>({
  tabs: [],
  activeTab: undefined,
  setTabs: () => { },
  setActiveTab:  () => {},
})

export const TabsProvider = (props: React.PropsWithChildren) => {
  const [ tabs, setTabs ] = React.useState<Tab[]>([ timelineTab ])
  const [ activeTab, setActiveTab ] = React.useState<{id: string, type: string} | undefined>()

  return <TabsContext.Provider value={{ tabs, setTabs, activeTab, setActiveTab }}>{props.children}</TabsContext.Provider>
}
