import { classNames } from '../utils'

export const Pill = ({ text, className }: { text: string; className?: string }) => {
  return (
    <span
      className={classNames(
        'bg-gray-100 text-gray-500 dark:text-gray-400 dark:bg-gray-700 rounded-full px-2.5 py-0.5 text-xs font-medium inline-block',
        className,
      )}
    >
      {text}
    </span>
  )
}
