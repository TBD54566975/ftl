import type { StoryObj } from '@storybook/react/*'
import { AttributeBadge } from '../components/AttributeBadge'

const meta = {
  title: 'Components/AttributeBadge',
  component: AttributeBadge,
}

export default meta
type Story = StoryObj<typeof meta>

export const Primary: Story = {
  args: {
    name: 'name',
    value: 'value',
  },
}
