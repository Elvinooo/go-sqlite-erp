<template>
  <main class="page customer-page">
    <section class="toolbar">
      <div class="filters">
        <el-input
          v-model="query.keyword"
          clearable
          placeholder="供应商编码、名称、联系人、电话、税号"
          :prefix-icon="Search"
          @keyup.enter="refresh"
          @clear="refresh"
        />
        <el-select v-model="query.sourceType" clearable placeholder="供应商类型" style="width: 160px" @change="refresh">
          <el-option v-for="item in supplierTypeOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="refresh">查询</el-button>
      </div>
      <div class="actions">
        <el-button type="primary" :icon="Plus" @click="openCreate">新增</el-button>
      </div>
    </section>

    <el-table v-if="!isMobile" v-loading="loading" :data="rows" border stripe class="data-table">
      <el-table-column prop="code" label="供应商编码" min-width="130" fixed />
      <el-table-column prop="name" label="供应商名称" min-width="180" show-overflow-tooltip />
      <el-table-column prop="supplierTypes" label="供应商类型" min-width="160" show-overflow-tooltip />
      <el-table-column prop="contactName" label="联系人" min-width="120" />
      <el-table-column prop="phone" label="电话" min-width="130" />
      <el-table-column prop="address" label="地址" min-width="220" show-overflow-tooltip />
      <el-table-column prop="payableBalance" label="应付余额" min-width="120" align="right" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : 'info'">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" :icon="Edit" @click="openEdit(row)">编辑</el-button>
          <el-popconfirm title="确定删除该供应商？" @confirm="remove(row)">
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
          <span>{{ row.contactName || '未填写联系人' }}</span>
          <span>{{ row.phone || '无电话' }}</span>
        </div>
        <p>{{ row.supplierTypes || '商品供应商' }}</p>
        <p>{{ row.address || '暂无地址' }}</p>
        <div class="card-meta">
          <span>应付 {{ row.payableBalance || '0.00' }}</span>
          <span>{{ row.taxNo || '无税号' }}</span>
        </div>
        <div class="card-actions">
          <el-button size="large" :icon="Edit" @click="openEdit(row)">编辑</el-button>
          <el-button size="large" type="danger" :icon="Delete" @click="remove(row)">删除</el-button>
        </div>
      </article>
      <el-empty v-if="!loading && rows.length === 0" description="暂无供应商" />
    </section>

    <el-pagination
      v-if="!isMobile"
      v-model:current-page="query.page"
      v-model:page-size="query.pageSize"
      :total="total"
      layout="total, sizes, prev, pager, next"
      @current-change="refresh"
      @size-change="refresh"
    />

    <el-dialog v-model="dialog.visible" :fullscreen="isMobile" :title="dialog.editing ? '编辑供应商' : '新增供应商'" width="560px">
      <el-form label-width="96px">
        <el-form-item label="编码"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="供应商类型">
          <el-select v-model="form.supplierTypeList" multiple collapse-tags collapse-tags-tooltip style="width: 100%">
            <el-option v-for="item in supplierTypeOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="联系人"><el-input v-model="form.contactName" /></el-form-item>
        <el-form-item label="电话"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item>
        <el-form-item label="税号"><el-input v-model="form.taxNo" /></el-form-item>
        <el-form-item label="地址"><el-input v-model="form.address" /></el-form-item>
        <el-form-item label="应付余额"><el-input-number v-model="form.payableBalance" :min="0" style="width: 100%" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width: 100%">
            <el-option label="启用" value="active" />
            <el-option label="停用" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </main>
</template>

<script setup>
import { computed, onActivated, onMounted, onUnmounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Delete, Edit, Plus, Search } from '@element-plus/icons-vue'
import { createSupplier, deleteSupplier, listSuppliers, updateSupplier } from '../../api/supplier'
import { getBreakpoint } from '../../utils/mobile'
import { useSearch } from '../../composables/useSearch'
import { formatDateRows } from '../../utils/date'
import { statusLabel } from '../../utils/status'

const loading = ref(false)
const saving = ref(false)
const rows = ref([])
const total = ref(0)
const breakpoint = ref(getBreakpoint())
const supplierTypeOptions = ['商品供应商', '外协服务商', '工程施工方', '综合供应商']
const { searchForm: query, resetSearch } = useSearch({ page: 1, pageSize: 20, keyword: '', sourceType: '', sortBy: 'id', order: 'desc' })
const dialog = reactive({ visible: false, editing: false, id: null })
const form = reactive(defaultForm())
const isMobile = computed(() => breakpoint.value === 'mobile')

onMounted(() => {
  window.addEventListener('resize', handleResize)
  resetSearch()
  refresh()
})
onActivated(() => {
  resetSearch()
  refresh()
})
onUnmounted(() => window.removeEventListener('resize', handleResize))

function handleResize() {
  breakpoint.value = getBreakpoint()
}

function defaultForm() {
  return { code: '', name: '', supplierTypes: '商品供应商', supplierTypeList: ['商品供应商'], contactName: '', phone: '', email: '', taxNo: '', address: '', payableBalance: 0, status: 'active', remark: '' }
}

async function refresh() {
  loading.value = true
  try {
    const data = await listSuppliers(query)
    rows.value = formatDateRows(data.list || [])
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

function resetForm(row = defaultForm()) {
  Object.assign(form, defaultForm(), row)
  form.supplierTypeList = String(form.supplierTypes || '商品供应商').split(/[、,，]/).map((item) => item.trim()).filter(Boolean)
}

function openCreate() {
  dialog.visible = true
  dialog.editing = false
  dialog.id = null
  resetForm()
}

function openEdit(row) {
  dialog.visible = true
  dialog.editing = true
  dialog.id = row.id
  resetForm(row)
}

async function save() {
  saving.value = true
  try {
    const payload = { ...form, supplierTypes: (form.supplierTypeList || []).join('、') || '商品供应商' }
    if (dialog.editing) {
      await updateSupplier(dialog.id, payload)
    } else {
      await createSupplier(payload)
    }
    ElMessage.success('保存成功')
    dialog.visible = false
    await refresh()
  } finally {
    saving.value = false
  }
}

async function remove(row) {
  await deleteSupplier(row.id)
  ElMessage.success('删除成功')
  await refresh()
}
</script>
