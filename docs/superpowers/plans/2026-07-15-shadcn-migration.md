# shadcn Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate the Foodo frontend to shadcn/ui components, fixing two mobile bugs in the process — dead kanban drag-and-drop on touch, and no mobile nav.

**Architecture:** Hand-author shadcn's standard "new-york" style primitives directly into `frontend/src/components/ui/` (no network-dependent CLI calls — deterministic, offline-safe). Build on Radix UI primitives + `class-variance-authority` + `clsx`/`tailwind-merge`. Re-skin every page/component to consume these primitives while preserving all existing accessible names, roles, routes, and `data-testid` attributes so the existing Playwright e2e suite keeps passing.

**Tech Stack:** React 19, TypeScript, Tailwind CSS v4, Radix UI primitives, class-variance-authority, lucide-react icons, @dnd-kit/core (kanban drag, gets a TouchSensor).

## Global Constraints

- Brand colors carry over exactly: primary `#5c3d1e`, background `#faf7f2`, card `#ffffff`, muted/border `#e8ddd0`, secondary `#f0e8de`, muted-foreground `#8a6a50`, foreground `#3d2b1a`. No default shadcn slate/zinc theme.
- Style: "new-york", CSS variables (not utility classes) for theme tokens, path alias `@/*` → `src/*`.
- Icons: `lucide-react` only. No emoji added; the one existing emoji (🛒 in Nav) is replaced.
- No dark mode (out of scope per spec).
- No backend changes.
- Every `getByRole`/`getByText`/`getByTestId` target currently asserted in `frontend/e2e/*.spec.ts` must resolve identically after migration, unless the plan explicitly says a spec file is being updated.
- `data-testid="kanban-column"`, `data-testid="column-count"`, `data-testid="order-card"`, `data-testid="loading-skeleton"`, `data-order-id`, `data-order-status`, `data-column-status` in `admin/Orders.tsx` are preserved verbatim.
- Component set is trimmed from the design spec's draft list to only what the app actually uses: `button`, `card`, `badge`, `input`, `label`, `textarea`, `switch`, `dialog`, `sheet`. (`select`, `form`, `separator`, `dropdown-menu` dropped — nothing in the current UI needs a native select, react-hook-form, a divider, or a dropdown menu; adding them would be unused surface area.)
- All `npm` commands run from `frontend/`.

---

## File Map

**Create:**
- `frontend/components.json`
- `frontend/src/lib/utils.ts`
- `frontend/src/components/ui/button.tsx` (+ `button.test.tsx`)
- `frontend/src/components/ui/card.tsx` (+ `card.test.tsx`)
- `frontend/src/components/ui/badge.tsx` (+ `badge.test.tsx`)
- `frontend/src/components/ui/label.tsx` (+ `label.test.tsx`)
- `frontend/src/components/ui/input.tsx` (+ `input.test.tsx`)
- `frontend/src/components/ui/textarea.tsx` (+ `textarea.test.tsx`)
- `frontend/src/components/ui/switch.tsx` (+ `switch.test.tsx`)
- `frontend/src/components/ui/dialog.tsx`
- `frontend/src/components/ui/sheet.tsx`
- `frontend/e2e/admin-products.spec.ts`

**Modify:**
- `frontend/vite.config.ts`, `frontend/tsconfig.app.json` (path alias)
- `frontend/src/index.css` (theme tokens)
- `frontend/src/components/Nav.tsx`, `frontend/e2e/nav.spec.ts`
- `frontend/src/pages/admin/Orders.tsx`
- `frontend/src/components/ProductCard.tsx`, `frontend/src/pages/Store.tsx`
- `frontend/src/pages/Product.tsx`
- `frontend/src/pages/Basket.tsx`
- `frontend/src/components/StatusBadge.tsx`, `frontend/src/pages/OrderHistory.tsx`, `frontend/src/pages/OrderStatus.tsx`
- `frontend/src/pages/admin/Products.tsx`

**Delete:**
- `frontend/src/App.css` (dead file — not imported anywhere, confirmed via grep)

---

## Task 1: Foundation — path alias, theme tokens, cn utility

**Files:**
- Modify: `frontend/vite.config.ts`, `frontend/tsconfig.app.json`, `frontend/src/index.css`
- Create: `frontend/components.json`, `frontend/src/lib/utils.ts`

**Interfaces:**
- Produces: `cn(...inputs: ClassValue[]): string` from `src/lib/utils.ts`, used by every component in Tasks 2–4. Path alias `@/*` → `src/*`, used by every new file from Task 2 onward.

- [ ] **Step 1: Install cn dependencies**

```bash
cd frontend && npm install clsx tailwind-merge
```

Expected: exit 0, `clsx` and `tailwind-merge` added to `package.json` dependencies.

- [ ] **Step 2: Add the `@` path alias**

Replace `frontend/vite.config.ts` entirely:

```ts
import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  test: {
    environment: 'jsdom',
    environmentOptions: {
      jsdom: {
        url: 'http://localhost',
      },
    },
    setupFiles: './src/test/setup.ts',
    globals: true,
    exclude: ['**/node_modules/**', '**/e2e/**'],
  },
})
```

Replace `frontend/tsconfig.app.json` entirely:

```json
{
  "compilerOptions": {
    "tsBuildInfoFile": "./node_modules/.tmp/tsconfig.app.tsbuildinfo",
    "target": "ES2023",
    "useDefineForClassFields": true,
    "lib": ["ES2023", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "types": ["vite/client", "vitest/globals"],
    "skipLibCheck": true,

    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "verbatimModuleSyntax": true,
    "moduleDetection": "force",
    "noEmit": true,
    "jsx": "react-jsx",
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    },

    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "erasableSyntaxOnly": true,
    "noFallthroughCasesInSwitch": true,
    "noUncheckedSideEffectImports": true
  },
  "include": ["src"]
}
```

- [ ] **Step 3: Add theme tokens**

Replace `frontend/src/index.css` entirely:

```css
@import "tailwindcss";

:root {
  --radius: 0.5rem;
  --background: #faf7f2;
  --foreground: #3d2b1a;
  --card: #ffffff;
  --card-foreground: #3d2b1a;
  --popover: #ffffff;
  --popover-foreground: #3d2b1a;
  --primary: #5c3d1e;
  --primary-foreground: #ffffff;
  --secondary: #f0e8de;
  --secondary-foreground: #5c3d1e;
  --muted: #f0e8de;
  --muted-foreground: #8a6a50;
  --accent: #e8ddd0;
  --accent-foreground: #3d2b1a;
  --destructive: #dc2626;
  --destructive-foreground: #ffffff;
  --border: #e8ddd0;
  --input: #e8ddd0;
  --ring: #5c3d1e;
}

@theme inline {
  --color-background: var(--background);
  --color-foreground: var(--foreground);
  --color-card: var(--card);
  --color-card-foreground: var(--card-foreground);
  --color-popover: var(--popover);
  --color-popover-foreground: var(--popover-foreground);
  --color-primary: var(--primary);
  --color-primary-foreground: var(--primary-foreground);
  --color-secondary: var(--secondary);
  --color-secondary-foreground: var(--secondary-foreground);
  --color-muted: var(--muted);
  --color-muted-foreground: var(--muted-foreground);
  --color-accent: var(--accent);
  --color-accent-foreground: var(--accent-foreground);
  --color-destructive: var(--destructive);
  --color-destructive-foreground: var(--destructive-foreground);
  --color-border: var(--border);
  --color-input: var(--input);
  --color-ring: var(--ring);
  --radius-sm: calc(var(--radius) - 4px);
  --radius-md: calc(var(--radius) - 2px);
  --radius-lg: var(--radius);
  --radius-xl: calc(var(--radius) + 4px);
}

@layer base {
  * {
    @apply border-border outline-ring/50;
  }
  body {
    @apply bg-background text-foreground;
  }
}
```

- [ ] **Step 4: Add components.json and the cn utility**

Create `frontend/components.json`:

```json
{
  "$schema": "https://ui.shadcn.com/schema.json",
  "style": "new-york",
  "rsc": false,
  "tsx": true,
  "tailwind": {
    "config": "",
    "css": "src/index.css",
    "baseColor": "neutral",
    "cssVariables": true,
    "prefix": ""
  },
  "aliases": {
    "components": "@/components",
    "utils": "@/lib/utils",
    "ui": "@/components/ui",
    "lib": "@/lib",
    "hooks": "@/hooks"
  },
  "iconLibrary": "lucide"
}
```

Create `frontend/src/lib/utils.ts`:

```ts
import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

- [ ] **Step 5: Verify build and existing tests still pass**

```bash
cd frontend && npm run build && npm run test
```

Expected: both PASS. `npm run build` confirms the path alias and new CSS compile; `npm run test` confirms nothing in the existing vitest suite (api/client, hooks/useBasket, store/basket) broke.

- [ ] **Step 6: Commit**

```bash
git add frontend/vite.config.ts frontend/tsconfig.app.json frontend/src/index.css frontend/components.json frontend/src/lib/utils.ts frontend/package.json frontend/package-lock.json
git commit -m "chore: add shadcn foundation (path alias, theme tokens, cn util)"
```

---

## Task 2: Core display primitives — Button, Card, Badge

**Files:**
- Create: `frontend/src/components/ui/button.tsx`, `frontend/src/components/ui/button.test.tsx`
- Create: `frontend/src/components/ui/card.tsx`, `frontend/src/components/ui/card.test.tsx`
- Create: `frontend/src/components/ui/badge.tsx`, `frontend/src/components/ui/badge.test.tsx`

**Interfaces:**
- Consumes: `cn` from `@/lib/utils` (Task 1).
- Produces: `Button`/`buttonVariants` (props: `variant?: 'default'|'destructive'|'outline'|'secondary'|'ghost'|'link'`, `size?: 'default'|'sm'|'lg'|'icon'`, `asChild?: boolean`), `Card`/`CardHeader`/`CardTitle`/`CardDescription`/`CardContent`/`CardFooter`, `Badge`/`badgeVariants` (props: `variant?: 'default'|'secondary'|'destructive'|'outline'`) — all from `frontend/src/components/ui/{button,card,badge}.tsx`. Consumed by every page task from Task 5 onward.

- [ ] **Step 1: Write failing tests**

Create `frontend/src/components/ui/button.test.tsx`:

```tsx
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { Button } from './button'

