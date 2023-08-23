import { createContext, useContext } from 'react'
import useLocalStorage from '../hooks/use-local-storage'

const DarkModeContext = createContext({ isDarkMode: false, setDarkMode: () => { } })

export const useDarkMode = () => {
  return useContext(DarkModeContext)
}

export const DarkModeProvider = ({ children }) => {
  const [ isDarkMode, setDarkMode ] = useLocalStorage('dark-mode', false)

  return (
    <DarkModeContext.Provider value={{ isDarkMode, setDarkMode }}>
      {children}
    </DarkModeContext.Provider>
  )
}
