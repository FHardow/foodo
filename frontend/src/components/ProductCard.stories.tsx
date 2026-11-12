import type { Meta, StoryObj } from '@storybook/react-vite'
import { MemoryRouter } from 'react-router-dom'
import { fn } from 'storybook/test'
import ProductCard from './ProductCard'
import type { Product } from '../types'

const base: Product = {
  id: '1',
  name: 'Heirloom Tomatoes',
  description: 'Sun-ripened, mixed varieties from the greenhouse.',
  unit: '500 g',
  available: true,
  categories: ['Vegetables'],
}

const meta = {
  title: 'Components/ProductCard',
  component: ProductCard,
  decorators: [
    (Story, ctx) => (
      <MemoryRouter>
        <div className={`bg-[#faf6f1] p-4 ${ctx.parameters.wide ? 'w-[1000px]' : 'w-60'}`}>
          <Story />
        </div>
      </MemoryRouter>
    ),
  ],
  args: { product: base, onAdd: fn() },
} satisfies Meta<typeof ProductCard>

export default meta
type Story = StoryObj<typeof meta>

export const WithCategory: Story = {}

export const MultipleCategories: Story = {
  args: { product: { ...base, categories: ['Vegetables', 'Organic', 'Greenhouse'] } },
}

export const Overflow: Story = {
  args: {
    product: { ...base, categories: ['Vegetables', 'Organic', 'Greenhouse', 'Seasonal', 'Local'] },
  },
}

export const WithoutCategory: Story = {
  args: { product: { ...base, categories: [] } },
}

export const LongCategory: Story = {
  args: { product: { ...base, name: 'Wildflower Honey', categories: ['Preserves & Spreads'], unit: '250 g jar' } },
}

export const Disabled: Story = {
  args: { disabled: true },
}

export const Grid: Story = {
  parameters: { wide: true },
  render: (args) => (
    <div className="grid grid-cols-4 gap-4">
      {[
        { ...base, categories: ['Vegetables', 'Organic'] },
        { ...base, id: '2', name: 'Strawberries', categories: ['Fruit', 'Seasonal'], unit: '250 g' },
        { ...base, id: '3', name: 'Raw Milk', categories: ['Dairy'], unit: '1 L' },
        { ...base, id: '4', name: 'Sourdough Loaf', categories: ['Bakery'], unit: '1 loaf' },
        { ...base, id: '5', name: 'Free-range Eggs', categories: ['Eggs', 'Local', 'Organic', 'Seasonal'], unit: '6 pcs' },
        { ...base, id: '6', name: 'Fresh Basil', categories: [], unit: 'bunch' },
      ].map((p) => (
        <ProductCard key={p.id} product={p} onAdd={args.onAdd} />
      ))}
    </div>
  ),
}
