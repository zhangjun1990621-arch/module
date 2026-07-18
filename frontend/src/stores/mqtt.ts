import { defineStore } from 'pinia'
import { ref } from 'vue'

interface CommandStatus {
  token: number
  type: string
  deviceId: string
  status: 'pending' | 'success' | 'error' | 'timeout'
  result?: any
}

export const useMqttStore = defineStore('mqtt', () => {
  const pendingCommands = ref<Map<number, CommandStatus>>(new Map())

  function registerCommand(token: number, type: string, deviceId: string) {
    pendingCommands.value.set(token, {
      token,
      type,
      deviceId,
      status: 'pending'
    })
  }

  function resolveCommand(token: number, result: any) {
    const cmd = pendingCommands.value.get(token)
    if (cmd) {
      cmd.status = result.confirm === 0 ? 'success' : 'error'
      cmd.result = result
    }
  }

  function timeoutCommand(token: number) {
    const cmd = pendingCommands.value.get(token)
    if (cmd) {
      cmd.status = 'timeout'
    }
  }

  function getCommandStatus(token: number) {
    return pendingCommands.value.get(token)
  }

  return {
    pendingCommands,
    registerCommand,
    resolveCommand,
    timeoutCommand,
    getCommandStatus
  }
})
