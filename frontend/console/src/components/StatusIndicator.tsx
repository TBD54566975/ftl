import type React from 'react'
import { classNames } from '../utils'

type StatusIndicatorProps = {
  state: 'success' | 'error' | 'idle'
  text?: string
}

export const StatusIndicator: React.FC<StatusIndicatorProps> = ({ state, text }) => {
  const backgrounds = {
    idle: 'text-gray-500 bg-gray-500/20 dark:bg-gray-100/10',
    success: 'text-green-500 bg-emerald-500/20 dark:text-green-400 dark:bg-green-400/10 ',
    error: 'text-rose-500 bg-rose-500/20 dark:text-rose-400 dark:bg-rose-400/10',
  }

  return (
    <div className='flex items-center gap-x-1.5'>
      <div className={classNames(backgrounds[state], 'flex-none rounded-full p-1')}>
        <div className='h-1.5 w-1.5 rounded-full bg-current' />
      </div>
      {text && <p className='text-xs leading-5 text-gray-500'>{text}</p>}
    </div>
  )
}
