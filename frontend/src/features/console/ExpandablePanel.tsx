import { ChevronDownIcon, ChevronRightIcon } from '@heroicons/react/24/outline'
import React, { useEffect, useState } from 'react'

export interface ExpandablePanelProps {
  icon?: React.ElementType
  title: string
  children: React.ReactNode
  expanded?: boolean
  padding?: string
}

export const ExpandablePanel: React.FC<ExpandablePanelProps> = ({ icon: Icon, title, children, expanded, padding }) => {
  const [isExpanded, setIsExpanded] = useState(expanded || false)

  useEffect(() => {
    setIsExpanded(expanded || false)
  }, [expanded])

  const toggleExpansion = () => {
    setIsExpanded(!isExpanded)
  }

  return (
    <div className='border-b border-gray-300 dark:border-gray-700'>
      <div className='bg-gray-100 dark:bg-gray-700 flex items-center cursor-pointer' onClick={toggleExpansion}>
        {isExpanded ? <ChevronDownIcon className='h-3 w-3 m-1' /> : <ChevronRightIcon className='h-3 w-3  m-1' />}
        {Icon ? <Icon className='h-4 w-4 mr-2' /> : null}
        <span className='text-xs py-1'>{title}</span>
      </div>
      {isExpanded && <div className={`overflow-auto max-h-64 ${padding ? padding : 'p-2'}`}>{children}</div>}
    </div>
  )
}
