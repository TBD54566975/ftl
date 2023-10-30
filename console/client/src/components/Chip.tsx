export const Chip = ({ name, className }: { name: string; className?: string }) => {
  return (
    <span
      className={`text-xs bg-indigo-100 text-indigo-800 rounded-full mr-1 px-2 py-0.5 border border-indigo-300 truncate min-w-0 flex-shrink dark:bg-indigo-900 dark:text-indigo-100 dark:border-indigo-700 ${className}`}
    >
      {name}
    </span>
  )
}
