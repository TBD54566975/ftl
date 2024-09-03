import { classNames } from '../utils'

type ListProps<T> = {
  items: T[]
  renderItem: (item: T) => React.ReactNode
  onClick?: (item: T) => void
  className?: string
}

export const List = <T,>({ items, renderItem, onClick, className }: ListProps<T>) => {
  return (
    <ul className={classNames('divide-y divide-gray-100 dark:divide-gray-700 overflow-hidden', className)}>
      {items.map((item, index) => (
        <li
          key={index}
          className={`relative flex justify-between gap-x-6 p-4 ${onClick ? 'cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700' : ''}`}
          onClick={onClick ? () => onClick(item) : undefined}
        >
          {renderItem(item)}
        </li>
      ))}
    </ul>
  )
}
