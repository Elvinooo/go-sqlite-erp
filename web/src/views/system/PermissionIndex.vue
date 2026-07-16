<template>
  <div class="permission-page">
    <div class="permission-head">
      <div>
        <h2>权限管理</h2>
        <p>管理员工账号、角色、菜单和接口/按钮权限，员工登录后按权限显示菜单与按钮。</p>
      </div>
      <el-button type="primary" :icon="Refresh" @click="loadAll">刷新</el-button>
    </div>

    <el-tabs v-model="activeTab" class="permission-tabs">
      <el-tab-pane label="员工用户" name="users">
        <div class="permission-toolbar">
          <el-input v-model="userQuery.keyword" clearable placeholder="搜索用户名、姓名、手机号" @keyup.enter="loadUsers" />
          <el-button :icon="Search" @click="loadUsers">搜索</el-button>
          <el-button type="primary" :icon="Plus" @click="openUser()">新增员工</el-button>
        </div>

        <el-table :data="users" class="permission-table" v-loading="loading.users">
          <el-table-column prop="username" label="用户名" min-width="120" />
          <el-table-column prop="realName" label="姓名" min-width="110" />
          <el-table-column prop="phone" label="手机" min-width="130" />
          <el-table-column prop="email" label="邮箱" min-width="160" show-overflow-tooltip />
          <el-table-column label="角色" min-width="180">
            <template #default="{ row }">
              <el-tag v-for="role in row.roles || []" :key="role.id" size="small">{{ role.name }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'active' ? 'success' : 'info'">{{ statusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="改密" width="100">
            <template #default="{ row }">
              <el-tag :type="row.mustChangePassword ? 'warning' : 'info'">{{ row.mustChangePassword ? '需改密' : '正常' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="lastLoginAt" label="最后登录" min-width="170" show-overflow-tooltip />
          <el-table-column label="操作" width="250" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" :icon="Edit" @click="openUser(row)">编辑</el-button>
              <el-button link :icon="Key" @click="openReset(row)">重置密码</el-button>
              <el-popconfirm v-if="row.username !== 'admin'" title="确定删除该员工？" @confirm="removeUser(row)">
                <template #reference>
                  <el-button link type="danger" :icon="Delete">删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="角色授权" name="roles">
        <div class="permission-toolbar">
          <el-input v-model="roleQuery.keyword" clearable placeholder="搜索角色编码、名称" @keyup.enter="loadRoles" />
          <el-button :icon="Search" @click="loadRoles">搜索</el-button>
          <el-button type="primary" :icon="Plus" @click="openRole()">新增角色</el-button>
        </div>

        <el-table :data="roles" class="permission-table" v-loading="loading.roles">
          <el-table-column prop="code" label="角色编码" min-width="140" />
          <el-table-column prop="name" label="角色名称" min-width="130" />
          <el-table-column label="继承角色" min-width="130">
            <template #default="{ row }">{{ row.parent?.name || '-' }}</template>
          </el-table-column>
          <el-table-column label="数据权限" width="120">
            <template #default="{ row }">{{ dataScopeText(row.dataScope) }}</template>
          </el-table-column>
          <el-table-column label="权限数" width="100">
            <template #default="{ row }">{{ row.permissions?.length || 0 }}</template>
          </el-table-column>
          <el-table-column label="菜单数" width="100">
            <template #default="{ row }">{{ row.menus?.length || 0 }}</template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'active' ? 'success' : 'info'">{{ statusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="190" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" :icon="Edit" @click="openRole(row)">授权</el-button>
              <el-popconfirm v-if="row.code !== 'super_admin'" title="确定删除该角色？" @confirm="removeRole(row)">
                <template #reference>
                  <el-button link type="danger" :icon="Delete">删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="权限点" name="permissions">
        <div class="permission-toolbar">
          <el-input v-model="permissionQuery.keyword" clearable placeholder="搜索权限编码、名称、模块" @keyup.enter="loadPermissions" />
          <el-button :icon="Search" @click="loadPermissions">搜索</el-button>
          <el-button type="primary" :icon="Plus" @click="openPermission()">新增权限</el-button>
        </div>

        <el-table :data="permissions" class="permission-table" v-loading="loading.permissions">
          <el-table-column prop="code" label="权限编码" min-width="180" show-overflow-tooltip />
          <el-table-column prop="name" label="权限名称" min-width="160" />
          <el-table-column prop="module" label="模块" width="110" />
          <el-table-column label="类型" width="100">
            <template #default="{ row }">{{ permissionTypeText(row.type) }}</template>
          </el-table-column>
          <el-table-column prop="method" label="方法" width="90" />
          <el-table-column prop="path" label="接口路径" min-width="220" show-overflow-tooltip />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'active' ? 'success' : 'info'">{{ statusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="170" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" :icon="Edit" @click="openPermission(row)">编辑</el-button>
              <el-popconfirm v-if="row.code !== '*'" title="确定删除该权限点？" @confirm="removePermission(row)">
                <template #reference>
                  <el-button link type="danger" :icon="Delete">删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="菜单管理" name="menus">
        <div class="permission-toolbar">
          <el-input v-model="menuQuery.keyword" clearable placeholder="搜索菜单名称、标题、路径" @keyup.enter="loadMenus" />
          <el-button :icon="Search" @click="loadMenus">搜索</el-button>
          <el-button type="primary" :icon="Plus" @click="openMenu()">新增菜单</el-button>
        </div>

        <el-table :data="menus" row-key="id" class="permission-table" v-loading="loading.menus">
          <el-table-column prop="title" label="菜单标题" min-width="150" />
          <el-table-column prop="name" label="菜单标识" min-width="130" />
          <el-table-column prop="path" label="路径" min-width="160" show-overflow-tooltip />
          <el-table-column label="类型" width="100">
            <template #default="{ row }">{{ menuTypeText(row.type) }}</template>
          </el-table-column>
          <el-table-column prop="permissionCode" label="权限编码" min-width="180" show-overflow-tooltip />
          <el-table-column label="显示" width="90">
            <template #default="{ row }">
              <el-tag :type="row.visible ? 'success' : 'info'">{{ row.visible ? '显示' : '隐藏' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'active' ? 'success' : 'info'">{{ statusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="170" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" :icon="Edit" @click="openMenu(row)">编辑</el-button>
              <el-popconfirm title="确定删除该菜单？" @confirm="removeMenu(row)">
                <template #reference>
                  <el-button link type="danger" :icon="Delete">删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="userDialog.visible" :title="userDialog.id ? '编辑员工' : '新增员工'" width="640px">
      <el-form label-width="96px" class="form-grid">
        <el-form-item label="用户名">
          <el-input v-model.trim="userForm.username" :disabled="!!userDialog.id" placeholder="登录账号" />
        </el-form-item>
        <el-form-item v-if="!userDialog.id" label="初始密码">
          <el-input v-model="userForm.password" show-password placeholder="至少 6 位" />
        </el-form-item>
        <el-form-item label="姓名">
          <el-input v-model.trim="userForm.realName" />
        </el-form-item>
        <el-form-item label="手机">
          <el-input v-model.trim="userForm.phone" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model.trim="userForm.email" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="userForm.status" :disabled="userForm.username === 'admin'" class="full">
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item label="强制改密">
          <el-switch v-model="userForm.mustChangePassword" />
        </el-form-item>
        <el-form-item label="角色" class="span-2">
          <el-select v-model="userForm.roleIds" multiple filterable collapse-tags class="full" placeholder="选择员工角色">
            <el-option v-for="role in roles" :key="role.id" :label="role.name" :value="role.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注" class="span-2">
          <el-input v-model="userForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="userDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveUser">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="resetDialog.visible" title="重置密码" width="420px">
      <el-form label-width="96px">
        <el-form-item label="员工">
          <el-input :model-value="resetDialog.user?.username" disabled />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="resetDialog.password" show-password placeholder="不少于8位，包含字母和数字" />
        </el-form-item>
        <el-form-item label="首次改密">
          <el-switch v-model="resetDialog.mustChangePassword" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveResetPassword">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="roleDialog.visible" :title="roleDialog.id ? '角色授权' : '新增角色'" width="860px">
      <el-form label-width="96px" class="form-grid">
        <el-form-item label="角色编码">
          <el-input v-model.trim="roleForm.code" :disabled="roleForm.code === 'super_admin'" />
        </el-form-item>
        <el-form-item label="角色名称">
          <el-input v-model.trim="roleForm.name" />
        </el-form-item>
        <el-form-item label="继承角色">
          <el-select v-model="roleForm.parentId" clearable filterable class="full" :disabled="roleForm.code === 'super_admin'">
            <el-option v-for="role in parentRoleOptions" :key="role.id" :label="role.name" :value="role.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="数据权限">
          <el-select v-model="roleForm.dataScope" class="full" :disabled="roleForm.code === 'super_admin'">
            <el-option label="全部数据" value="all" />
            <el-option label="本人数据" value="self" />
            <el-option label="本部门数据" value="dept" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="roleForm.sort" :min="0" class="full" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="roleForm.status" class="full" :disabled="roleForm.code === 'super_admin'">
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item label="菜单权限" class="span-2">
          <el-tree
            ref="menuTreeRef"
            :data="menuTree"
            node-key="id"
            show-checkbox
            default-expand-all
            :props="{ label: 'title', children: 'children' }"
            :disabled="roleForm.code === 'super_admin'"
          />
        </el-form-item>
        <el-form-item label="接口/按钮" class="span-2">
          <el-transfer
            v-model="roleForm.permissionIds"
            filterable
            :titles="['可选权限', '已授权']"
            :data="permissionTransferData"
            :disabled="roleForm.code === 'super_admin'"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveRole">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="permissionDialog.visible" :title="permissionDialog.id ? '编辑权限点' : '新增权限点'" width="620px">
      <el-form label-width="96px" class="form-grid">
        <el-form-item label="权限编码">
          <el-input v-model.trim="permissionForm.code" :disabled="permissionForm.code === '*'" />
        </el-form-item>
        <el-form-item label="权限名称">
          <el-input v-model.trim="permissionForm.name" />
        </el-form-item>
        <el-form-item label="模块">
          <el-input v-model.trim="permissionForm.module" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="permissionForm.type" class="full">
            <el-option label="接口权限" value="api" />
            <el-option label="按钮权限" value="button" />
            <el-option label="数据权限" value="data" />
          </el-select>
        </el-form-item>
        <el-form-item label="方法">
          <el-select v-model="permissionForm.method" clearable class="full">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="permissionForm.status" class="full">
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item label="接口路径" class="span-2">
          <el-input v-model.trim="permissionForm.path" placeholder="/api/v1/..." />
        </el-form-item>
        <el-form-item label="备注" class="span-2">
          <el-input v-model="permissionForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="permissionDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="savePermission">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="menuDialog.visible" :title="menuDialog.id ? '编辑菜单' : '新增菜单'" width="660px">
      <el-form label-width="96px" class="form-grid">
        <el-form-item label="上级菜单">
          <el-select v-model="menuForm.parentId" clearable filterable class="full">
            <el-option v-for="menu in parentMenuOptions" :key="menu.id" :label="menu.title" :value="menu.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="menuForm.type" class="full">
            <el-option label="目录" value="directory" />
            <el-option label="菜单" value="menu" />
            <el-option label="按钮" value="button" />
          </el-select>
        </el-form-item>
        <el-form-item label="菜单标识">
          <el-input v-model.trim="menuForm.name" />
        </el-form-item>
        <el-form-item label="菜单标题">
          <el-input v-model.trim="menuForm.title" />
        </el-form-item>
        <el-form-item label="路径">
          <el-input v-model.trim="menuForm.path" />
        </el-form-item>
        <el-form-item label="组件">
          <el-input v-model.trim="menuForm.component" />
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model.trim="menuForm.icon" />
        </el-form-item>
        <el-form-item label="权限编码">
          <el-select v-model="menuForm.permissionCode" clearable filterable allow-create class="full">
            <el-option v-for="item in permissions" :key="item.id" :label="`${item.name} / ${item.code}`" :value="item.code" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="menuForm.sort" :min="0" class="full" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="menuForm.status" class="full">
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item label="显示">
          <el-switch v-model="menuForm.visible" active-text="显示" inactive-text="隐藏" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="menuForm.remark" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="menuDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveMenu">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, nextTick, onActivated, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Delete, Edit, Key, Plus, Refresh, Search } from '@element-plus/icons-vue'
import {
  createMenu,
  createPermission,
  createRole,
  createUser,
  deleteMenu,
  deletePermission,
  deleteRole,
  deleteUser,
  listMenus,
  listMenuTree,
  listPermissions,
  listRoles,
  listUsers,
  resetUserPassword,
  updateMenu,
  updatePermission,
  updateRole,
  updateUser
} from '../../api/rbac'
import { useSearch } from '../../composables/useSearch'
import { formatDateRows } from '../../utils/date'

const activeTab = ref('users')
const saving = ref(false)
const menuTreeRef = ref()

const loading = reactive({ users: false, roles: false, permissions: false, menus: false })
const { searchForm: userQuery, resetSearch: resetUserSearch } = useSearch({ page: 1, pageSize: 100, keyword: '' })
const { searchForm: roleQuery, resetSearch: resetRoleSearch } = useSearch({ page: 1, pageSize: 300, keyword: '' })
const { searchForm: permissionQuery, resetSearch: resetPermissionSearch } = useSearch({ page: 1, pageSize: 500, keyword: '' })
const { searchForm: menuQuery, resetSearch: resetMenuSearch } = useSearch({ page: 1, pageSize: 500, keyword: '' })

const users = ref([])
const roles = ref([])
const permissions = ref([])
const menus = ref([])
const menuTree = ref([])

const userDialog = reactive({ visible: false, id: null })
const roleDialog = reactive({ visible: false, id: null })
const permissionDialog = reactive({ visible: false, id: null })
const menuDialog = reactive({ visible: false, id: null })
const resetDialog = reactive({ visible: false, user: null, password: 'Abc12345', mustChangePassword: true })

const userForm = reactive({})
const roleForm = reactive({})
const permissionForm = reactive({})
const menuForm = reactive({})

const permissionTransferData = computed(() => permissions.value.map((item) => ({
  key: item.id,
  label: `${item.name} / ${item.code}`,
  disabled: item.code === '*'
})))

const parentRoleOptions = computed(() => roles.value.filter((role) => role.id !== roleDialog.id && role.code !== 'super_admin'))
const parentMenuOptions = computed(() => menus.value.filter((menu) => menu.id !== menuDialog.id && menu.type !== 'button'))

onMounted(resetSearchAndLoad)
onActivated(resetSearchAndLoad)

function resetAllSearch() {
  resetUserSearch()
  resetRoleSearch()
  resetPermissionSearch()
  resetMenuSearch()
}

async function resetSearchAndLoad() {
  resetAllSearch()
  await loadAll()
}

async function loadAll() {
  await Promise.all([loadRoles(), loadPermissions(), loadMenus(), loadMenuTree()])
  await loadUsers()
}

async function loadUsers() {
  loading.users = true
  try {
    const data = await listUsers(userQuery)
    users.value = formatDateRows(data.list || [])
  } finally {
    loading.users = false
  }
}

async function loadRoles() {
  loading.roles = true
  try {
    const data = await listRoles(roleQuery)
    roles.value = formatDateRows(data.list || [])
  } finally {
    loading.roles = false
  }
}

async function loadPermissions() {
  loading.permissions = true
  try {
    const data = await listPermissions(permissionQuery)
    permissions.value = formatDateRows(data.list || [])
  } finally {
    loading.permissions = false
  }
}

async function loadMenus() {
  loading.menus = true
  try {
    const data = await listMenus(menuQuery)
    menus.value = formatDateRows(data.list || [])
  } finally {
    loading.menus = false
  }
}

async function loadMenuTree() {
  menuTree.value = await listMenuTree()
}

function openUser(row) {
  assign(userForm, {
    username: row?.username || '',
    password: '',
    realName: row?.realName || '',
    phone: row?.phone || '',
    email: row?.email || '',
    avatar: row?.avatar || '',
    status: row?.status || 'active',
    mustChangePassword: row?.mustChangePassword || false,
    roleIds: row?.roles?.map((role) => role.id) || [],
    remark: row?.remark || ''
  })
  userDialog.id = row?.id || null
  userDialog.visible = true
}

async function saveUser() {
  if (!userDialog.id && !isValidPassword(userForm.password)) {
    ElMessage.warning('初始密码不少于8位，且必须包含字母和数字')
    return
  }
  if (userDialog.id && userForm.password && !isValidPassword(userForm.password)) {
    ElMessage.warning('密码不少于8位，且必须包含字母和数字')
    return
  }
  saving.value = true
  try {
    if (userDialog.id) {
      await updateUser(userDialog.id, cleanPayload(userForm, ['username']))
      ElMessage.success('员工已更新')
    } else {
      await createUser(userForm)
      ElMessage.success('员工已创建')
    }
    userDialog.visible = false
    await loadUsers()
  } finally {
    saving.value = false
  }
}

async function removeUser(row) {
  await deleteUser(row.id)
  ElMessage.success('员工已删除')
  await loadUsers()
}

function openReset(row) {
  resetDialog.user = row
  resetDialog.password = 'Abc12345'
  resetDialog.mustChangePassword = true
  resetDialog.visible = true
}

async function saveResetPassword() {
  if (!isValidPassword(resetDialog.password)) {
    ElMessage.warning('密码不少于8位，且必须包含字母和数字')
    return
  }
  saving.value = true
  try {
    await resetUserPassword(resetDialog.user.id, {
      password: resetDialog.password,
      mustChangePassword: resetDialog.mustChangePassword
    })
    ElMessage.success('密码已重置')
    resetDialog.visible = false
  } finally {
    saving.value = false
  }
}

function isValidPassword(password) {
  return typeof password === 'string' && password.length >= 8 && /^(?=.*[A-Za-z])(?=.*\d).+$/.test(password)
}

async function openRole(row) {
  assign(roleForm, {
    parentId: row?.parentId || null,
    code: row?.code || '',
    name: row?.name || '',
    dataScope: row?.dataScope || 'all',
    dataScopeDeptIds: row?.dataScopeDeptIds || '',
    sort: row?.sort || 0,
    status: row?.status || 'active',
    permissionIds: row?.permissions?.map((item) => item.id) || [],
    menuIds: row?.menus?.map((item) => item.id) || [],
    remark: row?.remark || ''
  })
  roleDialog.id = row?.id || null
  roleDialog.visible = true
  await nextTick()
  menuTreeRef.value?.setCheckedKeys(roleForm.menuIds)
}

async function saveRole() {
  saving.value = true
  try {
    const payload = { ...roleForm, menuIds: menuTreeRef.value?.getCheckedKeys(false) || [] }
    if (roleDialog.id) {
      await updateRole(roleDialog.id, payload)
      ElMessage.success('角色权限已更新')
    } else {
      await createRole(payload)
      ElMessage.success('角色已创建')
    }
    roleDialog.visible = false
    await Promise.all([loadRoles(), loadUsers()])
  } finally {
    saving.value = false
  }
}

async function removeRole(row) {
  await deleteRole(row.id)
  ElMessage.success('角色已删除')
  await Promise.all([loadRoles(), loadUsers()])
}

function openPermission(row) {
  assign(permissionForm, {
    code: row?.code || '',
    name: row?.name || '',
    module: row?.module || '',
    type: row?.type || 'api',
    method: row?.method || '',
    path: row?.path || '',
    status: row?.status || 'active',
    remark: row?.remark || ''
  })
  permissionDialog.id = row?.id || null
  permissionDialog.visible = true
}

async function savePermission() {
  saving.value = true
  try {
    if (permissionDialog.id) {
      await updatePermission(permissionDialog.id, permissionForm)
      ElMessage.success('权限点已更新')
    } else {
      await createPermission(permissionForm)
      ElMessage.success('权限点已创建')
    }
    permissionDialog.visible = false
    await Promise.all([loadPermissions(), loadRoles()])
  } finally {
    saving.value = false
  }
}

async function removePermission(row) {
  await deletePermission(row.id)
  ElMessage.success('权限点已删除')
  await Promise.all([loadPermissions(), loadRoles()])
}

function openMenu(row) {
  assign(menuForm, {
    parentId: row?.parentId || null,
    name: row?.name || '',
    title: row?.title || '',
    path: row?.path || '',
    component: row?.component || '',
    icon: row?.icon || '',
    type: row?.type || 'menu',
    permissionCode: row?.permissionCode || '',
    sort: row?.sort || 0,
    visible: row?.visible ?? true,
    status: row?.status || 'active',
    remark: row?.remark || ''
  })
  menuDialog.id = row?.id || null
  menuDialog.visible = true
}

async function saveMenu() {
  saving.value = true
  try {
    if (menuDialog.id) {
      await updateMenu(menuDialog.id, menuForm)
      ElMessage.success('菜单已更新')
    } else {
      await createMenu(menuForm)
      ElMessage.success('菜单已创建')
    }
    menuDialog.visible = false
    await Promise.all([loadMenus(), loadMenuTree(), loadRoles()])
  } finally {
    saving.value = false
  }
}

async function removeMenu(row) {
  await deleteMenu(row.id)
  ElMessage.success('菜单已删除')
  await Promise.all([loadMenus(), loadMenuTree(), loadRoles()])
}

function assign(target, value) {
  for (const key of Object.keys(target)) delete target[key]
  Object.assign(target, value)
}

function cleanPayload(source, omit = []) {
  const payload = { ...source }
  for (const key of omit) delete payload[key]
  if (!payload.password) delete payload.password
  return payload
}

function statusText(status) {
  return status === 'active' ? '启用' : '禁用'
}

function dataScopeText(scope) {
  return { all: '全部', self: '本人', dept: '部门', custom: '自定义' }[scope] || scope || '-'
}

function permissionTypeText(type) {
  return { api: '接口', button: '按钮', data: '数据' }[type] || type || '-'
}

function menuTypeText(type) {
  return { directory: '目录', menu: '菜单', button: '按钮' }[type] || type || '-'
}
</script>

<style scoped>
.permission-page {
  display: grid;
  gap: 14px;
}

.permission-head,
.permission-tabs {
  max-width: 100%;
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 8px;
  box-shadow: var(--shadow);
}

.permission-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px;
}

.permission-head h2 {
  margin: 0;
  font-size: 22px;
}

.permission-head p {
  margin: 6px 0 0;
  color: var(--muted);
}

.permission-tabs {
  padding: 8px 14px 14px;
}

.permission-tabs :deep(.el-tabs__header) {
  margin-bottom: 12px;
}

.permission-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.permission-toolbar .el-input {
  max-width: 360px;
}

.permission-table {
  width: 100%;
}

.permission-table .el-tag {
  margin: 2px 4px 2px 0;
}

.span-2 {
  grid-column: 1 / -1;
}

:deep(.el-transfer) {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 44px minmax(0, 1fr);
  align-items: center;
}

:deep(.el-transfer-panel) {
  width: 100%;
}

@media (max-width: 767px) {
  .permission-head,
  .permission-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .permission-toolbar .el-input,
  .permission-toolbar .el-button {
    width: 100%;
    max-width: none;
  }

  .permission-tabs {
    padding: 6px 10px 12px;
    width: 100%;
    max-width: 100%;
    overflow: hidden;
  }

  .permission-tabs :deep(.el-tabs__nav-wrap) {
    overflow-x: auto;
    scrollbar-width: none;
  }

  .permission-tabs :deep(.el-tabs__nav-wrap::-webkit-scrollbar) {
    display: none;
  }

  .permission-tabs :deep(.el-tabs__content) {
    min-width: 0;
    max-width: 100%;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .permission-tabs :deep(.el-tab-pane) {
    min-width: 0;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  :deep(.el-dialog) {
    width: calc(100vw - 24px) !important;
  }

  :deep(.el-transfer) {
    grid-template-columns: 1fr;
    gap: 8px;
  }

  :deep(.el-transfer__buttons) {
    display: flex;
    justify-content: center;
    padding: 0;
    transform: rotate(90deg);
  }

  :deep(.el-transfer-panel) {
    max-width: 100%;
  }
}
</style>
