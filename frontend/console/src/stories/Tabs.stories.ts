import type { StoryObj } from '@storybook/react/*'
import { Tabs } from '../components/Tabs'

const meta = {
  title: 'Components/Tabs',
  component: Tabs,
}

export default meta
type Story = StoryObj<typeof meta>

export const Primary: Story = {
  args: {
    tabs: [
      { name: 'First Tab', id: 'first', count: 3 },
      { name: 'Second Tab', id: 'second', count: 1 },
      { name: 'Third Tab', id: 'third' },
    ],
    initialTabId: 'second',
    onTabClick: (tabId: string) => console.log(tabId),
  },
}
