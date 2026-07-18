<template>
  <div class="platform-mgmt">
    <div class="page-header">
      <div>
        <h2>平台管理</h2>
        <p class="header-desc">注册和管理所有能源管控平台，配置变更后前端自动生效</p>
      </div>
      <el-button type="primary" :icon="Plus" @click="openCreate">
        新增平台
      </el-button>
    </div>

    <div class="al-card">
      <el-table :data="platformStore.platforms" v-loading="loading" stripe>
        <el-table-column prop="icon" label="图标" width="60" align="center">
          <template #default="{ row }">
            <span style="font-size: 20px">
              <template v-if="isEmoji(row.icon)">{{ row.icon }}</template>
              <el-icon v-else><component :is="row.icon" /></el-icon>
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="id" label="平台ID" width="120" />
        <el-table-column prop="name" label="平台名称" width="180" />
        <el-table-column prop="schema" label="数据库Schema" width="150" />
        <el-table-column label="菜单数" width="80" align="center">
          <template #default="{ row }">
            {{ countNavItems(row.config?.navItems) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small" effect="dark">
              {{ row.status === 'active' ? '运行中' : '已停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sortOrder" label="排序" width="80" align="center" />
        <el-table-column label="操作" width="160" align="center" fixed="right">
          <template #default="{ row }">
            <el-button size="small" text type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" text type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 新增/编辑平台弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑平台' : '新增平台'"
      width="880px"
      :close-on-click-modal="false"
      draggable
      top="5vh"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="120px" label-position="right">
        <el-form-item label="平台ID" prop="id">
          <el-input v-model="form.id" placeholder="如 farm、factory" :disabled="isEdit" style="width: 200px" @input="onPlatformIdChange" />
          <span class="form-hint">菜单路径将自动以 /{{ form.id || '平台ID' }}/ 开头</span>
        </el-form-item>
        <el-form-item label="平台名称" prop="name">
          <el-input v-model="form.name" placeholder="如 农场监控平台" style="width: 260px" />
        </el-form-item>
        <el-form-item label="平台图标" prop="icon">
          <IconPicker v-model="form.icon" />
        </el-form-item>
        <el-form-item label="数据库Schema" prop="schema">
          <el-input v-model="form.schema" placeholder="如 schema_farm" style="width: 260px" />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="状态" prop="status">
              <el-select v-model="form.status" style="width: 160px">
                <el-option label="运行中" value="active" />
                <el-option label="已停用" value="disabled" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="排序" prop="sortOrder">
              <el-input-number v-model="form.sortOrder" :min="0" :max="999" disabled style="width: 160px" />
              <span class="form-hint">自动分配</span>
            </el-form-item>
          </el-col>
        </el-row>

        <!-- ===== 菜单 & 页面一体化配置 ===== -->
        <el-divider content-position="left">菜单 & 页面配置</el-divider>

        <div class="menu-editor">
          <div class="menu-editor-toolbar">
            <el-button type="primary" size="small" :icon="Plus" @click="addNavItem()">
              添加菜单
            </el-button>
            <span class="menu-hint">只需填写菜单名称和选择页面类型，路由路径自动生成</span>
          </div>

          <!-- 菜单列表 -->
          <div class="nav-list">
            <div
              v-for="(item, idx) in navItems"
              :key="idx"
              class="nav-item-block"
            >
              <!-- 一级菜单行 -->
              <div class="nav-item-row">
                <span class="nav-level-tag l1">L1</span>
                <el-input
                  v-model="item.label"
                  placeholder="菜单名称（如 设备管理）"
                  size="small"
                  style="width: 160px"
                  @input="onMenuLabelChange(item)"
                />
                <span class="auto-path">/{{ form.id || 'id' }}/{{ item._pageKey || '?' }}</span>
                <el-select
                  v-model="item._pageType"
                  size="small"
                  style="width: 120px"
                  placeholder="页面类型"
                  @change="onPageTypeChange(item)"
                >
                  <el-option label="仪表盘" value="dashboard" />
                  <el-option label="表格" value="table" />
                  <el-option label="图表" value="chart" />
                  <el-option label="自定义组件" value="custom" />
                </el-select>
                <IconPicker v-model="item.icon" />
                <el-button size="small" text type="primary" :icon="Plus" @click="addNavChild(item)">子菜单</el-button>
                <el-button size="small" text type="danger" :icon="Delete" @click="removeNavItem(idx)" />
              </div>

              <!-- 表格列配置（仅 table 类型显示） -->
              <div v-if="item._pageType === 'table' && !item.children?.length" class="page-columns-inline">
                <el-button size="small" text type="primary" :icon="Plus" @click="addColumn(item)">添加列</el-button>
                <div v-for="(col, ci) in item._columns" :key="ci" class="col-row">
                  <el-input v-model="col.field" placeholder="字段名" size="small" style="width: 100px" />
                  <el-input v-model="col.label" placeholder="列标题" size="small" style="width: 100px" />
                  <el-input-number v-model="col.width" placeholder="宽度" size="small" :min="60" :max="500" style="width: 90px" />
                  <el-select v-model="col.type" size="small" style="width: 90px" placeholder="类型">
                    <el-option label="文本" value="text" />
                    <el-option label="标签" value="tag" />
                    <el-option label="圆点" value="dot" />
                  </el-select>
                  <el-input
                    v-if="col.type === 'tag' || col.type === 'dot'"
                    v-model="col.optionsStr"
                    placeholder='{"online":"在线"}'
                    size="small"
                    style="width: 200px"
                  />
                  <el-button size="small" text type="danger" :icon="Delete" @click="item._columns?.splice(ci, 1)" />
                </div>
              </div>

              <!-- 自定义组件名（仅 custom 类型显示） -->
              <div v-if="item._pageType === 'custom' && !item.children?.length" class="custom-component-inline">
                <span class="inline-label">组件名：</span>
                <el-input
                  v-model="item._component"
                  placeholder="如 FarmDashboard（需在注册表中注册）"
                  size="small"
                  style="width: 280px"
                />
              </div>

              <!-- 二级菜单 -->
              <div v-if="item.children && item.children.length" class="nav-children">
                <div
                  v-for="(child, cidx) in item.children"
                  :key="cidx"
                  class="nav-item-row child-row"
                >
                  <span class="nav-level-tag l2">L2</span>
                  <el-input
                    v-model="child.label"
                    placeholder="子菜单名称"
                    size="small"
                    style="width: 140px"
                    @input="onMenuLabelChange(child)"
                  />
                  <span class="auto-path">/{{ form.id || 'id' }}/{{ child._pageKey || '?' }}</span>
                  <el-select
                    v-model="child._pageType"
                    size="small"
                    style="width: 110px"
                    placeholder="页面类型"
                    @change="onPageTypeChange(child)"
                  >
                    <el-option label="仪表盘" value="dashboard" />
                    <el-option label="表格" value="table" />
                    <el-option label="图表" value="chart" />
                    <el-option label="自定义组件" value="custom" />
                  </el-select>
                  <IconPicker v-model="child.icon" />
                  <el-button size="small" text type="danger" :icon="Delete" @click="removeNavChild(item, cidx)" />
                </div>
              </div>
            </div>
          </div>

          <el-empty v-if="navItems.length === 0" description="暂无菜单，点击上方按钮添加" :image-size="60" />
        </div>

        <!-- JSON 预览 -->
        <el-divider content-position="left">
          <el-link type="info" @click="showJsonPreview = !showJsonPreview">
            {{ showJsonPreview ? '收起' : '展开' }} JSON 预览
          </el-link>
        </el-divider>
        <el-input
          v-if="showJsonPreview"
          :model-value="generatedJson"
          type="textarea"
          :rows="10"
          readonly
          class="config-preview"
        />
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { Plus, Delete, Rank } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { usePlatformStore } from '@/stores/platform'
import { createPlatform, updatePlatform, deletePlatform } from '@/api/platform'
import { setupDynamicRoutes, resetDynamicRoutes } from '@/router'
import IconPicker from '@/components/IconPicker.vue'
import type { Platform, NavItem, ColumnDef } from '@/types/platform'

const platformStore = usePlatformStore()
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const showJsonPreview = ref(false)

const form = reactive({
  id: '',
  name: '',
  icon: '',
  schema: '',
  status: 'active',
  sortOrder: 0,
})

// ===== 编辑用扩展类型 =====
interface EditableColumn extends ColumnDef {
  optionsStr?: string
}

interface EditableNavItem extends NavItem {
  _pageKey?: string   // 页面标识，自动从菜单名生成
  _pageType?: string  // 页面类型
  _columns?: EditableColumn[]
  _component?: string
  _children?: EditableNavItem[]
}

// ===== 菜单数据 =====
const navItems = reactive<EditableNavItem[]>([])

const rules = {
  id: [{ required: true, message: '请输入平台ID', trigger: 'blur' }],
  name: [{ required: true, message: '请输入平台名称', trigger: 'blur' }],
  schema: [{ required: true, message: '请输入数据库Schema', trigger: 'blur' }],
}

// ===== 中文 -> 英文 key 映射表 =====
const CN_EN_MAP: Record<string, string> = {
  '概览': 'overview', '总览': 'overview', '首页': 'home', '仪表盘': 'dashboard',
  '设备': 'devices', '设备管理': 'devices', '设备列表': 'devices', '设备监控': 'devices',
  '告警': 'alarms', '告警管理': 'alarms', '告警列表': 'alarms', '报警': 'alarms',
  '事件': 'events', '事件告警': 'events', '事件列表': 'events',
  '历史': 'history', '历史数据': 'history', '历史记录': 'history',
  '实时': 'realtime', '实时数据': 'realtime', '实时监控': 'realtime',
  '统计': 'stats', '统计分析': 'stats',
  '电站': 'stations', '台区': 'stations', '站点': 'stations', '台区管理': 'stations',
  '升级': 'ota', 'OTA升级': 'ota', '固件': 'ota', '固件管理': 'ota',
  '日志': 'logs', '系统日志': 'logs',
  '用户': 'users', '用户管理': 'users',
  '设置': 'settings', '系统设置': 'settings',
  '监控': 'monitor', '电解槽': 'cells', '电解槽监控': 'cells',
  '点位': 'points', '点位实时': 'points', '测点': 'points',
  '发电': 'generation', '发电统计': 'generation',
  '分析': 'analysis', '报表': 'report', '报表分析': 'report',
  '地图': 'map', '拓扑': 'topology',
  '水泵': 'pumps', '水质': 'water-quality', '流量': 'flow',
  '温度': 'temperature', '压力': 'pressure', '能耗': 'energy',
  '机组': 'generators', '发电机': 'generators', '锅炉': 'boiler',
  '农场': 'farm', '温室': 'greenhouse', '灌溉': 'irrigation',
  '土壤': 'soil', '气象': 'weather', '摄像头': 'camera', '视频': 'video',
}

/** 从中文菜单名称生成英文 key */
function generatePageKey(label: string, existingKeys: Set<string>): string {
  // 1. 先查映射表
  if (CN_EN_MAP[label]) {
    let key = CN_EN_MAP[label]
    // 避免重复
    if (!existingKeys.has(key)) return key
  }

  // 2. 尝试匹配部分关键词
  for (const [cn, en] of Object.entries(CN_EN_MAP)) {
    if (label.includes(cn)) {
      if (!existingKeys.has(en)) return en
    }
  }

  // 3. 用拼音首字母（简易实现：取 Unicode 编码做 hash）
  if (/^[a-zA-Z0-9_-]+$/.test(label)) {
    // 纯英文直接用
    const key = label.toLowerCase().replace(/\s+/g, '-')
    if (!existingKeys.has(key)) return key
  }

  // 4. 兜底：用 page_N
  let n = 1
  while (existingKeys.has(`page_${n}`)) n++
  return `page_${n}`
}

/** 收集所有已使用的 pageKey */
function collectUsedKeys(exclude?: EditableNavItem): Set<string> {
  const keys = new Set<string>()
  for (const item of navItems) {
    if (item !== exclude && item._pageKey) keys.add(item._pageKey)
    if (item.children) {
      for (const child of item.children) {
        if (child !== exclude && child._pageKey) keys.add(child._pageKey)
      }
    }
  }
  return keys
}

/** 判断是否为 Emoji */
function isEmoji(str: string): boolean {
  return /[\u{1F000}-\u{1FFFF}]|[\u{2600}-\u{27BF}]/u.test(str)
}

/** 统计菜单数量 */
function countNavItems(items?: NavItem[]): number {
  if (!items) return 0
  let count = 0
  for (const item of items) {
    count++
    if (item.children) count += item.children.length
  }
  return count
}

// ===== 菜单操作 =====
function addNavItem() {
  const usedKeys = collectUsedKeys()
  const key = generatePageKey('新菜单', usedKeys)
  navItems.push({
    label: '新菜单',
    path: `/${form.id || 'id'}/${key}`,
    icon: '',
    _pageKey: key,
    _pageType: 'dashboard',
  })
}

function addNavChild(parent: EditableNavItem) {
  if (!parent.children) parent.children = []
  if (!parent._children) parent._children = []
  const usedKeys = collectUsedKeys()
  const key = generatePageKey('子菜单', usedKeys)
  const child: EditableNavItem = {
    label: '子菜单',
    path: `/${form.id || 'id'}/${key}`,
    icon: '',
    _pageKey: key,
    _pageType: 'table',
  }
  parent.children.push(child as NavItem)
  parent._children.push(child)
}

function removeNavItem(idx: number) {
  navItems.splice(idx, 1)
}

function removeNavChild(parent: EditableNavItem, idx: number) {
  parent.children?.splice(idx, 1)
  parent._children?.splice(idx, 1)
  if (parent.children && parent.children.length === 0) {
    delete parent.children
  }
}

/** 菜单名称变化时，重新生成 pageKey 和 path */
function onMenuLabelChange(item: EditableNavItem) {
  const usedKeys = collectUsedKeys(item)
  const newKey = generatePageKey(item.label, usedKeys)
  item._pageKey = newKey
  item.path = `/${form.id || 'id'}/${newKey}`
}

/** 平台 ID 变化时，重新生成所有路径 */
function onPlatformIdChange() {
  for (const item of navItems) {
    item.path = `/${form.id || 'id'}/${item._pageKey || ''}`
    if (item.children) {
      for (const child of item.children as EditableNavItem[]) {
        child.path = `/${form.id || 'id'}/${child._pageKey || ''}`
      }
    }
  }
}

/** 页面类型变化 */
function onPageTypeChange(item: EditableNavItem) {
  if (item._pageType === 'table' && !item._columns) {
    item._columns = []
  }
}

/** 添加列 */
function addColumn(item: EditableNavItem) {
  if (!item._columns) item._columns = []
  item._columns.push({
    field: '',
    label: '',
    width: 120,
    type: 'text',
  })
}

// ===== 生成最终 JSON =====
const generatedJson = computed(() => {
  const pagesOut: Record<string, any> = {}

  // 构建菜单输出 + 页面输出
  const navOut = navItems.map(item => {
    const nav: any = { label: item.label }
    if (item.path) nav.path = item.path
    if (item.icon) nav.icon = item.icon

    // 如果有子菜单，不生成页面（父菜单只做展开）
    if (item.children && item.children.length > 0) {
      nav.children = item.children.map(child => {
        const childNav: any = { label: child.label }
        if (child.path) childNav.path = child.path
        if (child.icon) childNav.icon = child.icon

        // 子菜单生成页面
        const childEditable = child as EditableNavItem
        if (childEditable._pageKey) {
          pagesOut[childEditable._pageKey] = buildPageConfig(childEditable)
        }
        return childNav
      })
    } else {
      // 无子菜单的一级菜单生成页面
      if (item._pageKey) {
        pagesOut[item._pageKey] = buildPageConfig(item)
      }
    }
    return nav
  })

  return JSON.stringify({
    navItems: navOut,
    pages: pagesOut,
  }, null, 2)
})

/** 构建单个页面配置 */
function buildPageConfig(item: EditableNavItem): any {
  const page: any = {
    type: item._pageType || 'dashboard',
    title: item.label,
  }
  if (item._pageType === 'table' || item._pageType === 'chart' || item._pageType === 'dashboard') {
    page.api = item._pageKey || 'data'
  }
  if (item._pageType === 'custom' && item._component) {
    page.component = item._component
  }
  if (item._pageType === 'table' && item._columns && item._columns.length) {
    page.columns = item._columns.map(col => {
      const out: any = { field: col.field, label: col.label }
      if (col.width) out.width = col.width
      if (col.type) out.type = col.type
      if (col.optionsStr) {
        try { out.options = JSON.parse(col.optionsStr) } catch { /* ignore */ }
      }
      return out
    })
  }
  return page
}

// ===== 打开弹窗 =====
function openCreate() {
  isEdit.value = false
  form.id = ''
  form.name = ''
  form.icon = ''
  form.schema = ''
  form.status = 'active'
  form.sortOrder = platformStore.platforms.length + 1

  // 默认菜单
  navItems.length = 0
  const defaultKey = 'overview'
  navItems.push({
    label: '概览',
    path: '/id/overview',
    icon: 'Odometer',
    _pageKey: defaultKey,
    _pageType: 'dashboard',
  })

  showJsonPreview.value = false
  dialogVisible.value = true
}

function openEdit(row: Platform) {
  isEdit.value = true
  form.id = row.id
  form.name = row.name
  form.icon = row.icon || ''
  form.schema = row.schema
  form.status = row.status
  form.sortOrder = row.sortOrder

  navItems.length = 0

  // 从已有配置反向解析
  if (row.config?.navItems) {
    for (const nav of row.config.navItems) {
      const editable: EditableNavItem = {
        label: nav.label,
        path: nav.path,
        icon: nav.icon,
        _pageKey: nav.path ? nav.path.split('/').pop() || '' : '',
        _pageType: 'dashboard',
        _children: [],
      }

      // 从 pages 中找到对应的页面配置
      if (editable._pageKey && row.config.pages?.[editable._pageKey]) {
        const page = row.config.pages[editable._pageKey]
        editable._pageType = page.type
        if (page.columns) {
          editable._columns = page.columns.map(col => ({
            ...col,
            optionsStr: col.options ? JSON.stringify(col.options) : undefined,
          })) as EditableColumn[]
        }
        if (page.component) editable._component = page.component
      }

      // 子菜单
      if (nav.children) {
        editable.children = []
        for (const child of nav.children) {
          const childEditable: EditableNavItem = {
            label: child.label,
            path: child.path,
            icon: child.icon,
            _pageKey: child.path ? child.path.split('/').pop() || '' : '',
            _pageType: 'dashboard',
          }
          if (childEditable._pageKey && row.config.pages?.[childEditable._pageKey]) {
            const page = row.config.pages[childEditable._pageKey]
            childEditable._pageType = page.type
            if (page.columns) {
              childEditable._columns = page.columns.map(col => ({
                ...col,
                optionsStr: col.options ? JSON.stringify(col.options) : undefined,
              })) as EditableColumn[]
            }
            if (page.component) childEditable._component = page.component
          }
          editable.children!.push(childEditable as NavItem)
          editable._children!.push(childEditable)
        }
      }
      navItems.push(editable)
    }
  }

  showJsonPreview.value = false
  dialogVisible.value = true
}

// ===== 提交 =====
async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid: boolean) => {
    if (!valid) return

    if (navItems.length === 0) {
      ElMessage.warning('请至少添加一个菜单')
      return
    }

    // 检查菜单名称
    const hasError = navItems.some(item => {
      if (!item.label) return true
      if (item.children) return item.children.some(c => !c.label)
      return false
    })
    if (hasError) {
      ElMessage.warning('菜单名称不能为空')
      return
    }

    submitting.value = true
    try {
      const config = JSON.parse(generatedJson.value)
      const data: Partial<Platform> = {
        name: form.name,
        icon: form.icon,
        schema: form.schema,
        status: form.status,
        sortOrder: form.sortOrder,
        config,
      }

      if (isEdit.value) {
        await updatePlatform(form.id, data)
        ElMessage.success('平台已更新')
      } else {
        data.id = form.id
        await createPlatform(data)
        ElMessage.success('平台已创建')
      }

      dialogVisible.value = false
      await platformStore.loadPlatforms()
      resetDynamicRoutes()
      setupDynamicRoutes(platformStore.platforms)
    } catch (e: any) {
      ElMessage.error(e?.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

async function handleDelete(row: Platform) {
  try {
    await ElMessageBox.confirm(
      `确定要删除平台「${row.name}」吗？此操作不可恢复。`,
      '删除确认',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deletePlatform(row.id)
    ElMessage.success('平台已删除')
    await platformStore.loadPlatforms()
    resetDynamicRoutes()
    setupDynamicRoutes(platformStore.platforms)
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

platformStore.loadPlatforms().then(() => {
  loading.value = false
})
</script>

<style scoped lang="scss">
.platform-mgmt {
  .page-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 20px;

    h2 {
      font-size: 20px;
      font-weight: 700;
      color: #e6edf3;
    }

    .header-desc {
      font-size: 13px;
      color: #8b949e;
      margin-top: 4px;
    }
  }

  .form-hint {
    margin-left: 8px;
    font-size: 12px;
    color: #6e7681;
  }
}

/* ===== 菜单编辑器 ===== */
.menu-editor {
  .menu-editor-toolbar {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;

    .menu-hint {
      font-size: 12px;
      color: #6e7681;
    }
  }

  .nav-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .nav-item-block {
    background: rgba(22, 27, 34, 0.6);
    border: 1px solid #30363d;
    border-radius: 8px;
    padding: 8px 12px;
  }

  .nav-item-row {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;

    &.child-row {
      margin-left: 24px;
    }
  }

  .nav-children {
    margin-top: 8px;
    padding-left: 8px;
    border-left: 2px solid #4b3fe3;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .nav-level-tag {
    font-size: 10px;
    font-weight: 700;
    padding: 1px 6px;
    border-radius: 3px;
    min-width: 24px;
    text-align: center;

    &.l1 { color: #7c6ff5; background: rgba(75, 63, 227, 0.12); }
    &.l2 { color: #f0883e; background: rgba(240, 136, 62, 0.12); }
  }

  .auto-path {
    font-size: 11px;
    font-family: 'Consolas', monospace;
    color: #7c6ff5;
    background: rgba(75, 63, 227, 0.08);
    padding: 2px 8px;
    border-radius: 4px;
    min-width: 140px;
    text-align: center;
  }

  .page-columns-inline {
    margin-top: 6px;
    padding: 8px;
    background: rgba(13, 17, 23, 0.6);
    border-radius: 6px;

    .col-row {
      display: flex;
      align-items: center;
      gap: 6px;
      margin-top: 4px;
      flex-wrap: wrap;
    }
  }

  .custom-component-inline {
    margin-top: 6px;
    padding: 8px;
    background: rgba(13, 17, 23, 0.6);
    border-radius: 6px;
    display: flex;
    align-items: center;
    gap: 8px;

    .inline-label {
      font-size: 12px;
      color: #8b949e;
      white-space: nowrap;
    }
  }
}

.config-preview {
  :deep(.el-textarea__inner) {
    font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
    font-size: 12px;
    line-height: 1.6;
    background: #0d1117;
    border-color: #30363d;
    color: #e6edf3;
  }
}
</style>
