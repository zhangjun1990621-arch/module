<template>
  <div class="permission-management">
    <!-- 页面标题与操作栏 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">权限管理</h2>
        <p class="page-desc">管理用户账号、角色与平台访问权限</p>
      </div>
      <div class="header-right">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索用户名"
          clearable
          style="width: 200px"
          @keyup.enter="loadUsers"
          @clear="loadUsers"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-select v-model="filterRole" placeholder="角色筛选" clearable style="width: 140px" @change="loadUsers">
          <el-option label="超级管理员" value="super_admin" />
          <el-option label="管理员" value="admin" />
          <el-option label="只读用户" value="viewer" />
        </el-select>
        <el-button type="primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          新增用户
        </el-button>
      </div>
    </div>

    <!-- 用户列表 -->
    <div class="al-card">
      <el-table :data="users" v-loading="loading" stripe style="width: 100%">
        <el-table-column type="index" label="#" width="50" align="center" />
        <el-table-column prop="username" label="用户名" min-width="140">
          <template #default="{ row }">
            <div class="user-cell">
              <div class="user-avatar">{{ row.username?.[0]?.toUpperCase() || 'U' }}</div>
              <span>{{ row.username }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="role" label="角色" width="130" align="center">
          <template #default="{ row }">
            <el-tag :type="getRoleTagType(row.role)" effect="dark" size="small">
              {{ getRoleLabel(row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="platforms" label="可访问平台" min-width="200">
          <template #default="{ row }">
            <div v-if="row.role === 'super_admin'" class="platform-all">
              <el-tag type="danger" effect="dark" size="small">全部平台</el-tag>
            </div>
            <div v-else-if="row.platforms" class="platform-tags">
              <el-tag
                v-for="pid in row.platforms.split(',').filter(Boolean)"
                :key="pid"
                size="small"
                effect="plain"
                style="margin-right: 4px; margin-bottom: 4px"
              >
                {{ getPlatformName(pid) }}
              </el-tag>
            </div>
            <span v-else class="text-muted">未分配</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" effect="dark" size="small">
              {{ row.status === 'active' ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="lastLogin" label="最后登录" width="170" align="center">
          <template #default="{ row }">
            <span v-if="row.lastLogin" class="text-muted">{{ formatTime(row.lastLogin) }}</span>
            <span v-else class="text-muted">从未登录</span>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="170" align="center">
          <template #default="{ row }">
            <span class="text-muted">{{ formatTime(row.createdAt) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" align="center" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" size="small" @click="openEditDialog(row)" :disabled="row.id === currentUserId">
              编辑
            </el-button>
            <el-button text type="warning" size="small" @click="openPasswordDialog(row)">
              重置密码
            </el-button>
            <el-popconfirm
              title="确定删除该用户吗？"
              confirm-button-text="删除"
              cancel-button-text="取消"
              @confirm="handleDelete(row)"
            >
              <template #reference>
                <el-button text type="danger" size="small" :disabled="row.id === currentUserId">
                  删除
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-bar">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          background
          @size-change="loadUsers"
          @current-change="loadUsers"
        />
      </div>
    </div>

    <!-- 新增用户对话框 -->
    <el-dialog v-model="createDialogVisible" title="新增用户" width="520px" :close-on-click-modal="false">
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="100px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="createForm.username" placeholder="请输入用户名（2-64字符）" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="createForm.password" type="password" show-password placeholder="请输入密码（至少6位）" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="createForm.role" placeholder="请选择角色" style="width: 100%" @change="onRoleChange">
            <el-option label="超级管理员" value="super_admin" />
            <el-option label="管理员" value="admin" />
            <el-option label="只读用户" value="viewer" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="createForm.role !== 'super_admin'" label="可访问平台" prop="platforms">
          <el-checkbox-group v-model="createPlatformChecks">
            <el-checkbox
              v-for="p in allPlatforms"
              :key="p.id"
              :value="p.id"
            >
              {{ p.icon }} {{ p.name }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="createForm.status">
            <el-radio value="active">正常</el-radio>
            <el-radio value="inactive">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleCreate">确认创建</el-button>
      </template>
    </el-dialog>

    <!-- 编辑用户对话框 -->
    <el-dialog v-model="editDialogVisible" title="编辑用户" width="520px" :close-on-click-modal="false">
      <el-form ref="editFormRef" :model="editForm" label-width="100px">
        <el-form-item label="用户名">
          <el-input :value="editForm.username" disabled />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="editForm.role" placeholder="请选择角色" style="width: 100%" @change="onEditRoleChange">
            <el-option label="超级管理员" value="super_admin" />
            <el-option label="管理员" value="admin" />
            <el-option label="只读用户" value="viewer" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="editForm.role !== 'super_admin'" label="可访问平台">
          <el-checkbox-group v-model="editPlatformChecks">
            <el-checkbox
              v-for="p in allPlatforms"
              :key="p.id"
              :value="p.id"
            >
              {{ p.icon }} {{ p.name }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="editForm.status">
            <el-radio value="active">正常</el-radio>
            <el-radio value="inactive">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleEdit">保存修改</el-button>
      </template>
    </el-dialog>

    <!-- 重置密码对话框 -->
    <el-dialog v-model="passwordDialogVisible" title="重置密码" width="420px" :close-on-click-modal="false">
      <el-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules" label-width="80px">
        <el-form-item label="用户名">
          <el-input :value="passwordForm.username" disabled />
        </el-form-item>
        <el-form-item label="新密码" prop="password">
          <el-input v-model="passwordForm.password" type="password" show-password placeholder="请输入新密码（至少6位）" />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input v-model="passwordForm.confirmPassword" type="password" show-password placeholder="请再次输入新密码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleResetPassword">确认重置</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import request from '@/api/request'
import { useAuthStore } from '@/stores/auth'
import { usePlatformStore } from '@/stores/platform'

const authStore = useAuthStore()
const platformStore = usePlatformStore()

/** 当前登录用户 ID（不允许操作自己） */
const currentUserId = computed(() => authStore.user?.id || '')

/** 所有平台列表（用于分配权限） */
const allPlatforms = computed(() => platformStore.platforms)

// ===== 列表相关 =====
const loading = ref(false)
const users = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const searchKeyword = ref('')
const filterRole = ref('')

/** 加载用户列表 */
async function loadUsers() {
  loading.value = true
  try {
    const params: any = {
      page: page.value,
      pageSize: pageSize.value
    }
    if (searchKeyword.value) params.search = searchKeyword.value
    if (filterRole.value) params.role = filterRole.value

    const res = await request.get('/users', { params })
    users.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch {
    // 拦截器已处理错误提示
  } finally {
    loading.value = false
  }
}

// ===== 新增用户 =====
const createDialogVisible = ref(false)
const createFormRef = ref<FormInstance>()
const submitting = ref(false)
const createPlatformChecks = ref<string[]>([])

const createForm = reactive({
  username: '',
  password: '',
  role: 'viewer',
  status: 'active'
})

const createRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 2, max: 64, message: '长度 2-64 字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 128, message: '密码至少 6 位', trigger: 'blur' }
  ],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }]
}

function openCreateDialog() {
  createForm.username = ''
  createForm.password = ''
  createForm.role = 'viewer'
  createForm.status = 'active'
  createPlatformChecks.value = []
  createDialogVisible.value = true
}

function onRoleChange() {
  if (createForm.role === 'super_admin') {
    createPlatformChecks.value = []
  }
}

async function handleCreate() {
  if (!createFormRef.value) return
  await createFormRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      const payload: any = {
        username: createForm.username,
        password: createForm.password,
        role: createForm.role,
        status: createForm.status,
        platforms: createForm.role === 'super_admin' ? '' : createPlatformChecks.value.join(',')
      }
      await request.post('/users', payload)
      ElMessage.success('用户创建成功')
      createDialogVisible.value = false
      loadUsers()
    } catch {
      // 拦截器已处理错误提示
    } finally {
      submitting.value = false
    }
  })
}

// ===== 编辑用户 =====
const editDialogVisible = ref(false)
const editFormRef = ref<FormInstance>()
const editPlatformChecks = ref<string[]>([])
const editingUserId = ref('')

const editForm = reactive({
  username: '',
  role: 'viewer',
  status: 'active'
})

function openEditDialog(row: any) {
  editingUserId.value = row.id
  editForm.username = row.username
  editForm.role = row.role
  editForm.status = row.status
  editPlatformChecks.value = row.platforms ? row.platforms.split(',').filter(Boolean) : []
  editDialogVisible.value = true
}

function onEditRoleChange() {
  if (editForm.role === 'super_admin') {
    editPlatformChecks.value = []
  }
}

async function handleEdit() {
  submitting.value = true
  try {
    const payload: any = {
      role: editForm.role,
      status: editForm.status,
      platforms: editForm.role === 'super_admin' ? '' : editPlatformChecks.value.join(',')
    }
    await request.put(`/users/${editingUserId.value}`, payload)
    ElMessage.success('用户信息已更新')
    editDialogVisible.value = false
    loadUsers()
  } catch {
    // 拦截器已处理错误提示
  } finally {
    submitting.value = false
  }
}

// ===== 重置密码 =====
const passwordDialogVisible = ref(false)
const passwordFormRef = ref<FormInstance>()
const passwordUserId = ref('')

const passwordForm = reactive({
  username: '',
  password: '',
  confirmPassword: ''
})

const passwordRules: FormRules = {
  password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 128, message: '密码至少 6 位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (value !== passwordForm.password) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

function openPasswordDialog(row: any) {
  passwordUserId.value = row.id
  passwordForm.username = row.username
  passwordForm.password = ''
  passwordForm.confirmPassword = ''
  passwordDialogVisible.value = true
}

async function handleResetPassword() {
  if (!passwordFormRef.value) return
  await passwordFormRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      await request.put(`/users/${passwordUserId.value}/password`, {
        password: passwordForm.password
      })
      ElMessage.success('密码已重置')
      passwordDialogVisible.value = false
    } catch {
      // 拦截器已处理错误提示
    } finally {
      submitting.value = false
    }
  })
}

// ===== 删除用户 =====
async function handleDelete(row: any) {
  try {
    await request.delete(`/users/${row.id}`)
    ElMessage.success('用户已删除')
    loadUsers()
  } catch {
    // 拦截器已处理错误提示
  }
}

// ===== 工具函数 =====
function getRoleLabel(role: string): string {
  const map: Record<string, string> = {
    super_admin: '超级管理员',
    admin: '管理员',
    viewer: '只读用户'
  }
  return map[role] || role
}

function getRoleTagType(role: string): 'danger' | 'warning' | 'info' {
  if (role === 'super_admin') return 'danger'
  if (role === 'admin') return 'warning'
  return 'info'
}

function getPlatformName(pid: string): string {
  const p = allPlatforms.value.find((item) => item.id === pid)
  return p ? p.name : pid
}

function formatTime(ts: any): string {
  if (!ts) return '-'
  const d = new Date(ts)
  if (isNaN(d.getTime())) return String(ts)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// ===== 初始化 =====
onMounted(() => {
  loadUsers()
})
</script>

<style scoped lang="scss">
.permission-management {
  .page-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 20px;

    .header-left {
      .page-title {
        font-size: 22px;
        font-weight: 700;
        color: #e6edf3;
        margin: 0 0 4px;
      }

      .page-desc {
        font-size: 13px;
        color: #8b949e;
        margin: 0;
      }
    }

    .header-right {
      display: flex;
      gap: 12px;
      align-items: center;
    }
  }
}

.al-card {
  background: #161b22;
  border: 1px solid #30363d;
  border-radius: 12px;
  overflow: hidden;
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 10px;

  .user-avatar {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: linear-gradient(135deg, #4b3fe3, #7c6ff5);
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
    font-weight: 700;
    flex-shrink: 0;
  }
}

.platform-tags {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}

.text-muted {
  color: #6e7681;
  font-size: 12px;
}

.pagination-bar {
  padding: 16px 20px;
  display: flex;
  justify-content: flex-end;
}

/* 深色主题表格适配 */
:deep(.el-table) {
  background: transparent;

  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-header-bg-color: #21262d;
  --el-table-border-color: #30363d;
  --el-table-header-text-color: #8b949e;
  --el-table-text-color: #c9d1d9;
  --el-table-row-hover-bg-color: rgba(75, 63, 227, 0.06);

  th.el-table__cell {
    background: #21262d !important;
    border-bottom: 1px solid #30363d !important;
  }

  .el-table__row:hover > td {
    background: rgba(75, 63, 227, 0.06) !important;
  }
}

:deep(.el-dialog) {
  background: #161b22;
  border: 1px solid #30363d;

  .el-dialog__header {
    border-bottom: 1px solid #30363d;

    .el-dialog__title {
      color: #e6edf3;
    }
  }

  .el-dialog__body {
    .el-form-item__label {
      color: #c9d1d9;
    }

    .el-input__wrapper,
    .el-select__wrapper {
      background: #0d1117;
      box-shadow: 0 0 0 1px #30363d inset;

      .el-input__inner {
        color: #e6edf3;
      }
    }

    .el-radio__label {
      color: #c9d1d9;
    }

    .el-checkbox__label {
      color: #c9d1d9;
    }
  }
}

:deep(.el-pagination) {
  --el-pagination-bg-color: #21262d;
  --el-pagination-text-color: #c9d1d9;
  --el-pagination-button-color: #c9d1d9;
  --el-pagination-hover-color: #7c6ff5;
}
</style>
