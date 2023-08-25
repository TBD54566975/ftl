import { createContext, useContext } from 'react'
import useLocalStorage from '../hooks/use-local-storage'

const DarkModeContext = createContext({
  isDarkMode: false,
  setDarkMode: () => {},
})

export const useDarkMode = () => {
  return useContext(DarkModeContext)
}

interface DarkModeProviderProps {
  children: React.ReactNode
}

export const DarkModeProvider = ({ children }: DarkModeProviderProps) => {
  const [isDarkMode, setDarkMode] = useLocalStorage('dark-mode', 'false')

  return <DarkModeContext.Provider value={{ isDarkMode, setDarkMode }}>{children}</DarkModeContext.Provider>
}
