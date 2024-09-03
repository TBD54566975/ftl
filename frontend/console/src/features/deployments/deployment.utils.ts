import hash from 'fnv1a'

const colorNames = [
  'text-amber-500 dark:text-amber-400',
  'text-blue-500 dark:text-blue-400',
  'text-cyan-500 dark:text-cyan-400',
  'text-emerald-500 dark:text-emerald-400',
  'text-fuchsia-500 dark:text-fuchsia-400',
  'text-green-500 dark:text-green-400',
  'text-indigo-500 dark:text-indigo-400',
  'text-lime-500 dark:text-lime-400',
  'text-orange-500 dark:text-orange-400',
  'text-pink-500 dark:text-pink-400',
  'text-purple-500 dark:text-purple-400',
  'text-sky-500 dark:text-sky-400',
  'text-slate-500 dark:text-slate-400',
  'text-teal-500 dark:text-teal-400',
  'text-violet-500 dark:text-violet-400',
  'text-yellow-500 dark:text-yellow-400',
]

export const deploymentTextColor = (name: string) => colorNames[hash(name) % colorNames.length]
