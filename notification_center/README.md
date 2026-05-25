# Notification Center Widget

An embeddable React widget that gives end-users a self-serve subscription preference page — list management, transactional notification settings, unsubscribe — without leaving the host site.

This is the customer-facing surface of Broadsheet; the **console** (under `/console`) is the operator-facing surface.

## Stack

- React 19 + TypeScript 5.8
- Vite 7 build pipeline
- Radix UI primitives via Shadcn/ui
- Tailwind CSS 4 with `next-themes` for dark/light
- Sonner for toasts, Lucide for icons

## Development

```bash
npm install
npm run dev      # local dev server
npm run build    # production bundle
npm run lint     # type-check + lint
```

The built bundle is embedded by Broadsheet instances at the `/notifications` route per workspace; see the Broadsheet docs for embed instructions and customization options.

## Internationalization

The widget ships 11 locales (`en`, `fr`, `es`, `de`, `zh`, `hi`, `ar`, `pt`, `ru`, `ja`, `pl`) in `src/translations.ts`. This set is independent from the console's locales — picked by browser language at runtime, falls back to English.
