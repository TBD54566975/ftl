interface DividerProps {
  vertical?: boolean
  className?: string
}

export const Divider: React.FC<DividerProps> = ({ vertical = false, className }) => {
  const baseClass = vertical ? 'border-l' : 'border-t'

  return <div className={`${baseClass} border-gray-100 dark:border-gray-700/50 ${className}`} />
}