describe('Button', () => {
  it('renders as a button with the given text', () => {
    render(<Button>Click me</Button>)
    expect(screen.getByRole('button', { name: 'Click me' })).toBeInTheDocument()
  })

  it('calls onClick when clicked', async () => {
    const onClick = vi.fn()
    render(<Button onClick={onClick}>Go</Button>)
    await userEvent.click(screen.getByRole('button', { name: 'Go' }))
    expect(onClick).toHaveBeenCalledOnce()
  })

  it('is disabled when the disabled prop is set', () => {
    render(<Button disabled>Go</Button>)
    expect(screen.getByRole('button', { name: 'Go' })).toBeDisabled()
  })
})
```

Create `frontend/src/components/ui/card.test.tsx`:

```tsx
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
```

Create `frontend/src/components/ui/badge.test.tsx`:

```tsx
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
```

- [ ] **Step 2: Run tests, verify they fail**

```bash
cd frontend && npx vitest run src/components/ui/button.test.tsx src/components/ui/card.test.tsx src/components/ui/badge.test.tsx
```

Expected: FAIL — `./button`, `./card`, `./badge` cannot be resolved (files don't exist yet).

- [ ] **Step 3: Install dependencies**

```bash
cd frontend && npm install class-variance-authority @radix-ui/react-slot
```

- [ ] **Step 4: Implement the three primitives**

Create `frontend/src/components/ui/button.tsx`:

```tsx
import * as React from 'react'
import { Slot } from '@radix-ui/react-slot'
import { cva, type VariantProps } from 'class-variance-authority'

import { cn } from '@/lib/utils'

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 outline-none focus-visible:ring-2 focus-visible:ring-ring",
  {
    variants: {
      variant: {
        default: 'bg-primary text-primary-foreground hover:bg-primary/90',
        destructive: 'bg-destructive text-destructive-foreground hover:bg-destructive/90',
        outline: 'border border-input bg-background hover:bg-accent hover:text-accent-foreground',
        secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80',
        ghost: 'hover:bg-accent hover:text-accent-foreground',
        link: 'text-primary underline-offset-4 hover:underline',
      },
      size: {
        default: 'h-9 px-4 py-2',
        sm: 'h-8 rounded-md px-3 text-xs',
        lg: 'h-10 rounded-md px-8',
        icon: 'h-9 w-9',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  },
)

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : 'button'
    return (
      <Comp className={cn(buttonVariants({ variant, size, className }))} ref={ref} {...props} />
    )
  },
)
Button.displayName = 'Button'

export { Button, buttonVariants }
```

Create `frontend/src/components/ui/card.tsx`:

```tsx
import * as React from 'react'

import { cn } from '@/lib/utils'

const Card = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn('rounded-lg border border-border bg-card text-card-foreground shadow-sm', className)}
      {...props}
    />
  ),
)
Card.displayName = 'Card'

const CardHeader = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={cn('flex flex-col gap-1.5 p-4', className)} {...props} />
  ),
)
CardHeader.displayName = 'CardHeader'

const CardTitle = React.forwardRef<HTMLHeadingElement, React.HTMLAttributes<HTMLHeadingElement>>(
  ({ className, ...props }, ref) => (
    <h3
      ref={ref}
      className={cn('font-semibold leading-none tracking-tight text-card-foreground', className)}
      {...props}
    />
  ),
)
CardTitle.displayName = 'CardTitle'

const CardDescription = React.forwardRef<HTMLParagraphElement, React.HTMLAttributes<HTMLParagraphElement>>(
  ({ className, ...props }, ref) => (
    <p ref={ref} className={cn('text-sm text-muted-foreground', className)} {...props} />
  ),
)
CardDescription.displayName = 'CardDescription'

const CardContent = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => <div ref={ref} className={cn('p-4 pt-0', className)} {...props} />,
)
CardContent.displayName = 'CardContent'

const CardFooter = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={cn('flex items-center p-4 pt-0', className)} {...props} />
  ),
)
CardFooter.displayName = 'CardFooter'

export { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter }
```

Create `frontend/src/components/ui/badge.tsx`:

```tsx
import * as React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'

import { cn } from '@/lib/utils'

const badgeVariants = cva(
  'inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium capitalize transition-colors focus:outline-none focus:ring-2 focus:ring-ring',
  {
    variants: {
      variant: {
        default: 'border-transparent bg-primary text-primary-foreground',
        secondary: 'border-transparent bg-secondary text-secondary-foreground',
        destructive: 'border-transparent bg-destructive text-destructive-foreground',
        outline: 'text-foreground',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  },
)

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement>, VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return <span className={cn(badgeVariants({ variant }), className)} {...props} />
}

export { Badge, badgeVariants }
```

- [ ] **Step 5: Run tests, verify they pass**

```bash
cd frontend && npx vitest run src/components/ui/button.test.tsx src/components/ui/card.test.tsx src/components/ui/badge.test.tsx
```

Expected: PASS, 5 tests total.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/ui/button.tsx frontend/src/components/ui/button.test.tsx frontend/src/components/ui/card.tsx frontend/src/components/ui/card.test.tsx frontend/src/components/ui/badge.tsx frontend/src/components/ui/badge.test.tsx frontend/package.json frontend/package-lock.json
git commit -m "feat: add shadcn Button, Card, Badge primitives"
```

---

## Task 3: Form primitives — Label, Input, Textarea, Switch

**Files:**
- Create: `frontend/src/components/ui/label.tsx`, `frontend/src/components/ui/label.test.tsx`
- Create: `frontend/src/components/ui/input.tsx`, `frontend/src/components/ui/input.test.tsx`
- Create: `frontend/src/components/ui/textarea.tsx`, `frontend/src/components/ui/textarea.test.tsx`
- Create: `frontend/src/components/ui/switch.tsx`, `frontend/src/components/ui/switch.test.tsx`

**Interfaces:**
- Consumes: `cn` from `@/lib/utils` (Task 1).
- Produces: `Label`, `Input`, `Textarea` (standard HTML input/textarea prop pass-through), `Switch` (props: `checked?: boolean`, `onCheckedChange?: (checked: boolean) => void`, `id?: string`, `aria-label?: string`) — from `frontend/src/components/ui/{label,input,textarea,switch}.tsx`. Consumed by Task 11 (admin/Products form).

- [ ] **Step 1: Write failing tests**

Create `frontend/src/components/ui/label.test.tsx`:

```tsx
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
```

Create `frontend/src/components/ui/input.test.tsx`:

```tsx
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
```

Create `frontend/src/components/ui/textarea.test.tsx`:

```tsx
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
```

Create `frontend/src/components/ui/switch.test.tsx`:

```tsx
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
```

- [ ] **Step 2: Run tests, verify they fail**

```bash
cd frontend && npx vitest run src/components/ui/label.test.tsx src/components/ui/input.test.tsx src/components/ui/textarea.test.tsx src/components/ui/switch.test.tsx
```

Expected: FAIL — `./label`, `./input`, `./textarea`, `./switch` cannot be resolved.

- [ ] **Step 3: Install dependencies**

```bash
cd frontend && npm install @radix-ui/react-label @radix-ui/react-switch
```

- [ ] **Step 4: Implement the four primitives**

Create `frontend/src/components/ui/label.tsx`:

```tsx
import * as React from 'react'
import * as LabelPrimitive from '@radix-ui/react-label'

import { cn } from '@/lib/utils'

const Label = React.forwardRef<
  React.ElementRef<typeof LabelPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof LabelPrimitive.Root>
>(({ className, ...props }, ref) => (
  <LabelPrimitive.Root
    ref={ref}
    className={cn(
      'text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70',
      className,
    )}
    {...props}
  />
))
Label.displayName = LabelPrimitive.Root.displayName

export { Label }
```

Create `frontend/src/components/ui/input.tsx`:

```tsx
import * as React from 'react'

import { cn } from '@/lib/utils'

const Input = React.forwardRef<HTMLInputElement, React.ComponentProps<'input'>>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(
          'flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
          className,
        )}
        ref={ref}
        {...props}
      />
    )
  },
)
Input.displayName = 'Input'

export { Input }
```

Create `frontend/src/components/ui/textarea.tsx`:

```tsx
import * as React from 'react'

import { cn } from '@/lib/utils'

const Textarea = React.forwardRef<HTMLTextAreaElement, React.ComponentProps<'textarea'>>(
  ({ className, ...props }, ref) => {
    return (
      <textarea
        className={cn(
          'flex min-h-16 w-full rounded-md border border-input bg-background px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
          className,
        )}
        ref={ref}
        {...props}
      />
    )
  },
)
Textarea.displayName = 'Textarea'

export { Textarea }
```

Create `frontend/src/components/ui/switch.tsx`:

```tsx
import * as React from 'react'
import * as SwitchPrimitive from '@radix-ui/react-switch'

import { cn } from '@/lib/utils'

const Switch = React.forwardRef<
  React.ElementRef<typeof SwitchPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof SwitchPrimitive.Root>
>(({ className, ...props }, ref) => (
  <SwitchPrimitive.Root
    className={cn(
      'peer inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 data-[state=checked]:bg-primary data-[state=unchecked]:bg-input',
      className,
    )}
    {...props}
    ref={ref}
  >
    <SwitchPrimitive.Thumb
      className="pointer-events-none block h-4 w-4 rounded-full bg-background shadow-lg ring-0 transition-transform data-[state=checked]:translate-x-4 data-[state=unchecked]:translate-x-0"
    />
  </SwitchPrimitive.Root>
))
Switch.displayName = SwitchPrimitive.Root.displayName

export { Switch }
```

- [ ] **Step 5: Run tests, verify they pass**

```bash
cd frontend && npx vitest run src/components/ui/label.test.tsx src/components/ui/input.test.tsx src/components/ui/textarea.test.tsx src/components/ui/switch.test.tsx
```

Expected: PASS, 4 tests total.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/ui/label.tsx frontend/src/components/ui/label.test.tsx frontend/src/components/ui/input.tsx frontend/src/components/ui/input.test.tsx frontend/src/components/ui/textarea.tsx frontend/src/components/ui/textarea.test.tsx frontend/src/components/ui/switch.tsx frontend/src/components/ui/switch.test.tsx frontend/package.json frontend/package-lock.json
git commit -m "feat: add shadcn Label, Input, Textarea, Switch primitives"
```

---

## Task 4: Overlay primitives — Dialog, Sheet

**Files:**
- Create: `frontend/src/components/ui/dialog.tsx`
- Create: `frontend/src/components/ui/sheet.tsx`

**Interfaces:**
- Consumes: `cn` from `@/lib/utils` (Task 1).
- Produces: `Dialog`/`DialogTrigger`/`DialogContent`/`DialogHeader`/`DialogFooter`/`DialogTitle`/`DialogDescription`/`DialogClose` from `ui/dialog.tsx`; `Sheet`/`SheetTrigger`/`SheetContent`/`SheetHeader`/`SheetTitle`/`SheetClose` (with `SheetContent` taking `side?: 'top'|'bottom'|'left'|'right'`) from `ui/sheet.tsx`. `Sheet` consumed by Task 5 (Nav), `Dialog` consumed by Task 11 (admin/Products).

No component-level tests here: both are Radix `Dialog` primitive wrappers that render into a portal and manage focus trapping — jsdom support for that is flaky and low-value to fight for a smoke test. Real coverage comes from the e2e tests these primitives get exercised through in Task 5 (Sheet, via `nav.spec.ts`) and Task 11 (Dialog, via the new `admin-products.spec.ts`). This task's own verification is a build/typecheck pass.

- [ ] **Step 1: Install dependency**

```bash
cd frontend && npm install @radix-ui/react-dialog lucide-react
```

- [ ] **Step 2: Implement Dialog**

Create `frontend/src/components/ui/dialog.tsx`:

```tsx
import * as React from 'react'
import * as DialogPrimitive from '@radix-ui/react-dialog'
import { X } from 'lucide-react'

import { cn } from '@/lib/utils'

const Dialog = DialogPrimitive.Root
const DialogTrigger = DialogPrimitive.Trigger
const DialogPortal = DialogPrimitive.Portal
const DialogClose = DialogPrimitive.Close

const DialogOverlay = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Overlay>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Overlay>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Overlay ref={ref} className={cn('fixed inset-0 z-50 bg-black/40', className)} {...props} />
))
DialogOverlay.displayName = DialogPrimitive.Overlay.displayName

