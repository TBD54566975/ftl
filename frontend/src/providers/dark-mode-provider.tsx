import React, { PropsWithChildren, useContext } from 'react'
import useLocalStorage from '../hooks/use-local-storage'

const DarkModeContext = React.createContext({
  isDarkMode: false,
  setDarkMode: (_: boolean) => {},
})

export const useDarkMode = () => {
  return useContext(DarkModeContext)
}

export const DarkModeProvider = ({ children }: PropsWithChildren) => {
  const [isDarkMode, setDarkMode] = useLocalStorage('dark-mode', false)

  return (
    <DarkModeContext.Provider value={{ isDarkMode: isDarkMode, setDarkMode: setDarkMode }}>
      {children}
    </DarkModeContext.Provider>
  )
}
