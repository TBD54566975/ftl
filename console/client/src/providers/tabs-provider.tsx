import React from 'react'
import { useSearchParams } from 'react-router-dom'

export interface Tab {
  id: string
  label: string
  isClosable: boolean
  component: React.ReactNode
}

interface TabsContextType {
  tabs: Tab[]
  activeTabId?: string
  openTab: (tab: Tab) => void
  closeTab: (tabId: string) => void
}

export const TabsContext = React.createContext<TabsContextType>({
  tabs: [],
  activeTabId: undefined,
  openTab: () => {},
  closeTab: () => {},
})

export const TabsProvider = (props: React.PropsWithChildren) => {
  const [searchParams, setSearchParams] = useSearchParams()
  const [tabs, setTabs] = React.useState<Tab[]>([])
  const [activeTabId, setActiveTabId] = React.useState<string>()

  const updateParams = (tab: Tab) => {
    if (tab.id !== 'timeline') {
      setSearchParams({
        ...Object.fromEntries(searchParams),
        verb: tab.id,
      })
    }
  }

  const openTab = (tab: Tab) => {
    setTabs((prevTabs) => {
      // Add the tab if it doesn't exist
      if (!prevTabs.some((existingTab) => existingTab.id === tab.id)) {
        return [...prevTabs, tab]
      }
      return prevTabs
    })

    setActiveTabId(tab.id)
    updateParams(tab)
  }

  const closeTab = (tabId: string) => {
    const newTabs = tabs.filter((tab) => tab.id !== tabId)
    const closedTabIndex = tabs.findIndex((tab) => tab.id === tabId)

    if (activeTabId === tabId) {
      const activeTab = newTabs[closedTabIndex - 1]
      setActiveTabId(activeTab?.id)
      updateParams(activeTab)
    }

    if (newTabs.length === 1 && newTabs[0].id === 'timeline') {
      searchParams.delete('verb')
      setSearchParams(searchParams)
    }
    setTabs(newTabs)
  }

  return <TabsContext.Provider value={{ tabs, activeTabId, openTab, closeTab }}>{props.children}</TabsContext.Provider>
}
