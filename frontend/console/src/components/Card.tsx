interface Props {
  topBarColor?: string
  className?: string
  onClick?: () => void
  children: React.ReactNode
}

export const Card = ({ topBarColor, className, onClick, children }: Props) => {
  return (
    <div
      onClick={onClick}
      className={`relative rounded-md border border-gray-200 dark:border-gray-500 ${onClick ? 'cursor-pointer hover:bg-gray-100 dark:hover:bg-slate-700' : ''} ${className}`}
    >
      {topBarColor && <div className='absolute top-0 left-0 right-0 h-1 bg-green-400 rounded-t-md -mt-px -ml-px -mr-px' />}

      <div className={`${topBarColor ? 'mt-1' : ''} p-2`}>{children}</div>
    </div>
  )
}
