const WS_PROTOCOL = 'atlas.v1'
const WS_TOKEN_PREFIX = 'bearer.'

function resolveWebSocketUrl(): string {
  const configured = import.meta.env.VITE_WS_URL?.trim()
  if (configured) {
    return configured
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}/ws`
}

export function createAuthenticatedWebSocket(token: string): WebSocket {
  return new WebSocket(resolveWebSocketUrl(), [WS_PROTOCOL, `${WS_TOKEN_PREFIX}${token}`])
}

export { WS_PROTOCOL, WS_TOKEN_PREFIX }
