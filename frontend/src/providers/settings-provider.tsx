import useLocalStorage from '../hooks/use-local-storage'
import { createContext, PropsWithChildren, useContext, useEffect } from 'react'

interface SettingsContextProps {
  isDarkMode: boolean;
  setDarkMode: (value: boolean) => void;
  language: string;
  setLanguage: (value: string) => void;
  // Add more settings as needed
}

const SettingsContext = createContext<SettingsContextProps | undefined>(undefined)

export const useSettings = () => {
  const context = useContext(SettingsContext)
  if (!context) {
    throw new Error('useSettings must be used within a SettingsProvider')
  }
  return context
}

export const SettingsProvider = ({ children }: PropsWithChildren) => {
  const [isDarkMode, setDarkMode] = useLocalStorage('dark-mode', false)
  const [language, setLanguage] = useLocalStorage('language', 'en') // Example of another setting

  // You can use useEffect to sync settings with other parts of the app if necessary
  useEffect(() => {
    document.documentElement.classList.toggle('dark', isDarkMode)
  }, [isDarkMode])

  return (
    <SettingsContext.Provider value={{ isDarkMode, setDarkMode, language, setLanguage }}>
      {children}
    </SettingsContext.Provider>
  )
}
