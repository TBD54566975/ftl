import React from 'react'

interface SidePanelContextType {
  isOpen: boolean
  component: React.ReactNode
  openPanel: (component: React.ReactNode) => void
  closePanel: () => void
}

const defaultContextValue: SidePanelContextType = {
  isOpen: false,
  component: null,
  openPanel: () => {},
  closePanel: () => {},
}

export const SidePanelContext =
  React.createContext<SidePanelContextType>(defaultContextValue)

export const SidePanelProvider = ({children}) => {
  const [isOpen, setIsOpen] = React.useState(false)
  const [component, setComponent] = React.useState<React.ReactNode>()

  const openPanel = (comp: React.ReactNode) => {
    setIsOpen(true)
    setComponent(comp)
  }

  const closePanel = () => {
    setIsOpen(false)
    setComponent(undefined)
  }

  return (
    <SidePanelContext.Provider
      value={{isOpen, openPanel, closePanel, component}}
    >
      {children}
    </SidePanelContext.Provider>
  )
}
