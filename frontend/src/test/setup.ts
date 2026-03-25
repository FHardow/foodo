import '@testing-library/jest-dom'
import { afterEach } from 'vitest'

// jsdom 29 does not expose a functional localStorage in vitest's environment.
// Provide a minimal in-memory implementation so zustand's persist middleware works.
const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, value: string) => { store[key] = value },
    removeItem: (key: string) => { delete store[key] },
    clear: () => { store = {} },
    get length() { return Object.keys(store).length },
    key: (index: number) => Object.keys(store)[index] ?? null,
  }
})()

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
  writable: true,
})

afterEach(() => localStorage.clear())
