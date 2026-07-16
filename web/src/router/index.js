import { createRouter, createWebHistory } from 'vue-router'
import CustomerIndex from '../views/customer/CustomerIndex.vue'
import Login from '../views/auth/Login.vue'
import SupplierIndex from '../views/supplier/SupplierIndex.vue'
import Dashboard from '../views/mobile/Dashboard.vue'
import FeaturePage from '../views/mobile/FeaturePage.vue'
import PermissionIndex from '../views/system/PermissionIndex.vue'
import AuditLogIndex from '../views/system/AuditLogIndex.vue'

const routes = [
  { path: '/', redirect: '/dashboard' },
  { path: '/login', component: Login },
  { path: '/dashboard', component: Dashboard },
  { path: '/customers', component: CustomerIndex },
  { path: '/suppliers', component: SupplierIndex },
  { path: '/products', component: FeaturePage, props: { type: 'products' } },
  { path: '/sales', component: FeaturePage, props: { type: 'sales' } },
  { path: '/purchase', component: FeaturePage, props: { type: 'purchase' } },
  { path: '/inventory', component: FeaturePage, props: { type: 'inventory' } },
  { path: '/inventory-movements', component: FeaturePage, props: { type: 'inventory-movements' } },
  { path: '/repair', component: FeaturePage, props: { type: 'repair' } },
  { path: '/project', component: FeaturePage, props: { type: 'project' } },
  { path: '/finance', component: FeaturePage, props: { type: 'finance' } },
  { path: '/finance-accounts', component: FeaturePage, props: { type: 'finance-accounts' } },
  { path: '/receivables', component: FeaturePage, props: { type: 'receivables' } },
  { path: '/customer-statements', component: FeaturePage, props: { type: 'customer-statements' } },
  { path: '/payables', component: FeaturePage, props: { type: 'payables' } },
  { path: '/profit-report', component: FeaturePage, props: { type: 'profit-report' } },
  { path: '/inventory-asset-report', component: FeaturePage, props: { type: 'inventory-asset-report' } },
  { path: '/permissions', component: PermissionIndex },
  { path: '/audit-logs', component: AuditLogIndex },
  { path: '/document-delete-records', component: FeaturePage, props: { type: 'document-delete-records' } },
  { path: '/profile', component: FeaturePage, props: { type: 'profile' } },
  { path: '/settings', component: FeaturePage, props: { type: 'settings' } }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to) => {
  const token = localStorage.getItem('accessToken')
  if (to.path === '/login') {
    return token ? '/dashboard' : true
  }
  if (!token) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  return true
})

export default router
