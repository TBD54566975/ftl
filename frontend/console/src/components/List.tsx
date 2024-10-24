import { ArrowRight01Icon } from 'hugeicons-react'
import { Link } from 'react-router-dom'
import { classNames } from '../utils'

type ListProps<T> = {
  items: T[]
  renderItem: (item: T) => React.ReactNode
  href?: (item: T) => string
  className?: string
}

export const List = <T,>({ items, renderItem, href, className }: ListProps<T>) => {
  const baseClasses = 'relative flex justify-between items-center gap-x-4 p-4'
  return (
    <ul className={classNames('divide-y divide-gray-100 dark:divide-gray-700 overflow-hidden', className)}>
      {items.map((item, index) =>
        href ? (
          <Link key={index} className={`${baseClasses} cursor-pointer hover:bg-gray-100/50 dark:hover:bg-gray-700/50`} to={href(item)}>
            {renderItem(item)}
            <ArrowRight01Icon className='size-5 text-gray-400' />
          </Link>
        ) : (
          <li key={index} className={baseClasses}>
            {renderItem(item)}
          </li>
        ),
      )}
    </ul>
  )
}
