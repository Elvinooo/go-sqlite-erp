<template>
  <div class="audit-page">
    <section class="audit-head">
      <div>
        <h2>操作日志</h2>
        <p>查看谁在什么时间、从哪个 IP、在哪个板块执行了什么操作。</p>
      </div>
      <el-button type="primary" :icon="Refresh" @click="refresh">刷新</el-button>
    </section>

    <section class="audit-panel">
      <div class="audit-filters">
        <el-input v-model="query.keyword" clearable placeholder="搜索用户、IP、板块、接口、操作内容" @keyup.enter="refresh" @clear="refresh" />
        <el-select v-model="query.module" clearable placeholder="板块" @change="refresh">
          <el-option v-for="item in moduleOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
        <el-select v-model="query.action" clearable placeholder="事件" @change="refresh">
          <el-option v-for="item in actionOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
        <el-select v-model="query.statusCode" clearable placeholder="结果" @change="refresh">
          <el-option label="成功 200" value="200" />
          <el-option label="创建 201" value="201" />
          <el-option label="无权限 403" value="403" />
          <el-option label="失败 500" value="500" />
        </el-select>
        <el-date-picker
          v-model="dateRange"
          type="datetimerange"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          value-format="YYYY-MM-DDTHH:mm:ssZ"
          @change="refresh"
        />
        <el-button :icon="Search" @click="refresh">筛选</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>

      <el-table :data="logs" v-loading="loading" class="audit-table" border stripe>
        <el-table-column prop="createdAt" label="时间" min-width="170" show-overflow-tooltip />
        <el-table-column label="用户" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">{{ row.username || '-' }}</template>
        </el-table-column>
        <el-table-column label="事件" width="110">
          <template #default="{ row }">
            <el-tag :type="actionTagType(row.action)">{{ actionText(row.action) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="IP" min-width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ row.ip || '-' }}</template>
        </el-table-column>
        <el-table-column label="板块" min-width="130" show-overflow-tooltip>
          <template #default="{ row }">{{ moduleText(row.module) }}</template>
        </el-table-column>
        <el-table-column label="操作内容" min-width="340" show-overflow-tooltip>
          <template #default="{ row }">{{ operationText(row) }}</template>
        </el-table-column>
        <el-table-column label="结果" width="95">
          <template #default="{ row }">
            <el-tag :type="row.statusCode >= 400 ? 'danger' : 'success'">{{ resultText(row.statusCode) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="详情" width="90" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" :icon="Document" @click="openDetail(row)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pager">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.pageSize"
          :total="total"
          :page-sizes="[20, 50, 100, 200]"
          layout="total, sizes, prev, pager, next"
          @current-change="refresh"
          @size-change="refresh"
        />
      </div>
    </section>

    <el-drawer v-model="detail.visible" size="560px" title="操作详情">
      <el-descriptions v-if="detail.row" :column="1" border>
        <el-descriptions-item label="时间">{{ detail.row.createdAt }}</el-descriptions-item>
        <el-descriptions-item label="用户">{{ detail.row.username || '-' }}</el-descriptions-item>
        <el-descriptions-item label="事件">{{ actionText(detail.row.action) }}</el-descriptions-item>
        <el-descriptions-item label="IP">{{ detail.row.ip || '-' }}</el-descriptions-item>
        <el-descriptions-item label="板块">{{ moduleText(detail.row.module) }}</el-descriptions-item>
        <el-descriptions-item label="操作">{{ operationText(detail.row) }}</el-descriptions-item>
        <el-descriptions-item label="接口">{{ detail.row.method }} {{ detail.row.path }}</el-descriptions-item>
        <el-descriptions-item label="结果">{{ resultText(detail.row.statusCode) }}</el-descriptions-item>
        <el-descriptions-item label="耗时">{{ detail.row.costMs }} ms</el-descriptions-item>
        <el-descriptions-item label="设备">{{ detail.row.userAgent || '-' }}</el-descriptions-item>
      </el-descriptions>
      <h3>原始请求内容</h3>
      <pre class="request-body">{{ prettyBody(detail.row?.requestBody) }}</pre>
    </el-drawer>
  </div>
</template>

<script setup>
import { onActivated, onMounted, reactive, ref } from 'vue'
import { Document, Refresh, Search } from '@element-plus/icons-vue'
import { listOperationLogs } from '../../api/audit'
import { useSearch } from '../../composables/useSearch'
import { formatDateRows } from '../../utils/date'

const loading = ref(false)
const logs = ref([])
const total = ref(0)
const dateRange = ref([])
const detail = reactive({ visible: false, row: null })
const { searchForm: query, resetSearch } = useSearch({
  page: 1,
  pageSize: 20,
  keyword: '',
  module: '',
  action: '',
  method: '',
  statusCode: '',
  startDate: '',
  endDate: ''
})

const actionOptions = [
  { label: '新增', value: 'create' },
  { label: '修改', value: 'update' },
  { label: '删除', value: 'delete' },
  { label: '上传/导入', value: 'upload' },
  { label: '业务动作', value: 'action' },
  { label: '重置密码', value: 'reset_password' },
  { label: '登录', value: 'login' },
  { label: '退出', value: 'logout' }
]

const moduleOptions = [
  { label: '销售管理', value: 'sales' },
  { label: '采购管理', value: 'purchase' },
  { label: '库存查询', value: 'inventory' },
  { label: '库存流水', value: 'inventory-movements' },
  { label: '维修管理', value: 'repair' },
  { label: '工程管理', value: 'project' },
  { label: '资金流水', value: 'finance' },
  { label: '资金账户', value: 'finance-accounts' },
  { label: '应收账款', value: 'receivables' },
  { label: '应付账款', value: 'payables' },
  { label: '客户对账单', value: 'customer-statements' },
  { label: '客户管理', value: 'customers' },
  { label: '供应商管理', value: 'suppliers' },
  { label: '商品管理', value: 'products' },
  { label: '用户管理', value: 'users' },
  { label: '角色管理', value: 'roles' },
  { label: '权限管理', value: 'permissions' },
  { label: '菜单管理', value: 'menus' },
  { label: '系统配置', value: 'settings' },
  { label: '登录认证', value: 'auth' }
]

onMounted(() => {
  resetFilters()
})
onActivated(() => {
  resetFilters()
})

async function refresh() {
  loading.value = true
  try {
    query.startDate = dateRange.value?.[0] || ''
    query.endDate = dateRange.value?.[1] || ''
    const data = await listOperationLogs(query)
    logs.value = formatDateRows(data.list || [])
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  resetSearch()
  dateRange.value = []
  refresh()
}

function openDetail(row) {
  detail.row = row
  detail.visible = true
}

function actionText(action) {
  return {
    create: '新增',
    update: '修改',
    delete: '删除',
    upload: '上传/导入',
    action: '业务动作',
    reset_password: '重置密码',
    login: '登录',
    logout: '退出'
  }[action] || action || '-'
}

function actionTagType(action) {
  return { create: 'success', update: 'warning', delete: 'danger', upload: 'info', reset_password: 'warning' }[action] || 'primary'
}

function moduleText(module) {
  const found = moduleOptions.find((item) => item.value === module)
  return found?.label || module || '-'
}

function resultText(statusCode) {
  if (!statusCode) return '-'
  return statusCode >= 400 ? `失败 ${statusCode}` : `成功 ${statusCode}`
}

function operationText(row) {
  const moduleName = moduleText(row.module)
  const eventName = actionText(row.action)
  const target = targetText(row)
  const body = requestSummary(row.requestBody)
  const targetPart = target ? `，对象：${target}` : ''
  const bodyPart = body && body !== '-' ? `，内容：${body}` : ''
  return `在${moduleName}执行${eventName}${targetPart}${bodyPart}`
}

function targetText(row) {
  const path = row.path || ''
  const parts = path.split('/').filter(Boolean)
  const last = parts[parts.length - 1]
  if (last && /^\d+$/.test(last)) return `ID ${last}`
  const parsed = parseBody(row.requestBody)
  const data = parsed?.data || parsed || {}
  return firstValue(data, ['id', 'orderNo', 'order_no', 'statementNo', 'statement_no', 'receivableNo', 'receivable_no', 'payableNo', 'payable_no', 'name', 'username'])
}

function requestSummary(body) {
  if (!body) return '-'
  const parsed = parseBody(body)
  if (parsed && typeof parsed === 'object') {
    const data = parsed.data && typeof parsed.data === 'object' ? parsed.data : parsed
    const action = parsed.action ? `动作=${parsed.action}` : ''
    const pairs = usefulKeys(data).map((key) => `${fieldText(key)}=${summaryValue(data[key])}`)
    return [action, ...pairs].filter(Boolean).slice(0, 6).join('，') || '-'
  }
  return String(body).slice(0, 120)
}

function usefulKeys(data) {
  const preferred = [
    'orderNo', 'order_no', 'statementNo', 'statement_no', 'receivableNo', 'receivable_no', 'payableNo', 'payable_no',
    'customerName', 'customer_name', 'supplierName', 'supplier_name', 'productName', 'product_name',
    'name', 'username', 'amount', 'totalAmount', 'total_amount', 'status', 'reason', 'remark'
  ]
  const keys = Object.keys(data || {}).filter((key) => !['password', 'oldPassword', 'newPassword', 'confirmPassword', 'token'].includes(key))
  const ordered = preferred.filter((key) => keys.includes(key))
  return [...ordered, ...keys.filter((key) => !ordered.includes(key))].slice(0, 6)
}

function fieldText(key) {
  return {
    orderNo: '单号',
    order_no: '单号',
    statementNo: '对账单号',
    statement_no: '对账单号',
    receivableNo: '应收单号',
    receivable_no: '应收单号',
    payableNo: '应付单号',
    payable_no: '应付单号',
    customerName: '客户',
    customer_name: '客户',
    supplierName: '供应商',
    supplier_name: '供应商',
    productName: '商品',
    product_name: '商品',
    totalAmount: '金额',
    total_amount: '金额',
    amount: '金额',
    status: '状态',
    reason: '原因',
    remark: '备注',
    username: '用户',
    name: '名称'
  }[key] || key
}

function firstValue(data, keys) {
  for (const key of keys) {
    if (data?.[key] !== undefined && data?.[key] !== null && data?.[key] !== '') return data[key]
  }
  return ''
}

function prettyBody(body) {
  if (!body) return '-'
  const parsed = parseBody(body)
  if (!parsed) return body
  return JSON.stringify(parsed, null, 2)
}

function parseBody(body) {
  try {
    return JSON.parse(body)
  } catch {
    return null
  }
}

function summaryValue(value) {
  if (Array.isArray(value)) return `[${value.length}项]`
  if (value && typeof value === 'object') return '{...}'
  return String(value ?? '').slice(0, 32)
}
</script>

<style scoped>
.audit-page {
  display: grid;
  gap: 14px;
}

.audit-head,
.audit-panel {
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 8px;
  box-shadow: var(--shadow);
}

.audit-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px;
}

.audit-head h2 {
  margin: 0;
  font-size: 22px;
}

.audit-head p {
  margin: 6px 0 0;
  color: var(--muted);
}

.audit-panel {
  padding: 14px;
}

.audit-filters {
  display: grid;
  grid-template-columns: minmax(220px, 1.4fr) repeat(3, minmax(130px, 0.8fr)) minmax(300px, 1.4fr) auto auto;
  gap: 8px;
  margin-bottom: 12px;
}

.audit-table {
  width: 100%;
}

.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}

.request-body {
  overflow: auto;
  max-height: 46vh;
  padding: 12px;
  color: var(--text);
  white-space: pre-wrap;
  word-break: break-word;
  background: var(--panel-soft);
  border: 1px solid var(--line);
  border-radius: 8px;
}

@media (max-width: 1199px) {
  .audit-filters {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 767px) {
  .audit-head {
    align-items: stretch;
    flex-direction: column;
  }

  .audit-filters {
    grid-template-columns: 1fr;
  }

  .audit-filters .el-button {
    width: 100%;
  }

  :deep(.el-drawer) {
    width: 92vw !important;
  }
}
</style>
