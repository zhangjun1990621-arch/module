import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getDeviceTree, getDevice } from '@/api/device'

interface Device {
  id: string
  deviceId: string
  platformId: string
  deviceType: string
  model: string
  serialNo: string
  capacity: number
  firmware: string
  software: string
  hardware: string
  stationId: string | null
  status: string
  lastSeen: string | null
  lastOnline: string | null
  signalStrength: number
  name: string
  metadata: any
}

interface StationTree {
  id: number
  name: string
  code: string
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

// 将后端设备数据适配为前端 Device interface
function adaptDevice(raw: any): Device {
  const meta = raw.metadata || {}
  return {
    id: raw.id || '',
    deviceId: raw.deviceId || raw.device_id || '',
    platformId: raw.platformId || raw.platform_id || '',
    deviceType: meta.deviceType || meta.device_type || raw.deviceType || 'pv',
    model: meta.model || raw.model || 'SG110CX',
    serialNo: meta.serialNo || meta.serial_no || raw.serialNo || raw.deviceId || '',
    capacity: meta.capacity || raw.capacity || 110000,
    firmware: meta.firmware || raw.firmware || 'V3.2.1',
    software: meta.software || raw.software || meta.sw || 'V3.2.1',
    hardware: meta.hardware || raw.hardware || meta.hw || 'V2.0',
    stationId: raw.stationId || raw.station_id || null,
    status: raw.status || 'offline',
    lastSeen: raw.lastSeen || raw.last_seen || null,
    lastOnline: raw.lastSeen || raw.last_seen || raw.lastOnline || null,
    signalStrength: meta.signalStrength || meta.signal_strength || meta.cs || 28,
    name: raw.name || '',
    metadata: meta,
  }
}

export const useDeviceStore = defineStore('device', () => {
  const deviceTree = ref<StationTree[]>([])
  const selectedDeviceId = ref<string | null>(null)
  const selectedDevice = ref<Device | null>(null)
  const realtimeData = ref<RealtimeData | null>(null)

  async function fetchDeviceTree() {
    try {
      const res: any = await getDeviceTree()
      const rawTree = res.data || res || []
      // 适配后端返回的设备树结构
      deviceTree.value = (Array.isArray(rawTree) ? rawTree : []).map((station: any) => ({
        id: station.id || 0,
        name: station.name || station.code || '默认分组',
        code: station.code || station.name || '',
        deviceCount: station.deviceCount || station.devices?.length || 0,
        devices: (station.devices || []).map((d: any) => adaptDevice(d)),
      }))
    } catch (e) {
      console.error('fetchDeviceTree failed:', e)
      deviceTree.value = []
    }
  }

  async function selectDevice(id: string) {
    selectedDeviceId.value = id
    try {
      const res: any = await getDevice(id)
      const raw = res.data || res
      selectedDevice.value = adaptDevice(raw)
    } catch (e) {
      console.error('selectDevice failed:', e)
      // 如果 API 失败，从已加载的设备树中查找
      for (const station of deviceTree.value) {
        const device = station.devices?.find((d: any) => d.id === id)
        if (device) {
          selectedDevice.value = device
          break
        }
      }
    }
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
