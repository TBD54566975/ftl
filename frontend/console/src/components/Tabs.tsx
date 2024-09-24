import { type FC, useEffect, useState } from 'react'
import { classNames } from '../utils'
import { Pill } from './Pill'

interface Tab {
  name: string
  id: string
  count?: number
}

interface TabsProps {
  tabs: Tab[]
  initialTabId?: string
  onTabClick?: (tabId: string) => void
}

export const Tabs: FC<TabsProps> = ({ tabs, initialTabId, onTabClick }) => {
  const [selectedTabId, setSelectedTabId] = useState<string | undefined>(initialTabId || tabs[0]?.id)

  useEffect(() => {
    if (initialTabId) {
      setSelectedTabId(initialTabId)
    }
  }, [initialTabId])

  const handleTabClick = (tabId: string) => {
    setSelectedTabId(tabId)
    if (onTabClick) {
      onTabClick(tabId)
    }
  }

  return (
    <div className='border-b border-gray-200 dark:border-white/10'>
      <nav aria-label='Tabs' className='-mb-px flex space-x-8'>
        {tabs.map((tab) => (
          <button
            key={tab.id}
            type='button'
            aria-current={selectedTabId === tab.id ? 'page' : undefined}
            onClick={() => handleTabClick(tab.id)}
            className={classNames(
              selectedTabId === tab.id
                ? 'border-indigo-500 text-indigo-500'
                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-indigo-400 hover:border-gray-200 dark:hover:border-indigo-400 ',
              'flex whitespace-nowrap border-b-2 px-1 py-4 text-sm font-medium',
            )}
          >
            {tab.name}
            {tab.count !== undefined && (
              <Pill
                text={`${tab.count}`}
                className={classNames(selectedTabId === tab.id ? 'bg-indigo-100 text-indigo-600 dark:bg-indigo-500 dark:text-indigo-100' : '', 'ml-2')}
              />
            )}
          </button>
        ))}
      </nav>
    </div>
  )
}
