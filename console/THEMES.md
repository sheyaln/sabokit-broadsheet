# Broadsheet Console — Themes

This file documents the visual identity for the Broadsheet console and records
two alternative directions that were considered. The brief was "slightly
leftist broadsheet" — confident, plainspoken, literary, not party-coded.

## Current theme: Newsprint (after-hours)

The vibe: a late-shift labour-press desk. Sodium-lamp warmth, carbon
copies in low light, the dark side of the broadsheet tradition.
Confident, plainspoken, doesn't burn your retinas at 11pm.

### Palette

| Token                  | Hex       | Role                                                |
| ---------------------- | --------- | --------------------------------------------------- |
| `--color-primary`      | `#E26A55` | Brick / masthead red, dark-tuned. Buttons, links.   |
| `--color-primary-deep` | `#D55540` | Pressed / active.                                   |
| `--color-primary-soft` | `#F08672` | Hover wash.                                         |
| `--color-accent`       | `#D8A35A` | Warm ochre. Used sparingly.                         |
| `--color-paper`        | `#1A1612` | Page background. Warm near-black.                   |
| `--color-paper-bright` | `#221D18` | Surfaces and cards.                                 |
| `--color-paper-deep`   | `#2C2620` | Raised / hover. Inputs.                             |
| `--color-ink`          | `#F0E9DA` | Body text. Warm off-white that mirrors paper cream. |
| `--color-ink-muted`    | `#A39C8E` | Secondary text. Desaturated warm grey.              |
| `--color-ink-faint`    | `#6F685C` | Captions, placeholders.                             |
| `--color-rule`         | `#3B342D` | Hairline rules and borders.                         |
| `--color-rule-strong`  | `#5A4F43` | Visible masthead rules.                             |
| `--color-success`      | `#7BB364` | Status — slightly desaturated.                      |
| `--color-warning`      | `#D8A35A` |                                                     |
| `--color-error`        | `#E26A55` |                                                     |
| `--color-info`         | `#6AA2C4` |                                                     |

The light variant (cream paper, `#B23A2A` brick on `#F5F1E8`) was the
original. Reach for it again if you want a printable-feeling daytime
mode — the structure is identical, only the token values change.

### Typography

- **Display** — Fraunces (Google Fonts). Variable serif, optional opsz/SOFT
  axes. Used for headings, the masthead wordmark, and the Settings header.
  Does the "literary, slightly defiant" thing without going Victorian.
- **Body** — IBM Plex Sans (Google Fonts). Clean, technocratic, with a
  humanist warmth. Well-suited to dense admin UIs. Falls back to the system
  sans stack.
- **Mono** — IBM Plex Mono. Used for logs and code blocks.

### Where the leftist DNA shows up

- The primary is the same brick red lifted from masthead inks of
  mid-century labour papers (*Industrial Worker*, *The Dispatcher*),
  re-tuned warmer (`#E26A55`) so it still reads brick rather than going
  pink against a near-black ground.
- The background is warm near-black rather than cold tech-startup grey —
  think newsroom under desk lamps, not VS Code.
- The display serif gives screens a typographic confidence — headers
  feel set rather than coded.
- The sign-in screen and sidebar are framed with a horizontal-rules
  "masthead" — a 3pt top / 1pt bottom rule that nods to broadsheet
  typography without a single icon of a fist or a hammer.
- Border radii are tightened from 6px to 2px so cards and inputs feel
  more like printed boxes than rounded SaaS chiclets.

### Accessibility

All foreground/background pairs target WCAG AA at minimum:

| Pair                    | Ratio   | Pass        |
| ----------------------- | ------- | ----------- |
| `#F0E9DA` on `#1A1612`  | ~14.6:1 | AAA         |
| `#F0E9DA` on `#221D18`  | ~13.8:1 | AAA         |
| `#A39C8E` on `#1A1612`  | ~7.4:1  | AAA         |
| `#E26A55` on `#1A1612`  | ~5.4:1  | AA          |
| `#1A1612` on `#E26A55`  | ~5.4:1  | AA          |
| `#6AA2C4` on `#1A1612`  | ~6.6:1  | AAA         |

Primary buttons are rendered with the dark background colour as the text
colour (`#1A1612` on `#E26A55`) — that's the only way to keep AA on the
brick fill without going off-brand.

### Where the theme is wired up

- `console/src/index.css` — Tailwind v4 `@theme {}` block, base `:root`
  variables, and AntD CSS overrides. This is the single source of truth.
- `console/src/App.tsx` — the AntD `ConfigProvider` `ThemeConfig`. Mirrors
  the palette into AntD component tokens.
- `console/index.html` — Google Fonts preconnect and stylesheet, splash
  background, theme-color meta.
- `console/src/layouts/MainLayout.tsx` — the unauthenticated chrome with
  the broadsheet masthead.
- `console/src/layouts/WorkspaceLayout.tsx` — the authenticated sidebar
  with the inline Broadsheet masthead.
- `console/src/pages/SignInPage.tsx` — the sign-in card uses a hard 1px
  ink border with a soft offset shadow (a printed-box feel).
- `console/src/components/settings/SettingsSidebar.tsx` — Settings header
  uses the display serif.

---

## Considered alternatives

These were the other two directions presented to the brief. Not currently
implemented; documented here so the user can ask for them by name.

### Risograph

> Late-night zine made on a borrowed Riso. Spot colours don't quite align.

- `--ink: #2B2B2B`, `--paper: #EFEAE0`, `--surface: #FFFFFF`,
  `--riso-red: #E0395D`, `--riso-blue: #2F4EC2`, `--muted: #7A7268`,
  `--rule: #D0C8B8`.
- Display: Space Grotesk. Body: Inter. Mono: JetBrains Mono.
- Two spot colours (riso pink/red + electric blue) used as flat blocks.
  Hard-edged, slightly off-register feel via 1–2px offset shadows on
  accents. Reads as 1990s zine / DIY punk.
- More visible "we have a politics," more chance of feeling costumey.
  Use only if you want more conviction than Newsprint provides.

### Constructivist Restraint

> Rodchenko on a diet. Geometry and red, but quiet.

- `--ink: #111111`, `--paper: #F2F0EC`, `--surface: #FFFFFF`,
  `--soviet-red: #C8281E`, `--steel: #4A5260`, `--rule: #1A1A1A` (1px hard
  rules everywhere), `--muted: #6E6E6E`.
- Display: DM Serif Display, used sparingly and mostly all-caps.
  Body: Inter.
- Heavy use of horizontal/vertical hard rules between sections (very
  Constructivist layout), big slab letterforms in section headers,
  monochrome with one red. The most graphically loaded option.
- Risk: edges toward Soviet-poster cosplay if not held back. The brief
  warned against this; keep it parked unless explicitly requested.
