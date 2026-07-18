import { ref, onMounted, onUnmounted } from 'vue'

export interface WSMessage {
  type: string
  data: any
}

export function useWebSocket() {
  const connected = ref(false)
  const lastMessage = ref<WSMessage | null>(null)
  let ws: WebSocket | null = null
  let reconnectTimer: number | null = null
  const listeners: Map<string, ((data: any) => void)[]> = new Map()

  function connect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const url = `${protocol}//${host}/ws`

    try {
      ws = new WebSocket(url)

      ws.onopen = () => {
        connected.value = true
      }

      ws.onmessage = (event) => {
        try {
          const msg: WSMessage = JSON.parse(event.data)
          lastMessage.value = msg
          const fns = listeners.get(msg.type)
          if (fns) {
            fns.forEach((fn) => fn(msg.data))
          }
        } catch (e) {
          // parse error, skip
        }
      }

      ws.onclose = () => {
        connected.value = false
        reconnectTimer = window.setTimeout(connect, 3000)
      }

      ws.onerror = () => {
        // silent
      }
    } catch (e) {
      // WebSocket not available, silently skip
    }
  }

  function on(type: string, callback: (data: any) => void) {
    if (!listeners.has(type)) {
      listeners.set(type, [])
    }
    listeners.get(type)!.push(callback)
  }

  function off(type: string, callback: (data: any) => void) {
    const fns = listeners.get(type)
    if (fns) {
      const idx = fns.indexOf(callback)
      if (idx >= 0) fns.splice(idx, 1)
    }
  }

  function disconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (ws) {
      ws.close()
      ws = null
    }
    connected.value = false
  }

  onMounted(() => {
    connect()
  })

  onUnmounted(() => {
    disconnect()
  })

  return { connected, lastMessage, on, off, connect, disconnect }
}
