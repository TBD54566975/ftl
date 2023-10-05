import React, { PropsWithChildren } from 'react'
import { SidePanel } from '../layout/SidePanel'

interface SidePanelContextType {
  isOpen: boolean
  component: React.ReactNode
  openPanel: (component: React.ReactNode, onClose?: () => void) => void
  closePanel: () => void
}

const defaultContextValue: SidePanelContextType = {
  isOpen: false,
  component: null,
  openPanel: () => {},
  closePanel: () => {},
}

export const SidePanelContext = React.createContext<SidePanelContextType>(defaultContextValue)

export const SidePanelProvider = ({ children }: PropsWithChildren) => {
  const [isOpen, setIsOpen] = React.useState(false)
  const [component, setComponent] = React.useState<React.ReactNode>()
  const [onCloseCallback, setOnCloseCallback] = React.useState<(() => void) | null>(null)

  const openPanel = React.useCallback((comp: React.ReactNode, onClose?: () => void) => {
    setIsOpen(true)
    setComponent(comp)
    if (onClose) {
      setOnCloseCallback(() => onClose)
    }
  }, [])

  const closePanel = React.useCallback(() => {
    setIsOpen(false)
    setComponent(undefined)
    if (onCloseCallback) {
      onCloseCallback()
    }
    setOnCloseCallback(null)
  }, [onCloseCallback])

  return (
    <SidePanelContext.Provider value={{ isOpen, openPanel, closePanel, component }}>
      {children}
      <SidePanel />
    </SidePanelContext.Provider>
  )
}
