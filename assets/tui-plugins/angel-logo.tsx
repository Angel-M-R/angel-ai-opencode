// @ts-nocheck
/** @jsxImportSource @opentui/solid */
import type { TuiPlugin, TuiPluginApi, TuiSidebarMcpItem } from "@opencode-ai/plugin/tui"
import { useTerminalDimensions } from "@opentui/solid"
import { createEffect, createMemo, createSignal, onCleanup } from "solid-js"
import { resolveMcpFooterState } from "./mcp-footer-state"

const id = "angel-logo"

const angelArt = [
  "    _                     _       _     _",
  "    / \\   _ __   __ _  ___| |     / \\   | |",
  "   / _ \\ | '_ \\ / _` |/ _ \\ |    / _ \\  | |",
  "  / ___ \\| | | | (_| |  __/ |   / ___ \\ | |",
  "  /_/   \\_\\_| |_|\\__, |\\___|_|  /_/   \\_\\|_| ",
  "                  |___/                         ",
]

const compactArt = [
  "   _   _     _ ",
  "  / \\ / \\   | |",
  "  / _ \\ _ \\  | | ",
  " / ___ \\__ \\ | | ",
  "/_/   \\_\\ \\_\\|_|",
]

const mcpStatusStyle = {
  connected: { icon: "●", label: "connected", tone: "success" },
  disabled: { icon: "○", label: "disabled", tone: "textMuted" },
  failed: { icon: "✕", label: "error", tone: "error" },
  needs_auth: { icon: "◆", label: "authentication required", tone: "warning" },
  needs_client_registration: { icon: "◆", label: "registration required", tone: "warning" },
} as const

const spinnerFrames = ["◐", "◓", "◑", "◒"]

const Logo = () => {
  const dim = useTerminalDimensions()
  const lines = createMemo(() => {
    const term = dim()
    return term.height >= angelArt.length + 6 && term.width >= 54 ? angelArt : compactArt
  })

  return (
    <box flexDirection="column" alignItems="center">
      {lines().map((line) => (
        <text fg="magenta">{line}</text>
      ))}
    </box>
  )
}

const McpStatus = ({ api }: { api: TuiPluginApi }) => {
  const [spinnerFrame, setSpinnerFrame] = createSignal(0)
  const [timedOut, setTimedOut] = createSignal(false)

  const configured = createMemo(() => {
    if (!api.state.ready) return undefined
    return Object.entries(api.state.config.mcp ?? {}).map(([name, config]) => ({
      name,
      enabled: config !== false && config?.enabled !== false,
    }))
  })
  const state = createMemo(() => resolveMcpFooterState(configured(), api.state.mcp(), timedOut()))
  createEffect(() => {
    if (state().status !== "loading") return
    const timer = setInterval(() => setSpinnerFrame((value) => (value + 1) % spinnerFrames.length), 120)
    const timeout = setTimeout(() => setTimedOut(true), 10_000)
    onCleanup(() => {
      clearInterval(timer)
      clearTimeout(timeout)
    })
  })
  const styleFor = (server: TuiSidebarMcpItem) =>
    mcpStatusStyle[server.status] ?? { icon: "?", label: server.status, tone: "textMuted" }
  const entries = createMemo(() => {
    const current = state()
    if (current.status !== "resolved") return []
    return current.servers.map((server) => {
      const style = styleFor(server)
      return {
        name: `${style.icon} ${server.name}`,
        status: `· ${style.label}`,
        tone: style.tone,
      }
    })
  })
  const nameColumnWidth = createMemo(() => Math.max(...entries().map((entry) => entry.name.length), 0) + 2)
  const statusColumnWidth = createMemo(() => Math.max(...entries().map((entry) => entry.status.length), 0))
  const tableWidth = createMemo(() => Math.max(nameColumnWidth() + statusColumnWidth(), 3))

  return (
    <box flexDirection="column" alignItems="center" marginTop={1}>
      <box flexDirection="row" width={tableWidth()} justifyContent="center">
        <text fg={api.theme.current.textMuted}>MCP</text>
      </box>
      {state().status === "loading" ? (
        <text fg={api.theme.current.textMuted}>{spinnerFrames[spinnerFrame()]} checking connections...</text>
      ) : state().status === "unavailable" ? (
        <text fg={api.theme.current.error}>✕ status unavailable</text>
      ) : state().status === "empty" ? (
        <text fg={api.theme.current.textMuted}>○ none configured</text>
      ) : (
        entries().map((entry) => (
          <box flexDirection="row" width={tableWidth()}>
            <box width={nameColumnWidth()}>
              <text fg={api.theme.current[entry.tone]}>{entry.name}</text>
            </box>
            <box width={statusColumnWidth()}>
              <text fg={api.theme.current[entry.tone]}>{entry.status}</text>
            </box>
          </box>
        ))
      )}
    </box>
  )
}

const tui: TuiPlugin = async (api) => {
  api.slots.register({
    id,
    order: 100,
    slots: {
      home_logo() {
        return (
          <box flexDirection="column" alignItems="center">
            <Logo />
            <McpStatus api={api} />
          </box>
        )
      },
    },
  })
}

const plugin = { id: "angel-logo", tui }
export default plugin
