import { useId, useMemo, useRef, useState } from 'react'
import type { Category } from '../types'
import { fuzzyFilter } from '../lib/fuzzy'
import CategoryPill from './CategoryPill'
import { categoryColour } from '../lib/categoryColour'

export const MAX_CATEGORIES = 10

interface Props {
  /** Category names currently assigned to the product. */
  value: string[]
  onChange: (names: string[]) => void
  /** All existing categories, used for suggestions. */
  categories: Category[]
  /**
   * Deletes a category everywhere (not just from this product). When omitted,
   * the delete button is not shown.
   */
  onDeleteCategory?: (category: Category) => void | Promise<void>
  id?: string
  placeholder?: string
}

type Option =
  | { kind: 'existing'; category: Category; indices: number[] }
  | { kind: 'create'; name: string }

const normalise = (s: string) => s.trim().replace(/\s+/g, ' ')

function Highlighted({ text, indices }: { text: string; indices: number[] }) {
  const set = new Set(indices)
  return (
    <span>
      {[...text].map((ch, i) =>
        set.has(i) ? (
          <mark key={i} className="bg-transparent text-inherit font-extrabold underline decoration-2 underline-offset-2">
            {ch}
          </mark>
        ) : (
          <span key={i}>{ch}</span>
        ),
      )}
    </span>
  )
}