const DialogContent = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Content>
>(({ className, children, ...props }, ref) => (
  <DialogPortal>
    <DialogOverlay />
    <DialogPrimitive.Content
      ref={ref}
      className={cn(
        'fixed left-1/2 top-1/2 z-50 grid w-full max-w-md -translate-x-1/2 -translate-y-1/2 gap-4 rounded-lg border border-border bg-card p-0 shadow-xl',
        className,
      )}
      {...props}
    >
      {children}
      <DialogPrimitive.Close className="absolute right-4 top-4 rounded-sm text-muted-foreground opacity-70 transition-opacity hover:opacity-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-ring">
        <X className="h-4 w-4" />
        <span className="sr-only">Close</span>
      </DialogPrimitive.Close>
    </DialogPrimitive.Content>
  </DialogPortal>
))
DialogContent.displayName = DialogPrimitive.Content.displayName

const DialogHeader = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={cn('flex items-center justify-between border-b border-border px-6 py-4', className)} {...props} />
)
DialogHeader.displayName = 'DialogHeader'

const DialogFooter = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={cn('flex justify-end gap-3 px-6 pb-6 pt-2', className)} {...props} />
)
DialogFooter.displayName = 'DialogFooter'

const DialogTitle = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Title>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Title>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Title ref={ref} className={cn('font-semibold text-foreground', className)} {...props} />
))
DialogTitle.displayName = DialogPrimitive.Title.displayName

const DialogDescription = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Description>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Description>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Description ref={ref} className={cn('text-sm text-muted-foreground', className)} {...props} />
))
DialogDescription.displayName = DialogPrimitive.Description.displayName

export {
  Dialog,
  DialogPortal,
  DialogOverlay,
  DialogTrigger,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogFooter,
  DialogTitle,
  DialogDescription,
}
```

- [ ] **Step 3: Implement Sheet**

Create `frontend/src/components/ui/sheet.tsx`:

```tsx
import * as React from 'react'
import * as DialogPrimitive from '@radix-ui/react-dialog'
import { cva, type VariantProps } from 'class-variance-authority'
import { X } from 'lucide-react'

import { cn } from '@/lib/utils'

const Sheet = DialogPrimitive.Root
const SheetTrigger = DialogPrimitive.Trigger
const SheetClose = DialogPrimitive.Close
const SheetPortal = DialogPrimitive.Portal

const SheetOverlay = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Overlay>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Overlay>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Overlay ref={ref} className={cn('fixed inset-0 z-50 bg-black/40', className)} {...props} />
))
SheetOverlay.displayName = DialogPrimitive.Overlay.displayName

const sheetVariants = cva('fixed z-50 flex flex-col gap-4 bg-card p-6 shadow-xl transition ease-in-out', {
  variants: {
    side: {
      top: 'inset-x-0 top-0 border-b border-border',
      bottom: 'inset-x-0 bottom-0 border-t border-border',
      left: 'inset-y-0 left-0 h-full w-3/4 border-r border-border sm:max-w-xs',
      right: 'inset-y-0 right-0 h-full w-3/4 border-l border-border sm:max-w-xs',
    },
  },
  defaultVariants: {
    side: 'right',
  },
})

interface SheetContentProps
  extends React.ComponentPropsWithoutRef<typeof DialogPrimitive.Content>,
    VariantProps<typeof sheetVariants> {}

const SheetContent = React.forwardRef<React.ElementRef<typeof DialogPrimitive.Content>, SheetContentProps>(
  ({ side = 'right', className, children, ...props }, ref) => (
    <SheetPortal>
      <SheetOverlay />
      <DialogPrimitive.Content ref={ref} className={cn(sheetVariants({ side }), className)} {...props}>
        {children}
        <DialogPrimitive.Close className="absolute right-4 top-4 rounded-sm text-muted-foreground opacity-70 transition-opacity hover:opacity-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-ring">
          <X className="h-4 w-4" />
          <span className="sr-only">Close</span>
        </DialogPrimitive.Close>
      </DialogPrimitive.Content>
    </SheetPortal>
  ),
)
SheetContent.displayName = DialogPrimitive.Content.displayName

const SheetHeader = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={cn('flex flex-col gap-1', className)} {...props} />
)
SheetHeader.displayName = 'SheetHeader'

const SheetTitle = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Title>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Title>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Title ref={ref} className={cn('font-semibold text-foreground', className)} {...props} />
))
SheetTitle.displayName = DialogPrimitive.Title.displayName

export { Sheet, SheetTrigger, SheetClose, SheetContent, SheetHeader, SheetTitle }
```

- [ ] **Step 4: Verify build**

```bash
cd frontend && npm run build
```

Expected: PASS (typecheck + bundle succeed; neither component is wired into a page yet, so this only proves they compile).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/ui/dialog.tsx frontend/src/components/ui/sheet.tsx frontend/package.json frontend/package-lock.json
git commit -m "feat: add shadcn Dialog, Sheet primitives"
```

---

## Task 5: Nav — mobile burger menu (fixes bug #2: no mobile nav)

**Files:**
- Modify: `frontend/src/components/Nav.tsx`
- Modify: `frontend/e2e/nav.spec.ts`

**Interfaces:**
- Consumes: `Button` (Task 2), `Sheet`/`SheetTrigger`/`SheetContent`/`SheetHeader`/`SheetTitle`/`SheetClose` (Task 4), `Menu`/`ShoppingCart` icons from `lucide-react` (Task 4 installed the package).
- No new exports — `Nav` remains a default export consumed by `App.tsx` unchanged.

Below the `sm` breakpoint, links (Store/History/Orders/Products) are currently rendered but CSS-hidden with no way to reach them. Fix: a `Menu` icon button (visible only below `sm`) opens a `Sheet` listing the same links. The cart link stays outside the sheet, always visible.

- [ ] **Step 1: Update nav.spec.ts — fix stale skip reasons, add mobile burger menu coverage**

Replace `frontend/e2e/nav.spec.ts` entirely:

