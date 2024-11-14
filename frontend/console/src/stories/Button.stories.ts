import type { Meta, StoryObj } from '@storybook/react'
import { Button } from '../components/Button'

// Meta information for the Button component
const meta = {
  title: 'Components/Button',
  component: Button,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['primary', 'secondary'],
    },
    size: {
      control: 'select',
      options: ['xs', 'sm', 'md', 'lg', 'xl'],
    },
  },
} satisfies Meta<typeof Button>

export default meta
type Story = StoryObj<typeof Button>

// Default button story
export const Primary: Story = {
  args: {
    variant: 'primary',
    children: 'Button',
    size: 'md',
  },
}

// Secondary button story
export const Secondary: Story = {
  args: {
    variant: 'secondary',
    children: 'Button',
    size: 'md',
  },
}

// Small button story
export const Small: Story = {
  args: {
    size: 'sm',
    children: 'Button',
    variant: 'primary',
  },
}

// Large button story
export const Large: Story = {
  args: {
    size: 'lg',
    children: 'Button',
    variant: 'primary',
  },
}
