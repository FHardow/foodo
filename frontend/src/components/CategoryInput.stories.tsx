import type { Meta, StoryObj } from '@storybook/react-vite'
import { useState } from 'react'
import { expect, fn, userEvent, within } from 'storybook/test'
import CategoryInput from './CategoryInput'
import type { Category } from '../types'

const ALL: Category[] = [
  'Bakery', 'Berries', 'Beverages', 'Dairy', 'Eggs', 'Fruit', 'Greenhouse', 'Herbs',
  'Honey', 'Local', 'Meat', 'Organic', 'Preserves & Spreads', 'Seasonal', 'Vegetables',
].map((name, i) => ({ id: String(i), name }))

/** Stateful wrapper so the story behaves like the real form. */
function Playground({
  initial,
  onChange,
  onDelete,
  withDelete = true,
}: {
  initial: string[]
  onChange: (v: string[]) => void
  onDelete: (c: Category) => void
  withDelete?: boolean
}) {
  const [value, setValue] = useState(initial)
  const [categories, setCategories] = useState(ALL)
  return (
    <div className="w-96 bg-white p-6 rounded-lg border border-[#e8ddd0] space-y-1">
      <label htmlFor="cats" className="block text-xs font-medium text-[#8a6a50] mb-1">
        Categories
      </label>
      <CategoryInput
        id="cats"
        value={value}
        categories={categories}
        onChange={(v) => {
          setValue(v)
          onChange(v)
        }}
        onDeleteCategory={
          withDelete
            ? (c) => {
                setCategories((cs) => cs.filter((x) => x.id !== c.id))
                setValue((v) => v.filter((n) => n !== c.name))
                onDelete(c)
              }
            : undefined
        }
      />
      <p className="text-[11px] text-[#a88a6e]">Enter, Tab or comma to add. Backspace twice removes the last one.</p>
      <div className="h-72" />
    </div>
  )
}

const meta = {
  title: 'Components/CategoryInput',
  component: Playground,
  args: { initial: ['Vegetables', 'Organic'], onChange: fn(), onDelete: fn() },
  parameters: { layout: 'centered' },
} satisfies Meta<typeof Playground>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const Empty: Story = { args: { initial: [] } }

export const WithoutDelete: Story = { args: { withDelete: false } }

export const Full: Story = {
  args: { initial: ['Bakery', 'Berries', 'Dairy', 'Eggs', 'Fruit', 'Herbs', 'Honey', 'Local', 'Meat', 'Organic'] },
}

/** Typing "vg" fuzzy-matches "Vegetables"; here it's already selected so other hits show. */
export const FuzzySearch: Story = {
  args: { initial: [] },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.type(canvas.getByRole('combobox'), 'vg')
    await expect(canvas.getByRole('option', { name: /Vegetables/ })).toBeInTheDocument()
  },
}

export const TypoTolerant: Story = {
  args: { initial: [] },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.type(canvas.getByRole('combobox'), 'furit')
    await expect(canvas.getByRole('option', { name: /Fruit/ })).toBeInTheDocument()
  },
}

export const CreateNew: Story = {
  args: { initial: [] },
  play: async ({ canvasElement, args }) => {
    const canvas = within(canvasElement)
    await userEvent.type(canvas.getByRole('combobox'), 'Gluten free')
    await expect(canvas.getByRole('option', { name: /Create/ })).toBeInTheDocument()
    await userEvent.keyboard('{ArrowUp}{Enter}') // last option = Create
    await expect(args.onChange).toHaveBeenLastCalledWith(['Gluten free'])
  },
}

export const BackspaceHighlightsThenRemoves: Story = {
  play: async ({ canvasElement, args }) => {
    const canvas = within(canvasElement)
    await userEvent.click(canvas.getByRole('combobox'))
    await userEvent.keyboard('{Backspace}')
    await expect(args.onChange).not.toHaveBeenCalled()
    await userEvent.keyboard('{Backspace}')
    await expect(args.onChange).toHaveBeenLastCalledWith(['Vegetables'])
  },
}
