import type { TuiSidebarMcpItem } from "@opencode-ai/plugin/tui"

export type ConfiguredMcpServer = {
  name: string
  enabled: boolean
}

export type McpFooterState =
  | { status: "loading" }
  | { status: "unavailable" }
  | { status: "empty" }
  | { status: "resolved"; servers: TuiSidebarMcpItem[] }

export function resolveMcpFooterState(
  configured: readonly ConfiguredMcpServer[] | undefined,
  returned: readonly TuiSidebarMcpItem[],
  timedOut = false,
): McpFooterState {
  if (!configured) return { status: timedOut ? "unavailable" : "loading" }
  if (configured.length === 0) return { status: "empty" }

  const returnedByName = new Map(returned.map((server) => [server.name, server]))
  const missingEnabled = configured.some((server) => server.enabled && !returnedByName.has(server.name))
  if (missingEnabled && !timedOut) {
    return { status: "loading" }
  }

  return {
    status: "resolved",
    servers: configured.map((server) =>
      server.enabled
        ? returnedByName.get(server.name) ?? { name: server.name, status: "failed", error: "Status unavailable" }
        : { name: server.name, status: "disabled" },
    ),
  }
}
