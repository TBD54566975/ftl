import React from 'react'
import { ChevronRightIcon } from '@heroicons/react/20/solid'
import { classNames } from '../utils'

const Icon: React.FC<{ children: React.ReactNode, className?: string}> = ({children, className}) => ( <span className={classNames(className, 'mt-1 text-indigo-500 mr-2 mb-1 h-5 w-5')}>{children}</span>)

const Breadcrumb: React.FC<{
  className?: string,
  data:{
  label: string
  link?: string
}[]}> = ({ className, data}) => (
  <nav className={classNames(className, 'flex pr-2 text-lg')} aria-label='Breadcrumb'>
  <ol role='list' className='flex items-center space-x-2'>
    {data.map((crumb, index) => (
      <li key={index}>
        <div className='flex items-center'>
          <a href={crumb.link || '#'} className='mr-2 hover:text-indigo-500'> 
            {crumb.label}
          </a>
          <ChevronRightIcon className='mt-0.5 h-5 w-5' />
        </div>
      </li>
    ))}
  </ol>
</nav>
)

const Title: React.FC<{ children: React.ReactNode, className?:string }> = ({ children, className }) => (<span className={classNames(className, 'text-lg')}>{children}</span>)

const Text: React.FC<{ children: React.ReactNode, className?:string }> = ({ children, className }) => (<span className={classNames(className, 'text-base')}>{children}</span>)

export const ControlBar: React.FC<{ className?: string, children?: React.ReactNode}> & {
  Icon: typeof Icon
  Breadcrumb: typeof Breadcrumb
  Title: typeof Title
  Text: typeof Text
} = ({className, children}) => {
  return (
  <div className={classNames(className, `shadow dark:shadow-md flex items-center py-2 px-4 text-gray-70`)}>
    <div className='flex items-center'>
      {children}
    </div>
  </div>
  )
} 

ControlBar.Icon = Icon
ControlBar.Breadcrumb = Breadcrumb
ControlBar.Title = Title
ControlBar.Text = Text