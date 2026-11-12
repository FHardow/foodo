import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { Input } from './input'

describe('Input', () => {
  it('accepts typed text', async () => {
    render(<Input aria-label="Name" />)
    const input = screen.getByLabelText('Name')
    await userEvent.type(input, 'Sourdough')
    expect(input).toHaveValue('Sourdough')
  })
})
