import { ChevronDownIcon, ChevronRightIcon, InformationCircleIcon } from '@heroicons/react/24/outline'
import React, { useState } from 'react'

export interface ExpandablePanelProps {
  title: string
  expanded?: boolean
  children: React.ReactNode
}

export const ExpandablePanel: React.FC<ExpandablePanelProps> = ({ title, expanded, children }) => {
  const [isExpanded, setIsExpanded] = useState(expanded || false)

  const toggleExpansion = () => {
    setIsExpanded(!isExpanded)
  }

  return (
    <div className='border border-gray-700'>
      <div className='bg-gray-700 flex items-center cursor-pointer' onClick={toggleExpansion}>
        {isExpanded ? (
          <ChevronDownIcon className='h-3 w-3 text-white m-1' />
        ) : (
          <ChevronRightIcon className='h-3 w-3 text-white m-1' />
        )}
        <InformationCircleIcon className='h-4 w-4 text-white mr-2' />
        <span className='text-white'>{title}</span>
      </div>
      {isExpanded && <div className='p-2'>{children}</div>}
    </div>
  )
}
