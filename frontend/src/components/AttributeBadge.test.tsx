import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { AttributeBadge } from './AttributeBadge'

describe('AttributeBadge', () => {
  it('renders the name and value correctly', () => {
    render(<AttributeBadge name='Role' value='Admin' />)
    expect(screen.getByText('Role')).toBeInTheDocument()
    expect(screen.getByText('Admin')).toBeInTheDocument()
    expect(screen.getByText(':')).toBeInTheDocument()
  })

  it('passes additional props to the span element', () => {
    const { container } = render(<AttributeBadge name='Role' value='Admin' data-testid='badge' />)
    const spanElement = container.querySelector('span')
    expect(spanElement).toHaveAttribute('data-testid', 'badge')
  })
})
