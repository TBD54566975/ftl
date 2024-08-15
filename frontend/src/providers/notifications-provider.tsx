import React, { type PropsWithChildren, useRef, useState } from 'react'

export enum NotificationType {
  Success = 0,
  Error = 1,
  Warning = 2,
  Info = 3,
}

interface Notification {
  title: string
  message: string
  type: NotificationType
}

interface NotificationContextType {
  isOpen: boolean
  notification?: Notification
  showNotification: (notification: Notification, duration?: number | 'indefinite') => void
  closeNotification: () => void
}

const defaultContextValue: NotificationContextType = {
  isOpen: false,
  showNotification: () => {},
  closeNotification: () => {},
}

export const NotificationsContext = React.createContext<NotificationContextType>(defaultContextValue)

export const NotificationsProvider = ({ children }: PropsWithChildren) => {
  const [isOpen, setIsOpen] = useState(false)
  const [notification, setNotification] = useState<Notification>()
  const timeoutRef = useRef<NodeJS.Timeout | null>(null)

  const closeNotification = () => {
    setIsOpen(false)
    setNotification(undefined)
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }
  }

  const showNotification = (notification: Notification, duration: number | 'indefinite' = 4000) => {
    setIsOpen(true)
    setNotification(notification)
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }
    if (duration !== 'indefinite') {
      timeoutRef.current = setTimeout(() => {
        closeNotification()
      }, duration)
    }
  }

  return <NotificationsContext.Provider value={{ isOpen, showNotification, closeNotification, notification }}>{children}</NotificationsContext.Provider>
}