```ts
import { test, expect, setupApiMocks, setRoles } from './fixtures'

test.describe('Navigation — owner', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
    await setupApiMocks(page)
  })

  test('shows Foodo brand link', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('link', { name: 'Foodo' })).toBeVisible()
  })

  test('shows Store and History links', async ({ page, isMobile }) => {
    test.skip(isMobile, 'Desktop-only row — on mobile these live inside the burger menu')
    await page.goto('/')

    await expect(page.getByRole('link', { name: 'Store' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'History' })).toBeVisible()
  })

  test('shows owner-only nav links for admin', async ({ page, isMobile }) => {
    test.skip(isMobile, 'Desktop-only row — on mobile these live inside the burger menu')
    await page.goto('/')

    await expect(page.getByRole('link', { name: 'Orders' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Products' })).toBeVisible()
  })

  test('shows basket button', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('link', { name: /Basket/ })).toBeVisible()
  })

  test('brand link navigates to store', async ({ page }) => {
    await page.goto('/basket')

    await page.getByRole('link', { name: 'Foodo' }).click()
    await expect(page).toHaveURL('/')
  })

  test('Orders link navigates to kanban board', async ({ page, isMobile }) => {
    test.skip(isMobile, 'Desktop-only row — on mobile these live inside the burger menu')
    await page.goto('/')

    await page.getByRole('link', { name: 'Orders' }).click()
    await expect(page).toHaveURL('/admin/orders')
  })

  test('History link navigates to order history', async ({ page, isMobile }) => {
    test.skip(isMobile, 'Desktop-only row — on mobile these live inside the burger menu')
    await page.goto('/')

    await page.getByRole('link', { name: 'History' }).click()
    await expect(page).toHaveURL('/orders')
  })

  test('basket link navigates to basket', async ({ page }) => {
    await page.goto('/')

    await page.getByRole('link', { name: /Basket/ }).click()
    await expect(page).toHaveURL('/basket')
  })
})

test.describe('Navigation — customer', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, [])
    await setupApiMocks(page)
  })

  test('does not show admin nav links for non-owner', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('link', { name: 'Orders' })).not.toBeVisible()
    await expect(page.getByRole('link', { name: 'Products' })).not.toBeVisible()
  })

  test('still shows Store, History and basket links', async ({ page, isMobile }) => {
    test.skip(isMobile, 'Desktop-only row — on mobile these live inside the burger menu')
    await page.goto('/')

    await expect(page.getByRole('link', { name: 'Store' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'History' })).toBeVisible()
    await expect(page.getByRole('link', { name: /Basket/ })).toBeVisible()
  })
})

test.describe('Navigation — mobile burger menu', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
    await setupApiMocks(page)
  })

  test('shows a menu button on mobile; desktop links stay hidden', async ({ page, isMobile }) => {
    test.skip(!isMobile, 'Burger menu only renders below the sm breakpoint')
    await page.goto('/')

    await expect(page.getByRole('button', { name: 'Open menu' })).toBeVisible()
    await expect(page.getByRole('navigation').getByRole('link', { name: 'Store' })).not.toBeVisible()
  })

  test('opens the menu and shows all nav links, including owner-only ones', async ({ page, isMobile }) => {
    test.skip(!isMobile, 'Burger menu only renders below the sm breakpoint')
    await page.goto('/')

    await page.getByRole('button', { name: 'Open menu' }).click()
    const menu = page.getByRole('dialog')
    await expect(menu.getByRole('link', { name: 'Store' })).toBeVisible()
    await expect(menu.getByRole('link', { name: 'History' })).toBeVisible()
    await expect(menu.getByRole('link', { name: 'Orders' })).toBeVisible()
    await expect(menu.getByRole('link', { name: 'Products' })).toBeVisible()
  })

  test('clicking a link in the menu navigates and closes it', async ({ page, isMobile }) => {
    test.skip(!isMobile, 'Burger menu only renders below the sm breakpoint')
    await page.goto('/')

    await page.getByRole('button', { name: 'Open menu' }).click()
    await page.getByRole('dialog').getByRole('link', { name: 'History' }).click()

    await expect(page).toHaveURL('/orders')
    await expect(page.getByRole('dialog')).not.toBeVisible()
  })

  test('non-owner does not see owner-only links in the menu', async ({ page, isMobile }) => {
    test.skip(!isMobile, 'Burger menu only renders below the sm breakpoint')
    await setRoles(page, [])
    await setupApiMocks(page)
    await page.goto('/')

    await page.getByRole('button', { name: 'Open menu' }).click()
    const menu = page.getByRole('dialog')
    await expect(menu.getByRole('link', { name: 'Orders' })).toHaveCount(0)
    await expect(menu.getByRole('link', { name: 'Products' })).toHaveCount(0)
  })
})

test.describe('Order history page', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, [])
    await setupApiMocks(page)
  })

  test('shows order history heading', async ({ page }) => {
    await page.goto('/orders')

    await expect(page).toHaveURL('/orders')
  })
})
```

- [ ] **Step 2: Run the new/updated spec against the current Nav, verify the new mobile tests fail**

```bash
cd frontend && npx playwright test e2e/nav.spec.ts --project=mobile-chrome
```

Expected: FAIL on the 4 new "mobile burger menu" tests (`getByRole('button', { name: 'Open menu' })` not found) — the button doesn't exist yet. The pre-existing tests (now with corrected skip reasons) still pass.

- [ ] **Step 3: Implement the burger menu**

Replace `frontend/src/components/Nav.tsx` entirely:

```tsx
import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Menu, ShoppingCart } from 'lucide-react'
import { getOrder } from '../api/orders'
import { useBasketStore } from '../store/basket'
import keycloak from '../auth/keycloak'
import { Button } from './ui/button'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from './ui/sheet'

export default function Nav() {
  const [menuOpen, setMenuOpen] = useState(false)
  const basketOrderId = useBasketStore((s) => s.basketOrderId)
  const { data: order } = useQuery({
    queryKey: ['order', basketOrderId],
    queryFn: () => getOrder(basketOrderId!),
    enabled: !!basketOrderId,
  })

  const itemCount = order?.items.reduce((sum, item) => sum + item.quantity, 0) ?? 0
  const isOwner = keycloak.hasRealmRole('owner')

  const links = [
    { to: '/', label: 'Store' },
    { to: '/orders', label: 'History' },
    ...(isOwner ? [{ to: '/admin/orders', label: 'Orders' }] : []),
    ...(isOwner ? [{ to: '/admin/products', label: 'Products' }] : []),
  ]

  return (
    <nav className="bg-white border-b border-border px-4 py-3 flex justify-between items-center sticky top-0 z-10">
      <div className="flex items-center gap-2">
        <Sheet open={menuOpen} onOpenChange={setMenuOpen}>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon" className="sm:hidden" aria-label="Open menu">
              <Menu />
            </Button>
          </SheetTrigger>
          <SheetContent side="left">
            <SheetHeader>
              <SheetTitle>Menu</SheetTitle>
            </SheetHeader>
            <div className="flex flex-col gap-1 mt-4">
              {links.map((link) => (
                <SheetClose asChild key={link.to}>
                  <Link to={link.to} className="rounded-md px-3 py-2 text-sm text-foreground hover:bg-accent">
                    {link.label}
                  </Link>
                </SheetClose>
              ))}
            </div>
          </SheetContent>
        </Sheet>
        <Link to="/" className="font-bold text-primary text-lg">
          Foodo
        </Link>
      </div>
      <div className="flex items-center gap-4">
        {links.map((link) => (
          <Link key={link.to} to={link.to} className="hidden sm:block text-sm text-primary hover:text-foreground">
            {link.label}
          </Link>
        ))}
        <Button variant="default" size="sm" className="rounded-full" asChild>
          <Link
            to="/basket"
            aria-label={`Basket${itemCount > 0 ? `, ${itemCount} item${itemCount !== 1 ? 's' : ''}` : ''}`}
          >
            <ShoppingCart className="h-4 w-4" />
            {itemCount > 0 ? itemCount : null}
          </Link>
        </Button>
      </div>
    </nav>
  )
}
```

- [ ] **Step 4: Run the full spec again, verify everything passes on both projects**

```bash
cd frontend && npx playwright test e2e/nav.spec.ts
```

