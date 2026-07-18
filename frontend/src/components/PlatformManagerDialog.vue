<template>
  <el-dialog
    v-model="show"
    title="🧩 平台管理中心"
    width="920px"
    top="6vh"
    :close-on-click-modal="false"
    append-to-body
  >
    <div class="pm-body">
      <!-- 顶部摘要 -->
      <div class="pm-summary">
        <span class="pm-sum-item"><b>{{ PLATFORM_CONFIGS.length }}</b> 平台</span>
        <span class="pm-sum-sep">·</span>
        <span class="pm-sum-item">当前用户：<b>{{ authStore.user?.username || '-' }}</b></span>
        <span class="pm-sum-sep">·</span>
        <span class="pm-sum-item">身份：<b>{{ currentRoleName }}</b></span>
        <span v-if="!authStore.isSuperAdmin" class="pm-sum-tip">仅超级管理员可管理用户账号</span>
      </div>

      <!-- ① 平台入口 -->
      <div class="pm-section">
        <div class="pm-section-tt">
          <span class="pm-num">1</span>平台入口
          <span class="pm-hint">点击「进入平台」跳转到对应平台，无权限将置灰</span>
        </div>
        <div class="pm-card-grid">
          <div
            v-for="p in PLATFORM_CONFIGS"
            :key="p.id"
            class="pm-plat-card"
            :class="{ 'is-locked': !canEnter(p) }"
            :style="{ '--card-accent': p.accent }"
          >
            <div class="pm-pc-header">
              <span class="pm-plat-icon" :style="{ background: p.accent + '22', color: p.accent }">{{ p.icon }}</span>
              <div class="pm-pc-name">{{ p.name }}</div>
              <span v-if="p.disabled" class="pm-tag warn">规划中</span>
              <span v-else class="pm-tag ok">已接入</span>
            </div>
            <div class="pm-pc-desc">{{ p.desc }}</div>
            <div class="pm-pc-footer">
              <span class="pm-pc-meta">{{ p.navItems.length }} 项菜单</span>
              <el-button v-if="canEnter(p)" type="primary" size="small" @click="enter(p)">进入平台 →</el-button>
              <el-button v-else type="info" size="small" disabled>{{ p.disabled ? '暂未开放' : '无权限' }}</el-button>
            </div>
          </div>
        </div>
      </div>

      <!-- ② 用户账号管理（仅超级管理员） -->
      <div v-if="authStore.isSuperAdmin" class="pm-section">
        <div class="pm-section-tt">
          <span class="pm-num">2</span>用户账号管理
          <span class="pm-hint">仅超级管理员可见 · 真实账号登录，不再使用模拟登录</span>
        </div>
        <div class="pm-toolbar">
          <el-button type="primary" size="small" @click="openCreate">+ 新增用户</el-button>
          <el-button size="small" @click="loadUsers">刷新</el-button>
          <span class="pm-toolbar-tip">共 {{ users.length }} 个账号</span>
        </div>
        <el-table
          :data="users"
          size="small"
          stripe
          v-loading="loading"
          class="pm-user-table"
        >
          <el-table-column prop="username" label="用户名" min-width="110" />
          <el-table-column label="角色" min-width="120">
            <template #default="{ row }">
              <span class="pm-role-badge" :style="roleBadgeStyle(row.role)">{{ roleLabel(row.role) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="可访问平台" min-width="160">
            <template #default="{ row }">
              <span
                v-for="p in userPlatformsOf(row)"
                :key="p.id"
                class="pm-plat-chip"
                :style="{ background: p.accent + '22', color: p.accent }"
                :title="p.name"
              >{{ p.icon }} {{ p.name.replace('平台', '').slice(0, 4) }}</span>
              <span v-if="!userPlatformsOf(row).length" class="pm-empty">无</span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small" effect="dark">
                {{ row.status === 'active' ? '正常' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="最后登录" width="170">
            <template #default="{ row }">{{ formatTime(row.lastLogin) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="130" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
              <el-button
                link
                type="danger"
                size="small"
                :disabled="row.id === authStore.user?.id"
                @click="deleteUser(row)"
              >删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- ③ 角色说明 -->
      <div class="pm-section">
        <div class="pm-section-tt">
          <span class="pm-num">{{ authStore.isSuperAdmin ? '3' : '2' }}</span>角色说明
          <span class="pm-hint">系统内置 7 种角色及其权限范围</span>
        </div>
        <table class="pm-table">
          <thead>
            <tr>
              <th style="width: 150px">角色</th>
              <th>说明</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in ROLE_DEFS" :key="r.id">
              <td>
                <span class="pm-role-badge" :style="roleBadgeStyle(r.id)">{{ r.name }}</span>
                <span v-if="r.id === store.currentRole" class="cur-mark">● 当前</span>
              </td>
              <td style="color: #8d9db8; font-size: 12px">{{ r.desc }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <template #footer>
      <span style="font-size: 11px; color: #6a7a90">当前角色：{{ currentRoleName }}</span>
      <el-button size="small" @click="show = false">关闭</el-button>
    </template>

    <!-- 新增 / 编辑用户弹窗 -->
    <el-dialog
      v-model="userDialogVisible"
      :title="formMode === 'create' ? '新增用户' : '编辑用户'"
      width="460px"
      append-to-body
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="92px">
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="form.username"
            :disabled="formMode === 'edit'"
            placeholder="3-50 个字符"
          />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            :placeholder="formMode === 'create' ? '至少 8 个字符' : '留空则不修改'"
          />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="form.role" style="width: 100%">
            <el-option v-for="r in ROLE_DEFS" :key="r.id" :label="r.name" :value="r.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="可访问平台" prop="platforms">
          <el-checkbox-group v-model="form.platforms" :disabled="platformsDisabled">
            <el-checkbox v-for="p in platformOptions" :key="p.id" :value="p.id">
              {{ p.icon }} {{ p.name.replace('平台', '') }}
            </el-checkbox>
          </el-checkbox-group>
          <div v-if="platformsDisabled" class="pm-form-tip">超级管理员默认拥有全部平台权限</div>
        </el-form-item>
        <el-form-item v-if="formMode === 'edit'" label="状态">
          <el-switch
            v-model="form.status"
            active-value="active"
            inactive-value="disabled"
            active-text="正常"
            inactive-text="禁用"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="userDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import request from '@/api/request'
import { useAuthStore } from '@/stores/auth'
import {
  usePlatformStore,
  PLATFORM_CONFIGS,
  ROLE_DEFS,
  getRoleColor,
  type PlatformConfig
} from '@/stores/platform'
import { ROLE_LABELS } from '@/utils/constants'

interface UserItem {
  id: number
  username: string
  role: string
  platforms: string
  status: string
  lastLogin: string | null
  createdAt?: string
}

const props = defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ (e: 'update:modelValue', v: boolean): void }>()
const show = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})

const router = useRouter()
const authStore = useAuthStore()
const store = usePlatformStore()

const currentRoleName = computed(() => ROLE_LABELS[store.currentRole] || store.currentRole)
const platformOptions = computed(() => PLATFORM_CONFIGS.filter((p) => !p.disabled))

function roleLabel(id: string) {
  return ROLE_LABELS[id] || id
}

/* ---------------- ① 平台入口 ---------------- */
function canEnter(p: PlatformConfig) {
  return !p.disabled && authStore.canAccessPlatform(p.id)
}
function enter(p: PlatformConfig) {
  if (!canEnter(p)) return
  store.setPlatform(p.id)
  show.value = false
  router.push(p.navItems[0]?.path || '/dashboard')
}

/* ---------------- ② 用户账号管理 ---------------- */
const users = ref<UserItem[]>([])
const loading = ref(false)

async function loadUsers() {
  if (!authStore.isSuperAdmin) return
  loading.value = true
  try {
    const res: any = await request.get('/users')
    users.value = (res.data || []) as UserItem[]
  } catch (e) {
    console.error('加载用户列表失败', e)
  } finally {
    loading.value = false
  }
}

function userPlatformsOf(u: UserItem): PlatformConfig[] {
  if (u.role === 'super_admin') return PLATFORM_CONFIGS.filter((p) => !p.disabled)
  const ids = (u.platforms || '').split(',').filter(Boolean)
  return PLATFORM_CONFIGS.filter((p) => !p.disabled && ids.includes(p.id))
}

function roleBadgeStyle(role: string) {
  const c = getRoleColor(role)
  return { background: c + '22', color: c, border: `1px solid ${c}44` }
}

function formatTime(t: string | null) {
  if (!t) return '—'
  return new Date(t).toLocaleString('zh-CN')
}

/* ---------------- 新增 / 编辑用户 ---------------- */
const userDialogVisible = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const submitting = ref(false)
const formRef = ref<FormInstance>()

const form = reactive({
  id: 0,
  username: '',
  password: '',
  role: 'ops',
  platforms: [] as string[],
  status: 'active'
})

const platformsDisabled = computed(() => form.role === 'super_admin')

const formRules = computed<FormRules>(() => ({
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名需 3-50 个字符', trigger: 'blur' }
  ],
  password:
    formMode.value === 'create'
      ? [
          { required: true, message: '请输入密码', trigger: 'blur' },
          { min: 8, message: '密码至少 8 个字符', trigger: 'blur' }
        ]
      : [{ min: 8, message: '密码至少 8 个字符', trigger: 'blur' }],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }],
  platforms: [
    { required: true, type: 'array', min: 1, message: '请至少选择一个平台', trigger: 'change' }
  ]
}))

