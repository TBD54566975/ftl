import { PropsWithChildren } from 'react'

export const Card = (props: PropsWithChildren) => {
  return (
    <div
      className={`relative flex items-center space-x-3 rounded-lg border border-gray-300 bg-white dark:bg-slate-800 dark:border-indigo-400 px-6 py-5 shadow-sm focus-within:ring-2 focus-within:ring-indigo-500 focus-within:ring-offset-2 dark:focus-within:ring-2 dark:focus-within:ring-indigo-400 dark:focus-within:ring-offset-2 hover:border-gray-400 dark:hover:border-indigo-200`}
    >
      {props.children}
    </div>
  )
}