Expected: PASS on both `chromium` and `mobile-chrome` projects, all tests.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/Nav.tsx frontend/e2e/nav.spec.ts
git commit -m "feat: add mobile burger menu to Nav via shadcn Sheet"
```

---

## Task 6: Kanban touch-drag fix + re-skin (fixes bug #1: dead touch drag)

**Files:**
- Modify: `frontend/src/pages/admin/Orders.tsx`

**Interfaces:**
- Consumes: `Card` (Task 2), `Badge` (Task 2).
- No exports change — `AdminOrders` remains the default export used by `App.tsx`.

Root cause (confirmed by reading the file): `useSensors` only registers `PointerSensor`. On a touchscreen, the browser's native scroll gesture wins the race against `PointerSensor`'s activation because nothing tells the browser to hand touch control to the drag interaction. Fix: add `TouchSensor` with a small activation delay (so a tap-to-scroll isn't mistaken for a drag start) and set `touch-action: none` on the draggable card so the browser doesn't intercept the gesture mid-drag.

`data-testid="kanban-column"`, `data-testid="column-count"`, `data-testid="order-card"`, `data-testid="loading-skeleton"`, `data-order-id`, `data-order-status`, `data-column-status` must all keep their exact current values — `kanban.spec.ts` is not being modified in this task and must pass unchanged.

Playwright's `page.mouse` API (used by `dragCardToColumn` in `e2e/fixtures/index.ts`) dispatches synthetic pointer/mouse events, not real OS-level touch events, even in the `mobile-chrome` project — so this fix cannot be proven by the existing e2e drag tests. They're the regression guard (drag must keep working via mouse/pointer simulation); real touch behavior needs a manual check.

- [ ] **Step 1: Confirm baseline — run the existing kanban e2e suite before touching the file**

```bash
cd frontend && npx playwright test e2e/kanban.spec.ts
```

Expected: PASS, all tests, both projects. This is the regression baseline for Step 3.

- [ ] **Step 2: Apply the touch-sensor fix and re-skin**

Replace `frontend/src/pages/admin/Orders.tsx` entirely:

```tsx
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import {
  DndContext,
  DragOverlay,
  useDroppable,
  useDraggable,
  type DragEndEvent,
  type DragStartEvent,
  PointerSensor,
  TouchSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import { CSS } from '@dnd-kit/utilities'
import { useState, useEffect } from 'react'
import type { CSSProperties } from 'react'
import keycloak from '../../auth/keycloak'
import { getAllOrders, acceptOrder, startOrder, finishOrder, unacceptOrder, stopOrder, unfinishOrder } from '../../api/orders'
import type { Order } from '../../types'
import { Card } from '../../components/ui/card'
import { Badge } from '../../components/ui/badge'

type KanbanStatus = 'created' | 'accepted' | 'ongoing' | 'finished'

const COLUMNS: { status: KanbanStatus; label: string }[] = [
  { status: 'created',  label: 'New Orders' },
  { status: 'accepted', label: 'Accepted' },
  { status: 'ongoing',  label: 'Ongoing' },
  { status: 'finished', label: 'Finished' },
]

const STATUS_ORDER: KanbanStatus[] = ['created', 'accepted', 'ongoing', 'finished']

// ---- Draggable card ----

function OrderCard({ order, isDragging }: { order: Order; isDragging?: boolean }) {
  const { attributes, listeners, setNodeRef, transform } = useDraggable({ id: order.id })

  const style: CSSProperties = {
    transform: CSS.Translate.toString(transform),
    opacity: isDragging ? 0.4 : 1,
    touchAction: 'none',
  }

  return (
    <Card
      ref={setNodeRef}
      style={style}
      {...listeners}
      {...attributes}
      data-testid="order-card"
      data-order-id={order.id}
      data-order-status={order.status}
      className="p-3 cursor-grab active:cursor-grabbing select-none hover:border-primary transition-colors"
    >
      <OrderCardBody order={order} />
    </Card>
  )
}

function OrderCardBody({ order }: { order: Order }) {
  return (
    <>
      <div className="flex items-start justify-between gap-2 mb-1">
        <span className="text-xs font-medium text-primary truncate">
          {order.user_name ?? 'Unknown customer'}
        </span>
        <span className="text-xs text-muted-foreground whitespace-nowrap">
          {new Date(order.created_at).toLocaleDateString()}
        </span>
      </div>
      <ul className="mt-1 space-y-0.5">
        {order.items.map((item) => (
          <li key={item.product_id} className="text-xs text-foreground">
            {item.product_name}
            {item.unit ? ` · ${item.unit} × ${item.quantity}` : ` × ${item.quantity}`}
          </li>
        ))}
      </ul>
    </>
  )
}

// ---- Droppable column ----

function KanbanColumn({
  status,
  label,
  orders,
  activeId,
}: {
  status: KanbanStatus
  label: string
  orders: Order[]
  activeId: string | null
}) {
  const { setNodeRef, isOver } = useDroppable({ id: status })

  return (
    <div
      ref={setNodeRef}
      data-testid="kanban-column"
      data-column-status={status}
      className={`flex flex-col min-h-[200px] rounded-xl p-3 transition-colors ${
        isOver ? 'bg-accent ring-2 ring-primary' : 'bg-secondary'
      }`}
    >
      <div className="flex items-center justify-between mb-3">
        <h2 className="font-semibold text-sm text-foreground">{label}</h2>
        <Badge data-testid="column-count" variant="secondary">
          {orders.length}
        </Badge>
      </div>
      <div className="flex flex-col gap-2 flex-1">
        {orders.map((order) => (
          <OrderCard key={order.id} order={order} isDragging={order.id === activeId} />
        ))}
        {orders.length === 0 && (
          <p className="text-xs text-muted-foreground text-center mt-4">No orders</p>
        )}
      </div>
    </div>
  )
}

// ---- Main page ----

export default function AdminOrders() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [activeId, setActiveId] = useState<string | null>(null)
  const isOwner = keycloak.hasRealmRole('owner')

  const { data: allOrders = [], isLoading, isError } = useQuery({
    queryKey: ['all-orders'],
    queryFn: getAllOrders,
    refetchInterval: 30_000,
    enabled: isOwner,
  })

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['all-orders'] })

  const accept   = useMutation({ mutationFn: acceptOrder,   onSuccess: invalidate, onError: () => toast.error('Failed to accept order') })
  const start    = useMutation({ mutationFn: startOrder,    onSuccess: invalidate, onError: () => toast.error('Failed to start order') })
  const finish   = useMutation({ mutationFn: finishOrder,   onSuccess: invalidate, onError: () => toast.error('Failed to finish order') })
  const unaccept = useMutation({ mutationFn: unacceptOrder, onSuccess: invalidate, onError: () => toast.error('Failed to move order back') })
  const stop     = useMutation({ mutationFn: stopOrder,     onSuccess: invalidate, onError: () => toast.error('Failed to move order back') })
  const unfinish = useMutation({ mutationFn: unfinishOrder, onSuccess: invalidate, onError: () => toast.error('Failed to move order back') })

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(TouchSensor, { activationConstraint: { delay: 150, tolerance: 5 } }),
  )

  useEffect(() => {
    if (!isOwner) navigate('/')
  }, [isOwner, navigate])

  if (!isOwner) return null

  // Only show non-pending orders in the kanban
  const kanbanOrders = allOrders.filter(
    (o): o is Order & { status: KanbanStatus } => o.status !== 'pending'
  )

  const columnOrders = (status: KanbanStatus) =>
    kanbanOrders.filter((o) => o.status === status)

  const activeOrder = activeId ? kanbanOrders.find((o) => o.id === activeId) : null

  function handleDragStart({ active }: DragStartEvent) {
    setActiveId(active.id as string)
  }

  function handleDragEnd({ active, over }: DragEndEvent) {
    setActiveId(null)
    if (!over) return

    const fromOrder = kanbanOrders.find((o) => o.id === active.id)
    if (!fromOrder) return

    const fromIdx = STATUS_ORDER.indexOf(fromOrder.status as KanbanStatus)
    const toIdx   = STATUS_ORDER.indexOf(over.id as KanbanStatus)

    const diff = toIdx - fromIdx
    if (diff !== 1 && diff !== -1) return // one step at a time, either direction

    const orderId = active.id as string
    if (diff === 1) {
      if (toIdx === 1) accept.mutate(orderId)
      else if (toIdx === 2) start.mutate(orderId)
      else if (toIdx === 3) finish.mutate(orderId)
    } else {
      if (fromIdx === 1) unaccept.mutate(orderId)
      else if (fromIdx === 2) stop.mutate(orderId)
      else if (fromIdx === 3) unfinish.mutate(orderId)
    }
  }

  if (isLoading) {
    return <div data-testid="loading-skeleton" className="animate-pulse bg-white rounded-lg h-48 border border-border" />
  }

  if (isError) {
    return (
      <div className="text-center py-16">
        <p className="text-muted-foreground">Could not load orders. Try again.</p>
      </div>
    )
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-foreground mb-6">Order Board</h1>
      <p className="text-sm text-muted-foreground mb-6">
        Drag orders left or right to change their status.
      </p>

      <DndContext sensors={sensors} onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {COLUMNS.map(({ status, label }) => (
            <KanbanColumn
              key={status}
              status={status}
              label={label}
              orders={columnOrders(status)}
              activeId={activeId}
            />
          ))}
        </div>

        <DragOverlay>
          {activeOrder && (
            <Card className="border-2 border-primary p-3 shadow-lg rotate-1 w-64">
              <OrderCardBody order={activeOrder} />
            </Card>
          )}
        </DragOverlay>
      </DndContext>
    </div>
  )
}
```

- [ ] **Step 3: Run the kanban e2e suite again, verify no regression**

```bash
cd frontend && npx playwright test e2e/kanban.spec.ts
```

Expected: PASS, all tests, both projects — identical result to Step 1. `kanban.spec.ts` was not modified.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/admin/Orders.tsx
git commit -m "fix: register TouchSensor and touch-action:none so kanban drag works on touchscreens"
```

- [ ] **Step 5 (manual, non-blocking): verify on a real touchscreen or Chrome DevTools touch emulation**

Run `npm run dev` in `frontend/`, open the app on a phone or with DevTools device toolbar + touch simulation enabled, go to `/admin/orders` as an owner, and drag a card between adjacent columns with a finger/touch pointer. This step exists because Playwright's `page.mouse`-based drag (used throughout `kanban.spec.ts`) does not exercise real touch input, so it's the only way to confirm the actual reported bug is fixed. Not required to mark the task complete, but should be done before considering the mobile bug closed.

---

## Task 7: ProductCard + Store re-skin

**Files:**
- Modify: `frontend/src/components/ProductCard.tsx`
- Modify: `frontend/src/pages/Store.tsx`

**Interfaces:**
- Consumes: `Card`, `CardContent`, `Button` (Task 2).
- No exports change.

`store.spec.ts` asserts `getByRole('button', { name: 'Add' })` count and product name/description/unit text — all preserved verbatim.

- [ ] **Step 1: Confirm baseline**

```bash
cd frontend && npx playwright test e2e/store.spec.ts
```

Expected: PASS, all tests.

- [ ] **Step 2: Re-skin ProductCard**

Replace `frontend/src/components/ProductCard.tsx` entirely:

```tsx
import { useNavigate } from 'react-router-dom'
import type { Product } from '../types'
import { Card, CardContent } from './ui/card'
import { Button } from './ui/button'

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

interface Props {
  product: Product
  onAdd: () => void
  disabled?: boolean
}

export default function ProductCard({ product, onAdd, disabled }: Props) {
  const navigate = useNavigate()

  return (
    <Card
      className="overflow-hidden flex flex-col cursor-pointer hover:border-primary transition-colors"
      onClick={() => navigate(`/products/${product.id}`)}
    >
      {product.image_url ? (
        <img
          src={`${BASE_URL}${product.image_url}`}
          alt={product.name}
          className="w-full h-36 object-cover"
        />
      ) : (
        <div className="w-full h-36 bg-secondary" />
      )}
      <CardContent className="p-4 flex flex-col gap-2 flex-1">
        <h3 className="font-semibold text-foreground">{product.name}</h3>
        <p className="text-sm text-muted-foreground flex-1">{product.description}</p>
        <p className="text-xs text-muted-foreground">{product.unit}</p>
        <Button
          onClick={(e) => { e.stopPropagation(); onAdd() }}
          disabled={disabled}
          size="sm"
          className="mt-2"
        >
          Add
        </Button>
      </CardContent>
    </Card>
  )
}
```

- [ ] **Step 3: Re-skin Store**

Replace `frontend/src/pages/Store.tsx` entirely:

```tsx
import { useQuery } from '@tanstack/react-query'
import { getProducts } from '../api/products'
import ProductCard from '../components/ProductCard'
import { useBasket } from '../hooks/useBasket'
import { Card } from '../components/ui/card'
import { Button } from '../components/ui/button'

function SkeletonGrid() {
  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
      {Array.from({ length: 6 }).map((_, i) => (
        <Card key={i} className="h-40 animate-pulse" />
      ))}
    </div>
  )
}

export default function Store() {
  const { data: products, isLoading, isError, refetch } = useQuery({
    queryKey: ['products'],
    queryFn: getProducts,
  })
  const { addItem, isValidating } = useBasket()

  if (isLoading) return <SkeletonGrid />

  if (isError) {
    return (
      <div className="text-center py-16">
        <p className="text-muted-foreground mb-4">Could not load products. Try again.</p>
        <Button onClick={() => refetch()}>Retry</Button>
      </div>
    )
  }

  const available = products?.filter((p) => p.available) ?? []

  if (available.length === 0) {
    return <p className="text-center py-16 text-muted-foreground">No products available yet.</p>
  }

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
      {available.map((p) => (
        <ProductCard
          key={p.id}
          product={p}
          onAdd={() => addItem(p)}
          disabled={isValidating}
        />
      ))}
    </div>
  )
}
```

