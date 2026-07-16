<template>
  <main class="mobile-page dashboard-page boss-dashboard">
    <section class="boss-hero">
      <div>
        <span>ERP Pro</span>
        <h2>老板驾驶舱</h2>
        <p>PC + 手机 + PWA 三端一体</p>
      </div>
      <el-button circle :icon="Refresh" :loading="loading" @click="load" />
    </section>

    <section class="metric-grid boss-metrics">
      <article v-for="item in metrics" :key="item.label" class="metric-card">
        <span>{{ item.label }}</span>
        <strong>{{ item.value }}</strong>
        <small>{{ item.hint }}</small>
      </article>
    </section>

    <section class="quick-panel">
      <div class="section-title">
        <h2>快捷处理</h2>
        <el-button link type="primary" @click="$router.push('/sales')">开单</el-button>
      </div>
      <div class="quick-grid">
        <button v-for="item in quickActions" :key="item.label" class="quick-action" :disabled="item.loading?.value" @click="item.action">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </button>
      </div>
    </section>

    <section class="chart-panel">
      <div class="section-title">
        <h2>利润趋势</h2>
        <span>近7天</span>
      </div>
      <div ref="profitChartRef" class="chart-box"></div>
    </section>

    <section class="price-panel">
      <div class="section-title">
        <h2>价格利润分析</h2>
        <span>近30天</span>
      </div>
      <div class="price-grid">
        <article>
          <span>平均售价</span>
          <strong>{{ money(priceAnalysis.summary.avgSalesPrice) }}</strong>
        </article>
        <article>
          <span>平均采购价</span>
          <strong>{{ money(priceAnalysis.summary.avgPurchasePrice) }}</strong>
        </article>
        <article>
          <span>平均成本价</span>
          <strong>{{ money(priceAnalysis.summary.avgCostPrice) }}</strong>
        </article>
        <article>
          <span>平均毛利率</span>
          <strong>{{ percent(priceAnalysis.summary.avgProfitRate) }}</strong>
        </article>
      </div>
      <div class="price-table">
        <div class="price-row price-row-head">
          <span>日期</span><span>销售额</span><span>采购额</span><span>利润</span><span>毛利率</span>
        </div>
        <div v-for="item in priceAnalysis.daily.slice(-7)" :key="item.date" class="price-row">
          <span>{{ item.date }}</span>
          <span>{{ money(item.salesAmount) }}</span>
          <span>{{ money(item.purchaseAmount) }}</span>
          <span>{{ money(item.profitAmount) }}</span>
          <span>{{ percent(item.avgProfitRate) }}</span>
        </div>
      </div>
    </section>

    <section class="rank-grid">
      <article class="rank-panel">
        <div class="section-title">
          <h2>销售排行榜</h2>
        </div>
        <div v-if="dashboard.salesRanking.length === 0" class="empty-line">暂无销售数据</div>
        <div v-for="(item, index) in dashboard.salesRanking" :key="item.name" class="rank-row">
          <em>{{ index + 1 }}</em>
          <span>{{ item.name }}</span>
          <strong>{{ money(item.value) }}</strong>
        </div>
      </article>

      <article class="rank-panel">
        <div class="section-title">
          <h2>员工业绩</h2>
        </div>
        <div v-if="dashboard.employeePerformance.length === 0" class="empty-line">暂无业绩数据</div>
        <div v-for="(item, index) in dashboard.employeePerformance" :key="item.name" class="rank-row">
          <em>{{ index + 1 }}</em>
          <span>{{ item.name }}</span>
          <strong>{{ money(item.value) }}</strong>
        </div>
      </article>
    </section>

    <section class="recent-panel">
      <div class="section-title">
        <h2>最近订单</h2>
      </div>
      <el-empty v-if="dashboard.recentOrders.length === 0" description="暂无订单" />
      <article v-for="order in dashboard.recentOrders" :key="order.id || order.orderNo" class="order-card">
        <div>
          <strong>{{ order.orderNo }}</strong>
          <span>{{ order.customer || '未关联客户' }}</span>
        </div>
        <em>{{ money(order.amount) }}</em>
      </article>
    </section>

    <section class="project-panel">
      <div class="section-title">
        <h2>工程项目进度</h2>
      </div>
      <el-empty v-if="dashboard.projectProgress.length === 0" description="暂无工程项目" />
      <article v-for="project in dashboard.projectProgress" :key="project.id" class="project-row">
        <div>
          <strong>{{ project.name }}</strong>
          <span>{{ statusLabel(project.status) }}</span>
        </div>
        <el-progress :percentage="project.progress" :stroke-width="8" />
      </article>
    </section>

    <section class="location-panel">
      <div class="section-title">
        <h2>工程人员当前位置</h2>
        <el-tag size="small">可选</el-tag>
      </div>
      <el-empty v-if="dashboard.engineerLocations.length === 0" description="暂无位置上报" />
      <article v-for="item in dashboard.engineerLocations" :key="item.userId" class="location-row">
        <span>{{ item.name }}</span>
        <strong>{{ formatLocation(item) }}</strong>
        <small>{{ item.updatedAt || '-' }}</small>
      </article>
    </section>

    <button class="fab" @click="$router.push('/sales')">
      <el-icon><Plus /></el-icon>
    </button>
  </main>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import { Camera, Goods, Location, Plus, Printer, Refresh, Sell, Tools, UploadFilled } from '@element-plus/icons-vue'
import { getBossDashboard, getPriceProfitAnalysis, signInDashboard } from '../../api/dashboard'
import { getLocation, vibrate } from '../../utils/mobile'
import { statusLabel } from '../../utils/status'

