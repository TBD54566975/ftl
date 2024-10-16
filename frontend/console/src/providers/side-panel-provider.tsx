import React, { type PropsWithChildren, useState } from 'react'
import { SidePanel } from '../layout/SidePanel'

interface SidePanelContextType {
  isOpen: boolean
  header?: React.ReactNode
  component?: React.ReactNode
  openPanel: (component: React.ReactNode, header?: React.ReactNode, onClose?: () => void) => void
  closePanel: () => void
}

const defaultContextValue: SidePanelContextType = {
  isOpen: false,
  openPanel: () => {},
  closePanel: () => {},
}

export const SidePanelContext = React.createContext<SidePanelContextType>(defaultContextValue)

export const SidePanelProvider = ({ children }: PropsWithChildren) => {
  const [isOpen, setIsOpen] = useState(false)
  const [header, setHeader] = useState<React.ReactNode>()
  const [component, setComponent] = useState<React.ReactNode>()
  const [onCloseCallback, setOnCloseCallback] = useState<() => void>()

  const openPanel = React.useCallback((component?: React.ReactNode, header?: React.ReactNode, onClose?: () => void) => {
    setIsOpen(true)
    setComponent(component)
    setHeader(header)
    if (onClose) {
      setOnCloseCallback(() => onClose)
    }
  }, [])

  const closePanel = React.useCallback(() => {
    setIsOpen(false)
    setComponent(undefined)
    setHeader(undefined)
    if (onCloseCallback) {
      onCloseCallback()
    }
    setOnCloseCallback(undefined)
  }, [onCloseCallback])

  return (
    <SidePanelContext.Provider value={{ isOpen, header, component, openPanel, closePanel }}>
      {children}
      <SidePanel />
    </SidePanelContext.Provider>
  )
}