- [ ] **Step 4: Run store.spec.ts again, verify no regression**

```bash
cd frontend && npx playwright test e2e/store.spec.ts
```

Expected: PASS, all tests — identical result to Step 1.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/ProductCard.tsx frontend/src/pages/Store.tsx
git commit -m "refactor: re-skin ProductCard and Store on shadcn Card/Button"
```

---

## Task 8: Product detail page re-skin

**Files:**
- Modify: `frontend/src/pages/Product.tsx`

**Interfaces:**
- Consumes: `Card` (Task 2), `Button` (Task 2), `ArrowLeft` icon from `lucide-react` (Task 4 installed the package).
- No exports change.

`store.spec.ts`'s "Product detail page" tests assert product name and description text only — preserved verbatim. No spec asserts on the "Back" button's exact text, so `← Back` becomes an icon + "Back" without changing behavior.

- [ ] **Step 1: Confirm baseline**

```bash
cd frontend && npx playwright test e2e/store.spec.ts --grep "Product detail"
```

Expected: PASS.

- [ ] **Step 2: Re-skin**

Replace `frontend/src/pages/Product.tsx` entirely:

```tsx
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { ArrowLeft } from 'lucide-react'
import { getProduct } from '../api/products'
import { useBasket } from '../hooks/useBasket'
import { Card } from '../components/ui/card'
import { Button } from '../components/ui/button'

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export default function Product() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { addItem, isValidating } = useBasket()

  const { data: product, isLoading, isError } = useQuery({
    queryKey: ['product', id],
    queryFn: () => getProduct(id!),
    enabled: !!id,
  })

  if (isLoading) {
    return (
      <div className="max-w-lg mx-auto mt-8 animate-pulse">
        <Card className="overflow-hidden">
          <div className="w-full h-64 bg-secondary" />
          <div className="p-6 space-y-3">
            <div className="h-6 bg-secondary rounded w-1/2" />
            <div className="h-4 bg-secondary rounded w-full" />
            <div className="h-4 bg-secondary rounded w-2/3" />
          </div>
        </Card>
      </div>
    )
  }

  if (isError || !product) {
    return (
      <div className="text-center py-16">
        <p className="text-muted-foreground mb-4">Product not found.</p>
        <Button variant="link" onClick={() => navigate('/')}>
          Back to store
        </Button>
      </div>
    )
  }

  return (
    <div className="max-w-lg mx-auto mt-4">
      <Button
        variant="ghost"
        size="sm"
        onClick={() => navigate(-1)}
        className="mb-4 text-muted-foreground hover:text-primary"
      >
        <ArrowLeft className="h-4 w-4" />
        Back
      </Button>

      <Card className="overflow-hidden">
        {product.image_url ? (
          <img
            src={`${BASE_URL}${product.image_url}`}
            alt={product.name}
            className="w-full h-64 object-cover"
          />
        ) : (
          <div className="w-full h-64 bg-secondary" />
        )}

        <div className="p-6 space-y-4">
          <div>
            <h1 className="text-2xl font-bold text-foreground">{product.name}</h1>
            <span className="text-xs text-muted-foreground uppercase tracking-wide">{product.unit}</span>
          </div>

          <p className="text-primary leading-relaxed">{product.description}</p>

          <div className="flex items-center gap-2 text-sm">
            <span
              className={`inline-block w-2 h-2 rounded-full ${product.available ? 'bg-green-500' : 'bg-muted'}`}
            />
            <span className="text-muted-foreground">
              {product.available ? 'Available' : 'Not available'}
            </span>
          </div>

          {product.available && (
            <Button onClick={() => addItem(product)} disabled={isValidating} className="w-full">
              Add to basket
            </Button>
          )}
        </div>
      </Card>
    </div>
  )
}
```

- [ ] **Step 3: Run the spec again, verify no regression**

```bash
cd frontend && npx playwright test e2e/store.spec.ts --grep "Product detail"
```

Expected: PASS — identical result to Step 1.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/Product.tsx
git commit -m "refactor: re-skin Product detail page on shadcn Card/Button"
```

---

## Task 9: Basket page re-skin

**Files:**
- Modify: `frontend/src/pages/Basket.tsx`

**Interfaces:**
- Consumes: `Card` (Task 2), `Button` (Task 2).
- No exports change.

`basket.spec.ts` strictly asserts `getByRole('button', { name: '−' })` ×2, `{ name: '+' }` ×2, `{ name: 'Remove' }` ×2, and `{ name: 'Place Order' }` (visible/enabled/disabled) plus the "Your Basket" heading. Every one of those exact strings is preserved.

- [ ] **Step 1: Confirm baseline**

```bash
cd frontend && npx playwright test e2e/basket.spec.ts
```

Expected: PASS, all tests.

- [ ] **Step 2: Re-skin**

Replace `frontend/src/pages/Basket.tsx` entirely:

```tsx
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getOrder } from '../api/orders'
import { getProducts } from '../api/products'
import { useBasket } from '../hooks/useBasket'
import { useBasketStore } from '../store/basket'
import { Card } from '../components/ui/card'
import { Button } from '../components/ui/button'

export default function Basket() {
  const navigate = useNavigate()
  const basketOrderId = useBasketStore((s) => s.basketOrderId)
  const { removeItem, updateQuantity, confirm } = useBasket()
  const [confirmError, setConfirmError] = useState<string | null>(null)
  const [confirming, setConfirming] = useState(false)

  const { data: order, isLoading } = useQuery({
    queryKey: ['order', basketOrderId],
    queryFn: () => getOrder(basketOrderId!),
    enabled: !!basketOrderId,
  })

  const { data: products } = useQuery({
    queryKey: ['products'],
    queryFn: getProducts,
  })

  if (!basketOrderId || (!isLoading && !order)) {
    return (
      <div className="text-center py-16">
        <p className="text-muted-foreground mb-4">Your basket is empty</p>
        <Link to="/" className="text-primary underline">
          Browse products
        </Link>
      </div>
    )
  }

  if (isLoading) {
    return <Card className="h-48 animate-pulse" />
  }

  // productMap used as fallback for items that may lack unit (e.g. older orders)
  const productMap = new Map(products?.map((p) => [p.id, p]) ?? [])

  async function handleConfirm() {
    setConfirmError(null)
    setConfirming(true)
    try {
      const id = await confirm()
      if (id) navigate(`/orders/${id}`)
    } catch {
      setConfirmError('Could not place order. Try again.')
    } finally {
      setConfirming(false)
    }
  }

  return (
    <div className="max-w-lg mx-auto">
      <h1 className="text-2xl font-bold text-foreground mb-6">Your Basket</h1>

      {order?.items.length === 0 ? (
        <p className="text-muted-foreground">
          No items yet.{' '}
          <Link to="/" className="underline text-primary">
            Add some
          </Link>
        </p>
      ) : (
        <ul className="space-y-4 mb-8">
          {order?.items.map((item) => {
            const unit = item.unit || productMap.get(item.product_id)?.unit
            return (
              <li key={item.product_id}>
                <Card className="flex items-center gap-4 p-4">
                  <div className="flex-1">
                    <p className="font-medium text-foreground">{item.product_name}</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="icon"
                      className="h-7 w-7"
                      onClick={() =>
                        item.quantity === 1
                          ? removeItem(item.product_id)
                          : updateQuantity(item.product_id, item.quantity - 1)
                      }
                    >
                      −
                    </Button>
                    <span className="text-center text-foreground min-w-[3rem]">
                      {unit ? `${unit} ${item.quantity}` : item.quantity}
                    </span>
                    <Button
                      variant="outline"
                      size="icon"
                      className="h-7 w-7"
                      onClick={() => updateQuantity(item.product_id, item.quantity + 1)}
                    >
                      +
                    </Button>
                    <Button
                      variant="link"
                      className="ml-2 text-red-400 hover:text-red-600 h-auto p-0"
                      onClick={() => removeItem(item.product_id)}
                    >
                      Remove
                    </Button>
                  </div>
                </Card>
              </li>
            )
          })}
        </ul>
      )}

      {confirmError && <p className="text-red-600 mb-4 text-sm">{confirmError}</p>}

      <Button
        onClick={handleConfirm}
        disabled={confirming || !order?.items.length}
        className="w-full py-3"
        size="lg"
      >
        {confirming ? 'Placing order…' : 'Place Order'}
      </Button>

      <div className="mt-8 text-center sm:hidden">
        <Link to="/orders" className="text-primary text-sm underline">
          View order history
        </Link>
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Run basket.spec.ts again, verify no regression**

```bash
cd frontend && npx playwright test e2e/basket.spec.ts
```

Expected: PASS, all tests — identical result to Step 1.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/Basket.tsx
git commit -m "refactor: re-skin Basket page on shadcn Card/Button"
```

---

## Task 10: StatusBadge, OrderHistory, OrderStatus re-skin

**Files:**
- Modify: `frontend/src/components/StatusBadge.tsx`
- Modify: `frontend/src/pages/OrderHistory.tsx`
- Modify: `frontend/src/pages/OrderStatus.tsx`

**Interfaces:**
- Consumes: `Badge` (Task 2), `Card` (Task 2), `ChevronRight` icon from `lucide-react` (Task 4 installed the package).
- `StatusBadge` keeps its existing signature: `({ status }: { status: Order['status'] }) => JSX.Element`, default export, consumed unchanged by `OrderHistory.tsx`, `OrderStatus.tsx`, and `admin/Orders.tsx` (not used there currently, no change needed).

Neither page has dedicated e2e coverage beyond `nav.spec.ts`'s "shows order history heading" smoke test — the exact five status colors (`amber`/`blue`/`purple`/`orange`/`green`) are preserved via `Badge`'s `className` override (Tailwind-merge lets the arbitrary background/text colors win over the default variant).

