<template>
  <main class="page customer-page" @touchstart="onTouchStart" @touchend="onTouchEnd">
    <section class="toolbar">
      <div class="filters">
        <el-input
          v-model="query.keyword"
          clearable
          placeholder="客户编码、名称、电话、税号、地址"
          :prefix-icon="Search"
          @keyup.enter="refresh"
          @clear="refresh"
        />
        <el-button type="primary" :icon="Search" @click="refresh">查询</el-button>
      </div>
      <div class="actions">
        <el-upload :show-file-list="false" :before-upload="handleImport">
          <el-button :icon="Upload">导入</el-button>
        </el-upload>
        <el-button :icon="Download" @click="handleExport">导出</el-button>
        <el-button v-if="canCreateCustomer" type="primary" :icon="Plus" @click="openCreate">新增</el-button>
      </div>
    </section>

    <el-table v-if="!isMobile" v-loading="loading" :data="rows" border stripe class="data-table">
      <el-table-column prop="code" label="客户编码" min-width="120" fixed />
      <el-table-column prop="name" label="客户名称" min-width="180" show-overflow-tooltip />
      <el-table-column prop="type" label="类型" width="100">
        <template #default="{ row }">{{ row.type === 'person' ? '个人' : '企业' }}</template>
      </el-table-column>
      <el-table-column prop="level" label="等级" width="100" />
      <el-table-column prop="phone" label="电话" min-width="130" />
      <el-table-column prop="address" label="地址" min-width="220" show-overflow-tooltip />
      <el-table-column prop="paymentMethod" label="结算方式" width="120">
        <template #default="{ row }">{{ paymentMethodText(row.paymentMethod) }}</template>
      </el-table-column>
      <el-table-column prop="creditLimit" label="信用额度" width="120" align="right" />
      <el-table-column prop="creditDays" label="信用天数" width="100" align="right" />
      <el-table-column prop="receivableBalance" label="应收余额" width="120" align="right" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : 'info'">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button v-if="authStore.hasPermission('customer.edit')" link type="primary" :icon="Edit" @click="openEdit(row)">编辑</el-button>
          <el-button link type="primary" :icon="Money" @click="openDebt(row)">欠款</el-button>
          <el-button link type="primary" :icon="Tickets" @click="openOrders(row)">订单</el-button>
          <el-popconfirm v-if="authStore.hasPermission('customer.delete')" title="确定删除该客户？" @confirm="remove(row)">
            <template #reference>
              <el-button link type="danger" :icon="Delete">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <section v-else class="mobile-card-list">
      <el-skeleton v-if="loading" :rows="4" animated />
      <article v-for="row in rows" :key="row.id" class="customer-card">
        <div class="card-head">
          <div>
            <strong>{{ row.name }}</strong>
            <span>{{ row.code }}</span>
          </div>
          <el-tag :type="row.status === 'active' ? 'success' : 'info'">{{ statusLabel(row.status) }}</el-tag>
        </div>
        <div class="card-meta">
          <span>{{ row.phone || '无电话' }}</span>
          <span>应收 {{ row.receivableBalance || '0.00' }}</span>
        </div>
        <p>{{ row.address || '暂无地址' }}</p>
        <div class="card-actions">
          <el-button v-if="authStore.hasPermission('customer.edit')" size="large" :icon="Edit" @click="openEdit(row)">编辑</el-button>
          <el-button size="large" :icon="Money" @click="openDebt(row)">欠款</el-button>
          <el-button size="large" :icon="Tickets" @click="openOrders(row)">订单</el-button>
        </div>
      </article>
      <el-empty v-if="!loading && rows.length === 0" description="暂无客户" />
      <el-button v-if="rows.length < total" class="load-more" :loading="loadingMore" @click="loadMore">加载更多</el-button>
    </section>

    <footer v-if="!isMobile" class="pager">
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="load"
        @current-change="load"
      />
    </footer>

    <el-dialog v-model="dialog.visible" :fullscreen="isMobile" :title="dialog.editing ? '编辑客户' : '新增客户'" width="720px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="96px">
        <div class="form-grid">
          <el-form-item label="客户编码" prop="code"><el-input v-model="form.code" size="large" /></el-form-item>
          <el-form-item label="客户名称" prop="name"><el-input v-model="form.name" size="large" /></el-form-item>
          <el-form-item label="类型">
            <el-select v-model="form.type" size="large">
              <el-option label="企业" value="company" />
              <el-option label="个人" value="person" />
            </el-select>
          </el-form-item>
          <el-form-item label="等级"><el-input v-model="form.level" size="large" /></el-form-item>
          <el-form-item label="电话"><el-input v-model="form.phone" size="large" /></el-form-item>
          <el-form-item label="邮箱"><el-input v-model="form.email" size="large" /></el-form-item>
          <el-form-item label="税号"><el-input v-model="form.taxNo" size="large" /></el-form-item>
          <el-form-item label="结算方式">
            <el-select v-model="form.paymentMethod" size="large" class="full">
              <el-option label="立即付款" value="immediate" />
              <el-option label="现金结算" value="cash" />
              <el-option label="月结" value="monthly" />
              <el-option label="季结" value="quarterly" />
              <el-option label="半年结" value="half_year" />
              <el-option label="项目验收结算" value="project_acceptance" />
              <el-option label="自定义账期" value="custom" />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="form.status" size="large">
              <el-option label="启用" value="active" />
              <el-option label="停用" value="disabled" />
            </el-select>
          </el-form-item>
          <el-form-item label="信用额度"><el-input-number v-model="form.creditLimit" :min="0" :precision="2" size="large" class="full" /></el-form-item>
          <el-form-item label="信用天数"><el-input-number v-model="form.creditDays" :min="0" :max="365" size="large" class="full" /></el-form-item>
          <el-form-item label="账期">
            <el-select v-model="form.billingCycle" size="large" class="full">
              <el-option label="无" value="none" />
              <el-option label="月结" value="monthly" />
              <el-option label="季结" value="quarterly" />
              <el-option label="半年结" value="half_year" />
              <el-option label="项目验收" value="project_acceptance" />
              <el-option label="自定义" value="custom" />
            </el-select>
          </el-form-item>
          <el-form-item label="付款日"><el-input-number v-model="form.paymentDay" :min="0" :max="28" size="large" class="full" /></el-form-item>
          <el-form-item label="应收余额"><el-input-number v-model="form.receivableBalance" :min="0" :precision="2" size="large" class="full" /></el-form-item>
        </div>
        <el-form-item label="地址"><el-input v-model="form.address" size="large" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button size="large" @click="dialog.visible = false">取消</el-button>
        <el-button size="large" type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="debt.visible" :fullscreen="isMobile" title="客户欠款" width="420px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="客户">{{ debt.data.customerName }}</el-descriptions-item>
        <el-descriptions-item label="欠款">{{ debt.data.receivableBalance }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <el-drawer v-model="orders.visible" title="客户历史订单" :size="isMobile ? '100%' : '720px'">
      <el-table :data="orders.rows" border>
        <el-table-column prop="orderNo" label="订单号" min-width="140" />
        <el-table-column prop="orderDate" label="日期" width="120" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">{{ statusLabel(row.status) }}</template>
        </el-table-column>
        <el-table-column prop="totalAmount" label="金额" width="110" align="right" />
        <el-table-column prop="receivableAmount" label="应收" width="110" align="right" />
      </el-table>
      <div class="pager">
        <el-pagination
          v-model:current-page="orders.query.page"
          v-model:page-size="orders.query.pageSize"
          :total="orders.total"
          layout="total, prev, pager, next"
          @current-change="loadOrders"
        />
      </div>
    </el-drawer>

    <button v-if="isMobile && canCreateCustomer" class="fab" @click="openCreate">
      <el-icon><Plus /></el-icon>
    </button>
  </main>
</template>

<script setup>
import { computed, onActivated, onMounted, onUnmounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Delete, Download, Edit, Money, Plus, Search, Tickets, Upload } from '@element-plus/icons-vue'
import {
  createCustomer,
  deleteCustomer,
  exportCustomers,
  getCustomerDebt,
  importCustomers,
  listCustomerOrders,
  listCustomers,
  updateCustomer
} from '../../api/customer'
import { getBreakpoint, vibrate } from '../../utils/mobile'
import { useAuthStore } from '../../stores/auth'
import { pinia } from '../../stores/pinia'
import { useSearch } from '../../composables/useSearch'
import { formatDateRows } from '../../utils/date'
import { statusLabel } from '../../utils/status'

const loading = ref(false)
const loadingMore = ref(false)
const saving = ref(false)
const rows = ref([])
const total = ref(0)
const formRef = ref()
const breakpoint = ref(getBreakpoint())
const authStore = useAuthStore(pinia)
const touchStartY = ref(0)
const { searchForm: query, resetSearch } = useSearch({ page: 1, pageSize: 20, keyword: '', sortBy: 'id', order: 'desc' })
const dialog = reactive({ visible: false, editing: false, id: null })
const form = reactive(defaultForm())
const debt = reactive({ visible: false, data: {} })
const orders = reactive({ visible: false, customerId: null, rows: [], total: 0, query: { page: 1, pageSize: 10 } })

const isMobile = computed(() => breakpoint.value === 'mobile')
const canCreateCustomer = computed(() => authStore.isLoggedIn || authStore.hasPermission('customer.create') || authStore.hasPermission('customer.manage'))
const rules = {
  code: [{ required: true, message: '请输入客户编码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入客户名称', trigger: 'blur' }]
}

function defaultForm() {
  return {
    code: '',
    name: '',
    type: 'company',
    level: '',
    phone: '',
    email: '',
    taxNo: '',
    address: '',
    paymentMethod: 'immediate',
    creditLimit: 0,
    creditDays: 0,
    billingCycle: 'none',
    paymentDay: 0,
    receivableBalance: 0,
    status: 'active',
    remark: ''
  }
}

function paymentMethodText(value) {
  return {
    immediate: '立即付款',
    cash: '现金结算',
    monthly: '月结',
    quarterly: '季结',
    half_year: '半年结',
    project_acceptance: '项目验收',
    custom: '自定义账期',
    none: '无'
  }[value] || value || '立即付款'
}

function assignForm(data) {
  Object.assign(form, defaultForm(), data)
}

async function load() {
  loading.value = true
  try {
    const data = await listCustomers(query)
    rows.value = formatDateRows(data.list || [])
    total.value = data.total
  } finally {
    loading.value = false
  }
}

async function refresh() {
  query.page = 1
  await load()
}

async function loadMore() {
  if (rows.value.length >= total.value) return
  loadingMore.value = true
  try {
    query.page += 1
    const data = await listCustomers(query)
    rows.value = rows.value.concat(formatDateRows(data.list || []))
    total.value = data.total
  } finally {
    loadingMore.value = false
  }
}

function openCreate() {
  vibrate()
  dialog.visible = true
  dialog.editing = false
  dialog.id = null
  assignForm({})
}

function openEdit(row) {
  dialog.visible = true
  dialog.editing = true
  dialog.id = row.id
  assignForm(row)
}

async function save() {
  await formRef.value.validate()
  saving.value = true
  try {
    if (dialog.editing) await updateCustomer(dialog.id, form)
    else await createCustomer(form)
    ElMessage.success('保存成功')
    dialog.visible = false
    await refresh()
  } finally {
    saving.value = false
  }
}

async function remove(row) {
  await deleteCustomer(row.id)
  ElMessage.success('删除成功')
  await refresh()
}

async function handleImport(file) {
  const result = await importCustomers(file)
  ElMessage.success(`导入完成：成功 ${result.success} 条，失败 ${result.failed} 条`)
  await refresh()
  return false
}

async function handleExport() {
  const response = await exportCustomers({ keyword: query.keyword })
  const blob = new Blob([response.data])
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `customers_${Date.now()}.xlsx`
  link.click()
  URL.revokeObjectURL(link.href)
}

async function openDebt(row) {
  debt.data = await getCustomerDebt(row.id)
  debt.visible = true
}

async function openOrders(row) {
  orders.customerId = row.id
  orders.query.page = 1
  orders.visible = true
  await loadOrders()
}

async function loadOrders() {
  const data = await listCustomerOrders(orders.customerId, orders.query)
  orders.rows = formatDateRows(data.list || [])
  orders.total = data.total
}

function onTouchStart(event) {
  touchStartY.value = event.touches[0].clientY
}

async function onTouchEnd(event) {
  if (!isMobile.value) return
  const endY = event.changedTouches[0].clientY
  if (window.scrollY === 0 && endY - touchStartY.value > 80) {
    ElMessage.info('正在刷新')
    await refresh()
  }
}

function handleResize() {
  breakpoint.value = getBreakpoint()
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
  resetSearch()
  load()
})
onActivated(() => {
  resetSearch()
  load()
})
onUnmounted(() => window.removeEventListener('resize', handleResize))
</script>
