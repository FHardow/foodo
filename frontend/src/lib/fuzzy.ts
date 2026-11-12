export interface FuzzyMatch {
  score: number
  /** Indices in the candidate string that matched the query (for highlighting). */
  indices: number[]
}

const isWordStart = (text: string, i: number) => i === 0 || /[\s\-_/&]/.test(text[i - 1])

/**
 * fzf-style subsequence match: every query character must appear in order in
 * the candidate. Consecutive runs and word-start hits score higher, gaps and
 * late starts score lower. If that fails and the query is 3+ characters, a
 * single typo (insert/delete/substitute/transpose) against the candidate's
 * prefix is tolerated with a low score, so "furit" still finds "Fruit".
 */
export function fuzzyMatch(query: string, candidate: string): FuzzyMatch | null {
  const q = query.trim().toLowerCase()
  const c = candidate.toLowerCase()
  if (!q) return { score: 0, indices: [] }

  const indices: number[] = []
  let score = 0
  let ci = 0
  let prev = -2
  for (const ch of q) {
    const found = c.indexOf(ch, ci)
    if (found === -1) {
      return typoMatch(q, c)
    }
    if (found === prev + 1) score += 8
    if (isWordStart(c, found)) score += 6
    score -= Math.min(found - ci, 5) // gap penalty
    indices.push(found)
    prev = found
    ci = found + 1
  }
  if (indices[0] === 0) score += 10
  if (c === q) score += 50
  else if (c.startsWith(q)) score += 20
  score -= (c.length - q.length) * 0.1 // prefer shorter candidates on ties
  return { score, indices }
}

function typoMatch(q: string, c: string): FuzzyMatch | null {
  if (q.length < 3) return null
  for (const len of [q.length, q.length - 1, q.length + 1]) {
    if (len < 1 || len > c.length) continue
    if (damerauLevenshtein(q, c.slice(0, len)) <= 1) {
      return { score: -10, indices: [] }
    }
  }
  return null
}

function damerauLevenshtein(a: string, b: string): number {
  const d: number[][] = Array.from({ length: a.length + 1 }, (_, i) =>
    Array.from({ length: b.length + 1 }, (_, j) => (i === 0 ? j : j === 0 ? i : 0)),
  )
  for (let i = 1; i <= a.length; i++) {
    for (let j = 1; j <= b.length; j++) {
      const cost = a[i - 1] === b[j - 1] ? 0 : 1
      d[i][j] = Math.min(d[i - 1][j] + 1, d[i][j - 1] + 1, d[i - 1][j - 1] + cost)
      if (i > 1 && j > 1 && a[i - 1] === b[j - 2] && a[i - 2] === b[j - 1]) {
        d[i][j] = Math.min(d[i][j], d[i - 2][j - 2] + 1)
      }
    }
  }
  return d[a.length][b.length]
}

/** Filters and ranks items by fuzzy match against `getText`, best first. */
export function fuzzyFilter<T>(
  query: string,
  items: T[],
  getText: (item: T) => string,
): { item: T; match: FuzzyMatch }[] {
  return items
    .map((item) => ({ item, match: fuzzyMatch(query, getText(item)) }))
    .filter((r): r is { item: T; match: FuzzyMatch } => r.match !== null)
    .sort((a, b) => b.match.score - a.match.score)
}
