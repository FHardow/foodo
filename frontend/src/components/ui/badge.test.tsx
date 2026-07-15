import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Badge } from './badge'

describe('Badge', () => {
  it('renders its label text', () => {
    render(<Badge>finished</Badge>)
    expect(screen.getByText('finished')).toBeInTheDocument()
  })

  it('merges a custom className with the variant classes', () => {
    render(<Badge className="bg-green-100 text-green-800">finished</Badge>)
    expect(screen.getByText('finished').className).toContain('bg-green-100')
  })
})
