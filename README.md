# HttpRepeater

**English** · [简体中文](README.zh-CN.md)

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Wails](https://img.shields.io/badge/Wails-v2-DF1D1D)](https://wails.io/)
[![Vue](https://img.shields.io/badge/Vue-3-42B883?logo=vuedotjs&logoColor=white)](https://vuejs.org/)
[![Platforms](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey)](#build-from-source)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

A focused, standalone desktop HTTP request replay tool inspired by **Burp Suite Repeater**.

Edit a raw HTTP request, send it exactly as written, inspect the response, and keep a per-tab
history of everything you sent — without launching a full interception proxy.

<p align="center">
  <img src="assets/screenshots/main-dark-en.png" alt="HttpRepeater main window" width="900">
</p>

Built with [Wails v2](https://wails.io/) (Go backend + system WebView), Vue 3, TypeScript,
Naive UI and CodeMirror 6. No JVM, no browser extension.

## Why another repeater?

Burp Suite's Repeater is excellent, but it comes as part of a large interception suite. Sometimes
you just want a small native window where you can paste a request, change a header, hit send, and
look at the response — without starting a proxy, configuring a browser, or loading a JVM.

HttpRepeater is that window. It deliberately does one thing: **the Repeater loop**.

- **Raw request fidelity** — requests are written to a TCP/TLS socket by hand, so header order,
  header case, duplicate headers and unusual bodies survive.
- **Fast input** — paste a raw request, a bare URL, or `curl` / `fetch` code copied straight from
  the browser's DevTools, including the Windows `cmd` flavour (**Copy as cURL (cmd)**).
- **Per-tab proxy chaining** — route a single tab through Burp, Yakit, mitmproxy or Clash with one
  toggle. The default proxy is `http://127.0.0.1:7890`; change it per tab or globally in Settings.
- **Per-tab history** — every send is recorded, including failures; jump back to a snapshot, edit
  and resend.
- **Optional AI help** — a built-in assistant backed by your own LLM endpoint can rewrite the
  request or draft payloads; you review and apply the result, nothing is sent automatically.
- **Small footprint** — one native app, no Java runtime.

## Features

- Raw HTTP/1.x request editor with HTTP syntax highlighting, line numbers and search (⌘F / Ctrl+F).
- Multi-tab workspace. Each tab keeps its own request, target, proxy settings and history.
- HTTP and HTTPS. Responses are decoded from `gzip`, `deflate` and `br`, and shown as
  **Pretty** (formatted JSON, status line and headers) or **Raw**.
- Response search: case-insensitive keyword search with every match highlighted, an `n/m` counter,
  and previous/next navigation (`Enter` / `Shift+Enter` or the arrow buttons). The current match is
  highlighted in a stronger colour and scrolled into view.
- Cancellable sends: while a request is in flight, **Send** turns into **Cancel**; the socket is
  closed and the attempt is recorded in history as cancelled.
- Request history with `‹ n/N ›` navigation and a dropdown list (status code and duration).
  Loading a snapshot restores the exact request text; unsent edits trigger a confirmation.
- Downstream HTTP proxy per tab, including `user:pass` Basic authentication:
  - `http://` targets use the absolute-URL request form;
  - `https://` targets use a `CONNECT` tunnel.
- `curl` / `fetch` import, including Chrome DevTools' **Copy as cURL** / **Copy as fetch** — both the
  bash flavour and the Windows `cmd` flavour (`^` line continuations, `^"` quoting, `curl.exe`).
  Paste the snippet and it is converted automatically; **Import** reads the clipboard.
- Built-in request templates: GET, POST form, POST JSON, multipart upload, empty `Host`, clear.
- Right-click menu in the request editor:
  - **Copy URL** — the full `scheme://host[:port]/path?query` URL, GET parameters included;
  - **Copy as curl** — a ready-to-run command: POSIX single-quote style on macOS/Linux,
    `cmd`-escaped (`^"..."^"`) style on Windows;
  - **Change request method** — Burp-style GET↔POST conversion that moves URL query parameters into
    an `application/x-www-form-urlencoded` body, and back.
- **AI assistant** (bring your own LLM): chat about the current request, ask for edits or payloads,
  then apply the generated raw request to the editor with one click — requests are never sent
  automatically. Built on the official **OpenAI Go SDK v3** with OpenAI-compatible **Chat
  Completions** and **Responses** APIs; base URL, API key, model, reasoning effort and system prompt
  are all configurable.
- **Follow Host** and **Auto CL** toggles:
  - Follow Host rewrites `Host` to match the target (on by default);
  - Auto CL recalculates `Content-Length` when a body is present (on by default).
  Turn both off to test Host validation or length-handling edge cases.
- Settings dialog (gear in the title bar): UI language, theme, default proxy and the LLM
  configuration, persisted to `~/.httpRepeater/setting.json` (written with `0600` permissions).
- Dark / light theme and English / Chinese UI, switchable from the title bar or Settings.
- Keyboard-first: send, new tab, close tab, editor search.

## Compared with Burp Suite Repeater

| Capability | Burp Suite Repeater | HttpRepeater |
| --- | --- | --- |
| Edit and resend raw HTTP requests | Yes | Yes |
| Per-tab request history and snapshots | Yes | Yes |
| Multiple tabs | Yes | Yes |
| Host header follow / Content-Length recalculation | Yes | Yes, as explicit toggles |
| gzip / deflate / br response decoding | Yes | Yes |
| HTTP/HTTPS with a proxy chain | Yes, via Burp's proxy | Yes, per-tab downstream proxy |
| curl / fetch import | Not built in | Built in |
| Change request method (GET ↔ POST) | Yes | Yes |
| AI-assisted request editing | No | Yes, bring your own LLM |
| Intercepting proxy, site map, scanner, Intruder | Part of Burp Suite | Not included |
| Extensions, macros, auth/session handling | Yes | Not included |
| Session persistence | Yes | In-memory only |
| Runtime | Java | Native Go + system WebView |

HttpRepeater is **not** a Burp Suite replacement. It intentionally covers only the Repeater loop;
use Burp, ZAP, mitmproxy or other tools for interception, crawling, scanning and automation.

## Use cases

- **Replay a captured request.** Copy a request from DevTools, Burp or a log, paste it, tweak a
  header, cookie or body field, and send it again. The previous attempt stays in history.
- **Turn curl into an editable request.** Paste `curl 'https://api.example.com/v1/login' ...` or
  browser `fetch(...)` code. HttpRepeater fills in the target and gives you the raw request.
- **Draft payloads with AI.** Ask the assistant to rewrite the request or produce a body, review the
  raw result, apply it to the editor and send it yourself.
- **Turn an advisory into a verification request.** Paste the CVE/advisory context and ask for a raw
  HTTP request against your local reproduction, for example: *"Write a verification request for
  CVE-2024-XXXX against http://127.0.0.1:8080 and return the complete HTTP/1.1 request."* The
  assistant replies with a ready-to-send request; click **Apply to request**, review it, and send it
  yourself.
- **Craft injection payloads for a parameter.** Describe the endpoint and the parameter you want to
  test: *"The endpoint is `GET /api/user?id=1`; help me test the `id` parameter for SQL injection in
  my authorized lab and return several candidate payloads with the full requests."* Each reply can
  be applied to the editor and replayed; nothing is sent automatically. The same workflow works for
  a local Nacos-style auth-bypass check: *"Write a request to verify the Nacos unauthorized user
  query issue against http://127.0.0.1:8848."*
- **Test through a proxy chain.** Enable the per-tab proxy and point it at Burp, Yakit, mitmproxy or
  Clash. HTTP requests use absolute-URL form; HTTPS requests use a CONNECT tunnel.
- **Probe header handling.** Disable **Follow Host** to send the `Host` header exactly as written,
  or **Auto CL** to send your own `Content-Length` — useful for Host validation and length-handling
  experiments on systems you are authorized to test.
- **Debug several endpoints at once.** Keep one tab per endpoint, each with its own history and
  proxy setting.
- **Compare responses over time.** Change one parameter, resend, and page through the history to
  compare status codes, headers and bodies.

### AI in practice

<p align="center">
  <img src="assets/screenshots/ai-poc-en.png" alt="AI turning an N-day advisory into a verification request" width="900"><br>
  <em>N-day advisory → a ready-to-send verification request</em>
</p>

<p align="center">
  <img src="assets/screenshots/ai-injection-en.png" alt="AI drafting SQL injection payloads for an id parameter" width="900"><br>
  <em>Injection testing → candidate payloads for the id parameter</em>
</p>

> Only use this tool against systems you own or have explicit permission to test.

## Build from source

### Prerequisites

| Dependency | Version |
| --- | --- |
| Go | 1.26+ |
| Node.js | 20.19+ (or 22.12+) |
| Wails CLI | v2.17+ |
| Platform toolchain | macOS: Xcode Command Line Tools · Windows: WebView2 · Linux: GTK3 + WebKit2GTK |

See the [Wails installation guide](https://wails.io/docs/gettingstarted/installation) for your
platform's packages.

```bash
# Clone the repository
git clone https://github.com/chenshiyi007/HttpRepeater.git
cd HttpRepeater

# Install the Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Build
wails dev      # hot-reload development
wails build    # release build, output in build/bin/
```

For a build whose header shows the git-derived version, use `make build` instead; see
[Versioning and releases](#versioning-and-releases).

Frontend-only development (no Wails backend calls, useful for UI work):

```bash
cd frontend
npm install
npm run dev
```

## Usage

### 1. Set a target

Enter the scheme, host and port in the target bar, or paste a request and let HttpRepeater parse it.
Click **Parse** to re-read the target from the request line / `Host` header. A bare URL is converted
into a `GET` request skeleton.

### 2. Edit the request

The editor works on raw HTTP: request line, headers, blank line, body. Use **Templates** for common
shapes, or **Import** to pull a `curl` / `fetch` snippet from the clipboard. Press **Send** (or
⌘Enter / Ctrl+Enter) when ready. While a request is in flight, the Send button becomes **Cancel**;
clicking it closes the socket immediately and the cancelled attempt still appears in history.

<p align="center">
  <img src="assets/screenshots/templates-en.png" alt="Request templates and raw editor" width="900">
</p>

### 3. Inspect the response

The response pane shows the status line, raw headers and the decoded body. Switch between **Pretty**
and **Raw** in the header; JSON bodies are indented in Pretty mode. The status area shows the status
line, elapsed time and transferred bytes.

Use the **Search response** box to find a keyword case-insensitively. All matches are highlighted,
the counter shows the current position (`n/m`), and the up/down buttons — or `Enter` / `Shift+Enter`
in the search box — jump between matches, scrolling each one into view.

<p align="center">
  <img src="assets/screenshots/response-search-en.png" alt="Response search with highlighted matches" width="900">
</p>

### 4. Move through history

Every send creates a history entry — failures included. Use `‹ n/N ›` to step through entries or
click the number to open a list with status code and duration. Loading a snapshot puts the exact
request text back into the editor; if you have unsent edits, HttpRepeater asks before overwriting them.

<p align="center">
  <img src="assets/screenshots/history-en.png" alt="Per-tab request history" width="900">
</p>

### 5. Work with tabs

`+` creates a tab; double-click a tab to rename it. Each tab is an independent workspace: request,
target, proxy state, Host/CL toggles and history. At least one tab always remains open.

### 6. Send through a downstream proxy

Click the proxy button to toggle it for the current tab, then open the gear to edit the URL. The
default is `http://127.0.0.1:7890`. Credentials can be embedded as `user:pass`:

```
http://user:pass@127.0.0.1:7890
```

HTTPS targets automatically use a `CONNECT` tunnel. The setting applies to the next send and is
remembered per tab while the app is running.

<p align="center">
  <img src="assets/screenshots/proxy-en.png" alt="Per-tab downstream proxy settings" width="900">
</p>

### 7. Import curl / fetch

The fastest path from a browser request: open DevTools → **Network**, right-click the request, and
choose **Copy as cURL** or **Copy as fetch**. Paste the copied text into the request editor —
HttpRepeater detects the format, converts it into an editable raw request, and fills in the target.
On Windows, **Copy as cURL (cmd)** works too: `^` line continuations, `^"` argument quoting and
`curl.exe` are handled transparently.

<p align="center">
  <img src="assets/screenshots/web-tools.png" alt="Copy as cURL / Copy as fetch in Chrome DevTools" width="620">
</p>

You can also paste a hand-written curl command or `fetch()` call; the **Import** button reads the
clipboard for you.

Supported curl options include `-X/--request`, `-H/--header`, `-d/--data*`, `--json`,
`-u/--user`, `-b/--cookie`, `-A/--user-agent`, `-e/--referer`, `-I/--head` and `--compressed`.
`-F`/`-T` (multipart / file upload) are not supported; use `-d` for the body or edit the request
manually. For `fetch()`, the URL, method, literal headers and body are imported; headers the browser
adds automatically (cookies, authorization, etc.) are not present and are reported as warnings.

### 8. Right-click tools

Right-click anywhere in the request editor:

- **Copy URL** — the full URL with scheme, host, port, path and query string (GET parameters
  included; POST bodies are left alone);
- **Copy as curl** — a curl command that reproduces the request. On Windows the arguments are
  quoted for `cmd` (`^"..."^"`, `^` continuations); on macOS/Linux they use POSIX single quotes;
- **Change to POST / Change to GET** — Burp-style method conversion. GET query parameters move into
  an `application/x-www-form-urlencoded` body (with `Content-Type` and `Content-Length` added); a
  form body moves back into the URL query. Non-form bodies (e.g. JSON) are kept as-is.

<p align="center">
  <img src="assets/screenshots/context-menu-en.png" alt="Request editor right-click menu" width="900">
</p>

### 9. AI assistant

Click **AI** next to **Import** to open the assistant. The request currently in the editor is sent
with every message, so you can ask for edits ("move the query parameters into a JSON body"), new
payloads, or explanations. When a reply contains a complete raw request, **Apply to request**
replaces the editor content with it — HttpRepeater never sends anything on the AI's behalf.

<p align="center">
  <img src="assets/screenshots/ai-en.png" alt="AI assistant editing the current request" width="900">
</p>

The assistant talks to the LLM endpoint configured in Settings (OpenAI-compatible **Chat
Completions** or **Responses** API). Conversation history lives in memory for the current app
session only. Send with **Ctrl/Cmd+Enter** — plain Enter inserts a new line, so IME candidate
selection never submits by accident. While a request is running, the send button turns into
**Stop**, which cancels the in-flight HTTP call. Drag the bottom-right corner of the dialog to
resize it.

### 10. Settings

The gear button in the title bar opens the settings dialog:

- UI language and theme;
- the default proxy URL used by new tabs and by "Reset to default";
- the LLM provider: base URL, API key, model, API type (Chat Completions / Responses), reasoning
  effort, and the system prompt that teaches the model how to output raw HTTP.

Everything is stored in `~/.httpRepeater/setting.json`, written with `0600` permissions. The file
never leaves your machine, and the API key is only used for calls to the endpoint you configured.
Theme and language are additionally mirrored in `localStorage` so the first frame always matches
your last choice.

<p align="center">
  <img src="assets/screenshots/settings-en.png" alt="Settings dialog" width="900">
</p>

## Keyboard shortcuts

| Shortcut | Action |
| --- | --- |
| `⌘Enter` / `Ctrl+Enter` | Send the current request |
| `⌘T` / `Ctrl+T` | New tab |
| `⌘W` / `Ctrl+W` | Close the current tab |
| `⌘F` / `Ctrl+F` | Search in the request editor |

## Design notes

- **Raw sending.** Requests are assembled and written directly to a `net.Conn` / `tls.Conn`; the
  standard `net/http` request serializer is not used, because it normalizes header names, drops
  duplicate headers and adds defaults. `net/http` is only used for chunked-response parsing.
- **Response decoding.** Bodies are read according to `Content-Length`, `Transfer-Encoding: chunked`
  or connection-close, then decoded in reverse order for `gzip`, `deflate` and `br`.
- **Limits.** Request 10 MB, response headers 1 MB, response body 64 MB.
- **LLM calls.** The Go backend uses the official OpenAI Go SDK v3
  (`github.com/openai/openai-go/v3`) with the configured base URL and API key, and selects Chat
  Completions or Responses per request. The current request is included as context with each
  message, and a reply is only written into the editor when you click **Apply to request**.
- **State.** Tabs, history, proxy settings and request text live in memory. Language, theme, the
  default proxy and the LLM configuration are persisted to `~/.httpRepeater/setting.json`; theme
  and language are also mirrored in `localStorage` for a flash-free start.

## Scope and limitations

HttpRepeater intentionally stops at the Repeater loop. It does **not** include:

- an intercepting proxy, site map, crawler, scanner or Intruder;
- persistent projects or saved request collections;
- scripting hooks, Nuclei templates or report export;
- LLM chat history persistence (conversations live in memory only);
- automatic redirect following, a cookie jar, or auth/session handling;
- HTTP/2 or a hex view.

If you need those, use Burp Suite, ZAP, mitmproxy, Yakit or another tool alongside HttpRepeater.

## Security notes

- TLS certificate verification is **disabled** so that self-signed test targets work. Do not use
  HttpRepeater for sensitive production traffic without understanding this.
- Proxy credentials remain in memory and are sent only to the configured proxy.
- The LLM API key is stored locally in `~/.httpRepeater/setting.json` (mode `0600`). Using the AI
  assistant sends the current request and your messages to the LLM endpoint you configured.
- Requests are sent as edited, apart from the optional `Host` / `Content-Length` rewrites described
  above; you are responsible for what you send.

**Use this tool only against systems you own or have explicit permission to test.**

## Development

```bash
# Go tests
go test ./...

# Frontend type-check and build
cd frontend && npm run build

# Package the desktop app
wails build
```

### Versioning and releases

The version badge in the header is injected at build time. `make build` and `make dev` derive it
from `git describe --tags --dirty`; when no tag is reachable it falls back to `dev+<commit>`, and
when there is no git metadata at all the badge is hidden. `0.0.0` is never shown.

To publish a GitHub release (requires a clean working tree, an `origin` remote and an authenticated
[`gh`](https://cli.github.com/) CLI):

```bash
gh auth login
make notes VERSION=v0.2.0        # generate build/release/notes.md
$EDITOR build/release/notes.md   # optional: edit the release notes
make release VERSION=v0.2.0      # build, package, tag and publish
```

`make release` builds and packages the current platform, creates and pushes the `v0.2.0` tag, and
publishes the archive with the notes. For additional platforms, build and package them first, then
upload the resulting archives to the same release:

```bash
make build-release VERSION=v0.2.0 PLATFORM=windows/amd64
make package-assets VERSION=v0.2.0 PLATFORM=windows/amd64
make publish VERSION=v0.2.0
```

Use `PLATFORM=darwin/universal` for a universal macOS build. Linux packages must be built on Linux
(or in a Linux container). Run `make help` for all targets.

### Project layout

| Path | Contents |
| --- | --- |
| `main.go`, `app.go` | Wails application entry point |
| `internal/repeater/` | Request parsing, raw socket sending, proxy, response decoding, curl/URL/method tools, settings & LLM, i18n |
| `frontend/src/` | Vue 3 UI, components, Pinia stores, editor and i18n messages |
| `build/` | App icons and platform packaging assets |
| `assets/screenshots/` | Images used in the README |
| `Makefile` | Build, packaging and GitHub release targets |
| `scripts/` | Release notes and `wails.json` version helpers |

## License

Licensed under the [Apache License, Version 2.0](LICENSE).

## Acknowledgements

- [Wails](https://wails.io/) — desktop shell
- [OpenAI Go SDK v3](https://github.com/openai/openai-go) — LLM client
- [Naive UI](https://www.naiveui.com/) — Vue 3 component library
- [CodeMirror](https://codemirror.net/) — request editor
- [Ionicons](https://ionic.io/ionicons) — interface icons (via `@vicons/ionicons5`)
- Burp Suite Repeater — the workflow this tool is modelled on
