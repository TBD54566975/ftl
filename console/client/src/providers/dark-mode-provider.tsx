import React, { useContext } from 'react'
import useLocalStorage from '../hooks/use-local-storage'

const DarkModeContext = React.createContext({
  isDarkMode: false,
  setDarkMode: (_: boolean) => {},
})

export const useDarkMode = () => {
  return useContext(DarkModeContext)
}

interface Props {
  children: React.ReactNode
}

export const DarkModeProvider = ({ children }: Props) => {
  const [isDarkMode, setDarkMode] = useLocalStorage('dark-mode', 'false')
  const setMode = (val: boolean) => {
    setDarkMode(`${val}`)
  }
  return (
    <DarkModeContext.Provider value={{ isDarkMode: isDarkMode === 'true', setDarkMode: setMode }}>
      {children}
    </DarkModeContext.Provider>
  )
}
