import { useEffect, useState } from 'react'

export interface KeyValuePair {
  id: string
  enabled: boolean
  key: string
  value: string
}

export const useKeyValuePairs = (storageKey: string | null) => {
  const [pairs, setPairs] = useState<KeyValuePair[]>([])

  useEffect(() => {
    if (storageKey) {
      const savedValue = localStorage.getItem(storageKey)
      try {
        const parsedValue = savedValue ? JSON.parse(savedValue) : []
        setPairs(parsedValue.length > 0 ? parsedValue : [])
      } catch {
        setPairs([])
      }
    }
  }, [storageKey])

  const updatePairs = (newPairs: KeyValuePair[]) => {
    setPairs(newPairs)
    if (storageKey) {
      localStorage.setItem(storageKey, JSON.stringify(newPairs))
    }
  }

  const getPairsObject = () =>
    pairs
      .filter((p) => p.enabled)
      .reduce<Record<string, string>>((acc, p) => {
        acc[p.key] = p.value
        return acc
      }, {})

  return {
    pairs,
    updatePairs,
    getPairsObject,
  }
}
