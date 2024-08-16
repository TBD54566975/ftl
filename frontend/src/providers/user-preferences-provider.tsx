import { type PropsWithChildren, createContext, useContext, useEffect } from 'react'
import useLocalStorage from '../hooks/use-local-storage'

interface UserPreferencesContextProps {
  isDarkMode: boolean
  setDarkMode: (value: boolean) => void
}

const UserPreferencesContext = createContext<UserPreferencesContextProps | undefined>(undefined)

export const useUserPreferences = () => {
  const context = useContext(UserPreferencesContext)
  if (!context) {
    throw new Error('useSettings must be used within a UserPreferencesProvider')
  }
  return context
}

export const UserPreferencesProvider = ({ children }: PropsWithChildren) => {
  const [isDarkMode, setDarkMode] = useLocalStorage('dark-mode', false)

  useEffect(() => {
    if (isDarkMode) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }, [isDarkMode])

  return <UserPreferencesContext.Provider value={{ isDarkMode, setDarkMode }}>{children}</UserPreferencesContext.Provider>
}
