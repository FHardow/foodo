import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { Textarea } from './textarea'

describe('Textarea', () => {
  it('accepts typed text', async () => {
    render(<Textarea aria-label="Description" />)
    const textarea = screen.getByLabelText('Description')
    await userEvent.type(textarea, 'Classic tangy sourdough')
    expect(textarea).toHaveValue('Classic tangy sourdough')
  })
})
