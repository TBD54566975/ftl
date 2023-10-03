import { ChevronRightIcon } from '@heroicons/react/20/solid'
import React from 'react'
import { classNames } from '../utils'

interface Breadcrumb {
  label: string
  link?: string
}

interface Props {
  icon?: React.ReactNode
  title: string
  children?: React.ReactNode
  breadcrumbs?: Breadcrumb[]
  className?: string
}

const Header = ({ icon, title, children, breadcrumbs, className }: Props) => {
  return (
    <div
      className={classNames(
        className,
        `flex-none w-full z-10 shadow dark:shadow-md flex justify-between items-center py-2 px-4 text-gray-70`,
      )}
    >
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

const Body: React.FC<{
  className?: string
  style?: React.CSSProperties
  children?: React.ReactNode
}> = ({ className, style, children }) => {
  return (
    <div className={classNames(className, 'flex-1')} style={{ height: 'calc(100% - 44px)', ...style }}>
      {children}
    </div>
  )
}

export const Page: React.FC<{
  className?: string
  style?: React.CSSProperties
  children?: React.ReactNode
}> & {
  Header: typeof Header
  Body: typeof Body
} = ({ className, style, children }) => {
  return (
    <div className={classNames(className, 'flex flex-col h-full')} style={style}>
      {children}
    </div>
  )
}

Page.Header = Header
Page.Body = Body
