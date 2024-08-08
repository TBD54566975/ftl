import { StoryObj } from '@storybook/react/*'
import { Badge } from '../components/Badge'

const meta = {
  title: 'Components/Badge',
  component: Badge,
}

export default meta
type Story = StoryObj<typeof meta>;

export const Primary: Story = {
  args: {
    name: 'name',
  },
}
