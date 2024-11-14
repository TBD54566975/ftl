import type { ButtonHTMLAttributes, ReactNode } from 'react'
import { classNames } from '../utils'

type ButtonSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode
  size?: ButtonSize
  variant?: 'primary' | 'secondary'
  fullWidth?: boolean
  title?: string
}

const sizeClasses: Record<ButtonSize, string> = {
  xs: 'rounded px-2 py-1 text-xs',
  sm: 'rounded px-2 py-1 text-sm',
  md: 'rounded-md px-2.5 py-1.5 text-sm',
  lg: 'rounded-md px-3 py-2 text-sm',
  xl: 'rounded-md px-3.5 py-2.5 text-sm',
}

export const Button = ({ children, size = 'md', variant = 'primary', fullWidth = false, className, title, ...props }: ButtonProps) => {
  const baseClasses = 'font-semibold shadow-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2'
  const variantClasses = {
    primary:
      'bg-indigo-600 dark:bg-indigo-500 text-white hover:bg-indigo-500 dark:hover:bg-indigo-400 focus-visible:outline-indigo-600 dark:focus-visible:outline-indigo-500',
    secondary:
      'bg-white text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus-visible:outline-gray-600 dark:bg-white/10 dark:text-white dark:hover:bg-white/20 dark:ring-0',
  }

  return (
    <button
      type='button'
      className={classNames(baseClasses, sizeClasses[size], variantClasses[variant], fullWidth ? 'w-full' : '', className)}
      title={title}
      {...props}
    >
      {children}
    </button>
  )
}
