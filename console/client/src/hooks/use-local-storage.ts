import React from 'react'

export default function useLocalStorage(
  key: string,
  initialValue: boolean
): [boolean, React.Dispatch<React.SetStateAction<boolean>>] {
  const [value, setValue] = React.useState<boolean>(() => {
    const jsonValue = localStorage.getItem(key)
    if (jsonValue != null) {
      const value = JSON.parse(jsonValue) as string
      return value === 'true'
    }
    return initialValue
  })

  React.useEffect(() => {
    localStorage.setItem(key, JSON.stringify(value))
  }, [key, value])

  return [value, setValue]
}