- [ ] **Step 1: Confirm baseline**

```bash
cd frontend && npx playwright test e2e/nav.spec.ts --grep "Order history page"
```

Expected: PASS.

- [ ] **Step 2: Re-skin StatusBadge**

Replace `frontend/src/components/StatusBadge.tsx` entirely:

```tsx
import type { Order } from '../types'
import { Badge } from './ui/badge'

const colours: Record<Order['status'], string> = {
  pending:  'bg-amber-100 text-amber-800',
  created:  'bg-blue-100 text-blue-800',
  accepted: 'bg-purple-100 text-purple-800',
  ongoing:  'bg-orange-100 text-orange-800',
  finished: 'bg-green-100 text-green-800',
}

export default function StatusBadge({ status }: { status: Order['status'] }) {
  return <Badge className={colours[status]}>{status}</Badge>
}
```

- [ ] **Step 3: Re-skin OrderHistory**

Replace `frontend/src/pages/OrderHistory.tsx` entirely:

```tsx
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { ChevronRight } from 'lucide-react'
import { getOrders } from '../api/orders'
import StatusBadge from '../components/StatusBadge'
import { Card } from '../components/ui/card'

export default function OrderHistory() {
  const { data: orders, isLoading, isError } = useQuery({
    queryKey: ['orders'],
    queryFn: getOrders,
  })

  if (isLoading) {
    return <Card className="h-48 animate-pulse" />
  }

  if (isError) {
    return (
      <div className="text-center py-16">
        <p className="text-muted-foreground">Could not load orders. Try again.</p>
      </div>
    )
  }

  const sorted = [...(orders ?? [])].sort(
    (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  )

  return (
    <div className="max-w-lg mx-auto">
      <h1 className="text-2xl font-bold text-foreground mb-6">Order History</h1>

      {sorted.length === 0 ? (
        <p className="text-muted-foreground">
          No orders yet.{' '}
          <Link to="/" className="underline text-primary">
            Place your first order
          </Link>
        </p>
      ) : (
        <ul className="space-y-3">
          {sorted.map((order) => (
            <li key={order.id}>
              <Link to={`/orders/${order.id}`}>
                <Card className="flex items-center justify-between p-4 hover:border-primary transition-colors">
                  <div>
                    <StatusBadge status={order.status} />
                    <p className="text-sm text-muted-foreground mt-1">
                      {new Date(order.created_at).toLocaleDateString()}
                    </p>
                  </div>
                  <ChevronRight className="h-4 w-4 text-primary" />
                </Card>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
```

- [ ] **Step 4: Re-skin OrderStatus**

Replace `frontend/src/pages/OrderStatus.tsx` entirely:

```tsx
import { useParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getOrder } from '../api/orders'
import StatusBadge from '../components/StatusBadge'
import { Card } from '../components/ui/card'
import type { Order } from '../types'

const TERMINAL = new Set<Order['status']>(['finished'])

export default function OrderStatus() {
  const { id } = useParams<{ id: string }>()

  const { data: order, isLoading, isError } = useQuery({
    queryKey: ['order', id],
    queryFn: () => getOrder(id!),
    enabled: !!id,
    refetchInterval: (query) => {
      const status = query.state.data?.status
      return status && TERMINAL.has(status) ? false : 10_000
    },
  })

  if (isLoading) {
    return <Card className="h-48 animate-pulse" />
  }

  if (isError) {
    return (
      <div className="text-center py-16">
        <p className="text-muted-foreground">Could not load order. Try again.</p>
      </div>
    )
  }

  if (!order) {
    return <p className="text-center py-16 text-muted-foreground">Order not found.</p>
  }

  return (
    <div className="max-w-lg mx-auto">
      <div className="flex items-center gap-3 mb-2">
        <h1 className="text-2xl font-bold text-foreground">Order</h1>
        <StatusBadge status={order.status} />
      </div>
      <p className="text-sm text-muted-foreground mb-6">
        Placed {new Date(order.created_at).toLocaleDateString()}
      </p>

      <ul className="space-y-3 mb-8">
        {order.items.map((item) => (
          <li key={item.product_id}>
            <Card className="flex items-center justify-between p-4">
              <span className="text-foreground">{item.product_name}</span>
              <span className="text-muted-foreground">
                {item.unit ? `${item.unit} × ${item.quantity}` : `× ${item.quantity}`}
              </span>
            </Card>
          </li>
        ))}
      </ul>

      <Link to="/orders" className="text-primary text-sm underline">
        ← Order history
      </Link>
    </div>
  )
}
```

- [ ] **Step 5: Run the baseline spec again, then full build, verify no regression**

```bash
cd frontend && npx playwright test e2e/nav.spec.ts --grep "Order history page" && npm run build
```

Expected: both PASS.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/StatusBadge.tsx frontend/src/pages/OrderHistory.tsx frontend/src/pages/OrderStatus.tsx
git commit -m "refactor: re-skin StatusBadge, OrderHistory, OrderStatus on shadcn Badge/Card"
```

---

## Task 11: admin/Products re-skin — Dialog modal, form primitives

**Files:**
- Modify: `frontend/src/pages/admin/Products.tsx`
- Create: `frontend/e2e/admin-products.spec.ts`

**Interfaces:**
- Consumes: `Button`, `Input`, `Label`, `Textarea`, `Switch` (Task 3), `Dialog`/`DialogContent`/`DialogHeader`/`DialogFooter`/`DialogTitle` (Task 4).
- No exports change.

This page has **no existing e2e coverage** (confirmed: `frontend/e2e/` only has `basket.spec.ts`, `kanban.spec.ts`, `nav.spec.ts`, `store.spec.ts`). Its custom `fixed inset-0` div-based modal is being replaced wholesale by Radix `Dialog`, and its checkbox/toggle-pill availability control by `Switch` — real behavioral changes worth locking down, so this task adds a new spec rather than relying on manual checks. `setupApiMocks` only mocks `GET /api/v1/products`; create/update/availability routes are mocked per-test via `page.route()` overrides, following the pattern already used in `kanban.spec.ts`.

- [ ] **Step 1: Write the new e2e spec (it will fail — page not migrated yet)**

Create `frontend/e2e/admin-products.spec.ts`:

```ts
import { test, expect, setupApiMocks, setRoles } from './fixtures'
import { PRODUCTS } from './mocks/data'

test.describe('Admin Products — layout', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
    await setupApiMocks(page)
  })

  test('shows Products heading and existing products', async ({ page }) => {
    await page.goto('/admin/products')

    await expect(page.getByRole('heading', { name: 'Products' })).toBeVisible()
    await expect(page.getByRole('cell', { name: 'Sourdough Loaf' })).toBeVisible()
    await expect(page.getByRole('cell', { name: 'Croissant' })).toBeVisible()
  })

  test('shows an availability switch matching each product state', async ({ page }) => {
    await page.goto('/admin/products')

    await expect(page.getByRole('switch', { name: 'Mark unavailable' })).toHaveCount(
      PRODUCTS.filter((p) => p.available).length,
    )
    await expect(page.getByRole('switch', { name: 'Mark available' })).toHaveCount(
      PRODUCTS.filter((p) => !p.available).length,
    )
  })
})

test.describe('Admin Products — create', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
    await setupApiMocks(page)
  })

  test('opens an empty dialog when adding a product', async ({ page }) => {
    await page.goto('/admin/products')

    await page.getByRole('button', { name: '+ Add product' }).click()

    const dialog = page.getByRole('dialog')
    await expect(dialog.getByRole('heading', { name: 'Add product' })).toBeVisible()
    await expect(dialog.getByLabel('Name')).toHaveValue('')
    await expect(dialog.getByLabel('Unit')).toHaveValue('')
  })

  test('submitting the form creates a product and closes the dialog', async ({ page }) => {
    let createCalled = false
    await page.route('http://localhost:8080/api/v1/products', (route) => {
      if (route.request().method() === 'POST') {
        createCalled = true
        route.fulfill({
          json: { id: 'prod-new', name: 'Rye Bread', description: '', unit: 'loaf', available: false },
        })
      } else {
        route.continue()
      }
    })

    await page.goto('/admin/products')
    await page.getByRole('button', { name: '+ Add product' }).click()

    const dialog = page.getByRole('dialog')
    await dialog.getByLabel('Name').fill('Rye Bread')
    await dialog.getByLabel('Unit').fill('loaf')
    await dialog.getByRole('button', { name: 'Create product' }).click()

    await expect(dialog).not.toBeVisible()
    expect(createCalled).toBe(true)
  })
})

test.describe('Admin Products — edit', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
    await setupApiMocks(page)
  })

  test('opens the dialog pre-filled with the product being edited', async ({ page }) => {
    await page.goto('/admin/products')

    await page
      .getByRole('row', { name: /Sourdough Loaf/ })
      .getByRole('button', { name: 'Edit' })
      .click()

    const dialog = page.getByRole('dialog')
    await expect(dialog.getByRole('heading', { name: 'Edit product' })).toBeVisible()
    await expect(dialog.getByLabel('Name')).toHaveValue('Sourdough Loaf')
    await expect(dialog.getByLabel('Unit')).toHaveValue('1 loaf')
  })
})

test.describe('Admin Products — availability toggle', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
    await setupApiMocks(page)
  })

  test('toggling the switch calls the availability endpoint', async ({ page }) => {
    let availabilityCalled = false
    await page.route('http://localhost:8080/api/v1/products/prod-1/availability', (route) => {
      availabilityCalled = true
      route.fulfill({ json: { ...PRODUCTS[0], available: false } })
    })

    await page.goto('/admin/products')

    await page.getByRole('switch', { name: 'Mark unavailable' }).first().click()

    expect(availabilityCalled).toBe(true)
  })
})

