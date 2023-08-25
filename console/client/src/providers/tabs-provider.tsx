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

export const isValidTab = ({ id, type }: {id?: string; type?: string}): boolean => {
  //P1 no ID or type
  if(!id || !type) {
    // throw new Error(`required tab field undefined: ${JSON.stringify({ type, id })}`)
    return false
  }
  //P2 invalid type
  const invalidType = Object.values(TabType).some(v => v === type)
  if(!invalidType) {
    // throw new Error(`invalid tab type: ${type}`)
    return false
  }

  //P3 type is timeline but id is wrong
  if(type === TabType.Timeline) {
    if(id !== timelineTab.id) {
      // throw new Error(`invalid timeline id: ${id}`)
      return false
    }
  }
  //P4 type is verb but invalid type
  if(type === TabType.Verb) {
    const verbIdArray = id.split('.')
    if(type === TabType.Verb && verbIdArray.length !== 2) {
      // throw new Error(`invalid verb ${id}`)
      return false
    }
  }
  return true
}

export const TabsProvider = (props: React.PropsWithChildren) => {
  const [ tabs, setTabs ] = React.useState<Tab[]>([ timelineTab ])
  const [ activeTab, setActiveTab ] = React.useState<{id: string, type: string}| undefined>()

  return <TabsContext.Provider value={{ tabs, setTabs, activeTab, setActiveTab }}>{props.children}</TabsContext.Provider>
}
