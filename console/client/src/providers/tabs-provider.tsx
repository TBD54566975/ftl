import { PropsWithChildren, createContext, useState } from 'react'
import { Module, Verb } from '../protos/xyz/block/ftl/v1/console/console_pb'

export enum TabType {
  Timeline = 'timeline',
  Verb = 'verb',
}

export type Tab = {
  id: string;
  label: string;
  type: TabType;
  module?: Module | null;
  verb?: Verb | null;
}

export const timelineTab = { id: 'timeline', label: 'Timeline', type: TabType.Timeline }

type TabsContextType = {
  tabs: Tab[];
  activeTab?: Tab | null;
  setTabs: React.Dispatch<React.SetStateAction<Tab[]>>;
  setActiveTab: React.Dispatch<React.SetStateAction<Tab>>;
}

export const TabsContext = createContext<TabsContextType>({ tabs: [], activeTab: null, setTabs: () => { }, setActiveTab: () => { } })

export const TabsProvider = (props: PropsWithChildren) => {
  const [ tabs, setTabs ] = useState<Tab[]>([ timelineTab ])
  const [ activeTab, setActiveTab ] = useState<Tab>(timelineTab)

  return <TabsContext.Provider value={{ tabs, setTabs, activeTab, setActiveTab }}>{props.children}</TabsContext.Provider>
}
