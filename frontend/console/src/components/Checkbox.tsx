import type { InputHTMLAttributes, ReactNode } from 'react'

interface CheckboxProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
  label?: ReactNode
}

export const Checkbox = ({ label, className = '', ...props }: CheckboxProps) => {
  return (
    <label className='inline-flex items-center w-full'>
      <div className='group grid size-4 grid-cols-1 flex-shrink-0'>
        <input
          type='checkbox'
          aria-describedby='comments-description'
          className='col-start-1 row-start-1 appearance-none rounded border border-gray-300 bg-white dark:border-gray-600 dark:bg-gray-700 checked:border-indigo-600 checked:bg-indigo-600 dark:checked:border-indigo-600 dark:checked:bg-indigo-600 hover:bg-white focus:bg-white focus-visible:bg-white active:bg-white dark:hover:bg-gray-700 dark:focus:bg-gray-700 dark:focus-visible:bg-gray-700 dark:active:bg-gray-700 checked:hover:bg-indigo-500 checked:focus:bg-indigo-600 dark:checked:hover:bg-indigo-500 dark:checked:focus:bg-indigo-600 disabled:border-gray-300 disabled:bg-gray-100 disabled:checked:bg-gray-100 dark:disabled:border-gray-700 dark:disabled:bg-gray-900 dark:disabled:checked:bg-gray-700 forced-colors:appearance-auto focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 focus:ring-0 focus:ring-offset-0'
          {...props}
        />
        <svg
          aria-hidden='true'
          fill='none'
          viewBox='0 0 14 14'
          className='pointer-events-none col-start-1 row-start-1 size-3.5 self-center justify-self-center stroke-white group-has-[:disabled]:stroke-gray-950/25'
        >
          <path d='M3 8L6 11L11 3.5' strokeWidth={2} strokeLinecap='round' strokeLinejoin='round' className='opacity-0 group-has-[:checked]:opacity-100' />
          <path d='M3 7H11' strokeWidth={2} strokeLinecap='round' strokeLinejoin='round' className='opacity-0 group-has-[:indeterminate]:opacity-100' />
        </svg>
      </div>
      {label && <div className='ml-2 text-sm dark:text-gray-300 flex-grow'>{label}</div>}
    </label>
  )
}