// 切换为超级管理员时自动勾选全部平台
watch(
  () => form.role,
  (r) => {
    if (r === 'super_admin') {
      form.platforms = platformOptions.value.map((p) => p.id)
    }
  }
)

function openCreate() {
  formMode.value = 'create'
  form.id = 0
  form.username = ''
  form.password = ''
  form.role = 'ops'
  form.platforms = ['pv']
  form.status = 'active'
  userDialogVisible.value = true
  nextTick(() => formRef.value?.clearValidate())
}

function openEdit(u: UserItem) {
  formMode.value = 'edit'
  form.id = u.id
  form.username = u.username
  form.password = ''
  form.role = u.role
  if (u.role === 'super_admin') {
    form.platforms = platformOptions.value.map((p) => p.id)
  } else {
    form.platforms = (u.platforms || '').split(',').filter(Boolean)
  }
  form.status = u.status || 'active'
  userDialogVisible.value = true
  nextTick(() => formRef.value?.clearValidate())
}

async function submitForm() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  submitting.value = true
  try {
    const platformsStr = form.platforms.join(',')
    if (formMode.value === 'create') {
      await request.post('/users', {
        username: form.username,
        password: form.password,
        role: form.role,
        platforms: platformsStr
      })
      ElMessage.success('用户创建成功')
    } else {
      const payload: Record<string, unknown> = {
        role: form.role,
        platforms: platformsStr,
        status: form.status
      }
      if (form.password) payload.password = form.password
      await request.put(`/users/${form.id}`, payload)
      ElMessage.success('用户修改成功')
    }
    userDialogVisible.value = false
    loadUsers()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function deleteUser(u: UserItem) {
  try {
    await ElMessageBox.confirm(
      `确认删除用户「${u.username}」？该操作不可恢复。`,
      '删除确认',
      {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        confirmButtonClass: 'el-button--danger'
      }
    )
  } catch {
    return
  }
  try {
    await request.delete(`/users/${u.id}`)
    ElMessage.success('删除成功')
    loadUsers()
  } catch (e: any) {
    ElMessage.error(e?.message || '删除失败')
  }
}

// 弹窗打开且为超管时自动加载用户列表
watch(show, (v) => {
  if (v && authStore.isSuperAdmin) loadUsers()
})
</script>

<style scoped lang="scss">
.pm-body {
  color: #d0d8e8;
  max-height: 72vh;
  overflow-y: auto;
  overflow-x: hidden;
  padding-right: 8px;
  scrollbar-width: thin;
  scrollbar-color: #353d50 transparent;
}
.pm-body::-webkit-scrollbar { width: 8px; }
.pm-body::-webkit-scrollbar-thumb { background: #353d50; border-radius: 4px; }
.pm-body::-webkit-scrollbar-thumb:hover { background: #4a5568; }
.pm-body::-webkit-scrollbar-track { background: transparent; }

/* 顶部摘要 banner */
.pm-summary {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding: 12px 16px;
  margin-bottom: 16px;
  border-radius: 12px;
  background: linear-gradient(135deg, rgba(77, 163, 255, 0.10), rgba(157, 122, 255, 0.06));
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(77, 163, 255, 0.22);
  font-size: 12px;
  color: #b8c4d8;
  b { color: #e6ecf5; font-weight: 700; margin: 0 2px; font-size: 14px; }
}
.pm-sum-sep { color: #4a5568; }
.pm-sum-tip { margin-left: auto; font-size: 11px; color: #6a7a90; }

/* 卡片分区 — 玻璃拟态 + 入场动画 */
.pm-section {
  margin-bottom: 16px;
  padding: 16px 18px;
  background: rgba(20, 26, 38, 0.6);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 12px;
  transition: border-color 0.3s;
  opacity: 0;
  transform: translateY(12px);
  animation: pmSectionIn 0.4s ease forwards;
  &:nth-child(2) { animation-delay: 0.05s; }
  &:nth-child(3) { animation-delay: 0.12s; }
  &:nth-child(4) { animation-delay: 0.19s; }
  &:hover { border-color: rgba(77, 163, 255, 0.2); }
}
@keyframes pmSectionIn {
  to { opacity: 1; transform: translateY(0); }
}
.pm-section-tt {
  font-size: 13px;
  font-weight: 600;
  color: #e6ecf5;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  display: flex;
  align-items: center;
  gap: 8px;
}
.pm-num {
  width: 20px; height: 20px;
  border-radius: 50%;
  background: linear-gradient(135deg, #4da3ff, #9d7aff);
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(77, 163, 255, 0.35);
}
.pm-hint { font-size: 11px; color: #6a7a90; font-weight: 400; margin-left: auto; }

/* ---- 平台卡片网格 ---- */
.pm-card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 12px;
}
.pm-plat-card {
  background: rgba(24, 36, 52, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-left: 3px solid var(--card-accent, #4da3ff);
  border-radius: 10px;
  padding: 14px;
  transition: all 0.25s ease;
  display: flex;
  flex-direction: column;
  gap: 10px;

  &:hover {
    transform: translateY(-3px);
    background: rgba(24, 36, 52, 0.7);
    border-color: var(--card-accent, #4da3ff);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.35), 0 0 16px rgba(77, 163, 255, 0.08);
  }
  &.is-locked {
    opacity: 0.62;
    filter: grayscale(0.3);
  }

  .pm-pc-header {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .pm-pc-name {
    flex: 1;
    font-size: 14px;
    font-weight: 600;
    color: #e6ecf5;
  }
  .pm-pc-desc {
    font-size: 11px;
    color: #6a7a90;
  }
  .pm-pc-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-top: 8px;
    border-top: 1px solid rgba(255, 255, 255, 0.04);
  }
  .pm-pc-meta {
    font-size: 11px;
    color: #566880;
  }
}

/* 平台图标盒 */
.pm-plat-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px; height: 28px;
  border-radius: 8px;
  font-size: 16px;
  flex-shrink: 0;
}

/* ---- 工具栏 ---- */
.pm-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}
.pm-toolbar-tip { font-size: 11px; color: #6a7a90; margin-left: 4px; }

/* ---- 用户表格（el-table 暗色主题） ---- */
.pm-user-table {
  border-radius: 8px;
  overflow: hidden;
}
:deep(.pm-user-table) {
  --el-table-bg-color: #182434;
  --el-table-tr-bg-color: #182434;
  --el-table-row-hover-bg-color: #1d2b3e;
  --el-table-header-bg-color: #1d2b3e;
  --el-table-border-color: #253650;
  --el-table-text-color: #e6ecf5;
  --el-table-header-text-color: #8d9db8;
}

/* ---- 角色说明表（HTML 表格） ---- */
.pm-table { width: 100%; border-collapse: collapse; font-size: 12px; border-radius: 8px; overflow: hidden; }
.pm-table th, .pm-table td { border: 1px solid rgba(255, 255, 255, 0.05); padding: 9px 11px; text-align: left; }
.pm-table th {
  background: rgba(29, 43, 62, 0.6);
  color: #8d9db8;
  font-weight: 600;
  font-size: 11px;
  letter-spacing: 0.3px;
}
.pm-table tbody tr { transition: background 0.15s; }
.pm-table tbody tr:nth-child(even) { background: rgba(255, 255, 255, 0.012); }
.pm-table tbody tr:hover { background: rgba(77, 163, 255, 0.06); }

/* 角色 / 平台徽标 */
.pm-role-badge {
  display: inline-flex;
  align-items: center;
  padding: 3px 11px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
}
.pm-plat-chip {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 9px;
  font-size: 10px;
  margin: 1px 2px;
  line-height: 16px;
}
.pm-empty { font-size: 11px; color: #6a7a90; }
.cur-mark { font-size: 10px; color: #4da3ff; margin-left: 6px; }

.pm-tag { font-size: 11px; padding: 1px 8px; border-radius: 9px; font-weight: 500; }
.pm-tag.ok { background: rgba(61, 214, 140, 0.12); color: #3dd68c; border: 1px solid rgba(61, 214, 140, 0.25); }
.pm-tag.warn { background: rgba(255, 193, 7, 0.12); color: #ffc107; border: 1px solid rgba(255, 193, 7, 0.25); }

/* 表单提示 */
.pm-form-tip { font-size: 11px; color: #6a7a90; margin-top: 4px; }

/* 内层弹窗暗色主题 */
:deep(.el-dialog) {
  --el-dialog-bg-color: #182434;
  --el-dialog-title-font-size: 15px;
  border: 1px solid #253650;
  border-radius: 12px;
}
:deep(.el-form-item__label) { color: #b8c4d8; }
</style>
