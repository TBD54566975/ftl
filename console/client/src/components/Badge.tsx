export const Badge = ({ name, className }: { name: string; className?: string }) => {
  return (
    <span
      className={`inline-flex items-center rounded-lg bg-gray-100 dark:bg-gray-700 px-2 py-1 text-xs font-medium text-gray-600 dark:text-gray-300 ${className}`}
    >
      {name}
    </span>
  )
}
