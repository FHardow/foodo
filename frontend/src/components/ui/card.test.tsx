import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Card, CardHeader, CardTitle, CardContent } from './card'

describe('Card', () => {
  it('renders title and content', () => {
    render(
      <Card>
        <CardHeader>
          <CardTitle>Sourdough Loaf</CardTitle>
        </CardHeader>
        <CardContent>1 loaf</CardContent>
      </Card>,
    )
    expect(screen.getByText('Sourdough Loaf')).toBeInTheDocument()
    expect(screen.getByText('1 loaf')).toBeInTheDocument()
  })
})
