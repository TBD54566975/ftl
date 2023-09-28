import { ChevronRightIcon } from '@heroicons/react/20/solid'
import React from 'react'

interface Breadcrumb {
  label: string
  link?: string
}

interface Props {
  icon?: React.ReactNode
  title: string
  children?: React.ReactNode
  breadcrumbs?: Breadcrumb[]
}

export const PageHeader = ({ icon, title, children, breadcrumbs }: Props) => {
  return (
    <div className={`sticky top-0 z-10 shadow dark:shadow-md flex justify-between items-center py-2 px-4 text-gray-70`}>
      <div className='flex items-center'>
        <span className='mt-1 text-indigo-500 mr-2 mb-1 h-5 w-5'>{icon}</span>
        {breadcrumbs && breadcrumbs.length > 0 && (
          <nav className='flex pr-2' aria-label='Breadcrumb'>
            <ol role='list' className='flex items-center space-x-2'>
              {breadcrumbs.map((crumb, index) => (
                <li key={index}>
                  <div className='flex items-center'>
                    <a href={crumb.link || '#'} className='text-lg mr-2 hover:text-indigo-500'>
                      {crumb.label}
                    </a>
                    <ChevronRightIcon className='mt-0.5 h-5 w-5' />
                  </div>
                </li>
              ))}
            </ol>
          </nav>
        )}
        <span className='text-lg'>{title}</span>
      </div>
      {children}
    </div>
  )
}
