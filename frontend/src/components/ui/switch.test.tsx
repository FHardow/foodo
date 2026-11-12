import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { Switch } from './switch'

describe('Switch', () => {
  it('reports the new checked state via onCheckedChange', async () => {
    const onCheckedChange = vi.fn()
    render(<Switch aria-label="Available" checked={false} onCheckedChange={onCheckedChange} />)
    await userEvent.click(screen.getByRole('switch', { name: 'Available' }))
    expect(onCheckedChange).toHaveBeenCalledWith(true)
  })
})
