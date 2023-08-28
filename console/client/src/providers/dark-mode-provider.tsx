import React from 'react'
import useLocalStorage from '../hooks/use-local-storage'

const DarkModeContext = React.createContext({
  isDarkMode: false,
  setDarkMode: (() => {}) as React.Dispatch<React.SetStateAction<boolean>>,
})

export const useDarkMode = () => {
  return React.useContext(DarkModeContext)
}

type DarkModeProviderProps = {
  children: React.ReactNode
}

export const DarkModeProvider = ({children}: DarkModeProviderProps) => {
  const [isDarkMode, setDarkMode] = useLocalStorage('dark-mode', false)

  return (
    <DarkModeContext.Provider value={{isDarkMode, setDarkMode}}>
      {children}
    </DarkModeContext.Provider>
  )
}
