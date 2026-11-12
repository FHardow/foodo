import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Label } from './label'
import { Input } from './input'

describe('Label', () => {
  it('associates with its input via htmlFor', () => {
    render(
      <>
        <Label htmlFor="name">Name</Label>
        <Input id="name" />
      </>,
    )
    expect(screen.getByLabelText('Name')).toBeInTheDocument()
  })
})