test.describe('Admin Products — access control', () => {
  test('non-owner is redirected away from admin products', async ({ page }) => {
    await setRoles(page, [])
    await setupApiMocks(page)
    await page.goto('/admin/products')

    await page.waitForURL('/')
    await expect(page).toHaveURL('/')
  })
})
```

- [ ] **Step 2: Run the new spec, verify it fails**

```bash
cd frontend && npx playwright test e2e/admin-products.spec.ts
```

Expected: FAIL on most tests — the switch/dialog roles don't exist yet against the current custom-modal/checkbox implementation. The access-control test may already pass (unrelated to the re-skin).

- [ ] **Step 3: Re-skin the page**

Replace `frontend/src/pages/admin/Products.tsx` entirely:

```tsx
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import keycloak from '../../auth/keycloak'
import {
  getProducts,
  createProduct,
  updateProduct,
  deleteProduct,
  setAvailability,
  uploadImage,
} from '../../api/products'
import type { ProductInput } from '../../api/products'
import type { Product } from '../../types'
import { Button } from '../../components/ui/button'
import { Input } from '../../components/ui/input'
import { Label } from '../../components/ui/label'
import { Textarea } from '../../components/ui/textarea'
import { Switch } from '../../components/ui/switch'
import { Dialog, DialogContent, DialogHeader, DialogFooter, DialogTitle } from '../../components/ui/dialog'

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

const emptyForm: ProductInput = { name: '', description: '', unit: '', available: false }

export default function AdminProducts() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const { data: products = [], isLoading, isError } = useQuery({
    queryKey: ['products'],
    queryFn: getProducts,
  })

  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Product | null>(null)
  const [form, setForm] = useState<ProductInput>(emptyForm)
  const [imageFile, setImageFile] = useState<File | null>(null)
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['products'] })

  const availabilityMutation = useMutation({
    mutationFn: ({ id, available }: { id: string; available: boolean }) =>
      setAvailability(id, available),
    onSuccess: invalidate,
    onError: () => toast.error('Failed to update availability'),
  })

  const deleteMutation = useMutation({
    mutationFn: deleteProduct,
    onSuccess: () => {
      invalidate()
      setDeleteConfirm(null)
      toast.success('Product deleted')
    },
    onError: () => toast.error('Failed to delete product'),
  })

  if (!keycloak.hasRealmRole('owner')) {
    navigate('/')
    return null
  }

  const openCreate = () => {
    setEditing(null)
    setForm(emptyForm)
    setImageFile(null)
    setModalOpen(true)
  }

  const openEdit = (p: Product) => {
    setEditing(p)
    setForm({ name: p.name, description: p.description, unit: p.unit, available: p.available })
    setImageFile(null)
    setModalOpen(true)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      let product: Product
      if (editing) {
        product = await updateProduct(editing.id, form)
        toast.success('Product updated')
      } else {
        product = await createProduct(form)
        toast.success('Product created')
      }
      if (imageFile) {
        await uploadImage(product.id, imageFile)
      }
      invalidate()
      setModalOpen(false)
    } catch {
      toast.error('Something went wrong')
    } finally {
      setSubmitting(false)
    }
  }

  if (isLoading) {
    return <div className="text-center py-16 text-muted-foreground">Loading…</div>
  }

  if (isError) {
    return <div className="text-center py-16 text-muted-foreground">Failed to load products.</div>
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-primary">Products</h1>
        <Button onClick={openCreate}>+ Add product</Button>
      </div>

      {products.length === 0 ? (
        <p className="text-muted-foreground text-center py-16">No products yet.</p>
      ) : (
        <div className="bg-card rounded-lg border border-border overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-left text-muted-foreground">
                <th className="px-4 py-3 w-12"></th>
                <th className="px-4 py-3">Name</th>
                <th className="px-4 py-3 hidden sm:table-cell">Unit</th>
                <th className="px-4 py-3 hidden md:table-cell">Description</th>
                <th className="px-4 py-3">Available</th>
                <th className="px-4 py-3"></th>
              </tr>
            </thead>
            <tbody>
              {products.map((p) => (
                <tr key={p.id} className="border-b border-border last:border-0">
                  <td className="px-4 py-3">
                    {p.image_url ? (
                      <img
                        src={`${BASE_URL}${p.image_url}`}
                        alt={p.name}
                        className="w-10 h-10 object-cover rounded"
                      />
                    ) : (
                      <div className="w-10 h-10 bg-secondary rounded flex items-center justify-center text-muted-foreground text-xs">
                        No img
                      </div>
                    )}
                  </td>
                  <td className="px-4 py-3 font-medium text-foreground">{p.name}</td>
                  <td className="px-4 py-3 hidden sm:table-cell text-muted-foreground">{p.unit}</td>
                  <td className="px-4 py-3 hidden md:table-cell text-muted-foreground max-w-xs truncate">
                    {p.description}
                  </td>
                  <td className="px-4 py-3">
                    <Switch
                      checked={p.available}
                      onCheckedChange={(checked) =>
                        availabilityMutation.mutate({ id: p.id, available: checked })
                      }
                      aria-label={p.available ? 'Mark unavailable' : 'Mark available'}
                    />
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2 justify-end">
                      {deleteConfirm === p.id ? (
                        <>
                          <span className="text-muted-foreground text-xs">Delete?</span>
                          <Button
                            variant="link"
                            className="text-red-600 text-xs h-auto p-0"
                            onClick={() => deleteMutation.mutate(p.id)}
                          >
                            Yes
                          </Button>
                          <Button
                            variant="link"
                            className="text-muted-foreground text-xs h-auto p-0"
                            onClick={() => setDeleteConfirm(null)}
                          >
                            No
                          </Button>
                        </>
                      ) : (
                        <>
                          <Button
                            variant="link"
                            className="text-primary text-xs h-auto p-0"
                            onClick={() => openEdit(p)}
                          >
                            Edit
                          </Button>
                          <Button
                            variant="link"
                            className="text-red-500 text-xs h-auto p-0"
                            onClick={() => setDeleteConfirm(p.id)}
                          >
                            Delete
                          </Button>
                        </>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <Dialog open={modalOpen} onOpenChange={setModalOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? 'Edit product' : 'Add product'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className="px-6 py-4 space-y-4">
            <div className="space-y-1">
              <Label htmlFor="product-name">
                Name <span className="text-red-400">*</span>
              </Label>
              <Input
                id="product-name"
                required
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              />
            </div>
            <div className="space-y-1">
              <Label htmlFor="product-unit">
                Unit <span className="text-red-400">*</span>
              </Label>
              <Input
                id="product-unit"
                required
                placeholder="e.g. loaf, dozen"
                value={form.unit}
                onChange={(e) => setForm((f) => ({ ...f, unit: e.target.value }))}
              />
            </div>
            <div className="space-y-1">
              <Label htmlFor="product-description">Description</Label>
              <Textarea
                id="product-description"
                rows={3}
                value={form.description}
                onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div className="flex items-center gap-2">
              <Switch
                id="available"
                checked={form.available}
                onCheckedChange={(checked) => setForm((f) => ({ ...f, available: checked }))}
              />
              <Label htmlFor="available">Available for ordering</Label>
            </div>
            <div>
              <Label className="block mb-1">Image</Label>
              {editing?.image_url && !imageFile && (
                <img
                  src={`${BASE_URL}${editing.image_url}`}
                  alt="current"
                  className="w-16 h-16 object-cover rounded mb-2"
                />
              )}
              <Label
                htmlFor="product-image"
                className="inline-flex items-center gap-2 cursor-pointer border border-primary text-primary rounded px-3 py-1.5 text-sm hover:bg-primary hover:text-primary-foreground transition-colors font-normal"
              >
                {imageFile ? imageFile.name : 'Choose image…'}
                <input
                  id="product-image"
                  type="file"
                  accept="image/*"
                  onChange={(e) => setImageFile(e.target.files?.[0] ?? null)}
                  className="sr-only"
                />
              </Label>
              {imageFile && (
                <Button
                  type="button"
                  variant="link"
                  className="ml-2 text-xs text-muted-foreground hover:text-red-500 h-auto p-0"
                  onClick={() => setImageFile(null)}
                >
                  Remove
                </Button>
              )}
            </div>
            <DialogFooter className="px-0 pb-0">
              <Button type="button" variant="ghost" onClick={() => setModalOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={submitting}>
                {submitting ? 'Saving…' : editing ? 'Save changes' : 'Create product'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
```

- [ ] **Step 4: Run the new spec again, verify it passes**

```bash
cd frontend && npx playwright test e2e/admin-products.spec.ts
```

Expected: PASS, all tests, both projects.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/admin/Products.tsx frontend/e2e/admin-products.spec.ts
git commit -m "refactor: re-skin admin Products on shadcn Dialog/Input/Switch, add e2e coverage"
```

---

## Task 12: Cleanup + full-suite regression run

**Files:**
- Delete: `frontend/src/App.css`

**Interfaces:** none — this task touches no component contracts, it only removes dead code and verifies the whole migration end to end.

`App.css` is a leftover from the original Vite template. Confirmed unused: `grep -rn "App.css" frontend/src frontend/*.html` returns zero matches — nothing imports it.

- [ ] **Step 1: Delete the dead file**

```bash
cd frontend && rm src/App.css
```

- [ ] **Step 2: Run the full lint + typecheck + build**

```bash
cd frontend && npm run lint && npm run build
```

Expected: both PASS with zero errors/warnings.

- [ ] **Step 3: Run the full vitest suite**

```bash
cd frontend && npm run test
```

Expected: PASS — all existing tests (api/client, hooks/useBasket, store/basket) plus the new `src/components/ui/*.test.tsx` files from Tasks 2–3.

- [ ] **Step 4: Run the full Playwright suite**

```bash
cd frontend && npm run test:e2e
```

Expected: PASS on both `chromium` and `mobile-chrome` projects — `basket.spec.ts`, `kanban.spec.ts`, `nav.spec.ts`, `store.spec.ts`, `admin-products.spec.ts`, all green.

- [ ] **Step 5: Commit**

```bash
git add -u frontend/src/App.css
git commit -m "chore: remove unused App.css leftover from Vite template"
```

- [ ] **Step 6: Re-index for GitNexus**

```bash
npx gitnexus analyze
```

This project's `CLAUDE.md` requires the GitNexus index be refreshed after committing code changes — it will be stale after 12 tasks' worth of commits.
