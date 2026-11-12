import { describe, it, expect } from 'vitest'
import { fuzzyMatch, fuzzyFilter } from './fuzzy'

describe('fuzzyMatch', () => {
  it('matches subsequences case-insensitively and returns indices', () => {
    expect(fuzzyMatch('vgt', 'Vegetables')?.indices).toEqual([0, 2, 4])
  })

  it('returns null when characters are missing', () => {
    expect(fuzzyMatch('xyz', 'Vegetables')).toBeNull()
  })

  it('tolerates a single typo for queries of 3+ chars', () => {
    expect(fuzzyMatch('furit', 'Fruit')).not.toBeNull()
    expect(fuzzyMatch('frt', 'Fruit')).not.toBeNull() // subsequence
    expect(fuzzyMatch('fx', 'Fruit')).toBeNull() // too short for typo tolerance
  })

  it('empty query matches everything', () => {
    expect(fuzzyMatch('', 'Dairy')).toEqual({ score: 0, indices: [] })
  })
})

describe('fuzzyFilter', () => {
  const cats = ['Bakery', 'Berries', 'Beverages', 'Dairy', 'Herbs']

  it('ranks prefix matches above scattered matches', () => {
    const names = fuzzyFilter('be', cats, (c) => c).map((r) => r.item)
    expect(names.slice(0, 2)).toEqual(['Berries', 'Beverages'])
    expect(names).toContain('Bakery') // b…e as a subsequence
    expect(names).not.toContain('Herbs') // no 'e' after the 'b'
  })

  it('ranks exact match first', () => {
    expect(fuzzyFilter('dairy', cats, (c) => c)[0].item).toBe('Dairy')
  })

  it('ranks word-start hits above mid-word hits', () => {
    const r = fuzzyFilter('br', ['Bakery Rolls', 'Fibre'], (c) => c)
    expect(r[0].item).toBe('Bakery Rolls')
  })
})
