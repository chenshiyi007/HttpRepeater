# HttpRepeater

[English](README.md) · **简体中文**

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Wails](https://img.shields.io/badge/Wails-v2-DF1D1D)](https://wails.io/)
[![Vue](https://img.shields.io/badge/Vue-3-42B883?logo=vuedotjs&logoColor=white)](https://vuejs.org/)
[![Platforms](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey)](#从源码构建)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

一个独立、专注的桌面端 HTTP 请求重放工具，参考 **Burp Suite Repeater** 的使用方式。

编辑原始 HTTP 报文，按原样发送，查看响应，并在每个标签页里保留发送历史——不需要启动一整套抓包代理。

<p align="center">
  <img src="assets/screenshots/main-dark-zh.png" alt="HttpRepeater 主界面" width="900">
</p>

技术栈：[Wails v2](https://wails.io/)（Go 后端 + 系统 WebView）、Vue 3、TypeScript、Naive UI、
CodeMirror 6。不依赖 JVM，也不需要浏览器插件。

## 为什么再做一个 Repeater？

Burp Suite 的 Repeater 很好用，但它属于一套大型抓包工具。很多时候你只是想要一个小窗口：
粘贴一段请求，改一个头，点发送，看响应——不用先启动代理、配置浏览器，也不用等 JVM 起来。

HttpRepeater 就是这个小窗口。它只专注一件事：**Repeater 循环**。

- **原始报文保真**：请求直接写入 TCP/TLS socket，保留 Header 顺序、大小写、重复头和非规范 body。
- **快速输入**：支持粘贴原始报文、裸 URL，或直接从浏览器 DevTools 复制的 `curl` / `fetch` 代码。
- **按标签页走代理链**：一个开关就能让当前标签走 Burp、Yakit、mitmproxy 或 Clash。默认代理地址是
  `http://127.0.0.1:7890`，每个标签可以单独修改。
- **按标签页记录历史**：每次发送都会记录（失败也记录），可以回到任意快照改完重发。
- **体积小**：单个原生桌面应用，不需要 Java 运行时。

## 功能特性

- 原始 HTTP/1.x 请求编辑器，带 HTTP 语法高亮、行号和搜索（⌘F / Ctrl+F）。
- 多标签工作区。每个标签独立保存请求、目标、代理设置和历史。
- 支持 HTTP / HTTPS；响应自动解码 `gzip`、`deflate`、`br`，提供 **Pretty**（格式化 JSON、状态行和头）
  和 **Raw** 两种视图。
- 请求历史：`‹ n/N ›` 前后切换，点数字可展开列表（状态码 + 耗时）。加载快照会把当时的请求原文放回
  编辑器；如果有未发送的修改，会先弹确认。
- 每个标签独立的下游 HTTP 代理，支持 `user:pass` Basic 认证：
  - `http://` 目标使用绝对 URL 请求行；
  - `https://` 目标自动走 `CONNECT` 隧道。
- `curl` / `fetch` 导入，支持 Chrome DevTools 的 **Copy as cURL** / **Copy as fetch**。
  粘贴后自动转换；也可以点 **导入** 读取剪贴板。
- 内置请求模板：GET、POST 表单、POST JSON、multipart 文件上传、Host 留空、清空。
- **Host 跟随** 与 **自动 CL** 开关：
  - Host 跟随会把 `Host` 头改写为目标地址（默认开）；
  - 自动 CL 会在有 body 时重算 `Content-Length`（默认开）。
  两个都关掉，就能测试 Host 校验、长度处理之类的边界场景。
- 深色 / 浅色主题，中文 / 英文界面，选择存 `localStorage`。
- 键盘优先：发送、新建标签、关闭标签、编辑器内搜索。

## 与 Burp Suite Repeater 对比

| 能力 | Burp Suite Repeater | HttpRepeater |
| --- | --- | --- |
| 编辑并重发原始 HTTP 请求 | 支持 | 支持 |
| 每个标签独立的历史与快照 | 支持 | 支持 |
| 多标签 | 支持 | 支持 |
| Host 头跟随 / Content-Length 重算 | 支持 | 支持，独立开关 |
| gzip / deflate / br 响应解码 | 支持 | 支持 |
| HTTP/HTTPS 代理链 | 支持（走 Burp 代理） | 支持，按标签配置下游代理 |
| curl / fetch 导入 | 非内置 | 内置 |
| 抓包代理、站点地图、扫描器、Intruder | Burp Suite 的一部分 | 不包含 |
| 插件、宏、认证/会话处理 | 支持 | 不包含 |
| 会话持久化 | 支持 | 仅内存 |
| 运行时 | Java | 原生 Go + 系统 WebView |

HttpRepeater **不是** Burp Suite 的替代品。它只覆盖 Repeater 这一块；抓包、爬虫、扫描和自动化
请继续使用 Burp、ZAP、mitmproxy 等工具。

## 使用场景

- **重放抓到的请求。** 从 DevTools、Burp 或日志里复制请求，粘贴进来，改 Header、Cookie 或 body
  字段，再次发送。上一次的结果留在历史里。
- **把 curl 变成可编辑报文。** 粘贴 `curl 'https://api.example.com/v1/login' ...` 或浏览器里的
  `fetch(...)` 代码，HttpRepeater 会自动填好目标并生成原始请求。
- **通过代理链测试。** 打开当前标签的代理，指向 Burp、Yakit、mitmproxy 或 Clash。HTTP 请求走
  绝对 URL，HTTPS 请求走 CONNECT 隧道。
- **测试 Header 处理。** 关闭 **Host 跟随** 可以按原文发送 `Host`；关闭 **自动 CL** 可以自己控制
  `Content-Length`——适合在获得授权的系统上验证 Host 校验或长度处理行为。
- **同时调试多个接口。** 一个接口一个标签，各自保留历史和代理设置。
- **对比多次响应。** 改一个参数重发，然后在历史里前后翻看状态码、响应头和 body 的变化。

> 请只在你拥有或已获得明确授权的系统上使用本工具。

## 从源码构建

### 环境要求

| 依赖 | 版本 |
| --- | --- |
| Go | 1.25+ |
| Node.js | 20.19+（或 22.12+） |
| Wails CLI | v2.17+ |
| 平台工具链 | macOS：Xcode Command Line Tools · Windows：WebView2 · Linux：GTK3 + WebKit2GTK |

各平台的具体依赖安装方式见 [Wails 安装文档](https://wails.io/docs/gettingstarted/installation)。

```bash
# 克隆仓库
git clone https://github.com/chenshiyi007/HttpRepeater.git
cd HttpRepeater

# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 构建
wails dev      # 热重载开发
wails build    # 发布构建，产物在 build/bin/
```

需要顶栏显示 git 版本号时用 `make build`，见 [版本与发布](#版本与发布)。

只调试前端（不调用 Wails 后端时可用）：

```bash
cd frontend
npm install
npm run dev
```

## 使用说明

### 1. 设置目标

在目标栏选择协议、填写 Host 和端口；也可以直接粘贴一段请求让 HttpRepeater 自动解析。点 **解析**
会从请求行 / `Host` 头重新读取目标。只粘贴一个裸 URL 时，会自动生成 `GET` 请求骨架。

### 2. 编辑请求

编辑器直接编辑原始 HTTP：请求行、Header、空行、body。常用结构可以从 **模板** 里选，或者用
**导入** 从剪贴板读取 `curl` / `fetch` 片段。准备好后点 **发送**（或 ⌘Enter / Ctrl+Enter）。

<p align="center">
  <img src="assets/screenshots/templates-zh.png" alt="请求模板与原始报文编辑器" width="900">
</p>

### 3. 查看响应

响应面板显示状态行、原始响应头和解码后的 body。右上角可以切换 **Pretty** / **Raw**；JSON 在
Pretty 模式下会缩进。状态区显示状态行、耗时和传输字节数。

### 4. 浏览历史

每次发送都会生成一条历史（失败也记录）。用 `‹ n/N ›` 前后切换，或点数字展开列表查看状态码和耗时。
加载快照会把当时的请求原文放回编辑器；如果当前有未发送的修改，会先确认是否覆盖。

<p align="center">
  <img src="assets/screenshots/history-zh.png" alt="按标签页保存的请求历史" width="900">
</p>

### 5. 使用多标签

点 `+` 新建标签，双击标签可以重命名。每个标签都是独立工作区：请求、目标、代理状态、Host/CL 开关和
历史互不影响。至少会保留一个标签。

### 6. 通过下游代理发送

点代理按钮切换当前标签的代理开关，点旁边的齿轮修改地址。默认是 `http://127.0.0.1:7890`，也可以把
认证信息写进地址：

```
http://user:pass@127.0.0.1:7890
```

HTTPS 目标会自动使用 `CONNECT` 隧道。设置在下一次发送时生效，并在本次运行期间按标签记住。

<p align="center">
  <img src="assets/screenshots/proxy-zh.png" alt="按标签配置下游代理" width="900">
</p>

### 7. 导入 curl / fetch

最顺手的方式是从浏览器里直接复制：打开 DevTools → **Network**，右键请求，选 **Copy as cURL**
或 **Copy as fetch**，回到 HttpRepeater 直接粘贴。工具会自动识别格式、转换成可编辑的原始报文，
并把目标栏填好。

<p align="center">
  <img src="assets/screenshots/web-tools.png" alt="在 Chrome DevTools 中复制为 cURL / fetch" width="620">
</p>

也可以粘贴手写的 curl 命令或 `fetch()` 调用；**导入** 按钮会从剪贴板读取。

curl 支持 `-X/--request`、`-H/--header`、`-d/--data*`、`--json`、`-u/--user`、`-b/--cookie`、
`-A/--user-agent`、`-e/--referer`、`-I/--head`、`--compressed` 等常用参数。`-F` / `-T`
（multipart / 文件上传）暂不支持，可以用 `-d` 提供 body 或手工构造。`fetch()` 会导入 URL、method、
字面量 headers 和 body；浏览器自动附加的 Cookie / Authorization 等头不会被导入，并会给出提示。

### 8. 主题与语言

右上角的太阳/月亮按钮切换深浅色；`EN` / `中` 按钮切换界面语言。两个选择都会保存在本地，下次启动
自动恢复。

## 快捷键

| 快捷键 | 功能 |
| --- | --- |
| `⌘Enter` / `Ctrl+Enter` | 发送当前请求 |
| `⌘T` / `Ctrl+T` | 新建标签 |
| `⌘W` / `Ctrl+W` | 关闭当前标签 |
| `⌘F` / `Ctrl+F` | 在请求编辑器中搜索 |

## 实现说明

- **原始发送。** 请求由代码拼装后直接写入 `net.Conn` / `tls.Conn`，不使用 `net/http` 的请求序列化，
  避免 Header 大小写规范化、重复头被吞掉和自动补默认头。`net/http` 只用于解析 chunked 响应。
- **响应解码。** 按 `Content-Length`、`Transfer-Encoding: chunked` 或连接关闭读取 body，再按逆序
  解码 `gzip`、`deflate`、`br`。
- **上限。** 请求 10 MB，响应头 1 MB，响应体 64 MB。
- **状态。** 标签、历史、代理设置和请求内容都只在内存中；只有主题和语言偏好会写入
  `localStorage`。

## 范围与限制

HttpRepeater 有意只做到 Repeater 这一层，不包含：

- 抓包代理、站点地图、爬虫、扫描器或 Intruder；
- 持久化项目或请求收藏；
- LLM 辅助、脚本 Hook、Nuclei 模板或报告导出；
- 自动跟随重定向、Cookie Jar、认证/会话处理；
- HTTP/2、响应内搜索或 Hex 视图。

需要这些能力时，请把 HttpRepeater 和 Burp Suite、ZAP、mitmproxy、Yakit 等工具配合使用。

## 安全提示

- 为了便于测试自签名证书的目标，TLS 证书校验当前是**关闭**的。请勿在未理解这一点的情况下用于敏感的生产流量。
- 代理认证信息只保存在内存中，并且只会发给配置的代理。
- 请求会按你编辑的内容发送（启用的 Host 跟随 / 自动 CL 会按上文说明改写对应头），请对自己发出的内容负责。

**请只在你拥有或已获得明确授权的系统上使用本工具。**

## 开发

```bash
# Go 测试
go test ./...

# 前端类型检查与构建
cd frontend && npm run build

# 打包桌面应用
wails build
```

### 版本与发布

顶栏的版本号在构建时注入。`make build` / `make dev` 会用 `git describe --tags --dirty` 取值；
没有可用的 tag 时退化为 `dev+<commit>`；完全没有 git 信息时隐藏，不会显示 `0.0.0`。

发布 GitHub Release（需要干净的工作区、已配置的 `origin` 远程，以及已登录的
[`gh`](https://cli.github.com/) CLI）：

```bash
gh auth login
make notes VERSION=v0.2.0        # 生成 build/release/notes.md
$EDITOR build/release/notes.md   # 可选：手动修改更新说明
make release VERSION=v0.2.0      # 构建、打包、打 tag 并发布
```

`make release` 会构建并打包当前平台，创建并推送 `v0.2.0` tag，然后用更新说明发布产物。
其他平台可以先单独构建打包，再把产物上传到同一个 Release：

```bash
make build-release VERSION=v0.2.0 PLATFORM=windows/amd64
make package-assets VERSION=v0.2.0 PLATFORM=windows/amd64
make publish VERSION=v0.2.0
```

macOS 通用包用 `PLATFORM=darwin/universal`；Linux 包需要在 Linux（或 Linux 容器）里构建。
完整目标列表见 `make help`。

### 项目结构

| 路径 | 内容 |
| --- | --- |
| `main.go`、`app.go` | Wails 应用入口 |
| `internal/repeater/` | 请求解析、原始 socket 发送、代理、响应解码、i18n |
| `frontend/src/` | Vue 3 界面、组件、Pinia store、编辑器与 i18n 文案 |
| `build/` | 应用图标与各平台打包资源 |
| `assets/screenshots/` | README 使用的截图 |
| `Makefile` | 构建、打包与 GitHub Release 目标 |
| `scripts/` | Release notes 与 `wails.json` 版本辅助脚本 |

## 许可证

本项目基于 [Apache License 2.0](LICENSE) 开源。

## 致谢

- [Wails](https://wails.io/) — 桌面应用外壳
- [Naive UI](https://www.naiveui.com/) — Vue 3 组件库
- [CodeMirror](https://codemirror.net/) — 请求编辑器
- [Ionicons](https://ionic.io/ionicons) — 界面图标（通过 `@vicons/ionicons5`）
- Burp Suite Repeater — 本工具参考的工作流
