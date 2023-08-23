import React from 'react'
import { Module, Verb } from '../protos/xyz/block/ftl/v1/console/console_pb'
import { ConsoleService } from '../protos/xyz/block/ftl/v1/console/console_connect'
import { useClient } from '../hooks/use-client'
import { useSearchParams } from 'react-router-dom'

export const TabType  ={
  Timeline: 'timeline',
  Verb: 'verb',
} as const

export type Tab = {
  id: string;
  label: string;
  type: typeof TabType[keyof typeof TabType];
  module?: Module | null;
  verb?: Verb | null;
}

export const TabSearchParams = {
  verb: 'active-verb',
} as const

export const timelineTab = { id: 'timeline', label: 'Timeline', type: TabType.Timeline }

type TabsContextType = {
  tabs: Tab[];
  activeTab?: Tab | null;
  setTabs: React.Dispatch<React.SetStateAction<Tab[]>>;
  setActiveTab: React.Dispatch<React.SetStateAction<Tab>>;
}

export const TabsContext = React.createContext<TabsContextType>({ tabs: [], activeTab: null, setTabs: () => { }, setActiveTab: () => { } })

export const TabsProvider = (props: React.PropsWithChildren) => {
  const [ tabs, setTabs ] = React.useState<Tab[]>([ timelineTab ])
  const [ activeTab, setActiveTab ] = React.useState<Tab>(timelineTab)
  const [ searchParams ] = useSearchParams()
  // const client = useClient(ConsoleService)

  React.useEffect(() => {
    async function fetchModules() {
      // const modules = await client.getModules({})
      const activeVerb = searchParams.get(TabSearchParams.verb)
      if(activeVerb) {
       
      
      }

      return
    }
    fetchModules()
  })
  return <TabsContext.Provider value={{ tabs, setTabs, activeTab, setActiveTab }}>{props.children}</TabsContext.Provider>
}
