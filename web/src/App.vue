<template>
  <router-view v-if="route.path === '/login'" />
  <div v-else class="app-shell" :class="`is-${breakpoint}`">
    <aside v-if="breakpoint !== 'mobile'" class="side-nav">
      <div class="brand">
        <div class="brand-mark">E</div>
        <div>
          <strong>ERP Pro</strong>
          <span>设备 · 工程 · 维修</span>
        </div>
      </div>
      <div class="side-menu-scroll">
        <el-menu :default-active="$route.path" :default-openeds="defaultOpeneds" router>
          <template v-for="item in menuItems" :key="item.path">
            <el-sub-menu v-if="item.children?.length" :index="item.path">
              <template #title>
                <el-icon><component :is="item.icon" /></el-icon>
                <span>{{ item.label }}</span>
              </template>
              <el-menu-item v-for="child in item.children" :key="child.path" :index="child.path">
                <el-icon><component :is="child.icon" /></el-icon>
                <span>{{ child.label }}</span>
              </el-menu-item>
            </el-sub-menu>
            <el-menu-item v-else :index="item.path">
              <el-icon><component :is="item.icon" /></el-icon>
              <span>{{ item.label }}</span>
            </el-menu-item>
          </template>
        </el-menu>
      </div>
    </aside>

    <section class="main-frame">
      <header class="top-bar">
        <div>
          <h1>{{ routeTitle }}</h1>
          <p>{{ breakpointText }}</p>
        </div>
        <div class="top-actions">
          <el-button circle :icon="appStore.theme === 'dark' ? Sunny : Moon" @click="appStore.toggleTheme()" />
          <el-dropdown trigger="click" @command="handleUserCommand">
            <el-button class="user-menu-button">
              <el-icon><User /></el-icon>
              <span>{{ authStore.user?.realName || authStore.user?.username || '用户' }}</span>
              <el-icon><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="password" :icon="Key">修改密码</el-dropdown-item>
                <el-dropdown-item command="logout" :icon="SwitchButton">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-button v-if="breakpoint === 'mobile'" circle :icon="Menu" @click="drawer = true" />
        </div>
      </header>

      <router-view />
    </section>

    <nav v-if="breakpoint === 'mobile'" class="bottom-tabbar">
      <router-link v-for="item in mobileTabs" :key="item.path" :to="item.path" class="tab-item">
        <el-icon><component :is="item.icon" /></el-icon>
        <span>{{ item.label }}</span>
      </router-link>
    </nav>

    <el-drawer v-model="drawer" direction="rtl" size="78%" title="全部功能">
      <div class="drawer-menu">
        <template v-for="item in menuItems" :key="item.path">
          <div v-if="item.children?.length" class="drawer-menu-group">
            <div class="drawer-menu-title">
              <el-icon><component :is="item.icon" /></el-icon>
              <span>{{ item.label }}</span>
            </div>
            <router-link v-for="child in item.children" :key="child.path" :to="child.path" @click="drawer = false">
              <el-icon><component :is="child.icon" /></el-icon>
              <span>{{ child.label }}</span>
            </router-link>
          </div>
          <router-link v-else :to="item.path" @click="drawer = false">
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </router-link>
        </template>
      </div>
    </el-drawer>

    <el-dialog
      v-model="passwordDialog.visible"
      title="修改密码"
      width="420px"
      :close-on-click-modal="!authStore.user?.mustChangePassword"
      :close-on-press-escape="!authStore.user?.mustChangePassword"
      :show-close="!authStore.user?.mustChangePassword"
    >
      <el-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules" label-width="96px">
        <el-form-item label="旧密码" prop="oldPassword">
          <el-input v-model="passwordForm.oldPassword" type="password" show-password autocomplete="current-password" />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input v-model="passwordForm.newPassword" type="password" show-password autocomplete="new-password" />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input v-model="passwordForm.confirmPassword" type="password" show-password autocomplete="new-password" @keyup.enter="submitPassword" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button v-if="!authStore.user?.mustChangePassword" @click="passwordDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="passwordDialog.saving" @click="submitPassword">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  ArrowDown,
  Box,
  Briefcase,
  Document,
  Goods,
  HomeFilled,
  Key,
  Lock,
  Menu,
  Money,
  Moon,
  OfficeBuilding,
  Sell,
  Setting,
  ShoppingCart,
  Sunny,
  SwitchButton,
  Tools,
  User
} from '@element-plus/icons-vue'
import { changePassword, logout } from './api/auth'
import { useAppStore } from './stores/app'
import { useAuthStore } from './stores/auth'
import { pinia } from './stores/pinia'
import { getBreakpoint } from './utils/mobile'

const route = useRoute()
const router = useRouter()
const appStore = useAppStore(pinia)
const authStore = useAuthStore(pinia)
const breakpoint = ref(getBreakpoint())
const drawer = ref(false)
const passwordFormRef = ref()
const passwordDialog = reactive({ visible: false, saving: false })
const passwordForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })

