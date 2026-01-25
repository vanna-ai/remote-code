# Vanna Design System & Style Guide

## Brand Color Palette

| Color | Hex | Usage |
|-------|-----|-------|
| **Navy** | `#023d60` | Primary dark color - headings, text, dark accents |
| **Cream** | `#e7e1cf` | Light warm backgrounds, visual contrast with navy |
| **Teal** | `#15a8a8` | Primary accent - CTAs, highlights, interactive elements |
| **Orange** | `#fe5d26` | Secondary accent - alerts, progress indicators |
| **Magenta** | `#bf1363` | Tertiary accent - decorative elements, visual hierarchy |

Extended colors use Tailwind's slate scale for neutral UI elements.

---

## Typography

| Type | Font | Usage |
|------|------|-------|
| **Serif** | Roboto Slab | Headlines, titles, branded text (`font-serif`) |
| **Sans** | Space Grotesk | Body text, UI labels, navigation (`font-sans`) |
| **Mono** | Disket Mono / Space Mono | Code snippets, technical content (`font-mono`) |

**Sizing patterns:**
- H1: `text-5xl sm:text-6xl lg:text-7xl font-bold`
- H2/Subheading: `text-2xl sm:text-3xl font-serif`
- Body: `text-base` to `text-lg` in `slate-600`
- Captions: `text-xs` to `text-sm` in `slate-500`

---

## Shadows & Elevation

Shadows use **teal tinted glows** for brand consistency:

```css
/* Subtle */
shadow-[0_4px_15px_rgba(21,168,168,0.2)]

/* Medium card */
shadow-[0_15px_40px_rgba(15,23,42,0.08)]

/* Feature cards */
shadow-[0_25px_55px_rgba(21,168,168,0.18)]

/* Large showcase */
shadow-[0_30px_80px_rgba(59,130,246,0.15)]
```

**Hover:** Cards lift with `hover:-translate-y-1` + enhanced shadow.

---

## Border Radius

| Size | Class | Usage |
|------|-------|-------|
| Full | `rounded-full` | Buttons, badges (pill shape) |
| 3xl | `rounded-3xl` | Large cards, pricing sections |
| 2xl | `rounded-2xl` | Feature cards, containers (most common) |
| xl | `rounded-xl` | Medium components, inputs |
| lg/md | `rounded-lg/md` | Small UI elements |

---

## Gradients

**Text gradients:**
```css
bg-gradient-to-r from-vanna-navy via-vanna-teal to-vanna-navy
bg-clip-text text-transparent
```

**Background gradients:**
```css
bg-gradient-to-br from-white/90 via-vanna-cream/80 to-vanna-teal/15
bg-gradient-to-b from-vanna-cream via-white to-vanna-cream
```

**Decorative radials:**
```css
radial-gradient(circle at top_left, rgba(21,168,168,0.12), transparent 60%)
```

---

## Buttons (VannaButton)

| Variant | Style |
|---------|-------|
| **Primary** | `bg-vanna-teal text-white` + teal glow shadow |
| **Secondary** | `bg-vanna-cream text-vanna-navy border-vanna-teal/30` |
| **Outline** | `border-vanna-navy/30 bg-transparent text-vanna-navy` |
| **Ghost** | `text-vanna-navy hover:bg-vanna-cream/50` |

All buttons: `rounded-full` with `transition-all`.

**Sizes:** sm (`h-9`), md (`h-11`), lg (`h-12`)

---

## Cards

**Base card:**
```css
rounded-2xl border border-slate-200/60 bg-white/80 backdrop-blur-sm
shadow-[0_15px_40px_rgba(15,23,42,0.08)]
hover:-translate-y-1 hover:shadow-[...]
```

**Color theme variants** (applied via gradient overlays):
- Teal: `to-vanna-teal/15`
- Orange: `to-vanna-orange/15`
- Magenta: `to-vanna-magenta/15`

---

## Glass-morphism Effects

```css
backdrop-blur-sm
bg-white/60
bg-vanna-cream/40
```

Used on overlays, navigation, and floating cards.

---

## Navigation

```css
fixed top-0 z-50 w-full
bg-vanna-cream/90 backdrop-blur
border-b border-vanna-teal/30
h-16

/* Links */
text-vanna-navy/70 hover:text-vanna-teal
```

---

## Decorative Elements

**Floating accent cards:**
- Positioned absolutely with rotations (`rotate-12`, `-rotate-6`)
- `rounded-3xl border-vanna-teal/40 bg-vanna-cream/60 backdrop-blur`

**Background blobs:**
- Large colored blurs: `blur-[180px]` to `blur-[220px]`
- Positioned off-screen for ambient effect

**Dot patterns:**
```css
background-image: radial-gradient(circle at 2px 2px, rgba(2,61,96,0.3) 1px, transparent 0)
background-size: 32px 32px
```

**Accent lines:**
```css
bg-gradient-to-r from-transparent via-vanna-teal to-transparent
```

---

## Status Indicators

| State | Style |
|-------|-------|
| Completed/Active | `text-vanna-teal bg-vanna-teal/10` |
| Error | `text-vanna-orange bg-vanna-orange/10` |
| Pending | `text-slate-400 bg-slate-100` |

Animations: `animate-pulse`, `animate-spin`, `animate-ping`

---

## Responsive Breakpoints

Mobile-first with standard Tailwind breakpoints:
- `sm:` 640px
- `md:` 768px
- `lg:` 1024px
- `xl:` 1280px

Grids: 1 col → 2 col → 3 col

---

## Key Files

| File | Purpose |
|------|---------|
| `src/app.css` | Global CSS with `@theme` color definitions |
| `tailwind.config.js` | Extended colors, fonts, plugins |
| `src/lib/components/` | Reusable styled components |
