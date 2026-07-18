<template>
  <div class="dynamic-page">
    <!-- 页面标题（custom 类型由组件内部自行渲染标题，不重复显示） -->
    <div class="page-header" v-if="pageConfig && pageConfig.type !== 'custom'">
      <h2>{{ pageConfig.title }}</h2>
    </div>

    <!-- 根据页面类型渲染对应通用组件 -->
    <div class="page-body" v-if="pageConfig">
      <!-- 表格类型 -->
      <DynamicTable
        v-if="pageConfig.type === 'table'"
        :columns="pageConfig.columns || []"
        :api="pageConfig.api || ''"
        :platform="platformId"
      />

      <!-- 仪表盘类型 -->
      <DynamicDashboard
        v-else-if="pageConfig.type === 'dashboard'"
        :platform="platformId"
        :config="pageConfig.layout"
      />

      <!-- 图表类型 -->
      <DynamicChart
        v-else-if="pageConfig.type === 'chart'"
        :api="pageConfig.api || ''"
        :chart-type="pageConfig.chartType || 'line'"
        :platform="platformId"
        :title="pageConfig.title"
      />

      <!-- 自定义组件类型 -->
      <component
        v-else-if="pageConfig.type === 'custom' && customComponent"
        :is="customComponent"
        :platform="platformId"
      />

      <!-- 自定义组件未注册 -->
      <el-empty
        v-else-if="pageConfig.type === 'custom' && !customComponent"
        :description="`自定义组件未注册: ${pageConfig.component}`"
      />

      <!-- 未知类型 -->
      <el-empty v-else :description="`未支持的页面类型: ${pageConfig.type}`" />
    </div>

    <!-- 页面配置不存在 -->
    <el-empty v-else description="页面配置不存在" />
  </div>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue'
import { usePlatformStore } from '@/stores/platform'
import DynamicTable from './DynamicTable.vue'
import DynamicDashboard from './DynamicDashboard.vue'
import DynamicChart from './DynamicChart.vue'
import { getCustomComponent } from './customComponentRegistry'
import type { PageConfig } from '@/types/platform'

const props = defineProps<{
  platformId: string
  pagePath: string
}>()

const platformStore = usePlatformStore()

/** 当前平台 */
const platform = computed(() => platformStore.getPlatformById(props.platformId))

/** 当前页面配置 */
const pageConfig = computed<PageConfig | undefined>(() => {
  return platform.value?.config?.pages?.[props.pagePath]
})

/** 自定义组件（异步加载） */
const customComponent = computed(() => {
  const componentName = pageConfig.value?.component
  if (!componentName) return null
  const loader = getCustomComponent(componentName)
  if (!loader) return null
  return defineAsyncComponent(loader)
})
</script>

<style scoped lang="scss">
.dynamic-page {
  .page-header {
    margin-bottom: 20px;

    h2 {
      font-size: 20px;
      font-weight: 700;
      color: #e6edf3;
    }
  }

  .page-body {
    min-height: 400px;
  }
}
</style>
