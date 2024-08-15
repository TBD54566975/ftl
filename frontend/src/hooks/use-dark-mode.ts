import { useEffect } from 'react'
import { useLocalStorage } from 'react-use'

export const useDarkMode = () => {
  const [isDarkMode, setDarkMode] = useLocalStorage('dark-mode', false)

  useEffect(() => {
    if (isDarkMode) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }, [isDarkMode])

  return { isDarkMode, setDarkMode }
}
