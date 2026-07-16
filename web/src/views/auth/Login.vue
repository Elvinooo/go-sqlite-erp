<template>
  <main class="login-page">
    <section class="login-panel">
      <div class="login-brand">
        <div class="brand-mark">E</div>
        <div>
          <strong>ERP Pro</strong>
          <span>企业经营管理系统</span>
        </div>
      </div>

      <el-form ref="formRef" :model="form" :rules="rules" size="large" @submit.prevent="submit">
        <el-form-item prop="username">
          <el-input v-model.trim="form.username" placeholder="用户名" :prefix-icon="User" @keyup.enter="submit" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            placeholder="密码"
            :prefix-icon="Lock"
            show-password
            type="password"
            @keyup.enter="submit"
          />
        </el-form-item>
        <el-button class="login-button" type="primary" size="large" :loading="loading" @click="submit">
          登录
        </el-button>
      </el-form>
    </section>
  </main>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Lock, User } from '@element-plus/icons-vue'
import { login } from '../../api/auth'
import { useAuthStore } from '../../stores/auth'
import { pinia } from '../../stores/pinia'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore(pinia)
const formRef = ref()
const loading = ref(false)
const form = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

async function submit() {
  await formRef.value?.validate()
  loading.value = true
  try {
    const data = await login(form)
    localStorage.setItem('accessToken', data.accessToken)
    localStorage.setItem('refreshToken', data.refreshToken)
    await authStore.loadCurrentUser()
    ElMessage.success(data.mustChangePassword ? '请先修改密码' : '登录成功')
    router.replace(data.mustChangePassword ? '/profile' : (route.query.redirect || '/dashboard'))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
  background:
    linear-gradient(135deg, rgba(33, 150, 243, 0.12), rgba(0, 150, 136, 0.1)),
    var(--bg);
}

.login-panel {
  width: min(420px, 100%);
  padding: 32px;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--panel);
  box-shadow: var(--shadow);
}

.login-brand {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 28px;
}

.brand-mark {
  width: 42px;
  height: 42px;
  display: grid;
  place-items: center;
  border-radius: 8px;
  background: var(--primary);
  color: #fff;
  font-weight: 800;
}

.login-brand strong,
.login-brand span {
  display: block;
}

.login-brand strong {
  font-size: 22px;
}

.login-brand span {
  color: var(--muted);
}

.login-button {
  width: 100%;
}
</style>
