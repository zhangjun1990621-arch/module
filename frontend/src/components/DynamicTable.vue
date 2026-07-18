<template>
  <div class="dynamic-table">
    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <el-input
          v-model="searchText"
          placeholder="搜索..."
          prefix-icon="Search"
          clearable
          style="width: 240px"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
        <el-button :icon="Refresh" @click="fetchData">刷新</el-button>
      </div>
    </div>

    <!-- 表格 -->
    <el-table
      :data="tableData"
      v-loading="loading"
      stripe
      border
      style="width: 100%"
      @sort-change="handleSortChange"
    >
      <el-table-column type="index" label="#" width="50" align="center" />

      <el-table-column
        v-for="col in columns"
        :key="col.field"
        :prop="col.field"
        :label="col.label"
        :width="col.width"
        :align="col.align || 'left'"
        :sortable="col.type === 'number' ? 'custom' : false"
        show-overflow-tooltip
      >
        <template #default="{ row }">
          <!-- 文本类型 -->
          <span v-if="!col.type || col.type === 'text'">
            {{ formatText(row[col.field], col) }}
          </span>

          <!-- 标签类型 -->
          <el-tag
            v-else-if="col.type === 'tag'"
            :type="getTagType(row[col.field])"
            size="small"
            effect="dark"
          >
            {{ col.options?.[row[col.field]] || row[col.field] || '-' }}
          </el-tag>

          <!-- 状态点类型 -->
          <span v-else-if="col.type === 'dot'" class="dot-cell">
            <span class="status-dot" :class="getDotClass(row[col.field])"></span>
            {{ col.options?.[row[col.field]] || row[col.field] || '-' }}
          </span>

          <!-- 温度类型 -->
          <span
            v-else-if="col.type === 'temperature'"
            :style="{ color: getTempColor(row[col.field]), fontWeight: 600 }"
          >
            {{ row[col.field] != null ? row[col.field] : '-' }}{{ col.unit || '°C' }}
          </span>

          <!-- 数字类型 -->
          <span
            v-else-if="col.type === 'number'"
            style="font-variant-numeric: tabular-nums; font-weight: 600"
          >
            {{ formatNumber(row[col.field]) }}{{ col.unit ? ' ' + col.unit : '' }}
          </span>
        </template>
      </el-table-column>

      <el-table-column label="操作" width="120" align="center" fixed="right">
        <template #default="{ row }">
          <el-button size="small" text type="primary" @click="$emit('view', row)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        background
        @size-change="fetchData"
        @current-change="fetchData"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getResources } from '@/api/resource'
import type { ColumnDef } from '@/types/platform'

const props = defineProps<{
  columns: ColumnDef[]
  api: string
  platform: string
}>()

defineEmits<{
  view: [row: any]
}>()

const loading = ref(false)
const tableData = ref<any[]>([])
const searchText = ref('')

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const sortConfig = ref<{ field: string; order: string } | null>(null)

/** 格式化文本 */
function formatText(value: any, col: ColumnDef): string {
  if (value == null || value === '') return '-'
  if (col.options && col.options[value]) return col.options[value]
  return String(value)
}

/** 格式化数字 */
function formatNumber(value: any): string {
  if (value == null || value === '') return '-'
  const num = Number(value)
  if (isNaN(num)) return String(value)
  return num.toLocaleString()
}

/** 标签类型映射 */
function getTagType(value: any): 'success' | 'warning' | 'danger' | 'info' | 'primary' {
  const map: Record<string, 'success' | 'warning' | 'danger' | 'info' | 'primary'> = {
    online: 'success',
    active: 'success',
    running: 'success',
    offline: 'info',
    inactive: 'info',
    alarm: 'danger',
    critical: 'danger',
    warning: 'warning',
    pending: 'warning'
  }
  return map[String(value)] || 'primary'
}

/** 状态点样式 */
function getDotClass(value: any): string {
  const map: Record<string, string> = {
    online: 'dot-green',
    active: 'dot-green',
    running: 'dot-green',
    offline: 'dot-gray',
    inactive: 'dot-gray',
    alarm: 'dot-red',
    critical: 'dot-red',
    warning: 'dot-yellow',
    pending: 'dot-yellow'
  }
  return map[String(value)] || 'dot-gray'
}

/** 温度着色 */
function getTempColor(temp: any): string {
  if (temp == null) return '#6e7681'
  const t = Number(temp)
  if (isNaN(t)) return '#6e7681'
  if (t >= 80) return '#f85149'
  if (t >= 60) return '#f0883e'
  if (t >= 40) return '#d29922'
  if (t >= 20) return '#3fb950'
  return '#58a6ff'
}

/** 获取数据 */
async function fetchData() {
  if (!props.api || !props.platform) return
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: pagination.page,
      pageSize: pagination.pageSize
    }
    if (searchText.value) {
      params.search = searchText.value
    }
    if (sortConfig.value) {
      params.sortBy = sortConfig.value.field
      params.sortOrder = sortConfig.value.order
    }

    const res = await getResources(props.platform, props.api, params)
    const data = res.data

    // 兼容分页结构和平铺数组
    if (Array.isArray(data)) {
      tableData.value = data
      pagination.total = data.length
    } else if (data && data.list) {
      tableData.value = data.list
      pagination.total = data.total || 0
    } else {
      tableData.value = data ? [data] : []
      pagination.total = tableData.value.length
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '数据加载失败')
    tableData.value = []
  } finally {
    loading.value = false
  }
}

/** 搜索 */
function handleSearch() {
  pagination.page = 1
  fetchData()
}

/** 排序 */
function handleSortChange({ prop, order }: { prop: string; order: string | null }) {
  if (order) {
    sortConfig.value = {
      field: prop,
      order: order === 'ascending' ? 'asc' : 'desc'
    }
  } else {
    sortConfig.value = null
  }
  fetchData()
}

/** 监听 props 变化重新加载 */
watch(
  () => [props.api, props.platform],
  () => {
    pagination.page = 1
    fetchData()
  }
)

onMounted(() => {
  fetchData()
})

defineExpose({ fetchData })
</script>

<style scoped lang="scss">
.dynamic-table {
  .toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;

    .toolbar-left {
      display: flex;
      align-items: center;
      gap: 12px;
    }
  }

  .pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }

  .dot-cell {
    display: inline-flex;
    align-items: center;
    gap: 6px;

    .status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      flex-shrink: 0;

      &.dot-green {
        background: #3fb950;
        box-shadow: 0 0 6px rgba(63, 185, 80, 0.5);
      }

      &.dot-red {
        background: #f85149;
        box-shadow: 0 0 6px rgba(248, 81, 73, 0.5);
        animation: pulse 1.5s infinite;
      }

      &.dot-yellow {
        background: #d29922;
      }

      &.dot-gray {
        background: #6e7681;
      }
    }
  }
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.4;
  }
}
</style>