export default function CategoryInput({
  value,
  onChange,
  categories,
  onDeleteCategory,
  id,
  placeholder = 'Add a category…',
}: Props) {
  const [query, setQuery] = useState('')
  const [open, setOpen] = useState(false)
  const [active, setActive] = useState(0)
  const [armedRemove, setArmedRemove] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const listId = useId()

  const selectedKeys = useMemo(() => new Set(value.map((v) => v.toLowerCase())), [value])
  const full = value.length >= MAX_CATEGORIES
  const trimmed = normalise(query)

  const options: Option[] = useMemo(() => {
    const available = categories.filter((c) => !selectedKeys.has(c.name.toLowerCase()))
    const matches = fuzzyFilter(trimmed, available, (c) => c.name)
      .slice(0, 8)
      .map<Option>(({ item, match }) => ({ kind: 'existing', category: item, indices: match.indices }))
    const exists =
      categories.some((c) => c.name.toLowerCase() === trimmed.toLowerCase()) ||
      selectedKeys.has(trimmed.toLowerCase())
    if (trimmed && !exists) matches.push({ kind: 'create', name: trimmed })
    return matches
  }, [categories, selectedKeys, trimmed])

  const activeIndex = Math.min(active, Math.max(options.length - 1, 0))
  const showList = open && !full && options.length > 0

  const add = (name: string) => {
    const n = normalise(name)
    if (!n || full || selectedKeys.has(n.toLowerCase())) return
    // Reuse an existing category's casing if the name matches one.
    const existing = categories.find((c) => c.name.toLowerCase() === n.toLowerCase())
    onChange([...value, existing?.name ?? n])
    setQuery('')
    setActive(0)
    setArmedRemove(false)
  }

  const remove = (name: string) => {
    onChange(value.filter((v) => v !== name))
    setArmedRemove(false)
    inputRef.current?.focus()
  }

  const choose = (opt: Option) => add(opt.kind === 'existing' ? opt.category.name : opt.name)

  const onKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        setOpen(true)
        setActive((i) => (options.length ? (i + 1) % options.length : 0))
        break
      case 'ArrowUp':
        e.preventDefault()
        setOpen(true)
        setActive((i) => (options.length ? (i - 1 + options.length) % options.length : 0))
        break
      case 'Enter':
      case 'Tab':
      case ',':
        if (!trimmed && !(e.key === 'Enter' && showList)) return
        e.preventDefault() // don't submit the surrounding form / leave the field
        if (showList && options[activeIndex]) choose(options[activeIndex])
        else add(trimmed)
        break
      case 'Backspace':
        if (query || value.length === 0) return
        e.preventDefault()
        // First press highlights the last pill, second press removes it.
        if (armedRemove) remove(value[value.length - 1])
        else setArmedRemove(true)
        break
      case 'Escape':
        if (open) {
          e.preventDefault()
          setOpen(false)
        }
        break
      default:
        setArmedRemove(false)
    }
  }

  return (
    <div className="relative">
      <div
        onClick={() => inputRef.current?.focus()}
        className="w-full min-h-[38px] border border-[#e8ddd0] rounded px-2 py-1.5 flex flex-wrap items-center gap-1.5 cursor-text focus-within:border-[#5c3d1e] bg-white"
      >
        {value.map((name, i) => (
          <CategoryPill
            key={name}
            category={name}
            onRemove={() => remove(name)}
            highlighted={armedRemove && i === value.length - 1}
          />
        ))}
        <input
          ref={inputRef}
          id={id}
          role="combobox"
          aria-expanded={showList}
          aria-controls={listId}
          aria-autocomplete="list"
          aria-activedescendant={showList ? `${listId}-${activeIndex}` : undefined}
          value={query}
          disabled={full}
          placeholder={full ? `Max ${MAX_CATEGORIES} categories` : value.length ? '' : placeholder}
          onChange={(e) => {
            setQuery(e.target.value)
            setActive(0)
            setOpen(true)
            setArmedRemove(false)
          }}
          onFocus={() => setOpen(true)}
          onBlur={() => {
            setOpen(false)
            setArmedRemove(false)
            setConfirmDelete(null)
          }}
          onKeyDown={onKeyDown}
          className="flex-1 min-w-[8rem] text-sm outline-none bg-transparent px-1 py-0.5 disabled:cursor-not-allowed placeholder:text-[#c4a882]"
        />
      </div>

      {showList && (
        <ul
          id={listId}
          role="listbox"
          // Keep focus in the input while clicking inside the list.
          onMouseDown={(e) => e.preventDefault()}
          className="absolute z-10 mt-1 w-full max-h-64 overflow-auto bg-white border border-[#e8ddd0] rounded shadow-lg py-1 text-sm"
        >
          {options.map((opt, i) => {
            const isActive = i === activeIndex
            const key = opt.kind === 'existing' ? opt.category.id : `create:${opt.name}`
            const base = `flex items-center gap-2 px-3 py-1.5 cursor-pointer ${isActive ? 'bg-[#f5ede3]' : ''}`

            if (opt.kind === 'create') {
              return (
                <li
                  key={key}
                  id={`${listId}-${i}`}
                  role="option"
                  aria-selected={isActive}
                  onMouseEnter={() => setActive(i)}
                  onClick={() => choose(opt)}
                  className={`${base} text-[#5c3d1e]`}
                >
                  <span className="text-base leading-none">+</span>
                  <span>Create</span>
                  <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${categoryColour(opt.name)}`}>
                    {opt.name}
                  </span>
                </li>
              )
            }

            const cat = opt.category
            return (
              <li
                key={key}
                id={`${listId}-${i}`}
                role="option"
                aria-selected={isActive}
                onMouseEnter={() => setActive(i)}
                onClick={() => confirmDelete !== cat.id && choose(opt)}
                className={`${base} group`}
              >
                <span className="flex-1">
                  <span className={`inline-block rounded-full px-2.5 py-0.5 text-xs font-medium ${categoryColour(cat.name)}`}>
                    <Highlighted text={cat.name} indices={opt.indices} />
                  </span>
                </span>
                {onDeleteCategory &&
                  (confirmDelete === cat.id ? (
                    <span className="flex items-center gap-2 text-xs" onClick={(e) => e.stopPropagation()}>
                      <span className="text-[#8a6a50]">Delete everywhere?</span>
                      <button
                        type="button"
                        className="text-red-600 font-medium hover:underline"
                        onClick={async () => {
                          setConfirmDelete(null)
                          await onDeleteCategory(cat)
                        }}
                      >
                        Yes
                      </button>
                      <button
                        type="button"
                        className="text-[#8a6a50] hover:underline"
                        onClick={() => setConfirmDelete(null)}
                      >
                        No
                      </button>
                    </span>
                  ) : (
                    <button
                      type="button"
                      aria-label={`Delete category ${cat.name}`}
                      title="Delete category everywhere"
                      onClick={(e) => {
                        e.stopPropagation()
                        setConfirmDelete(cat.id)
                      }}
                      className={`text-xs text-[#8a6a50] hover:text-red-600 px-1 ${
                        isActive ? 'visible' : 'invisible group-hover:visible'
                      }`}
                    >
                      🗑
                    </button>
                  ))}
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