const loading = ref(false)
const signInLoading = ref(false)
const profitChartRef = ref()
const router = useRouter()
let chart = null

const dashboard = reactive({
  todaySales: 0,
  monthProfit: 0,
  pendingReceivables: 0,
  inventoryValue: 0,
  inventoryQuantity: 0,
  inventoryAlerts: 0,
  pendingRepairs: 0,
  todayNewCustomers: 0,
  projectProgress: [],
  recentOrders: [],
  salesRanking: [],
  employeePerformance: [],
  profitTrend: [],
  engineerLocations: []
})

const priceAnalysis = reactive({
  summary: {},
  daily: [],
  products: []
})

const metrics = computed(() => [
  { label: '今日销售额', value: money(dashboard.todaySales), hint: '实时经营收入' },
  { label: '本月利润', value: money(dashboard.monthProfit), hint: '本月毛利统计' },
  { label: '库存资产', value: money(dashboard.inventoryValue), hint: `${Number(dashboard.inventoryQuantity || 0).toLocaleString('zh-CN')} 件库存` },
  { label: '待收款', value: money(dashboard.pendingReceivables), hint: '客户欠款余额' },
  { label: '库存预警', value: dashboard.inventoryAlerts, hint: '低库存商品' },
  { label: '待维修设备', value: dashboard.pendingRepairs, hint: '未完成维修' },
  { label: '今日新增客户', value: dashboard.todayNewCustomers, hint: '新增客户数' }
])

const quickActions = [
  { label: '快速销售', icon: Sell, action: () => go('/sales') },
  { label: '快速采购', icon: Goods, action: () => go('/purchase') },
  { label: '库存查询', icon: Camera, action: () => go('/inventory') },
  { label: '维修登记', icon: Tools, action: () => go('/repair') },
  { label: '工程签到', icon: Location, loading: signInLoading, action: signIn },
  { label: '现场图片', icon: UploadFilled, action: () => go('/project') },
  { label: '订单打印', icon: Printer, action: () => go('/sales') }
]

async function load() {
  loading.value = true
  try {
    const [data, priceData] = await Promise.all([getBossDashboard(), getPriceProfitAnalysis(30)])
    Object.assign(dashboard, normalize(data))
    Object.assign(priceAnalysis, normalizePriceAnalysis(priceData))
    await nextTick()
    renderChart()
  } finally {
    loading.value = false
  }
}

function normalizePriceAnalysis(data = {}) {
  return {
    summary: data.summary || {},
    daily: data.daily || [],
    products: data.products || []
  }
}

function normalize(data = {}) {
  return {
    todaySales: data.todaySales || 0,
    monthProfit: data.monthProfit || 0,
    pendingReceivables: data.pendingReceivables || 0,
    inventoryValue: data.inventoryValue || 0,
    inventoryQuantity: data.inventoryQuantity || 0,
    inventoryAlerts: data.inventoryAlerts || 0,
    pendingRepairs: data.pendingRepairs || 0,
    todayNewCustomers: data.todayNewCustomers || 0,
    projectProgress: data.projectProgress || [],
    recentOrders: data.recentOrders || [],
    salesRanking: data.salesRanking || [],
    employeePerformance: data.employeePerformance || [],
    profitTrend: data.profitTrend || [],
    engineerLocations: data.engineerLocations || []
  }
}

function renderChart() {
  if (!profitChartRef.value) return
  if (!chart) chart = echarts.init(profitChartRef.value)
  chart.setOption({
    grid: { left: 8, right: 8, top: 20, bottom: 22, containLabel: true },
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: dashboard.profitTrend.map((item) => item.date)
    },
    yAxis: { type: 'value' },
    series: [
      {
        type: 'line',
        smooth: true,
        symbolSize: 7,
        areaStyle: {},
        data: dashboard.profitTrend.map((item) => Number(item.value || 0))
      }
    ]
  })
}

function money(value) {
  const number = Number(value || 0)
  return number.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function percent(value) {
  return `${Number(value || 0).toFixed(2)}%`
}

function go(path) {
  vibrate()
  router.push(path)
}

async function signIn() {
	if (signInLoading.value) return
	signInLoading.value = true
	try {
		const position = await getLocation()
		const latitude = position.coords.latitude
		const longitude = position.coords.longitude
		const address = await reverseGeocode(latitude, longitude)
		const location = await signInDashboard({
			latitude,
			longitude,
			address,
			device: navigator.userAgent
		}) || {}
		ElMessage.success(`签到成功：${location.address || address || formatCoords({ latitude, longitude })}`)
		await load()
	} catch (error) {
		ElMessage.error(error.message || '签到失败')
	} finally {
    signInLoading.value = false
  }
}

function formatLocation(item) {
	if (item.address) return item.address
	return formatCoords(item)
}

function formatCoords(item) {
	const lat = Number(item.latitude || 0).toFixed(5)
	const lng = Number(item.longitude || 0).toFixed(5)
	return `${lat}, ${lng}`
}

async function reverseGeocode(latitude, longitude) {
	try {
		const params = new URLSearchParams({
			format: 'jsonv2',
			lat: String(latitude),
			lon: String(longitude),
			'accept-language': 'zh-CN'
		})
		const res = await fetch(`https://nominatim.openstreetmap.org/reverse?${params.toString()}`)
		if (!res.ok) return ''
		const data = await res.json()
		return data.display_name || data.name || ''
	} catch {
		return ''
	}
}

function resizeChart() {
  chart?.resize()
}

onMounted(() => {
  load()
  window.addEventListener('resize', resizeChart)
})

onUnmounted(() => {
  window.removeEventListener('resize', resizeChart)
  chart?.dispose()
  chart = null
})
</script>