const allMenuItems = [
  { path: '/dashboard', label: '首页', icon: HomeFilled, permission: 'dashboard.boss.view' },
  { path: '/sales', label: '销售', icon: Sell, permission: 'sales.manage' },
  { path: '/purchase', label: '采购', icon: ShoppingCart, permission: 'purchase.manage' },
  { path: '/inventory', label: '库存查询', icon: Box, permission: 'inventory.manage' },
  { path: '/repair', label: '维修', icon: Tools, permission: 'repair.manage' },
  { path: '/project', label: '工程', icon: Briefcase, permission: 'project.manage' },
  { path: '/customers', label: '客户', icon: OfficeBuilding, permission: 'customer.manage' },
  { path: '/suppliers', label: '供应商', icon: OfficeBuilding, permission: 'supplier.manage' },
  { path: '/products', label: '商品', icon: Goods, permission: 'product.manage' },
  {
    path: '/finance-group',
    label: '财务',
    icon: Money,
    permission: 'finance.manage',
    children: [
      { path: '/finance', label: '资金流水', icon: Money, permission: 'finance.manage' },
      { path: '/finance-accounts', label: '资金账户', icon: Money, permission: 'finance.manage' },
      { path: '/receivables', label: '应收账款', icon: Money, permission: 'finance.manage' },
      { path: '/customer-statements', label: '客户对账单', icon: Money, permission: 'finance.manage' },
      { path: '/payables', label: '应付账款', icon: Money, permission: 'finance.manage' },
      { path: '/profit-report', label: '利润报表', icon: Money, permission: 'finance.manage' },
      { path: '/inventory-asset-report', label: '库存资产报表', icon: Money, permission: 'finance.manage' }
    ]
  },
  { path: '/permissions', label: '权限', icon: Lock, permission: 'auth.user.manage' },
  {
    path: '/logs-group',
    label: '日志',
    icon: Document,
    permission: 'system.audit.view',
    children: [
      { path: '/audit-logs', label: '操作日志', icon: Document, permission: 'system.audit.view' },
      { path: '/document-delete-records', label: '单据删除记录', icon: Document, permission: 'system.audit.view' }
    ]
  },
  { path: '/profile', label: '我的', icon: User },
  { path: '/settings', label: '设置', icon: Setting, permission: 'system.setting.manage' }
]

const menuItems = computed(() => {
  if (!authStore.isLoggedIn) return allMenuItems
  return filterMenuItems(allMenuItems)
})

const flatMenuItems = computed(() => flattenMenuItems(menuItems.value))
const allFlatMenuItems = computed(() => flattenMenuItems(allMenuItems))
const mobileTabs = computed(() => flatMenuItems.value.filter((item) => ['/dashboard', '/sales', '/inventory', '/customers', '/profile'].includes(item.path)).slice(0, 5))
const defaultOpeneds = computed(() => menuItems.value.filter((item) => item.children?.some((child) => child.path === route.path)).map((item) => item.path))

const routeTitle = computed(() => allFlatMenuItems.value.find((item) => item.path === route.path)?.label || 'ERP')
const breakpointText = computed(() => {
  if (breakpoint.value === 'pc') return 'PC后台模式'
  if (breakpoint.value === 'pad') return '平板自适应模式'
  return '移动端触摸模式'
})

const passwordRules = {
  oldPassword: [{ required: true, message: '请输入旧密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 8, message: '新密码不少于8位', trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (value && !/^(?=.*[A-Za-z])(?=.*\d).+$/.test(value)) {
          callback(new Error('新密码必须包含字母和数字'))
          return
        }
        if (value && value === passwordForm.oldPassword) {
          callback(new Error('新密码不能与旧密码一致'))
          return
        }
        callback()
      },
      trigger: 'blur'
    }
  ],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (value !== passwordForm.newPassword) {
          callback(new Error('两次密码必须一致'))
          return
        }
        callback()
      },
      trigger: 'blur'
    }
  ]
}

function handleResize() {
  breakpoint.value = getBreakpoint()
}

function filterMenuItems(items) {
  return items
    .map((item) => {
      const children = item.children ? filterMenuItems(item.children) : []
      const visible = authStore.hasPermission(item.permission) || children.length > 0
      if (!visible) return null
      return children.length > 0 ? { ...item, children } : { ...item, children: undefined }
    })
    .filter(Boolean)
}

function flattenMenuItems(items) {
  return items.flatMap((item) => item.children?.length ? flattenMenuItems(item.children) : [item])
}

async function handleLogout() {
  try {
    await logout()
  } catch {
    // Local token cleanup still allows the user to leave the session.
  }
  authStore.clear()
  ElMessage.success('已退出登录')
  router.replace('/login')
}

function handleUserCommand(command) {
  if (command === 'password') {
    openPasswordDialog()
    return
  }
  if (command === 'logout') {
    handleLogout()
  }
}

function openPasswordDialog() {
  passwordForm.oldPassword = ''
  passwordForm.newPassword = ''
  passwordForm.confirmPassword = ''
  passwordFormRef.value?.clearValidate()
  passwordDialog.visible = true
}

async function submitPassword() {
  await passwordFormRef.value?.validate()
  passwordDialog.saving = true
  try {
    await changePassword({
      oldPassword: passwordForm.oldPassword,
      newPassword: passwordForm.newPassword,
      confirmPassword: passwordForm.confirmPassword
    })
    ElMessage.success('密码修改成功，请重新登录')
    passwordDialog.visible = false
    await logout().catch(() => {})
    authStore.clear()
    router.replace('/login')
  } finally {
    passwordDialog.saving = false
  }
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
  authStore.loadCurrentUser()
    .then(() => {
      if (authStore.user?.mustChangePassword) openPasswordDialog()
    })
    .catch(() => authStore.clear())
})
onUnmounted(() => window.removeEventListener('resize', handleResize))
</script>
