# HttpRepeater frontend

Vue 3 + TypeScript + Vite UI for [HttpRepeater](../README.md), using Naive UI and CodeMirror 6.

## Development

```bash
npm install
npm run dev      # Vite dev server
npm run build    # type-check + production build
```

The normal desktop workflow is to run `wails dev` from the repository root; that starts the Go
backend and the Vite dev server together.

Running the frontend alone works for layout and styling work, but Wails runtime calls (for example
clipboard access and the Go request bindings) are unavailable in a plain browser.

## Layout

| Path | Contents |
| --- | --- |
| `src/components/` | Tab bar, target bar, request/response panels, editor, history navigation |
| `src/stores/` | Pinia stores for tabs, history and settings |
| `src/i18n/` | Chinese/English message dictionary |
| `src/api/` | Thin wrappers around the generated Wails bindings |
| `wailsjs/` | Generated Wails bindings (do not edit by hand) |
