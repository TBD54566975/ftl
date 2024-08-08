import { StoryObj } from '@storybook/react/*'
import { Chip } from '../components/Chip'

const meta = {
  title: 'Components/Chip',
  component: Chip,
}

export default meta
type Story = StoryObj<typeof meta>;

export const Primary: Story = {
  args: {
    name: 'name',
  },
}
