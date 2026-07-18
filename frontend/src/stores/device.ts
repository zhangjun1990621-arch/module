import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getDeviceTree, getDevice } from '@/api/device'

interface Device {
  id: string
  deviceType: string
  model: string
  serialNo: string
  capacity: number
  firmware: string
  software: string
  hardware: string
  stationId: number | null
  status: string
  lastOnline: string | null
  signalStrength: number
}

interface StationTree {
  id: number
  name: string
  code: string
  region: string
  deviceCount: number
  devices: Device[]
}

interface RealtimeData {
  ac: {
    ph: number
    v: number[]
    c: number[]
    p: number
    q: number
    pf: number
    f: number
  } | null
  dc: {
    v: number[]
    c: number[]
    p: number[]
  } | null
  cs: number
}

export const useDeviceStore = defineStore('device', () => {
  const deviceTree = ref<StationTree[]>([])
  const selectedDeviceId = ref<string | null>(null)
  const selectedDevice = ref<Device | null>(null)
  const realtimeData = ref<RealtimeData | null>(null)

  async function fetchDeviceTree() {
    const res: any = await getDeviceTree()
    deviceTree.value = res.data || []
  }

  async function selectDevice(id: string) {
    selectedDeviceId.value = id
    const res: any = await getDevice(id)
    selectedDevice.value = res.data
  }

  function updateRealtimeData(data: RealtimeData) {
    realtimeData.value = data
  }

  function updateDeviceStatus(deviceId: string, status: string) {
    for (const station of deviceTree.value) {
      const device = station.devices?.find((d: any) => d.id === deviceId)
      if (device) {
        device.status = status
        break
      }
    }
  }

  return {
    deviceTree,
    selectedDeviceId,
    selectedDevice,
    realtimeData,
    fetchDeviceTree,
    selectDevice,
    updateRealtimeData,
    updateDeviceStatus
  }
})
