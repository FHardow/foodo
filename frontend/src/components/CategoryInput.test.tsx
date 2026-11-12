import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import CategoryInput, { MAX_CATEGORIES } from './CategoryInput'
import type { Category } from '../types'

const cats: Category[] = ['Bakery', 'Dairy', 'Fruit', 'Vegetables'].map((name, i) => ({ id: String(i), name }))

function Harness(props: { initial?: string[]; onChange?: (v: string[]) => void; onDelete?: (c: Category) => void }) {
  const [value, setValue] = useState(props.initial ?? [])
  return (
    <CategoryInput
      value={value}
      categories={cats}
      onChange={(v) => {
        setValue(v)
        props.onChange?.(v)
      }}
      onDeleteCategory={props.onDelete}
    />
  )
}

describe('CategoryInput', () => {
  it('fuzzy-matches and adds the top suggestion on Enter', async () => {
    const onChange = vi.fn()
    render(<Harness onChange={onChange} />)
    await userEvent.type(screen.getByRole('combobox'), 'vgt{Enter}')
    expect(onChange).toHaveBeenLastCalledWith(['Vegetables'])
    expect(screen.getByText('Vegetables')).toBeInTheDocument()
  })

  it('reuses existing casing when typing an exact name', async () => {
    const onChange = vi.fn()
    render(<Harness onChange={onChange} />)
    await userEvent.type(screen.getByRole('combobox'), 'dairy,')
    expect(onChange).toHaveBeenLastCalledWith(['Dairy'])
  })

  it('offers to create a new category', async () => {
    const onChange = vi.fn()
    render(<Harness onChange={onChange} />)
    await userEvent.type(screen.getByRole('combobox'), 'Zzz')
    await userEvent.click(screen.getByRole('option', { name: /Create/ }))
    expect(onChange).toHaveBeenLastCalledWith(['Zzz'])
  })

  it('does not suggest already-selected categories', async () => {
    render(<Harness initial={['Fruit']} />)
    await userEvent.type(screen.getByRole('combobox'), 'fru')
    expect(screen.queryByRole('option', { name: /Fruit/ })).not.toBeInTheDocument()
  })

  it('removes a pill with its × button', async () => {
    const onChange = vi.fn()
    render(<Harness initial={['Fruit', 'Dairy']} onChange={onChange} />)
    await userEvent.click(screen.getByRole('button', { name: 'Remove Fruit' }))
    expect(onChange).toHaveBeenLastCalledWith(['Dairy'])
  })

  it('backspace on empty input arms then removes the last pill', async () => {
    const onChange = vi.fn()
    render(<Harness initial={['Fruit', 'Dairy']} onChange={onChange} />)
    await userEvent.click(screen.getByRole('combobox'))
    await userEvent.keyboard('{Backspace}')
    expect(onChange).not.toHaveBeenCalled()
    await userEvent.keyboard('{Backspace}')
    expect(onChange).toHaveBeenLastCalledWith(['Fruit'])
  })

  it('Enter with text does not submit the surrounding form', async () => {
    const onSubmit = vi.fn((e: React.FormEvent) => e.preventDefault())
    render(
      <form onSubmit={onSubmit}>
        <Harness />
      </form>,
    )
    await userEvent.type(screen.getByRole('combobox'), 'Bak{Enter}')
    expect(onSubmit).not.toHaveBeenCalled()
  })

  it('deletes a category everywhere after confirmation', async () => {
    const onDelete = vi.fn()
    render(<Harness onDelete={onDelete} />)
    await userEvent.type(screen.getByRole('combobox'), 'bak')
    await userEvent.click(screen.getByRole('button', { name: 'Delete category Bakery' }))
    expect(onDelete).not.toHaveBeenCalled()
    await userEvent.click(screen.getByRole('button', { name: 'Yes' }))
    expect(onDelete).toHaveBeenCalledWith(cats[0])
  })

  it('disables input when the maximum is reached', () => {
    const many = Array.from({ length: MAX_CATEGORIES }, (_, i) => `c${i}`)
    render(<Harness initial={many} />)
    expect(screen.getByRole('combobox')).toBeDisabled()
  })
})
