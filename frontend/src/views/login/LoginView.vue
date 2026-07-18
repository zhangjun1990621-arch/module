<template>
  <div class="login-scene">
    <div class="login-card">
      <div class="login-header">
        <div class="icon">&#9889;</div>
        <h1>综合能源云控平台</h1>
        <div class="subtitle">Integrated Energy Cloud Control Platform</div>
      </div>
      <div class="login-body">
        <el-form ref="formRef" :model="form" :rules="rules" @submit.prevent="handleLogin">
          <el-form-item prop="username">
            <el-input
              v-model="form.username"
              placeholder="用户名"
              prefix-icon="User"
              size="large"
            />
          </el-form-item>
          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="密码"
              prefix-icon="Lock"
              size="large"
              show-password
              @keyup.enter="handleLogin"
            />
          </el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            class="login-btn"
            @click="handleLogin"
          >
            登 录
          </el-button>
        </el-form>
      </div>
      <div class="login-footer">
        <span class="warn">登录失败5次将锁定30分钟</span>
        <span class="sep">|</span>
        <span>超时无操作2小时后自动登出</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePlatformStore } from '@/stores/platform'
import { setupDynamicRoutes, resetDynamicRoutes } from '@/router'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()
const platformStore = usePlatformStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

async function handleLogin() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid: boolean) => {
    if (!valid) return
    loading.value = true
    try {
      await authStore.login(form.username, form.password)

      // 登录成功后立即加载平台配置并注册动态路由
      resetDynamicRoutes()
      await platformStore.loadPlatforms()
      setupDynamicRoutes(platformStore.platforms)

      ElMessage.success('登录成功')
      router.push('/dashboard')
    } catch (e: any) {
      ElMessage.error(e?.message || '登录失败')
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped lang="scss">
.login-scene {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: #0d1117;
  background-image: radial-gradient(circle, rgba(75, 63, 227, 0.08) 1px, transparent 1px);
  background-size: 40px 40px;
}

.login-card {
  width: 420px;
  background: #161b22;
  border-radius: 16px;
  border: 1px solid #30363d;
  box-shadow: 0 8px 50px rgba(0, 0, 0, 0.5);
  overflow: hidden;
}

.login-header {
  padding: 32px 36px 22px;
  text-align: center;
  background: linear-gradient(180deg, rgba(75, 63, 227, 0.08), transparent);

  .icon {
    font-size: 38px;
    margin-bottom: 8px;
    filter: drop-shadow(0 0 10px rgba(75, 63, 227, 0.4));
  }

  h1 {
    font-size: 22px;
    font-weight: 700;
    color: #e6edf3;
    letter-spacing: 1px;
  }

  .subtitle {
    font-size: 11px;
    color: #6e7681;
    margin-top: 4px;
    letter-spacing: 2px;
    text-transform: uppercase;
  }
}

.login-body {
  padding: 24px 36px 28px;

  .login-btn {
    width: 100%;
    height: 44px;
    font-size: 15px;
    font-weight: 600;
    margin-top: 8px;
  }
}

.login-footer {
  padding: 14px 36px 18px;
  text-align: center;
  font-size: 11px;
  color: #6e7681;

  .warn {
    color: #f85149;
  }

  .sep {
    margin: 0 6px;
    color: #30363d;
  }
}

:deep(.el-input__wrapper) {
  background: #0d1117;
  border: 1px solid #30363d;
  box-shadow: none;

  &:hover,
  &.is-focus {
    border-color: #4b3fe3;
    box-shadow: 0 0 0 2px rgba(75, 63, 227, 0.1);
  }
}

:deep(.el-input__inner) {
  color: #e6edf3;
}

:deep(.el-input__prefix .el-icon) {
  color: #6e7681;
}
</style>
