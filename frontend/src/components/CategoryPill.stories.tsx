import type { Meta, StoryObj } from '@storybook/react-vite'
import CategoryPill from './CategoryPill'

const meta = {
  title: 'Components/CategoryPill',
  component: CategoryPill,
  args: { category: 'Vegetables' },
} satisfies Meta<typeof CategoryPill>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const AllColours: Story = {
  render: () => (
    <div className="flex flex-wrap gap-2">
      {['Vegetables', 'Fruit', 'Dairy', 'Bakery', 'Meat', 'Eggs', 'Honey', 'Herbs', 'Drinks', 'Preserves'].map((c) => (
        <CategoryPill key={c} category={c} />
      ))}
    </div>
  ),
}
