<template>
  <main class="mobile-page feature-page">
    <section class="feature-hero">
      <div>
        <h2>{{ config.title }}</h2>
        <p>{{ config.subtitle }}</p>
      </div>
      <el-button v-if="canCreate" type="primary" size="large" :icon="Plus" @click="openCreate">{{ createButtonText }}</el-button>
    </section>

    <section v-if="props.type === 'profile'" class="profile-page">
      <div v-loading="profileLoading" class="profile-card">
        <div class="profile-avatar">{{ profileInitial }}</div>
        <div>
          <h3>{{ authStore.user?.realName || authStore.user?.username || '当前用户' }}</h3>
          <p>账号：{{ authStore.user?.username || '-' }}</p>
          <p>用户ID：{{ authStore.user?.id || '-' }}</p>
        </div>
      </div>
      <div v-loading="merchantInfo.loading" class="profile-log-panel">
        <div class="section-title">
          <h2>商户信息设置</h2>
          <el-button type="primary" :loading="merchantInfo.saving" @click="saveMerchantInfo">保存</el-button>
        </div>
        <el-form label-width="88px" class="merchant-form">
          <el-form-item label="公司名"><el-input v-model="merchantInfo.form.companyName" placeholder="打印销售单中显示的供应商名称" /></el-form-item>
          <el-form-item label="联系人"><el-input v-model="merchantInfo.form.contactName" placeholder="联系人" /></el-form-item>
          <el-form-item label="联系方式"><el-input v-model="merchantInfo.form.contactPhone" placeholder="电话或手机" /></el-form-item>
        </el-form>
      </div>
      <div class="profile-log-panel">
        <div class="section-title">
          <h2>当前账户登录记录</h2>
          <el-button :icon="Refresh" :loading="profileLogs.loading" @click="loadProfileLoginLogs">刷新</el-button>
        </div>
        <el-table v-if="!isMobile" v-loading="profileLogs.loading" :data="profileLogs.rows" border stripe>
          <el-table-column prop="createdAt" label="时间" min-width="170" />
          <el-table-column prop="ip" label="IP" min-width="130" />
          <el-table-column label="结果" width="110">
            <template #default="{ row }"><el-tag :type="row.success ? 'success' : 'danger'">{{ row.success ? '成功' : '失败' }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="message" label="说明" min-width="160" show-overflow-tooltip />
          <el-table-column prop="userAgent" label="设备/浏览器" min-width="260" show-overflow-tooltip />
        </el-table>
        <section v-else class="mobile-card-list">
          <article v-for="row in profileLogs.rows" :key="row.id" class="customer-card">
            <div class="card-head">
              <div>
                <strong>{{ row.createdAt }}</strong>
                <span>{{ row.ip || '-' }}</span>
              </div>
              <el-tag :type="row.success ? 'success' : 'danger'">{{ row.success ? '成功' : '失败' }}</el-tag>
            </div>
            <p>{{ row.message || '登录记录' }}</p>
            <p>{{ row.userAgent || '-' }}</p>
          </article>
        </section>
        <el-empty v-if="!profileLogs.loading && profileLogs.rows.length === 0" description="暂无登录记录" />
      </div>
      <div class="profile-log-panel">
        <div class="section-title">
          <h2>工程签到历史</h2>
          <el-button :icon="Refresh" :loading="signInHistory.loading" @click="loadSignInHistory">刷新</el-button>
        </div>
        <el-table v-if="!isMobile" v-loading="signInHistory.loading" :data="signInHistory.rows" border stripe>
          <el-table-column prop="checkInAt" label="签到时间" min-width="170" />
          <el-table-column label="地址" min-width="260" show-overflow-tooltip>
            <template #default="{ row }">{{ signInAddress(row) }}</template>
          </el-table-column>
          <el-table-column label="坐标" min-width="150">
            <template #default="{ row }">{{ signInCoords(row) }}</template>
          </el-table-column>
          <el-table-column prop="device" label="设备/浏览器" min-width="260" show-overflow-tooltip />
        </el-table>
        <section v-else class="mobile-card-list">
          <article v-for="row in signInHistory.rows" :key="row.id" class="customer-card">
            <div class="card-head">
              <div>
                <strong>{{ row.checkInAt }}</strong>
                <span>{{ signInAddress(row) }}</span>
              </div>
              <el-tag type="success">已签到</el-tag>
            </div>
            <p>{{ signInCoords(row) }}</p>
            <p>{{ row.device || '-' }}</p>
          </article>
        </section>
        <el-empty v-if="!signInHistory.loading && signInHistory.rows.length === 0" description="暂无签到记录" />
        <el-pagination
          v-if="signInHistory.total > signInHistory.pageSize"
          v-model:current-page="signInHistory.page"
          v-model:page-size="signInHistory.pageSize"
          :total="signInHistory.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          class="profile-pagination"
          @current-change="loadSignInHistory"
          @size-change="reloadSignInHistory"
        />
      </div>
    </section>

    <section v-if="props.type !== 'profile'" class="toolbar">
      <div class="filters">
        <el-input v-if="props.type !== 'customer-statements'" v-model="query.keyword" clearable :placeholder="searchPlaceholder" :prefix-icon="Search" @keyup.enter="refresh" @clear="refresh" />
        <template v-if="props.type === 'inventory-movements' || props.type === 'document-delete-records'">
          <el-input v-if="props.type === 'inventory-movements'" v-model="query.productName" clearable placeholder="商品名称" @keyup.enter="refresh" @clear="refresh" />
          <el-select v-model="query.sourceType" clearable :placeholder="props.type === 'document-delete-records' ? '单据类型' : '业务类型'" @change="refresh">
            <el-option v-for="item in sourceFilterOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
          <el-input v-model="query.operatorName" clearable :placeholder="props.type === 'document-delete-records' ? '删除人' : '操作人'" @keyup.enter="refresh" @clear="refresh" />
          <el-date-picker v-model="movementDates" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始日期" end-placeholder="结束日期" @change="changeMovementDates" />
        </template>
        <template v-else-if="props.type === 'customer-statements'">
          <el-select v-model="query.customerId" filterable clearable placeholder="选择客户" @change="changeCustomerFilter">
            <el-option v-for="item in customers" :key="item.id" :label="customerOptionLabel(item)" :value="item.id" />
          </el-select>
          <el-input v-model="query.keyword" clearable placeholder="搜索对账单号" :prefix-icon="Search" @keyup.enter="refresh" @clear="refresh" />
          <el-select v-model="query.sourceType" clearable placeholder="对账状态" @change="refresh">
            <el-option v-for="item in statementStatusOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
          <el-date-picker v-model="movementDates" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始日期" end-placeholder="结束日期" @change="changeMovementDates" />
        </template>
        <el-button type="primary" :icon="Search" @click="refresh">搜索</el-button>
      </div>
    </section>

    <section v-if="props.type === 'settings'" class="settings-panel">
      <div>
        <h3>测试数据</h3>
        <p>清空销售、采购、库存、维修、工程、客户、供应商、商品和财务业务数据，并重新生成中文基础测试资料。</p>
      </div>
      <el-button type="danger" :loading="restoreLoading" @click="confirmRestoreTestData">一键恢复测试数据</el-button>
    </section>

    <section v-if="isReportPage" class="report-panel" v-loading="reportLoading">
      <div class="report-actions">
        <el-button @click="exportReportCsv">导出Excel</el-button>
        <el-button @click="printReport">打印/PDF</el-button>
      </div>
      <div class="report-card-grid">
        <article v-for="item in reportSummaryCards" :key="item.label" class="report-card">
          <span>{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
          <small>{{ item.hint }}</small>
        </article>
      </div>
      <div class="report-analysis-grid">
        <article v-for="block in reportBlocks" :key="block.title" class="report-block">
          <div class="section-title">
            <h2>{{ block.title }}</h2>
          </div>
          <div v-if="block.rows.length === 0" class="empty-line">暂无数据</div>
          <div v-for="row in block.rows" :key="row.name || row.ageBucket || row.date || row.id" class="report-row">
            <span>{{ row.name || row.ageBucket || row.date || row.productName || '-' }}</span>
            <strong>{{ reportRowValue(row, block.type) }}</strong>
          </div>
        </article>
      </div>
    </section>

    <input ref="photoInput" class="hidden-file" type="file" accept="image/*" capture="environment" @change="handlePhoto" />
    <input ref="albumInput" class="hidden-file" type="file" accept="image/*" multiple @change="handlePhoto" />

    <el-table v-if="props.type !== 'profile' && !isMobile" v-loading="loading" :data="rows" border stripe class="data-table">
      <el-table-column v-for="field in displayColumns" :key="field.prop" :prop="field.prop" :label="field.label" min-width="130" show-overflow-tooltip>
        <template #default="{ row }">
          <el-button v-if="field.type === 'businessLink' && row.businessNo" link type="primary" @click="goMovementBusiness(row)">{{ row.businessNo }}</el-button>
          <el-button v-else-if="field.type === 'salesLink' && row.salesOrderId" link type="primary" @click="openBusinessDetail('sales', row.salesOrderId)">{{ row[field.prop] }}</el-button>
          <el-button v-else-if="field.type === 'purchaseLink' && row.purchaseOrderId" link type="primary" @click="openBusinessDetail('purchase', row.purchaseOrderId)">{{ row[field.prop] }}</el-button>
          <el-button v-else-if="field.type === 'inventoryLink' && row.productCode" link type="primary" @click="openInventoryAssetDetail(row)">{{ row[field.prop] }}</el-button>
          <span v-else-if="field.prop === 'profitRate'">{{ percent(row[field.prop]) }}</span>
          <el-button v-else-if="isBusinessNoField(field.prop) && row[field.prop]" link type="primary" @click="openBusinessDetail(props.type, row.id)">{{ row[field.prop] }}</el-button>
          <span v-else-if="field.prop === 'businessType' || field.prop === 'sourceType'">{{ sourceTypeLabel(row[field.prop]) }}</span>
          <span v-else-if="isStatusField(field.prop)">{{ statusLabel(row[field.prop]) }}</span>
          <span v-else-if="field.prop === 'accountType'">{{ statusLabel(row[field.prop]) }}</span>
          <span v-else>{{ row[field.prop] ?? '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" :width="actionWidth" fixed="right">
        <template #default="{ row }">
          <template v-if="props.type === 'sales'">
            <template v-if="!query.deletedOnly">
              <el-button link type="primary" @click="openView(row)">查看</el-button>
              <el-button link type="primary" @click="salesAction(row, 'print')">打印</el-button>
              <el-button v-if="canDelete" link type="danger" @click="openDelete(row)">删除</el-button>
            </template>
            <el-tag v-else type="info">已删除</el-tag>
          </template>
          <template v-else>
            <template v-if="props.type === 'inventory'">
              <el-button link type="primary" @click="openView(row)">查看</el-button>
            </template>
            <template v-else-if="props.type === 'purchase'">
              <el-button link type="primary" @click="openView(row)">查看</el-button>
              <el-button v-if="canDelete" link type="danger" @click="openDelete(row)">删除</el-button>
            </template>
            <el-button v-else-if="props.type === 'finance'" link type="primary" @click="openView(row)">查看</el-button>
            <el-button v-else-if="props.type === 'document-delete-records'" link type="primary" @click="openView(row)">查看</el-button>
            <el-button v-else-if="props.type === 'inventory-movements'" link type="primary" @click="goMovementBusiness(row)">查看单据</el-button>
            <el-button v-else-if="props.type === 'profit-report'" link type="primary" @click="openBusinessDetail('sales', row.salesOrderId)">查看销售单</el-button>
            <el-button v-else-if="props.type === 'inventory-asset-report'" link type="primary" @click="openInventoryAssetDetail(row)">库存详情</el-button>
            <template v-else-if="props.type === 'receivables'">
              <el-button v-if="isAccountSettled(row)" link type="primary" @click="openView(row)">查看</el-button>
              <el-button v-else link type="primary" @click="openSettlement(row)">收款</el-button>
              <el-button link type="primary" @click="printSettlement(row)">打印</el-button>
            </template>
            <template v-else-if="props.type === 'customer-statements'">
              <el-button link type="primary" @click="openView(row)">查看</el-button>
              <el-button v-if="row.status === 'unconfirmed'" link type="primary" @click="confirmCustomerStatement(row)">确认</el-button>
              <el-button v-if="row.status === 'confirmed'" link type="primary" @click="openSettlement(row)">结算</el-button>
              <el-button link type="primary" @click="printCustomerStatement(row)">打印</el-button>
            </template>
            <template v-else-if="props.type === 'payables'">
              <el-button v-if="isAccountSettled(row)" link type="primary" @click="openView(row)">查看</el-button>
              <el-button v-else link type="primary" @click="openSettlement(row)">付款</el-button>
              <el-button link type="primary" @click="printSettlement(row)">打印</el-button>
            </template>
            <el-button v-else-if="props.type === 'repair'" link type="primary" @click="openView(row)">查看</el-button>
            <el-button v-else link type="primary" :icon="EditPen" @click="openEdit(row)">编辑</el-button>
            <template v-if="props.type === 'repair'">
              <el-button link type="primary" @click="repairPrint(row, 'repair-print')">维修单</el-button>
              <el-button link type="primary" @click="repairPrint(row, 'quote-print')">报价单</el-button>
              <el-button link type="primary" @click="repairPrint(row, 'settlement-print')">结算单</el-button>
            </template>
            <el-button v-if="supportsPhotos" link type="primary" :icon="Camera" @click="openPhotos(row)">照片</el-button>
          </template>
          <el-button v-if="canDelete && !['sales', 'purchase'].includes(props.type)" link type="danger" @click="openDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <section v-else-if="props.type !== 'profile'" class="mobile-card-list">
      <el-skeleton v-if="loading" :rows="4" animated />
      <article v-for="row in rows" :key="row.id" class="customer-card">
        <div class="card-head">
          <div>
            <strong>{{ primaryText(row) }}</strong>
            <span>{{ secondaryText(row) }}</span>
          </div>
          <el-tag>{{ statusText(row) }}</el-tag>
        </div>
        <div class="card-meta">
          <span>{{ amountText(row) }}</span>
          <span>{{ dateText(row) }}</span>
        </div>
        <p>{{ row.remark || '暂无备注' }}</p>
        <div class="card-actions">
          <template v-if="props.type === 'sales'">
            <template v-if="!query.deletedOnly">
              <el-button size="large" @click="openView(row)">查看</el-button>
              <el-button size="large" @click="salesAction(row, 'print')">打印</el-button>
              <el-button v-if="canDelete" size="large" type="danger" @click="openDelete(row)">删除</el-button>
            </template>
            <el-tag v-else type="info">已删除</el-tag>
          </template>
          <template v-else>
            <template v-if="props.type === 'inventory'">
              <el-button size="large" @click="openView(row)">查看</el-button>
            </template>
            <template v-else-if="props.type === 'purchase'">
              <el-button size="large" @click="openView(row)">查看</el-button>
              <el-button v-if="canDelete" size="large" type="danger" @click="openDelete(row)">删除</el-button>
            </template>
            <el-button v-else-if="props.type === 'finance'" size="large" @click="openView(row)">查看</el-button>
            <el-button v-else-if="props.type === 'document-delete-records'" size="large" @click="openView(row)">查看</el-button>
            <el-button v-else-if="props.type === 'inventory-movements'" size="large" @click="goMovementBusiness(row)">查看单据</el-button>
            <el-button v-else-if="props.type === 'profit-report'" size="large" @click="openBusinessDetail('sales', row.salesOrderId)">查看销售单</el-button>
            <el-button v-else-if="props.type === 'inventory-asset-report'" size="large" @click="openInventoryAssetDetail(row)">库存详情</el-button>
            <template v-else-if="props.type === 'receivables'">
              <el-button v-if="isAccountSettled(row)" size="large" @click="openView(row)">查看</el-button>
              <el-button v-else size="large" @click="openSettlement(row)">收款</el-button>
              <el-button size="large" @click="printSettlement(row)">打印</el-button>
            </template>
            <template v-else-if="props.type === 'customer-statements'">
              <el-button size="large" @click="openView(row)">查看</el-button>
              <el-button v-if="row.status === 'unconfirmed'" size="large" @click="confirmCustomerStatement(row)">确认</el-button>
              <el-button v-if="row.status === 'confirmed'" size="large" @click="openSettlement(row)">结算</el-button>
              <el-button size="large" @click="printCustomerStatement(row)">打印</el-button>
            </template>
            <template v-else-if="props.type === 'payables'">
              <el-button v-if="isAccountSettled(row)" size="large" @click="openView(row)">查看</el-button>
              <el-button v-else size="large" @click="openSettlement(row)">付款</el-button>
              <el-button size="large" @click="printSettlement(row)">打印</el-button>
            </template>
            <el-button v-else size="large" :icon="EditPen" @click="openEdit(row)">编辑</el-button>
            <template v-if="props.type === 'repair'">
              <el-button size="large" @click="repairPrint(row, 'repair-print')">维修单</el-button>
              <el-button size="large" @click="repairPrint(row, 'quote-print')">报价单</el-button>
              <el-button size="large" @click="repairPrint(row, 'settlement-print')">结算单</el-button>
            </template>
            <el-button v-if="supportsPhotos" size="large" :icon="Camera" @click="openPhotos(row)">照片</el-button>
            <el-button v-if="canDelete && !['sales', 'purchase'].includes(props.type)" size="large" type="danger" @click="openDelete(row)">删除</el-button>
          </template>
        </div>
      </article>
      <el-empty v-if="!loading && rows.length === 0" :description="props.type === 'customer-statements' ? '暂无客户对账单，可先选择客户生成' : '暂无业务数据'">
        <el-button v-if="props.type === 'customer-statements'" type="primary" @click="openStatementDialog">生成对账单</el-button>
      </el-empty>
      <el-button v-if="rows.length < total" class="load-more" :loading="loadingMore" @click="loadMore">加载更多</el-button>
    </section>

    <el-pagination
      v-if="props.type !== 'profile' && !isMobile"
      v-model:current-page="query.page"
      v-model:page-size="query.pageSize"
      :total="total"
      layout="total, sizes, prev, pager, next"
      @current-change="refresh"
      @size-change="refresh"
    />

    <el-dialog v-model="dialog.visible" :title="dialog.editing ? '编辑' : '新建'" width="620px">
      <el-form label-width="96px">
        <el-form-item v-for="field in visibleFormFields" :key="field.prop" :label="field.label">
          <el-select v-if="field.type === 'customer'" v-model="form.customerId" filterable clearable placeholder="选择客户" style="width: 100%" @change="selectCustomer">
            <el-option v-for="item in customers" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
          <el-select v-else-if="field.type === 'supplier'" v-model="form.supplierId" filterable clearable placeholder="选择供应商" style="width: 100%" @change="selectSupplier">
            <el-option v-for="item in suppliers" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
          <el-select
            v-else-if="field.type === 'product'"
            v-model="form.productId"
            filterable
            clearable
            allow-create
            default-first-option
            placeholder="选择或输入商品名称"
            style="width: 100%"
            @change="selectProduct"
          >
            <el-option v-for="item in products" :key="item.id" :label="productLabel(item)" :value="item.id" />
          </el-select>
          <div v-else-if="field.type === 'purchaseSource'" class="purchase-source-picker">
            <el-select v-model="form.inventoryBatchIds" multiple filterable collapse-tags collapse-tags-tooltip placeholder="选择采购来源" style="width: 100%" @change="selectPurchaseSourceIds">
              <el-option v-for="item in purchaseSources" :key="item.id" :label="purchaseSourceLabel(item)" :value="item.id" />
            </el-select>
            <el-table v-if="purchaseSources.length" ref="sourceTableRef" :data="purchaseSources" border class="source-table" row-key="id" @selection-change="selectPurchaseSourceRows" @row-click="selectPurchaseSourceRow">
              <el-table-column type="selection" width="48" />
              <el-table-column prop="purchaseOrderNo" label="采购单号" min-width="150" />
              <el-table-column prop="supplierName" label="供应商" min-width="150" />
              <el-table-column prop="purchaseDate" label="采购日期" min-width="140" />
              <el-table-column prop="productName" label="商品名称" min-width="150" />
              <el-table-column prop="inboundQuantity" label="采购数量" width="100" />
              <el-table-column prop="soldQuantity" label="已销售" width="90" />
              <el-table-column prop="remainingQuantity" label="剩余" width="90" />
              <el-table-column prop="purchasePrice" label="采购成本" width="110" />
              <el-table-column label="库存状态" width="110">
                <template #default="{ row }">{{ statusLabel(row.status) }}</template>
              </el-table-column>
            </el-table>
            <el-empty v-else description="选择商品后显示可用采购库存" />
          </div>
          <el-select v-else-if="field.type === 'select'" v-model="form[field.prop]" clearable filterable style="width: 100%">
            <el-option v-for="item in field.options" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
          <el-input-number v-else-if="field.type === 'number'" v-model="form[field.prop]" :min="0" style="width: 100%" />
          <el-date-picker v-else-if="field.type === 'date'" v-model="form[field.prop]" value-format="YYYY-MM-DDTHH:mm:ssZ" type="datetime" style="width: 100%" />
          <el-input v-else v-model="form[field.prop]" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="3" />
        </el-form-item>
        <template v-if="props.type === 'repair'">
          <el-form-item label="维修配件">
            <div class="inline-editor">
              <div v-for="(row, index) in form.repairParts" :key="index" class="repair-line">
                <div class="repair-line-head">
                  <strong>配件 {{ index + 1 }}</strong>
                  <el-button link type="danger" @click="removeRepairPart(index)">删除</el-button>
                </div>
                <div class="repair-line-grid">
                  <label>
                    <span>配件</span>
                    <el-select v-model="row.productId" filterable clearable placeholder="选择有库存配件" @change="(id) => selectRepairPart(row, id)">
                      <el-option v-for="item in inventoryProducts" :key="item.id" :label="inventoryProductLabel(item)" :value="item.productId || item.id" />
                    </el-select>
                  </label>
                  <label>
                    <span>库存批次</span>
                    <el-select v-model="row.inventoryBatchId" filterable clearable placeholder="选择批次" @change="(id) => selectRepairPartBatch(row, id)">
                      <el-option v-for="item in row.purchaseSources || []" :key="item.id" :label="purchaseSourceLabel(item)" :value="item.id" />
                    </el-select>
                  </label>
                  <label>
                    <span>数量</span>
                    <el-input-number v-model="row.quantity" :min="1" />
                  </label>
                  <label>
                    <span>报价</span>
                    <el-input-number v-model="row.price" :min="0" />
                  </label>
                </div>
              </div>
              <el-button :icon="Plus" @click="addRepairPart">新增配件</el-button>
            </div>
          </el-form-item>
          <el-form-item label="外协费用">
            <div class="inline-editor">
              <div v-for="(row, index) in form.outsourceItems" :key="index" class="repair-line">
                <div class="repair-line-head">
                  <strong>外协 {{ index + 1 }}</strong>
                  <el-button link type="danger" @click="removeRepairOutsource(index)">删除</el-button>
                </div>
                <div class="repair-line-grid outsource-grid">
                  <label>
                    <span>服务商</span>
                    <el-select v-model="row.supplierId" filterable clearable placeholder="选择供应商" @change="(id) => selectOutsourceSupplier(row, id)">
                      <el-option v-for="item in outsourceSuppliers" :key="item.id" :label="item.name" :value="item.id" />
                    </el-select>
                  </label>
                  <label>
                    <span>项目</span>
                    <el-input v-model="row.serviceProject" placeholder="上门维修" />
                  </label>
                  <label>
                    <span>金额</span>
                    <el-input-number v-model="row.amount" :min="0" />
                  </label>
                </div>
              </div>
              <el-button :icon="Plus" @click="addRepairOutsource">新增外协</el-button>
            </div>
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="photoDialog.visible" :title="photoTitle" width="720px">
      <div class="photo-toolbar">
        <el-button type="primary" :icon="Camera" @click="triggerPhoto('camera')">拍照</el-button>
        <el-button :icon="UploadFilled" @click="triggerPhoto('album')">相册</el-button>
      </div>
      <section class="photo-grid">
        <a v-for="item in photos" :key="item.id || item.fileUrl" :href="apiOrigin + item.fileUrl" target="_blank" class="photo-card">
          <img :src="apiOrigin + item.fileUrl" :alt="item.fileName" />
          <span>{{ item.scene }}</span>
        </a>
      </section>
      <el-empty v-if="photos.length === 0" description="暂无照片记录" />
    </el-dialog>

    <el-dialog v-model="detailDialog.visible" :title="detailTitle" width="980px" class="sales-detail-dialog">
      <section v-if="detail && detailType === 'inventory'" class="sales-detail inventory-detail">
        <div class="detail-section">
          <h3>商品基础信息</h3>
          <div class="detail-grid">
            <span>商品编码：{{ detail.productCode || '-' }}</span>
            <span>商品名称：{{ detail.productName || '-' }}</span>
            <span>规格型号：{{ detail.spec || '-' }}</span>
            <span>品牌：{{ detail.brand || '-' }}</span>
            <span>单位：{{ detail.unit || '-' }}</span>
            <span>当前库存：{{ detail.quantity || 0 }}</span>
            <span>可销售库存：{{ detail.availableQuantity || detail.quantity || 0 }}</span>
            <span>库存金额：{{ money(detail.amount) }}</span>
          </div>
        </div>
        <el-tabs v-model="inventoryTab.active" class="inventory-tabs" @tab-change="loadInventoryTab">
          <el-tab-pane label="采购来源" name="purchaseSources">
            <div class="tab-filter">
              <el-input v-model="inventoryTab.filters.purchaseSources.keyword" clearable placeholder="搜索采购单号、供应商" @change="reloadInventoryTab" />
              <el-select v-model="inventoryTab.filters.purchaseSources.status" clearable placeholder="批次状态" @change="reloadInventoryTab">
                <el-option label="未销售" value="未销售" />
                <el-option label="销售中" value="销售中" />
                <el-option label="已销售完成" value="已销售完成" />
              </el-select>
              <el-date-picker v-model="inventoryTab.filters.purchaseSources.dates" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始日期" end-placeholder="结束日期" @change="reloadInventoryTab" />
            </div>
            <el-table :data="inventoryTab.purchaseSources.list" border @sort-change="onInventoryTabSort">
              <el-table-column label="采购单号" min-width="160">
                <template #default="{ row }"><el-button link type="primary" @click="openBusinessDetail('purchase', row.purchaseOrderId)">{{ row.purchaseOrderNo || '-' }}</el-button></template>
              </el-table-column>
              <el-table-column prop="supplierName" label="供应商" min-width="150" />
              <el-table-column prop="purchaseDate" label="采购日期" min-width="150" sortable="custom" />
              <el-table-column prop="inboundQuantity" label="采购数量" width="110" sortable="custom" />
              <el-table-column prop="purchasePrice" label="采购成本" width="110" sortable="custom" />
              <el-table-column prop="soldQuantity" label="已销售" width="90" />
              <el-table-column prop="remainingQuantity" label="剩余数量" width="110" sortable="custom" />
              <el-table-column label="批次状态" width="120">
                <template #default="{ row }">{{ statusLabel(row.status) }}</template>
              </el-table-column>
            </el-table>
            <el-pagination v-model:current-page="inventoryTab.purchaseSources.page" v-model:page-size="inventoryTab.purchaseSources.pageSize" :total="inventoryTab.purchaseSources.total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next" @current-change="loadInventoryTab" @size-change="reloadInventoryTab" />
          </el-tab-pane>
          <el-tab-pane label="销售追踪" name="salesTrace">
            <div class="tab-filter">
              <el-input v-model="inventoryTab.filters.salesTrace.keyword" clearable placeholder="搜索销售单号、客户、采购单号" @change="reloadInventoryTab" />
              <el-input v-model="inventoryTab.filters.salesTrace.customerName" clearable placeholder="客户" @change="reloadInventoryTab" />
              <el-date-picker v-model="inventoryTab.filters.salesTrace.dates" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始日期" end-placeholder="结束日期" @change="reloadInventoryTab" />
            </div>
            <el-table :data="inventoryTab.salesTrace.list" border @sort-change="onInventoryTabSort">
              <el-table-column label="销售单号" min-width="160">
                <template #default="{ row }"><el-button link type="primary" @click="openBusinessDetail('sales', row.salesOrderId)">{{ row.salesOrderNo || '-' }}</el-button></template>
              </el-table-column>
              <el-table-column prop="customerName" label="客户" min-width="160" />
              <el-table-column prop="orderDate" label="销售日期" min-width="150" sortable="custom" />
              <el-table-column prop="quantity" label="销售数量" width="110" />
              <el-table-column prop="salesPrice" label="销售价格" width="110" />
              <el-table-column label="采购来源" min-width="160">
                <template #default="{ row }">
                  <el-button v-if="row.purchaseOrderNo" link type="primary" @click="openBusinessDetail('purchase', row.purchaseOrderId)">{{ row.purchaseOrderNo }}</el-button>
                  <span v-else>-</span>
                </template>
              </el-table-column>
              <el-table-column prop="costPrice" label="采购成本" width="110" />
              <el-table-column prop="profitAmount" label="销售利润" width="120" sortable="custom" />
            </el-table>
            <el-pagination v-model:current-page="inventoryTab.salesTrace.page" v-model:page-size="inventoryTab.salesTrace.pageSize" :total="inventoryTab.salesTrace.total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next" @current-change="loadInventoryTab" @size-change="reloadInventoryTab" />
          </el-tab-pane>
          <el-tab-pane label="库存流水" name="inventoryMovements">
            <div class="tab-filter">
              <el-input v-model="inventoryTab.filters.inventoryMovements.keyword" clearable placeholder="搜索流水号、单号" @change="reloadInventoryTab" />
              <el-select v-model="inventoryTab.filters.inventoryMovements.sourceType" clearable placeholder="业务类型" @change="reloadInventoryTab">
                <el-option label="采购入库" value="purchase" />
                <el-option label="销售出库" value="sales" />
                <el-option label="维修领料" value="repair" />
                <el-option label="工程领料" value="project" />
              </el-select>
              <el-date-picker v-model="inventoryTab.filters.inventoryMovements.dates" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始日期" end-placeholder="结束日期" @change="reloadInventoryTab" />
            </div>
            <el-table :data="inventoryTab.inventoryMovements.list" border @sort-change="onInventoryTabSort">
              <el-table-column prop="productCode" label="商品编码" min-width="130" />
              <el-table-column prop="productName" label="商品名称" min-width="150" />
              <el-table-column prop="spec" label="规格型号" min-width="120" />
              <el-table-column label="业务类型" width="110">
                <template #default="{ row }">{{ sourceTypeLabel(row.businessType || row.sourceType) }}</template>
              </el-table-column>
              <el-table-column label="业务单号" min-width="160">
                <template #default="{ row }"><el-button v-if="row.businessNo" link type="primary" @click="goMovementBusiness(row)">{{ row.businessNo }}</el-button><span v-else>-</span></template>
              </el-table-column>
              <el-table-column prop="quantityChange" label="数量变化" width="110" sortable="custom" />
              <el-table-column prop="beforeQuantity" label="变更前库存" width="120" />
              <el-table-column prop="afterQuantity" label="变更后库存" width="120" />
              <el-table-column prop="operatorName" label="操作人" width="120" />
              <el-table-column prop="occurredAt" label="操作时间" min-width="150" sortable="custom" />
            </el-table>
            <el-pagination v-model:current-page="inventoryTab.inventoryMovements.page" v-model:page-size="inventoryTab.inventoryMovements.pageSize" :total="inventoryTab.inventoryMovements.total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next" @current-change="loadInventoryTab" @size-change="reloadInventoryTab" />
          </el-tab-pane>
        </el-tabs>
      </section>
      <section v-else-if="detail && detailType === 'purchase'" class="sales-detail">
        <div class="detail-section">
          <h3>采购信息</h3>
          <div class="detail-grid">
            <span>采购单号：{{ detail.orderNo }}</span>
            <span>采购日期：{{ detail.orderDate || '-' }}</span>
            <span>供应商：{{ detail.supplierName || '-' }}</span>
            <span>采购员：{{ detail.buyerName || detail.createdBy || '-' }}</span>
            <span>采购状态：{{ statusLabel(detail.status) }}</span>
            <span>采购金额：{{ money(detail.totalAmount) }}</span>
            <span>已付款：{{ money(detail.paidAmount) }}</span>
            <span>待付款：{{ money(detail.payableAmount) }}</span>
            <span>备注：{{ detail.remark || '-' }}</span>
          </div>
        </div>
        <div class="detail-section">
          <h3>商品明细</h3>
          <el-table :data="detail.items || []" border>
            <el-table-column prop="productCode" label="商品编码" min-width="130" />
            <el-table-column prop="productName" label="商品名称" min-width="180" />
            <el-table-column prop="spec" label="规格型号" min-width="120" />
            <el-table-column prop="quantity" label="采购数量" width="110" />
            <el-table-column prop="price" label="采购单价" width="110" />
            <el-table-column prop="amount" label="采购金额" width="120" />
          </el-table>
        </div>
        <div class="detail-section">
          <h3>入库批次</h3>
          <el-table :data="detail.inventoryBatches || []" border>
            <el-table-column prop="batchNo" label="批次号" min-width="160" />
            <el-table-column prop="productName" label="商品" min-width="180" />
            <el-table-column prop="supplierName" label="供应商" min-width="150" />
            <el-table-column prop="purchaseDate" label="采购日期" min-width="150" />
            <el-table-column prop="purchasePrice" label="采购价" width="110" />
            <el-table-column prop="inboundQuantity" label="入库数量" width="110" />
          </el-table>
        </div>
      </section>
      <section v-else-if="detail && detailType === 'sales'" class="sales-detail">
        <div class="detail-section">
          <h3>客户信息</h3>
          <div class="detail-grid">
            <span>销售单号：{{ detail.orderNo }}</span>
            <span>客户：{{ detail.customerName || '-' }}</span>
            <span>销售时间：{{ detail.orderDate || '-' }}</span>
            <span>状态：{{ statusLabel(detail.status) }}</span>
            <span>销售金额：{{ money(detail.totalAmount) }}</span>
            <span>成本金额：{{ money(detail.costAmount) }}</span>
            <span>利润金额：{{ money(detail.profitAmount) }}</span>
            <span>毛利率：{{ percent(detail.profitRate) }}</span>
          </div>
        </div>
        <div class="detail-section">
          <h3>商品明细</h3>
          <el-table :data="detail.items || []" border>
            <el-table-column prop="productCode" label="商品编码" min-width="130" />
            <el-table-column prop="productName" label="商品名称" min-width="160" />
            <el-table-column prop="quantity" label="数量" width="90" />
            <el-table-column prop="price" label="销售价" width="100" />
            <el-table-column label="采购单" min-width="150">
              <template #default="{ row }">
                <el-button v-if="row.purchaseOrderNo" link type="primary" @click="openBusinessDetail('purchase', row.purchaseOrderId)">{{ row.purchaseOrderNo }}</el-button>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column prop="costPrice" label="成本" width="110" />
            <el-table-column prop="profitAmount" label="利润" width="110" />
            <el-table-column prop="supplierName" label="供应商" min-width="150" />
            <el-table-column prop="purchaseDate" label="采购日期" min-width="150" />
            <el-table-column prop="purchasePrice" label="采购价" width="100" />
          </el-table>
        </div>
      </section>
      <section v-else-if="detail && detailType === 'customer-statements'" class="sales-detail">
        <div class="detail-section">
          <h3>客户信息</h3>
          <div class="detail-grid">
            <span>对账单号：{{ detail.statementNo || '-' }}</span>
            <span>客户：{{ detail.customerName || '-' }}</span>
            <span>联系人：{{ detail.contactName || '-' }}</span>
            <span>联系电话：{{ detail.contactPhone || '-' }}</span>
            <span>对账期间：{{ detail.startDate || '-' }} 至 {{ detail.endDate || '-' }}</span>
            <span>状态：{{ statusLabel(detail.status) }}</span>
            <span>确认时间：{{ detail.confirmedAt || '-' }}</span>
            <span>结算时间：{{ detail.settledAt || '-' }}</span>
          </div>
        </div>
        <div class="detail-section">
          <h3>汇总</h3>
          <div class="detail-grid">
            <span>期间销售金额：{{ money(detail.totalAmount) }}</span>
            <span>已收金额：{{ money(detail.receivedAmount) }}</span>
            <span>本期应收金额：{{ money(detail.unpaidAmount) }}</span>
            <span>累计欠款：{{ money(detail.cumulativeDebt) }}</span>
          </div>
        </div>
        <div class="detail-section">
          <h3>销售明细</h3>
          <el-table :data="detail.items || []" border>
            <el-table-column label="销售单号" min-width="150">
              <template #default="{ row }">
                <el-button v-if="row.saleId" link type="primary" @click="openBusinessDetail('sales', row.saleId)">{{ row.saleNo }}</el-button>
                <span v-else>{{ row.saleNo || '-' }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="saleDate" label="销售日期" min-width="130" />
            <el-table-column prop="productName" label="商品名称" min-width="180" />
            <el-table-column prop="quantity" label="数量" width="90" />
            <el-table-column prop="totalAmount" label="销售金额" width="120" />
            <el-table-column prop="receivedAmount" label="已收金额" width="120" />
            <el-table-column prop="unpaidAmount" label="未收金额" width="120" />
            <el-table-column label="结算状态" width="120">
              <template #default="{ row }">{{ statusLabel(row.settlementStatus) }}</template>
            </el-table-column>
          </el-table>
        </div>
      </section>
      <section v-else-if="detail && (detailType === 'receivables' || detailType === 'payables')" class="sales-detail">
        <div class="detail-section">
          <h3>{{ detailType === 'receivables' ? '应收账款' : '应付账款' }}</h3>
          <div class="detail-grid">
            <span>{{ detailType === 'receivables' ? '应收单号' : '应付单号' }}：{{ detail.receivableNo || detail.payableNo || '-' }}</span>
            <span>{{ detailType === 'receivables' ? '客户' : '供应商' }}：{{ detail.customerName || detail.supplierName || '-' }}</span>
            <span>来源单号：
              <el-button v-if="detail.sourceId" link type="primary" @click="openBusinessDetail(billSourceModule(detail), detail.sourceId)">{{ detail.sourceNo || '-' }}</el-button>
              <span v-else>{{ detail.sourceNo || '-' }}</span>
            </span>
            <span>{{ detailType === 'receivables' ? '应收金额' : '应付金额' }}：{{ money(detail.totalAmount) }}</span>
            <span>{{ detailType === 'receivables' ? '已收金额' : '已付金额' }}：{{ money(detail.receivedAmount || detail.paidAmount) }}</span>
            <span>余额：{{ money(detail.balanceAmount) }}</span>
            <span>单据日期：{{ detail.invoiceDate || detail.billDate || '-' }}</span>
            <span>到期日期：{{ detail.dueDate || '-' }}</span>
            <span>结算方式：{{ statusLabel(detail.settlementMode) }}</span>
            <span>状态：{{ statusLabel(detail.status) }}</span>
          </div>
        </div>
        <div class="detail-section">
          <h3>账单摘要</h3>
          <el-table :data="[detail]" border>
            <el-table-column :prop="detailType === 'receivables' ? 'receivableNo' : 'payableNo'" label="账单号" min-width="150" />
            <el-table-column :prop="detailType === 'receivables' ? 'customerName' : 'supplierName'" :label="detailType === 'receivables' ? '客户' : '供应商'" min-width="160" />
            <el-table-column prop="sourceNo" label="来源单号" min-width="150" />
            <el-table-column prop="totalAmount" :label="detailType === 'receivables' ? '应收金额' : '应付金额'" width="120" />
            <el-table-column :prop="detailType === 'receivables' ? 'receivedAmount' : 'paidAmount'" :label="detailType === 'receivables' ? '已收' : '已付'" width="110" />
            <el-table-column prop="balanceAmount" label="余额" width="110" />
            <el-table-column prop="dueDate" label="到期日" min-width="130" />
            <el-table-column label="状态" width="110">
              <template #default="{ row }">{{ statusLabel(row.status) }}</template>
            </el-table-column>
          </el-table>
        </div>
      </section>
      <section v-else-if="detail && detailType === 'document-delete-records'" class="sales-detail">
        <div class="detail-section">
          <h3>删除信息</h3>
          <div class="detail-grid">
            <span>单据编号：{{ detail.documentNo || '-' }}</span>
            <span>单据类型：{{ detail.documentType || '-' }}</span>
            <span>删除人：{{ detail.deleteUserName || detail.deleteUserId || '-' }}</span>
            <span>删除时间：{{ detail.deleteTime || '-' }}</span>
            <span>删除理由：{{ detail.deleteReason || '-' }}</span>
            <span>状态：{{ detail.deleteStatus || '-' }}</span>
            <span>库存处理：{{ detail.stockProcessed ? '已处理' : '未处理' }}</span>
            <span>财务处理：{{ detail.financeProcessed ? '已处理' : '未处理' }}</span>
            <span>IP：{{ detail.ipAddress || '-' }}</span>
          </div>
        </div>
        <div class="detail-section">
          <h3>影响明细</h3>
          <el-table :data="detail.details || []" border>
            <el-table-column prop="detailType" label="类型" width="100" />
            <el-table-column prop="skuCode" label="SKU" min-width="120" />
            <el-table-column prop="skuName" label="商品" min-width="160" />
            <el-table-column prop="warehouse" label="仓库" min-width="120" />
            <el-table-column prop="quantity" label="数量" width="100" />
            <el-table-column prop="stockChange" label="库存变化" width="110" />
            <el-table-column prop="financeNo" label="财务流水" min-width="140" />
            <el-table-column prop="amount" label="金额" width="110" />
            <el-table-column prop="remark" label="备注" min-width="160" />
          </el-table>
        </div>
        <div class="detail-section">
          <h3>库存变化记录</h3>
          <el-table :data="detail.stockMovements || []" border>
            <el-table-column prop="movementNo" label="流水号" min-width="150" />
            <el-table-column prop="productCode" label="SKU" min-width="120" />
            <el-table-column prop="productName" label="商品" min-width="160" />
            <el-table-column prop="warehouse" label="仓库" min-width="120" />
            <el-table-column prop="direction" label="方向" width="90" />
            <el-table-column prop="quantity" label="数量" width="100" />
            <el-table-column prop="occurredAt" label="时间" min-width="150" />
          </el-table>
        </div>
        <div class="detail-section">
          <h3>财务处理记录</h3>
          <el-table :data="detail.financeRecords || []" border>
            <el-table-column prop="recordNo" label="流水号" min-width="150" />
            <el-table-column prop="recordType" label="类型" width="100" />
            <el-table-column prop="accountName" label="账户" min-width="120" />
            <el-table-column prop="targetName" label="对象" min-width="140" />
            <el-table-column prop="amount" label="金额" width="110" />
            <el-table-column prop="businessType" label="业务类型" min-width="140" />
            <el-table-column prop="occurredAt" label="时间" min-width="150" />
          </el-table>
        </div>
        <div class="detail-section">
          <h3>删除前完整数据</h3>
          <pre class="json-preview">{{ prettyJSON(detail.beforeData) }}</pre>
        </div>
      </section>
      <section v-else-if="detail && detailType === 'finance'" class="sales-detail">
        <div class="detail-section">
          <h3>资金流水</h3>
          <div class="detail-grid">
            <span>流水号：{{ detail.recordNo || '-' }}</span>
            <span>类型：{{ statusLabel(detail.recordType) }}</span>
            <span>资金账户：{{ detail.accountName || '-' }}</span>
            <span>往来对象：{{ detail.targetName || '-' }}</span>
            <span>金额：{{ money(detail.amount) }}</span>
            <span>来源：{{ sourceTypeLabel(detail.businessType || detail.sourceType) }}</span>
            <span>业务单号：{{ detail.businessNo || '-' }}</span>
            <span>状态：{{ statusLabel(detail.status) }}</span>
            <span>发生时间：{{ detail.occurredAt || '-' }}</span>
          </div>
        </div>
        <div v-if="detail.purchaseOrder" class="detail-section">
          <h3>关联采购订单</h3>
          <div class="detail-grid">
            <span>采购单号：<el-button link type="primary" @click="openBusinessDetail('purchase', detail.purchaseOrder.id)">{{ detail.purchaseOrder.orderNo }}</el-button></span>
            <span>供应商：{{ detail.purchaseOrder.supplierName || '-' }}</span>
            <span>采购时间：{{ detail.purchaseOrder.orderDate || '-' }}</span>
            <span>采购金额：{{ money(detail.purchaseOrder.totalAmount) }}</span>
            <span>已付款：{{ money(detail.purchaseOrder.paidAmount) }}</span>
            <span>待付款：{{ money(detail.purchaseOrder.payableAmount) }}</span>
            <span>状态：{{ statusLabel(detail.purchaseOrder.status) }}</span>
          </div>
        </div>
        <div v-if="detail.salesOrder" class="detail-section">
          <h3>关联销售订单</h3>
          <div class="detail-grid">
            <span>销售单号：<el-button link type="primary" @click="openBusinessDetail('sales', detail.salesOrder.id)">{{ detail.salesOrder.orderNo }}</el-button></span>
            <span>客户：{{ detail.salesOrder.customerName || '-' }}</span>
            <span>销售时间：{{ detail.salesOrder.orderDate || '-' }}</span>
            <span>销售金额：{{ money(detail.salesOrder.totalAmount) }}</span>
            <span>已收款：{{ money(detail.salesOrder.receivedAmount) }}</span>
            <span>应收款：{{ money(detail.salesOrder.receivableAmount) }}</span>
            <span>利润：{{ money(detail.salesOrder.profitAmount) }}</span>
            <span>状态：{{ statusLabel(detail.salesOrder.status) }}</span>
          </div>
        </div>
      </section>
      <section v-else-if="detail && (detailType === 'repair' || detailType === 'project')" class="sales-detail">
        <div class="detail-section">
          <h3>{{ detailType === 'repair' ? '维修单信息' : '工程项目信息' }}</h3>
          <div class="detail-grid">
            <span>单据编号：{{ detail.orderNo || detail.projectNo || '-' }}</span>
            <span>名称：{{ detail.name || detail.deviceName || detail.productName || '-' }}</span>
            <span>客户：{{ detail.customerName || '-' }}</span>
            <span>状态：{{ statusLabel(detail.status || detail.repairStatus) }}</span>
            <span>开始时间：{{ detail.startDate || detail.registeredAt || '-' }}</span>
            <span>结束时间：{{ detail.endDate || '-' }}</span>
            <span>收入：{{ money(detail.totalAmount || detail.settleAmount || detail.serviceAmount) }}</span>
            <span v-if="detailType === 'repair'">配件收入：{{ money(detail.partsAmount) }}</span>
            <span v-if="detailType === 'repair'">配件成本：{{ money(detail.partsCost) }}</span>
            <span v-if="detailType === 'repair'">外协支出：{{ money(detail.outsourceCost) }}</span>
            <span v-if="detailType === 'repair'">总成本：{{ money(detail.costAmount) }}</span>
            <span v-if="detailType === 'repair'">利润：{{ money(detail.profitAmount) }}</span>
            <span>备注：{{ detail.remark || detail.faultDesc || '-' }}</span>
          </div>
        </div>
        <div v-if="detail.items?.length" class="detail-section">
          <h3>明细</h3>
          <el-table :data="detail.items" border>
            <el-table-column prop="productCode" label="商品编码" min-width="120" />
            <el-table-column prop="productName" label="商品名称" min-width="160" />
            <el-table-column prop="quantity" label="数量" width="100" />
            <el-table-column prop="price" label="单价" width="110" />
            <el-table-column prop="amount" label="金额" width="110" />
            <el-table-column prop="supplierName" label="供应商" min-width="140" />
            <el-table-column prop="costAmount" label="成本" width="110" />
          </el-table>
        </div>
      </section>
      <template #footer>
        <el-button @click="detailDialog.visible = false">关闭</el-button>
        <el-button v-if="detailType === 'purchase'" type="primary" @click="printDetail('purchase')">打印采购单</el-button>
        <el-button v-if="detailType === 'sales'" type="primary" @click="salesAction(detail, 'print')">打印</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="deleteDialog.visible" :title="`删除${config.title}`" width="480px">
      <el-form label-width="88px">
        <el-form-item label="单据编号">
          <el-input :model-value="deleteDialogNo" disabled />
        </el-form-item>
        <el-form-item label="删除原因" required>
          <el-input v-model.trim="deleteDialog.reason" type="textarea" :rows="4" maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="deleteDialog.visible = false">取消</el-button>
        <el-button type="danger" :loading="deleteDialog.saving" @click="confirmDelete">确认删除</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="settlementDialog.visible" :title="settlementTitle" width="520px">
      <el-form label-width="96px">
        <el-form-item :label="props.type === 'payables' ? '供应商' : '客户'">
          <el-input v-model="settlementForm.targetName" disabled />
        </el-form-item>
        <el-form-item :label="props.type === 'customer-statements' ? '对账单号' : '来源单号'">
          <el-input v-model="settlementForm.sourceNo" disabled />
        </el-form-item>
        <el-form-item label="账户">
          <el-select v-model="settlementForm.accountName" filterable style="width: 100%">
            <el-option v-for="item in financeAccounts" :key="item.id || item.name" :label="financeAccountLabel(item)" :value="item.name" />
          </el-select>
        </el-form-item>
        <el-form-item :label="props.type === 'payables' ? '付款金额' : '收款金额'">
          <el-input-number v-model="settlementForm.amount" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="settlementForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="settlementDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="settlementDialog.saving" @click="submitSettlement">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="statementDialog.visible" title="生成客户对账单" width="980px" class="statement-dialog">
      <el-form label-width="84px" class="statement-filter">
        <el-form-item label="客户" required>
          <el-select v-model="statementForm.customerId" filterable clearable placeholder="选择客户" style="width: 100%" @change="reloadStatementSales">
            <el-option v-for="item in customers" :key="item.id" :label="customerOptionLabel(item)" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="日期范围" required>
          <el-date-picker v-model="statementForm.dates" type="daterange" value-format="YYYY-MM-DD" start-placeholder="开始日期" end-placeholder="结束日期" style="width: 100%" @change="reloadStatementSales" />
        </el-form-item>
        <el-form-item label="业务单号">
          <el-input v-model="statementForm.keyword" clearable placeholder="搜索销售单号、维修单号" @keyup.enter="reloadStatementSales" @clear="reloadStatementSales" />
        </el-form-item>
        <el-form-item label="付款状态">
          <el-select v-model="statementForm.paymentStatus" clearable placeholder="全部" style="width: 100%" @change="reloadStatementSales">
            <el-option label="未收款" value="unpaid" />
            <el-option label="部分收款" value="partial" />
            <el-option label="已结清" value="paid" />
          </el-select>
        </el-form-item>
      </el-form>
      <div class="statement-actions">
        <el-button type="primary" :icon="Search" :loading="statementDialog.loading" @click="reloadStatementSales">查询业务单据</el-button>
        <span>已选择 {{ statementSelection.length }} 张业务单据</span>
      </div>
      <el-table ref="statementTableRef" v-loading="statementDialog.loading" :data="statementSales" border height="360" @selection-change="handleStatementSelection">
        <el-table-column type="selection" width="48" :selectable="statementSaleSelectable" />
        <el-table-column prop="saleNo" label="业务单号" min-width="140" />
        <el-table-column prop="saleDate" label="业务日期" min-width="120" />
        <el-table-column prop="customerName" label="客户" min-width="160" />
        <el-table-column prop="productName" label="商品" min-width="160" />
        <el-table-column prop="quantity" label="数量" width="90" />
        <el-table-column prop="totalAmount" label="业务金额" width="120" />
        <el-table-column prop="receivedAmount" label="已收金额" width="120" />
        <el-table-column prop="unpaidAmount" label="未收金额" width="120" />
        <el-table-column label="结算状态" width="120">
          <template #default="{ row }">{{ statusLabel(row.settlementStatus) }}</template>
        </el-table-column>
        <el-table-column label="对账状态" width="110">
          <template #default="{ row }">{{ row.reconciled ? '已对账' : '未对账' }}</template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="statementDialog.page"
        v-model:page-size="statementDialog.pageSize"
        :total="statementDialog.total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @current-change="loadStatementSales"
        @size-change="loadStatementSales"
      />
      <template #footer>
        <el-button @click="statementDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="statementDialog.saving" @click="generateStatement">生成对账单</el-button>
      </template>
    </el-dialog>

    <button v-if="canCreate" class="fab" :title="createButtonText" @click="openCreate">
      <el-icon><Plus /></el-icon>
    </button>
  </main>
</template>

<script setup>
import { computed, nextTick, onActivated, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Camera,
  EditPen,
  Plus,
  Refresh,
  Search,
  UploadFilled
} from '@element-plus/icons-vue'
import {
  createBusiness,
  deleteBusiness,
  getBusiness,
  listBusiness,
  listBusinessPhotos,
  runBusinessAction,
  updateBusiness,
  uploadBusinessPhoto
} from '../../api/business'
import { listCustomers } from '../../api/customer'
import { listSuppliers } from '../../api/supplier'
import { getMerchantInfo, restoreTestData, updateMerchantInfo } from '../../api/system'
import { listLoginLogs } from '../../api/audit'
import { listSignInHistory } from '../../api/dashboard'
import { useSearch } from '../../composables/useSearch'
import { useAuthStore } from '../../stores/auth'
import { formatDateFields, formatDateRows, formatDateTime } from '../../utils/date'
import { getBreakpoint } from '../../utils/mobile'
import { statusLabel } from '../../utils/status'

const props = defineProps({ type: { type: String, required: true } })
const authStore = useAuthStore()
const apiOrigin = (import.meta.env.VITE_API_BASE_URL || '/api/v1').replace('/api/v1', '')
const photoInput = ref()
const albumInput = ref()
const breakpoint = ref(getBreakpoint())
const loading = ref(false)
const loadingMore = ref(false)
const reportLoading = ref(false)
const saving = ref(false)
const rows = ref([])
const customers = ref([])
const suppliers = ref([])
const products = ref([])
const inventoryProducts = ref([])
const financeAccounts = ref([])
const photos = ref([])
const purchaseSources = ref([])
const reportSummary = ref({})
const reportRanking = ref({})
const reportTrend = ref([])
const reportAging = ref([])
const reportSlowMoving = ref([])
const restoreLoading = ref(false)
const total = ref(0)
function defaultQueryForType(type = props.type) {
  const sortBy = type === 'inventory-movements' ? 'occurredAt' : (type === 'profit-report' ? 'salesDate' : (type === 'inventory-asset-report' ? 'latestPurchaseDate' : 'id'))
  return { page: 1, pageSize: type === 'inventory-movements' ? 10 : 20, keyword: '', productName: '', customerId: '', sourceType: '', operatorName: '', startDate: '', endDate: '', deletedOnly: false, sortBy, order: 'desc' }
}

const { searchForm: query, resetSearch } = useSearch(defaultQueryForType())
const movementDates = ref([])
const dialog = reactive({ visible: false, editing: false, id: null })
const photoDialog = reactive({ visible: false, row: null, scene: 'general' })
const detailDialog = reactive({ visible: false })
const settlementDialog = reactive({ visible: false, saving: false, row: null })
const statementDialog = reactive({ visible: false, loading: false, saving: false, page: 1, pageSize: 10, total: 0 })
const deleteDialog = reactive({ visible: false, saving: false, row: null, reason: '' })
const form = reactive({})
const settlementForm = reactive({ targetName: '', sourceNo: '', accountName: '', amount: 0, remark: '' })
const statementForm = reactive({ customerId: '', dates: [], keyword: '', paymentStatus: '' })
const statementSales = ref([])
const statementSelection = ref([])
const statementTableRef = ref()
const sourceTableRef = ref()
const profileLogs = reactive({ loading: false, rows: [], total: 0 })
const signInHistory = reactive({ loading: false, rows: [], total: 0, page: 1, pageSize: 10 })
const merchantInfo = reactive({ loading: false, saving: false, form: { companyName: '', contactName: '', contactPhone: '' } })
const detail = ref(null)
const detailType = ref(props.type)
const inventoryTab = reactive({
  active: 'purchaseSources',
  purchaseSources: { list: [], total: 0, page: 1, pageSize: 10, sortBy: 'purchaseDate', order: 'desc' },
  salesTrace: { list: [], total: 0, page: 1, pageSize: 10, sortBy: 'orderDate', order: 'desc' },
  inventoryMovements: { list: [], total: 0, page: 1, pageSize: 10, sortBy: 'occurredAt', order: 'desc' },
  filters: {
    purchaseSources: { keyword: '', status: '', dates: [] },
    salesTrace: { keyword: '', customerName: '', dates: [] },
    inventoryMovements: { keyword: '', sourceType: '', dates: [] }
  }
})
let photoTarget = null

const salesStatusOptions = [
  { label: '草稿', value: '草稿' },
  { label: '已报价', value: '已报价' },
  { label: '待发货', value: '待发货' },
  { label: '已发货', value: '已发货' },
  { label: '待收款', value: '待收款' },
  { label: '已完成', value: '已完成' },
  { label: '已取消', value: '已取消' }
]

const productStatusOptions = [
  { label: '启用', value: 'active' },
  { label: '停用', value: 'disabled' }
]

const purchaseStatusOptions = [
  { label: '草稿', value: '草稿' },
  { label: '已下单', value: '已下单' },
  { label: '已入库', value: '已入库' },
  { label: '待付款', value: '待付款' },
  { label: '已完成', value: '已完成' },
  { label: '已退货', value: '已退货' },
  { label: '已取消', value: '已取消' }
]

const inventoryStatusOptions = [
  { label: '正常', value: '正常' },
  { label: '预警', value: '预警' },
  { label: '冻结', value: '冻结' },
  { label: '停用', value: '停用' }
]

const repairStatusOptions = [
  { label: '已完成', value: '已完成' }
]

const projectStatusOptions = [
  { label: '计划中', value: '计划中' },
  { label: '施工中', value: '施工中' },
  { label: '验收中', value: '验收中' },
  { label: '已验收', value: '已验收' },
  { label: '已结算', value: '已结算' },
  { label: '已暂停', value: '已暂停' },
  { label: '已取消', value: '已取消' }
]

const movementSourceOptions = [
  { label: '采购入库', value: 'purchase' },
  { label: '销售出库', value: 'sales' },
  { label: '维修领料', value: 'repair' },
  { label: '工程领料', value: 'project' },
  { label: '销售退货', value: 'sales_return' },
  { label: '采购退货', value: 'purchase_return' },
  { label: '库存盘点', value: 'stocktake' },
  { label: '库存调整', value: 'adjust' },
  { label: '销售删除恢复', value: 'ORDER_DELETE' },
  { label: '采购删除冲销', value: 'PURCHASE_DELETE' }
]

const sourceTypeLabels = {
  income: '收入',
  expense: '支出',
  daily: '日常记账',
  sales: '销售出库',
  purchase: '采购入库',
  repair: '维修领料',
  project: '工程领料',
  receivables: '应收账款',
  payables: '应付账款',
  '应收收款': '应收收款',
  '应付付款': '应付付款',
  '客户对账单收款': '客户对账单收款',
  '销售收款': '销售收款',
  '采购付款': '采购付款',
  '删除冲销': '删除冲销',
  '日常记账': '日常记账',
  '日常收入': '日常收入',
  '日常支出': '日常支出'
}

const documentTypeOptions = [
  { label: '销售订单', value: 'SALES_ORDER' },
  { label: '采购订单', value: 'PURCHASE_ORDER' },
  { label: '入库单', value: 'STOCK_IN' },
  { label: '出库单', value: 'STOCK_OUT' }
]

const statementStatusOptions = [
  { label: '未确认', value: 'unconfirmed' },
  { label: '已确认', value: 'confirmed' },
  { label: '已结算', value: 'settled' }
]

const financeAccountTypeOptions = [
  { label: '现金', value: 'cash' },
  { label: '微信', value: 'wechat' },
  { label: '支付宝', value: 'alipay' },
  { label: '银行卡', value: 'bank' },
  { label: '其他', value: 'other' }
]

const configs = {
  products: {
    title: '商品',
    subtitle: '商品编码、名称、分类、品牌、规格、单位、条码和库存预警资料管理',
    columns: [
      { prop: 'code', label: '商品编码' }, { prop: 'name', label: '商品名称' }, { prop: 'category', label: '商品分类' },
      { prop: 'brand', label: '品牌' }, { prop: 'spec', label: '规格型号' }, { prop: 'unit', label: '单位' },
      { prop: 'barcode', label: '条码' }, { prop: 'qrCode', label: '二维码' }, { prop: 'imageUrl', label: '图片' },
      { prop: 'minStock', label: '库存预警数量' }, { prop: 'status', label: '商品状态' }, { prop: 'createdAt', label: '创建时间' }, { prop: 'updatedAt', label: '更新时间' }
    ],
    formFields: [
      { prop: 'code', label: '商品编码' }, { prop: 'name', label: '商品名称' }, { prop: 'category', label: '商品分类' },
      { prop: 'brand', label: '品牌' }, { prop: 'spec', label: '规格型号' }, { prop: 'unit', label: '单位' },
      { prop: 'barcode', label: '条码' }, { prop: 'qrCode', label: '二维码' }, { prop: 'imageUrl', label: '图片' },
      { prop: 'minStock', label: '库存预警数量', type: 'number' },
      { prop: 'status', label: '商品状态', type: 'select', options: productStatusOptions }
    ]
  },
  sales: {
    title: '销售',
    subtitle: '快速销售、报价、送货单',
    photoScene: 'delivery',
    columns: [
      { prop: 'orderNo', label: '销售单号' }, { prop: 'customerName', label: '客户' }, { prop: 'productName', label: '商品' },
      { prop: 'quantity', label: '数量' }, { prop: 'price', label: '售价' }, { prop: 'costPrice', label: '成本价' }, { prop: 'orderDate', label: '销售时间' },
      { prop: 'totalAmount', label: '总额' }, { prop: 'profitAmount', label: '利润' }, { prop: 'profitRate', label: '毛利率' }, { prop: 'status', label: '状态' }
    ],
    formFields: [
      { prop: 'customerId', label: '客户', type: 'customer' },
      { prop: 'productId', label: '商品', type: 'product' },
      { prop: 'orderDate', label: '销售时间', type: 'date' },
      { prop: 'quantity', label: '数量', type: 'number' },
      { prop: 'price', label: '售价', type: 'number' },
      { prop: 'inventoryBatchId', label: '采购来源', type: 'purchaseSource' }
    ]
  },
  purchase: {
    title: '采购',
    subtitle: '快速采购、付款',
    photoScene: 'purchase',
    columns: [
      { prop: 'orderNo', label: '采购单号' }, { prop: 'supplierName', label: '供应商' }, { prop: 'productName', label: '商品' },
      { prop: 'quantity', label: '数量' }, { prop: 'orderDate', label: '采购时间' }, { prop: 'totalAmount', label: '总额' }, { prop: 'payableAmount', label: '待付' }, { prop: 'status', label: '状态' }
    ],
    formFields: [
      { prop: 'supplierId', label: '供应商', type: 'supplier' },
      { prop: 'productId', label: '商品', type: 'product' },
      { prop: 'orderDate', label: '采购时间', type: 'date' },
      { prop: 'quantity', label: '数量', type: 'number' },
      { prop: 'price', label: '采购价', type: 'number' }
    ]
  },
  inventory: {
    title: '库存',
    subtitle: '库存查询',
    columns: [
      { prop: 'productCode', label: '商品编码' }, { prop: 'productName', label: '商品名称' }, { prop: 'category', label: '分类' },
      { prop: 'brand', label: '品牌' }, { prop: 'unit', label: '单位' }, { prop: 'quantity', label: '当前库存' },
      { prop: 'availableQuantity', label: '可销售库存' }, { prop: 'occupiedQuantity', label: '占用数量' }, { prop: 'amount', label: '库存金额' }, { prop: 'status', label: '库存状态' }
    ],
    formFields: []
  },
  'inventory-movements': {
    title: '库存流水',
    subtitle: '按商品编码查询采购入库、销售出库、维修领料、工程领料等库存变化',
    columns: [
      { prop: 'productCode', label: '商品编码' }, { prop: 'productName', label: '商品名称' }, { prop: 'spec', label: '规格型号' },
      { prop: 'warehouse', label: '仓库' },
      { prop: 'businessType', label: '业务类型' }, { prop: 'businessNo', label: '业务单号', type: 'businessLink' },
      { prop: 'quantityChange', label: '数量变化' }, { prop: 'beforeQuantity', label: '变更前库存' }, { prop: 'afterQuantity', label: '变更后库存' },
      { prop: 'purchaseOrderNo', label: '关联采购单' }, { prop: 'salesOrderNo', label: '关联销售单' },
      { prop: 'operatorName', label: '操作人' }, { prop: 'occurredAt', label: '操作时间' }
    ],
    formFields: []
  },
  repair: {
    title: '维修',
    subtitle: '维修工单、客户设备、维修配件和应收结算',
    columns: [
      { prop: 'orderNo', label: '维修单号' }, { prop: 'customerName', label: '客户' }, { prop: 'deviceName', label: '设备' },
      { prop: 'repairStatus', label: '状态' }, { prop: 'registeredAt', label: '登记时间' },
      { prop: 'partsAmount', label: '配件收入' }, { prop: 'outsourceCost', label: '外协支出' }, { prop: 'totalAmount', label: '应收金额' }, { prop: 'profitAmount', label: '利润' }
    ],
    formFields: [
      { prop: 'customerId', label: '客户', type: 'customer' },
      { prop: 'deviceName', label: '设备' },
      { prop: 'registeredAt', label: '登记时间', type: 'date' },
      { prop: 'repairStatus', label: '状态', type: 'select', options: repairStatusOptions },
      { prop: 'faultDesc', label: '故障描述' },
      { prop: 'onsiteFee', label: '上门费用', type: 'number' },
      { prop: 'detectionFee', label: '检测费用', type: 'number' },
      { prop: 'installationFee', label: '安装费用', type: 'number' }
    ]
  },
  project: {
    title: '工程',
    subtitle: '签到、施工日志、材料使用、验收照片',
    columns: [
      { prop: 'projectNo', label: '项目编号' }, { prop: 'name', label: '项目名称' }, { prop: 'customerName', label: '客户' },
      { prop: 'status', label: '状态' }, { prop: 'progress', label: '进度' }, { prop: 'startDate', label: '开始时间' }, { prop: 'endDate', label: '结束时间' }, { prop: 'settleAmount', label: '结算金额' }
    ],
    formFields: [
      { prop: 'name', label: '项目名称' }, { prop: 'customerName', label: '客户' }, { prop: 'startDate', label: '开始时间', type: 'date' },
      { prop: 'endDate', label: '结束时间', type: 'date' }, { prop: 'status', label: '状态', type: 'select', options: projectStatusOptions },
      { prop: 'progress', label: '进度', type: 'number' }, { prop: 'budgetAmount', label: '预算', type: 'number' }
    ]
  },
  finance: {
    title: '资金流水',
    subtitle: '销售、采购、维修、工程和日常记账自动生成的统一资金流水',
    columns: [
      { prop: 'recordNo', label: '流水号' }, { prop: 'recordType', label: '类型' }, { prop: 'accountName', label: '账户' },
      { prop: 'targetName', label: '对象' }, { prop: 'amount', label: '金额' }, { prop: 'businessType', label: '来源' }, { prop: 'businessNo', label: '业务单号' }, { prop: 'status', label: '状态' }
    ],
    formFields: [
      { prop: 'recordType', label: '类型', type: 'select', options: [{ label: '日常支出', value: '日常支出' }, { label: '日常收入', value: '日常收入' }] },
      { prop: 'accountName', label: '账户', type: 'select', options: [{ label: '现金', value: '现金' }, { label: '微信', value: '微信' }, { label: '支付宝', value: '支付宝' }, { label: '银行卡', value: '银行卡' }] },
      { prop: 'targetName', label: '对象' },
      { prop: 'amount', label: '金额', type: 'number' }
    ]
  },
  'finance-accounts': {
    title: '资金账户',
    subtitle: '维护现金、微信、支付宝、银行卡等账户资料，余额由资金流水自动计算',
    columns: [
      { prop: 'code', label: '账户编码' }, { prop: 'name', label: '账户名称' }, { prop: 'accountType', label: '账户类型' },
      { prop: 'openingBalance', label: '期初余额' }, { prop: 'balance', label: '当前余额' }, { prop: 'status', label: '状态' }
    ],
    formFields: [
      { prop: 'code', label: '账户编码' },
      { prop: 'name', label: '账户名称' },
      { prop: 'accountType', label: '账户类型', type: 'select', options: financeAccountTypeOptions },
      { prop: 'openingBalance', label: '期初余额', type: 'number', createOnly: true },
      { prop: 'status', label: '状态', type: 'select', options: productStatusOptions }
    ]
  },
  receivables: {
    title: '应收账款',
    subtitle: '应收查询、账龄分析、客户对账、收款登记、自动核销、欠款提醒',
    columns: [
      { prop: 'receivableNo', label: '应收单号' }, { prop: 'customerName', label: '客户' }, { prop: 'sourceNo', label: '来源单号' },
      { prop: 'totalAmount', label: '应收金额' }, { prop: 'receivedAmount', label: '已收' }, { prop: 'balanceAmount', label: '余额' },
      { prop: 'dueDate', label: '到期日' }, { prop: 'status', label: '状态' }
    ],
    formFields: [
      { prop: 'customerName', label: '客户' }, { prop: 'sourceNo', label: '来源单号' }, { prop: 'invoiceDate', label: '单据日期', type: 'date' },
      { prop: 'dueDate', label: '到期日', type: 'date' }, { prop: 'totalAmount', label: '应收金额', type: 'number' },
      { prop: 'receivedAmount', label: '已收', type: 'number' }, { prop: 'balanceAmount', label: '余额', type: 'number' }
    ]
  },
  'customer-statements': {
    title: '客户对账单',
    subtitle: '按客户和期间选择销售单，生成月结、季结客户往来对账单',
    columns: [
      { prop: 'statementNo', label: '对账单号' }, { prop: 'customerName', label: '客户' },
      { prop: 'startDate', label: '开始日期' }, { prop: 'endDate', label: '结束日期' },
      { prop: 'totalAmount', label: '销售金额' }, { prop: 'receivedAmount', label: '已收金额' },
      { prop: 'unpaidAmount', label: '本期应收' }, { prop: 'cumulativeDebt', label: '累计欠款' },
      { prop: 'status', label: '状态' }
    ],
    formFields: []
  },
  payables: {
    title: '应付账款',
    subtitle: '采购完成自动生成，应付款登记后自动生成资金流水并核销',
    columns: [
      { prop: 'payableNo', label: '应付单号' }, { prop: 'supplierName', label: '供应商' }, { prop: 'sourceNo', label: '来源单号' },
      { prop: 'totalAmount', label: '应付金额' }, { prop: 'paidAmount', label: '已付' }, { prop: 'balanceAmount', label: '余额' },
      { prop: 'dueDate', label: '到期日' }, { prop: 'status', label: '状态' }
    ],
    formFields: []
  },
  'profit-report': {
    title: '利润报表',
    subtitle: '按销售单关联采购单成本实时统计销售、维修、工程和费用利润',
    columns: [
      { prop: 'salesDate', label: '销售日期' }, { prop: 'salesOrderNo', label: '销售单号', type: 'salesLink' },
      { prop: 'customerName', label: '客户' }, { prop: 'productCode', label: '商品编码' }, { prop: 'productName', label: '商品名称' },
      { prop: 'quantity', label: '销售数量' }, { prop: 'salesPrice', label: '销售单价' }, { prop: 'salesAmount', label: '销售金额' },
      { prop: 'costAmount', label: '采购成本' }, { prop: 'profitAmount', label: '利润' }, { prop: 'profitRate', label: '利润率' },
      { prop: 'purchaseOrderNo', label: '采购来源', type: 'purchaseLink' }
    ],
    formFields: []
  },
  'inventory-asset-report': {
    title: '库存资产报表',
    subtitle: '按采购批次剩余库存实时统计库存数量、库龄和采购成本资产价值',
    columns: [
      { prop: 'productCode', label: '商品编码', type: 'inventoryLink' }, { prop: 'productName', label: '商品名称' },
      { prop: 'brand', label: '品牌' }, { prop: 'category', label: '分类' }, { prop: 'spec', label: '规格型号' },
      { prop: 'quantity', label: '当前库存数量' }, { prop: 'availableQuantity', label: '可销售数量' },
      { prop: 'purchaseCost', label: '采购成本' }, { prop: 'inventoryAmount', label: '库存金额' },
      { prop: 'latestPurchaseDate', label: '最近采购日期' }, { prop: 'latestSalesDate', label: '最近销售日期' },
      { prop: 'inventoryStatus', label: '库存状态' }
    ],
    formFields: []
  },
  'document-delete-records': {
    title: '单据删除记录',
    subtitle: '查询单据删除审计、删除原因、库存反处理和财务冲销记录',
    columns: [
      { prop: 'documentNo', label: '单据编号' }, { prop: 'documentType', label: '单据类型' },
      { prop: 'deleteUserName', label: '删除人' }, { prop: 'deleteTime', label: '删除时间' },
      { prop: 'deleteReason', label: '删除理由' }, { prop: 'stockProcessed', label: '库存处理' },
      { prop: 'financeProcessed', label: '财务处理' }, { prop: 'deleteStatus', label: '状态' }
    ],
    formFields: []
  },
  profile: { title: '我的', subtitle: '个人中心、主题、安装到桌面', columns: [], formFields: [] },
  settings: { title: '设置', subtitle: '主题、打印、PWA、系统配置', columns: [], formFields: [] }
}

const config = computed(() => configs[props.type] || configs.profile)
const isReportPage = computed(() => ['profit-report', 'inventory-asset-report'].includes(props.type))
const outsourceSuppliers = computed(() => suppliers.value.filter((item) => String(item.supplierTypes || '').includes('外协服务商') || String(item.supplierTypes || '').includes('综合供应商')))
const sourceFilterOptions = computed(() => props.type === 'document-delete-records' ? documentTypeOptions : movementSourceOptions)
const reportSummaryCards = computed(() => {
  const s = reportSummary.value || {}
  if (props.type === 'profit-report') {
    return [
      { label: '销售收入', value: money(s.salesIncome), hint: '已关联采购成本销售' },
      { label: '销售成本', value: money(s.salesCost), hint: '采购来源成本' },
      { label: '销售毛利润', value: money(s.salesProfit), hint: `毛利率 ${percent(s.grossProfitRate)}` },
      { label: '维修收入', value: money(s.repairIncome), hint: '维修单收入' },
      { label: '维修利润', value: money(s.repairProfit), hint: '收入减配件成本' },
      { label: '工程收入', value: money(s.projectIncome), hint: '工程结算/合同金额' },
      { label: '工程利润', value: money(s.projectProfit), hint: '收入减项目成本' },
      { label: '其他收入', value: money(s.otherIncome), hint: '日常收入' },
      { label: '费用支出', value: money(s.expenseAmount), hint: '日常费用支出' },
      { label: '净利润', value: money(s.netProfit), hint: `净利率 ${percent(s.netProfitRate)}` }
    ]
  }
  return [
    { label: '库存商品数量', value: s.productCount || 0, hint: '有剩余库存的商品' },
    { label: '库存总数量', value: Number(s.totalQuantity || 0).toLocaleString('zh-CN'), hint: '当前可销售数量' },
    { label: '库存总价值', value: money(s.totalValue), hint: '采购成本口径' },
    { label: '平均周转天数', value: s.avgTurnoverDays || 0, hint: '后续按销售节奏扩展' },
    { label: '库存预警数量', value: s.warningCount || 0, hint: '低于预警线批次' },
    { label: '缺货商品数量', value: s.outOfStockCount || 0, hint: '当前库存为 0' },
    { label: '呆滞库存数量', value: s.slowMovingCount || 0, hint: '默认 180 天未销售' }
  ]
})
const reportBlocks = computed(() => {
  if (props.type === 'profit-report') {
    return [
      { title: '商品利润排行', type: 'profit', rows: reportRanking.value.products || [] },
      { title: '客户利润排行', type: 'profit', rows: reportRanking.value.customers || [] },
      { title: '品牌利润排行', type: 'profit', rows: reportRanking.value.brands || [] },
      { title: '每日利润趋势', type: 'trendProfit', rows: reportTrend.value || [] }
    ]
  }
  return [
    { title: '库龄分析', type: 'aging', rows: reportAging.value || [] },
    { title: '呆滞库存', type: 'asset', rows: reportSlowMoving.value || [] },
    { title: '库存价值趋势', type: 'assetTrend', rows: reportTrend.value || [] }
  ]
})
const displayColumns = computed(() => {
  if (props.type !== 'sales' || !query.deletedOnly) return config.value.columns
  return [
    { prop: 'orderNo', label: '销售单号' },
    { prop: 'customerName', label: '客户' },
    { prop: 'productName', label: '商品' },
    { prop: 'quantity', label: '数量' },
    { prop: 'totalAmount', label: '总额' },
    { prop: 'deleteReason', label: '删除原因' },
    { prop: 'deletedBy', label: '删除人ID' },
    { prop: 'deletedAt', label: '删除时间' }
  ]
})
const visibleFormFields = computed(() => config.value.formFields.filter((field) => !field.createOnly || !dialog.editing))
const isMobile = computed(() => breakpoint.value === 'mobile')
const profileInitial = computed(() => (authStore.user?.realName || authStore.user?.username || '用').slice(0, 1).toUpperCase())
const profileLoading = computed(() => props.type === 'profile' && (profileLogs.loading || signInHistory.loading) && !authStore.user)
const supportsPhotos = computed(() => ['sales', 'purchase', 'repair'].includes(props.type))
const photoTitle = computed(() => `${config.value.title}照片记录`)
const createButtonText = computed(() => props.type === 'customer-statements' ? '生成对账单' : '新建')
const searchPlaceholder = computed(() => {
  if (props.type === 'inventory-movements') return '输入商品编码查询库存流水'
  if (props.type === 'document-delete-records') return '搜索单据编号、删除原因'
  if (props.type === 'customer-statements') return '搜索对账单号'
  return '搜索单号、名称、客户、状态'
})
const canCreate = computed(() => !['inventory', 'inventory-movements', 'document-delete-records', 'finance', 'receivables', 'payables', 'profit-report', 'inventory-asset-report', 'profile'].includes(props.type))
const canDelete = computed(() => {
  if (query.deletedOnly) return false
  if (['inventory', 'inventory-movements', 'document-delete-records', 'finance', 'finance-accounts', 'receivables', 'customer-statements', 'payables', 'profit-report', 'inventory-asset-report'].includes(props.type)) return false
  if (['sales', 'purchase'].includes(props.type)) return authStore.isLoggedIn || authStore.hasPermission('document_delete')
  return authStore.hasPermission('document_delete')
})
const deleteDialogNo = computed(() => deleteDialog.row?.orderNo || deleteDialog.row?.documentNo || deleteDialog.row?.projectNo || deleteDialog.row?.recordNo || deleteDialog.row?.id || '')
const actionWidth = computed(() => {
  if (props.type === 'sales') return 260
  if (props.type === 'inventory') return 300
  if (props.type === 'customer-statements') return 300
  if (props.type === 'receivables' || props.type === 'payables') return 220
  return 220
})
const detailTitle = computed(() => {
  if (detailType.value === 'inventory') return '库存详情'
  if (detailType.value === 'purchase') return '采购单详情'
  if (detailType.value === 'sales') return '销售单详情'
  if (detailType.value === 'repair') return '维修单详情'
  if (detailType.value === 'project') return '工程项目详情'
  if (detailType.value === 'finance') return '资金流水详情'
  if (detailType.value === 'customer-statements') return '客户对账单详情'
  if (detailType.value === 'receivables') return '应收账款详情'
  if (detailType.value === 'payables') return '应付账款详情'
  return '销售单详情'
})
const settlementTitle = computed(() => {
  if (props.type === 'payables') return '付款登记'
  if (props.type === 'customer-statements') return '对账单结算'
  return '收款登记'
})

watch(() => props.type, () => {
  resetSearchState()
  rows.value = []
  if (props.type === 'profile') {
    loadProfile()
  } else {
    loadOptions()
    refresh()
  }
})

watch(() => form.quantity, () => {
  if (props.type === 'sales' && purchaseSources.value.length > 0 && form.inventoryBatchIds?.length) {
    updatePurchaseSourceSummary()
  }
})

onMounted(() => {
  window.addEventListener('resize', handleResize)
  resetSearchState()
  if (props.type === 'profile') {
    loadProfile()
  } else {
    loadOptions()
    refresh()
  }
})
onActivated(() => {
  resetSearchState()
  if (props.type === 'profile') loadProfile()
  else refresh()
})
onUnmounted(() => window.removeEventListener('resize', handleResize))

function handleResize() {
  breakpoint.value = getBreakpoint()
}

function resetSearchState() {
  resetSearch(defaultQueryForType(props.type))
  movementDates.value = []
}

function changeMovementDates(value) {
  query.startDate = value?.[0] || ''
  query.endDate = value?.[1] || ''
  query.page = 1
  refresh()
}

function changeCustomerFilter() {
  query.page = 1
  refresh()
}

async function refresh() {
  if (props.type === 'profile') {
    await loadProfile()
    return
  }
  if (!config.value.columns.length) return
  loading.value = true
  try {
    const data = await listBusiness(props.type, query)
    rows.value = formatDateRows(data.list || [])
    total.value = data.total || 0
    if (isReportPage.value) await loadReportData()
  } finally {
    loading.value = false
  }
}

async function loadReportData() {
  reportLoading.value = true
  try {
    const payload = { ...query }
    if (props.type === 'profit-report') {
      const [summary, ranking, trend] = await Promise.all([
        runBusinessAction(props.type, 'summary', payload),
        runBusinessAction(props.type, 'ranking', payload),
        runBusinessAction(props.type, 'trend', payload)
      ])
      reportSummary.value = summary.summary || {}
      reportRanking.value = ranking.ranking || {}
      reportTrend.value = trend.trend || []
      reportAging.value = []
      reportSlowMoving.value = []
    } else if (props.type === 'inventory-asset-report') {
      const [summary, aging, slowMoving, trend] = await Promise.all([
        runBusinessAction(props.type, 'summary', payload),
        runBusinessAction(props.type, 'aging', payload),
        runBusinessAction(props.type, 'slow-moving', { ...payload, days: 180 }),
        runBusinessAction(props.type, 'trend', payload)
      ])
      reportSummary.value = summary.summary || {}
      reportRanking.value = {}
      reportAging.value = aging.aging || []
      reportSlowMoving.value = slowMoving.list || []
      reportTrend.value = trend.trend || []
    }
  } finally {
    reportLoading.value = false
  }
}

async function loadProfile() {
  await ensureCurrentUser()
  await Promise.all([loadMerchantInfo(), loadProfileLoginLogs(), loadSignInHistory()])
}

async function ensureCurrentUser() {
  if (!authStore.isLoggedIn) return
  await authStore.loadCurrentUser()
}

async function loadProfileLoginLogs() {
  profileLogs.loading = true
  try {
    const data = await listLoginLogs({
      page: 1,
      pageSize: 20,
      userId: authStore.user?.id || '',
      username: authStore.user?.username || ''
    })
    profileLogs.rows = formatDateRows(data.list || [])
    profileLogs.total = data.total || 0
  } finally {
    profileLogs.loading = false
  }
}

async function loadSignInHistory() {
  signInHistory.loading = true
  try {
    const data = await listSignInHistory({
      page: signInHistory.page,
      pageSize: signInHistory.pageSize
    })
    signInHistory.rows = formatDateRows(data.list || [])
    signInHistory.total = data.total || 0
  } finally {
    signInHistory.loading = false
  }
}

function reloadSignInHistory() {
  signInHistory.page = 1
  loadSignInHistory()
}

function signInCoords(row) {
  const lat = Number(row.latitude || 0).toFixed(5)
  const lng = Number(row.longitude || 0).toFixed(5)
  return `${lat}, ${lng}`
}

function signInAddress(row) {
  return row.address || signInCoords(row)
}

async function loadMerchantInfo() {
  merchantInfo.loading = true
  try {
    const data = await getMerchantInfo()
    Object.assign(merchantInfo.form, {
      companyName: data.companyName || '',
      contactName: data.contactName || '',
      contactPhone: data.contactPhone || ''
    })
  } finally {
    merchantInfo.loading = false
  }
}

async function saveMerchantInfo() {
  merchantInfo.saving = true
  try {
    const data = await updateMerchantInfo(merchantInfo.form)
    Object.assign(merchantInfo.form, data)
    ElMessage.success('商户信息已保存')
  } finally {
    merchantInfo.saving = false
  }
}

async function loadOptions() {
  const [customerData, supplierData, productData, inventoryData, accountData] = await Promise.all([
    listCustomers({ page: 1, pageSize: 200 }).catch(() => ({ list: [] })),
    listSuppliers({ page: 1, pageSize: 200 }).catch(() => ({ list: [] })),
    listBusiness('products', { page: 1, pageSize: 500 }).catch(() => ({ list: [] })),
    listBusiness('inventory', { page: 1, pageSize: 500 }).catch(() => ({ list: [] })),
    listBusiness('finance-accounts', { page: 1, pageSize: 200 }).catch(() => ({ list: [] }))
  ])
  customers.value = customerData.list || []
  suppliers.value = supplierData.list || []
  products.value = productData.list || []
  inventoryProducts.value = (inventoryData.list || []).filter((item) => Number(item.availableQuantity ?? item.quantity ?? 0) > 0)
  financeAccounts.value = (accountData.list || []).filter((item) => item.status !== 'disabled')
}

async function loadMore() {
  loadingMore.value = true
  try {
    query.page += 1
    const data = await listBusiness(props.type, query)
    rows.value.push(...formatDateRows(data.list || []))
    total.value = data.total || 0
  } finally {
    loadingMore.value = false
  }
}

function openCreate() {
  if (props.type === 'customer-statements') {
    openStatementDialog()
    return
  }
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

function resetForm(row = {}) {
  for (const key of Object.keys(form)) delete form[key]
  purchaseSources.value = []
  for (const field of config.value.formFields) {
    const value = row[field.prop]
    form[field.prop] = value ?? defaultFieldValue(field)
  }
  if (props.type === 'sales') {
    const batchIds = Array.isArray(row.inventoryBatchIds) ? row.inventoryBatchIds : []
    form.inventoryBatchIds = batchIds.length ? batchIds : (row.inventoryBatchId ? [row.inventoryBatchId] : [])
  }
  if (props.type === 'repair') {
    form.repairParts = Array.isArray(row.repairParts) ? [...row.repairParts] : []
    form.outsourceItems = Array.isArray(row.outsourceItems) ? [...row.outsourceItems] : []
  }
  form.remark = row.remark || ''
}

function selectCustomer(id) {
  const item = customers.value.find((customer) => customer.id === id)
  form.customerName = item?.name || ''
}

function selectSupplier(id) {
  const item = suppliers.value.find((supplier) => supplier.id === id)
  form.supplierName = item?.name || ''
}

function addRepairPart() {
  if (!Array.isArray(form.repairParts)) form.repairParts = []
  form.repairParts.push({ productId: '', productCode: '', productName: '', inventoryBatchId: '', purchaseSources: [], quantity: 1, price: 0, amount: 0, costPrice: 0, costAmount: 0 })
}

function removeRepairPart(index) {
  form.repairParts.splice(index, 1)
}

async function selectRepairPart(row, id) {
  const item = inventoryProducts.value.find((product) => Number(product.productId || product.id) === Number(id))
  row.productId = item?.productId || item?.id || ''
  row.productCode = item?.productCode || ''
  row.productName = item?.productName || ''
  row.inventoryBatchId = ''
  row.purchaseSources = []
  if (!row.productId && !row.productCode) return
  const data = await runBusinessAction('sales', 'purchase-sources', {
    productId: row.productId,
    productCode: row.productCode,
    warehouse: row.warehouse || '主仓库'
  }).catch(() => ({ list: [] }))
  row.purchaseSources = (data.list || []).filter((source) => Number(source.remainingQuantity || 0) > 0)
  if (row.purchaseSources.length === 1) {
    selectRepairPartBatch(row, row.purchaseSources[0].id)
  }
}

function selectRepairPartBatch(row, id) {
  const item = (row.purchaseSources || []).find((source) => Number(source.id) === Number(id))
  row.inventoryBatchId = item?.id || ''
  row.costPrice = Number(item?.purchasePrice || 0)
  row.costAmount = Number(row.quantity || 0) * row.costPrice
  row.amount = 0
}

function addRepairOutsource() {
  if (!Array.isArray(form.outsourceItems)) form.outsourceItems = []
  form.outsourceItems.push({ supplierId: '', supplierName: '', serviceProject: '', amount: 0, remark: '' })
}

function removeRepairOutsource(index) {
  form.outsourceItems.splice(index, 1)
}

function selectOutsourceSupplier(row, id) {
  const item = suppliers.value.find((supplier) => supplier.id === id)
  row.supplierName = item?.name || ''
}

function normalizeRepairMoneyItems(items) {
  return (items || []).map((item) => {
    const quantity = Number(item.quantity || 0)
    const price = Number(item.price || 0)
    const amount = Number(item.amount || 0) || quantity * price
    const { purchaseSources: _purchaseSources, ...payload } = item
    return { ...payload, quantity, price, amount }
  }).filter((item) => item.amount > 0)
}

async function selectProduct(id) {
  const item = products.value.find((product) => product.id === id)
  if (item) {
    form.productName = item.name
    form.productCode = item.code
    form.inventoryBatchId = ''
    form.inventoryBatchIds = []
    purchaseSources.value = []
    if (props.type === 'sales') await loadPurchaseSources()
    return
  }
  if (typeof id === 'string') {
    form.productId = ''
    form.productName = id
    form.productCode = ''
    form.inventoryBatchId = ''
    form.inventoryBatchIds = []
    purchaseSources.value = []
  }
}

async function loadPurchaseSources() {
  if (props.type !== 'sales' || (!form.productId && !form.productCode)) return
  const data = await runBusinessAction('sales', 'purchase-sources', {
    productId: form.productId,
    productCode: form.productCode,
    warehouse: form.warehouse || '主仓库'
  }).catch(() => ({ list: [] }))
  purchaseSources.value = data.list || []
  if (purchaseSources.value.length === 1) {
    selectPurchaseSourceIds([purchaseSources.value[0].id])
  }
}

function selectPurchaseSourceRow(row) {
  sourceTableRef.value?.toggleRowSelection(row)
}

function selectPurchaseSourceRows(rows) {
  selectPurchaseSourceIds(rows.map((row) => row.id), false)
}

function selectPurchaseSourceIds(ids = [], syncTable = true) {
  const selectedIds = Array.from(new Set((Array.isArray(ids) ? ids : [ids]).map((id) => Number(id)).filter(Boolean)))
  form.inventoryBatchIds = selectedIds
  form.inventoryBatchId = selectedIds[0] || ''
  if (syncTable) syncPurchaseSourceTableSelection(selectedIds)
  updatePurchaseSourceSummary()
}

function syncPurchaseSourceTableSelection(ids) {
  nextTick(() => {
    const table = sourceTableRef.value
    if (!table) return
    table.clearSelection()
    for (const item of purchaseSources.value) {
      if (ids.includes(Number(item.id))) {
        table.toggleRowSelection(item, true)
      }
    }
  })
}

function updatePurchaseSourceSummary() {
  const selected = purchaseSources.value.filter((source) => (form.inventoryBatchIds || []).includes(Number(source.id)))
  if (selected.length === 0) {
    form.costPrice = 0
    form.purchasePrice = 0
    form.purchaseOrderId = ''
    form.purchaseOrderNo = ''
    form.supplierId = ''
    form.supplierName = ''
    return
  }
  let remainingQty = Number(form.quantity || 0)
  let totalCost = 0
  let totalQty = 0
  for (const item of selected) {
    if (remainingQty <= 0) break
    const available = Number(item.remainingQuantity || 0)
    const take = remainingQty > 0 ? Math.min(remainingQty, available) : available
    totalQty += take
    totalCost += take * Number(item.purchasePrice || 0)
    remainingQty -= take
  }
  const first = selected[0]
  form.costPrice = totalQty > 0 ? totalCost / totalQty : Number(first.purchasePrice || 0)
  form.purchasePrice = form.costPrice
  form.purchaseOrderId = first.purchaseOrderId
  form.purchaseOrderNo = selected.map((item) => item.purchaseOrderNo).filter(Boolean).join('、')
  form.supplierId = first.supplierId
  form.supplierName = selected.map((item) => item.supplierName).filter(Boolean).join('、')
}

function purchaseSourceLabel(item) {
  return `${item.purchaseOrderNo || '-'} / ${item.supplierName || '-'} / 剩余${item.remainingQuantity || 0} / 成本${item.purchasePrice || 0}`
}

function productLabel(item) {
  const code = item.code ? `${item.code} - ` : ''
  return `${code}${item.name}`
}

function inventoryProductLabel(item) {
  const code = item.productCode ? `${item.productCode} - ` : ''
  const qty = item.availableQuantity ?? item.quantity ?? 0
  return `${code}${item.productName || item.name}（库存 ${qty}）`
}

async function save() {
  if (props.type === 'sales' && (!form.inventoryBatchIds || form.inventoryBatchIds.length === 0)) {
    ElMessage.error('销售商品必须选择采购来源')
    return
  }
  if (props.type === 'repair') {
    form.repairStatus = '已完成'
    form.repairParts = normalizeRepairMoneyItems(form.repairParts).map((item) => ({ ...item, quantity: item.quantity || 1, costAmount: Number(item.costAmount || 0) || Number(item.quantity || 1) * Number(item.costPrice || 0) }))
    if (form.repairParts.some((item) => !item.inventoryBatchId)) {
      ElMessage.error('维修配件必须选择库存批次')
      return
    }
    form.outsourceItems = normalizeRepairMoneyItems(form.outsourceItems)
  }
  saving.value = true
  try {
    if (dialog.editing) {
      await updateBusiness(props.type, dialog.id, form)
      ElMessage.success('更新成功')
    } else {
      await createBusiness(props.type, form)
      ElMessage.success('创建成功')
    }
    dialog.visible = false
    query.page = 1
    await loadOptions()
    await refresh()
  } finally {
    saving.value = false
  }
}

async function confirmRestoreTestData() {
  await ElMessageBox.confirm(
    '该操作会清空销售、采购、库存、维修、工程、客户、供应商、商品、财务业务数据、登录/操作日志以及账户最后登录信息，并重新生成中文测试资料。确认继续？',
    '一键恢复测试数据',
    { type: 'warning', confirmButtonText: '确认恢复', cancelButtonText: '取消' }
  )
  restoreLoading.value = true
  try {
    const result = await restoreTestData()
    const clearedLogs = (result.loginLogs || 0) + (result.operationLogs || 0)
    ElMessage.success(`测试数据已恢复：客户${result.customers || 0}个，商品${result.products || 0}个，供应商${result.suppliers || 0}个，资金账户${result.financeAccounts || 0}个，已清除日志${clearedLogs}条，账户登录信息${result.accountLoginLogs || 0}条`)
    query.page = 1
    await loadOptions()
    await refresh()
  } finally {
    restoreLoading.value = false
  }
}

function changeDeletedOnly() {
  query.page = 1
  refresh()
}

function openDelete(row) {
  deleteDialog.row = row
  deleteDialog.reason = ''
  deleteDialog.visible = true
}

async function confirmDelete() {
  if (!deleteDialog.row?.id || !deleteDialog.reason) {
    ElMessage.error('请输入删除原因')
    return
  }
  deleteDialog.saving = true
  try {
    await deleteBusiness(props.type, deleteDialog.row.id, { reason: deleteDialog.reason })
    deleteDialog.visible = false
    ElMessage.success('删除成功，反处理已完成')
    await refresh()
  } finally {
    deleteDialog.saving = false
  }
}

function openSettlement(row) {
  settlementDialog.visible = true
  settlementDialog.row = row
  settlementForm.targetName = props.type === 'payables' ? (row.supplierName || '') : (row.customerName || '')
  settlementForm.sourceNo = row.statementNo || row.sourceNo || row.receivableNo || row.payableNo || ''
  settlementForm.accountName = financeAccounts.value[0]?.name || ''
  settlementForm.amount = Number(row.unpaidAmount || row.balanceAmount || 0)
  settlementForm.remark = ''
}

async function submitSettlement() {
  const row = settlementDialog.row
  if (!row?.id) return
  if (!settlementForm.amount || settlementForm.amount <= 0) {
    ElMessage.error(`${props.type === 'payables' ? '付款' : '收款'}金额必须大于0`)
    return
  }
  if (!settlementForm.accountName) {
    ElMessage.error('请选择资金账户')
    return
  }
  settlementDialog.saving = true
  try {
    if (props.type === 'customer-statements') {
      await runBusinessAction('customer-statements', 'settle', {
        id: row.id,
        accountName: settlementForm.accountName,
        amount: settlementForm.amount,
        remark: settlementForm.remark
      })
      ElMessage.success('对账单结算成功，已生成资金流水')
    } else if (props.type === 'receivables') {
      await runBusinessAction('receivables', 'receive', {
        receivableId: row.id,
        customerId: row.customerId,
        customerName: row.customerName,
        amount: settlementForm.amount,
        accountName: settlementForm.accountName,
        businessNo: row.receivableNo || row.sourceNo,
        businessType: '应收收款',
        remark: settlementForm.remark
      })
      ElMessage.success('收款登记成功')
    } else if (props.type === 'payables') {
      await runBusinessAction('payables', 'pay', {
        payableId: row.id,
        supplierId: row.supplierId,
        supplierName: row.supplierName,
        amount: settlementForm.amount,
        accountName: settlementForm.accountName,
        businessNo: row.payableNo || row.sourceNo,
        businessType: '应付付款',
        remark: settlementForm.remark
      })
      ElMessage.success('付款登记成功')
    }
    settlementDialog.visible = false
    await refresh()
  } finally {
    settlementDialog.saving = false
  }
}

async function printSettlement(row) {
  if (!row?.id) return
  openPrintPage(props.type, row.id)
}

function customerOptionLabel(item) {
  const code = item.code ? `${item.code} - ` : ''
  return `${code}${item.name}`
}

function financeAccountLabel(item) {
  const balance = item.balance === undefined || item.balance === null ? '' : `（余额 ${item.balance}）`
  return `${item.name || '-'}${balance}`
}

function customerNameById(id) {
  return customers.value.find((customer) => customer.id === id)?.name || ''
}

function currentMonthRange() {
  const now = new Date()
  const start = new Date(now.getFullYear(), now.getMonth(), 1)
  const end = new Date(now.getFullYear(), now.getMonth() + 1, 0)
  const format = (date) => {
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    return `${date.getFullYear()}-${month}-${day}`
  }
  return [format(start), format(end)]
}

function openStatementDialog() {
  statementDialog.visible = true
  statementDialog.page = 1
  statementDialog.pageSize = 10
  statementDialog.total = 0
  statementForm.customerId = query.customerId || ''
  statementForm.dates = movementDates.value?.length ? [...movementDates.value] : currentMonthRange()
  statementForm.keyword = ''
  statementForm.paymentStatus = ''
  statementSales.value = []
  statementSelection.value = []
  if (statementForm.customerId) loadStatementSales()
}

function statementActionPayload() {
  return {
    customerId: statementForm.customerId,
    customerName: customerNameById(statementForm.customerId),
    startDate: statementForm.dates?.[0] || '',
    endDate: statementForm.dates?.[1] || '',
    keyword: statementForm.keyword,
    paymentStatus: statementForm.paymentStatus,
    reconciled: 'unreconciled',
    page: statementDialog.page,
    pageSize: statementDialog.pageSize,
    sortBy: 'orderDate',
    order: 'desc'
  }
}

async function reloadStatementSales() {
  statementDialog.page = 1
  await loadStatementSales()
}

async function loadStatementSales() {
  if (!statementForm.customerId) {
    statementSales.value = []
    statementDialog.total = 0
    statementSelection.value = []
    return
  }
  statementDialog.loading = true
  try {
    const data = await runBusinessAction('customer-statements', 'sales-candidates', statementActionPayload())
    statementSales.value = formatDateRows(data.list || [])
    statementDialog.total = data.total || 0
    statementSelection.value = []
  } finally {
    statementDialog.loading = false
  }
}

function handleStatementSelection(selection) {
  statementSelection.value = selection
}

function statementSaleSelectable(row) {
  return !row.reconciled
}

async function generateStatement() {
  if (!statementForm.customerId) {
    ElMessage.error('请选择客户')
    return
  }
  if (!statementForm.dates?.[0] || !statementForm.dates?.[1]) {
    ElMessage.error('请选择对账期间')
    return
  }
  const receivableIds = statementSelection.value.map((row) => row.receivableId).filter(Boolean)
  const saleIds = statementSelection.value.map((row) => row.saleId).filter(Boolean)
  if (receivableIds.length === 0 && saleIds.length === 0) {
    ElMessage.error('请至少选择一张业务单据')
    return
  }
  statementDialog.saving = true
  try {
    const created = await runBusinessAction('customer-statements', 'generate', {
      customerId: statementForm.customerId,
      startDate: statementForm.dates[0],
      endDate: statementForm.dates[1],
      receivableIds,
      saleIds
    })
    statementDialog.visible = false
    ElMessage.success('客户对账单已生成')
    await refresh()
    if (created?.id) await openBusinessDetail('customer-statements', created.id)
  } finally {
    statementDialog.saving = false
  }
}

async function confirmCustomerStatement(row) {
  if (!row?.id) return
  await runBusinessAction('customer-statements', 'confirm', { id: row.id })
  ElMessage.success('对账单已确认')
  await refresh()
}

async function settleCustomerStatement(row) {
  if (!row?.id) return
  await runBusinessAction('customer-statements', 'settle', { id: row.id })
  ElMessage.success('对账单已结算')
  await refresh()
}

async function printCustomerStatement(row) {
  if (!row?.id) return
  openPrintPage('customer-statements', row.id)
}

async function openView(row) {
  await openBusinessDetail(props.type, row.id)
}

async function openBusinessDetail(module, id) {
  if (!id) return
  const targetModule = normalizeBusinessModule(module)
  if (authStore.isLoggedIn && !authStore.user) {
    await authStore.loadCurrentUser().catch(() => {})
  }
  try {
    detailType.value = targetModule
    detail.value = formatDateFields(await getBusiness(targetModule, id))
    detailDialog.visible = true
    if (targetModule === 'inventory') {
      inventoryTab.active = 'purchaseSources'
      resetInventoryTabs()
      await loadInventoryTab()
    }
    return true
  } catch (error) {
    if (error.response?.status === 403) {
      ElMessage.warning('您没有查看该业务单据的权限。')
    }
    return false
  }
}

function billSourceModule(row) {
  const source = String(row?.sourceType || '').toLowerCase()
  if (source.includes('repair')) return 'repair'
  if (source.includes('purchase')) return 'purchase'
  if (source.includes('sales')) return 'sales'
  return detailType.value === 'receivables' ? 'sales' : 'purchase'
}

async function openInventoryAssetDetail(row) {
  if (!row?.productCode && !row?.productName) return
  const data = await listBusiness('inventory', {
    page: 1,
    pageSize: 1,
    keyword: row.productCode || row.productName
  }).catch(() => ({ list: [] }))
  const target = data.list?.[0]
  if (!target?.id) {
    ElMessage.warning('未找到对应库存详情')
    return
  }
  await openBusinessDetail('inventory', target.id)
}

function resetInventoryTabs() {
  for (const key of ['purchaseSources', 'salesTrace', 'inventoryMovements']) {
    inventoryTab[key].page = 1
    inventoryTab[key].pageSize = inventoryTab[key].pageSize || 10
    inventoryTab[key].list = []
    inventoryTab[key].total = 0
  }
}

function inventoryTabParams(tabName = inventoryTab.active) {
  const state = inventoryTab[tabName]
  const filters = inventoryTab.filters[tabName] || {}
  const dates = filters.dates || []
  return {
    id: detail.value?.id,
    productId: detail.value?.productId,
    productCode: detail.value?.productCode,
    tab: tabName,
    page: state.page,
    pageSize: state.pageSize,
    sortBy: state.sortBy,
    order: state.order,
    keyword: filters.keyword || '',
    status: filters.status || '',
    customerName: filters.customerName || '',
    sourceType: filters.sourceType || '',
    startDate: dates[0] || '',
    endDate: dates[1] || ''
  }
}

async function loadInventoryTab() {
  if (!detail.value || detailType.value !== 'inventory') return
  const tabName = inventoryTab.active
  const data = await runBusinessAction('inventory', 'detail-tab', inventoryTabParams(tabName))
  inventoryTab[tabName].list = formatDateRows(data.list || [])
  inventoryTab[tabName].total = data.total || 0
}

async function reloadInventoryTab() {
  inventoryTab[inventoryTab.active].page = 1
  await loadInventoryTab()
}

async function onInventoryTabSort({ prop, order }) {
  const tab = inventoryTab[inventoryTab.active]
  tab.sortBy = prop || tab.sortBy
  tab.order = order === 'ascending' ? 'asc' : 'desc'
  await reloadInventoryTab()
}

async function salesAction(row, action) {
  if (!row?.id) return
  if (action === 'print') {
    openPrintPage('sales', row.id)
    return
  }
  ElMessage.warning('该销售操作已停用')
}

async function repairPrint(row, action) {
  if (!row?.id) return
  const printType = { 'quote-print': 'quote-print', 'settlement-print': 'settlement-print', 'repair-print': 'repair-print' }[action] || 'repair-print'
  const label = { 'repair-print': '维修单', 'quote-print': '维修报价单', 'settlement-print': '维修结算单' }[action] || '维修单'
  openPrintPage('repair', row.id, printType)
  ElMessage.success(`${label}已打开`)
}

function copySales(row) {
  dialog.visible = true
  dialog.editing = false
  dialog.id = null
  resetForm({ ...row, orderNo: '', status: '草稿' })
}

function goPurchase(orderNoOrId) {
  openBusinessDetail('purchase', orderNoOrId)
}

function goSales(orderNoOrId) {
  openBusinessDetail('sales', orderNoOrId)
}

function goMovementBusiness(row) {
  const module = normalizeBusinessModule(row.sourceType || row.businessType)
  const id = row.sourceId || row.businessId || row.purchaseOrderId || row.salesOrderId
  if (!module || !id) return
  openBusinessDetail(module, id)
}

function normalizeBusinessModule(value) {
  const type = String(value || '').toLowerCase()
  const map = {
    purchase: 'purchase',
    '采购入库': 'purchase',
    '采购退货': 'purchase',
    purchase_delete: 'purchase',
    sales: 'sales',
    sale: 'sales',
    '销售出库': 'sales',
    '销售退货': 'sales',
    sales_return: 'sales',
    order_delete: 'sales',
    '销售删除恢复': 'sales',
    '采购删除冲销': 'purchase',
    repair: 'repair',
    '维修领料': 'repair',
    project: 'project',
    '工程领料': 'project'
  }
  return map[type] || value
}

function canViewBusiness(module) {
  const permissionMap = {
    sales: 'sales.manage',
    purchase: 'purchase.manage',
    inventory: 'inventory.manage',
    repair: 'repair.manage',
    project: 'project.manage',
    finance: 'finance.manage',
    'customer-statements': 'finance.manage',
    'document-delete-records': 'system.audit.view'
  }
  const permission = permissionMap[module]
  return !permission || authStore.hasPermission(permission)
}

async function printDetail(type) {
  if (detail.value?.id) {
    openPrintPage(type, detail.value.id)
  }
}

function openPrintPage(module, id, type = '') {
  if (!id) return
  const params = new URLSearchParams({ paper: '241mm 140mm' })
  if (type) params.set('type', type)
  window.open(`${apiOrigin}/print/${module}/${id}?${params.toString()}`, '_blank')
}

function sourceTypeLabel(value) {
  if (value === undefined || value === null || value === '') return '-'
  if (sourceTypeLabels[value]) return sourceTypeLabels[value]
  const item = movementSourceOptions.find((option) => option.value === value)
  return item?.label || value || '-'
}

function isStatusField(prop) {
  return ['status', 'repairStatus'].includes(prop)
}

function isBusinessNoField(prop) {
  return ['orderNo', 'projectNo', 'recordNo', 'receivableNo', 'statementNo', 'payableNo', 'documentNo'].includes(prop)
}

function isAccountSettled(row) {
  const status = String(row?.status || '').toLowerCase()
  return Number(row?.balanceAmount || 0) <= 0 || ['paid', 'settled', 'closed', 'written_off', '已收款', '已付款', '已结清', '已核销'].includes(status)
}

function money(value) {
  return Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function reportRowValue(row, type) {
  if (type === 'profit') return `利润 ${money(row.profitAmount)} / 销售 ${money(row.salesAmount)}`
  if (type === 'trendProfit') return `${money(row.profitAmount)}`
  if (type === 'aging') return `${row.quantity || 0} 件 / ${money(row.amount)} / ${percent(row.ratio)}`
  if (type === 'assetTrend') return `${row.quantity || 0} 件 / ${money(row.inventoryAmount)}`
  if (type === 'asset') return `${row.quantity || 0} 件 / ${money(row.inventoryAmount)}`
  return money(row.value || row.amount || row.inventoryAmount || row.profitAmount)
}

function exportReportCsv() {
  if (!isReportPage.value) return
  const headers = displayColumns.value
  const lines = [
    headers.map((field) => csvCell(field.label)).join(','),
    ...rows.value.map((row) => headers.map((field) => {
      const value = field.prop === 'profitRate' ? percent(row[field.prop]) : (row[field.prop] ?? '')
      return csvCell(value)
    }).join(','))
  ]
  const blob = new Blob(['\ufeff' + lines.join('\r\n')], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `${config.value.title}-${new Date().toISOString().slice(0, 10)}.csv`
  link.click()
  URL.revokeObjectURL(link.href)
}

function csvCell(value) {
  return `"${String(value ?? '').replace(/"/g, '""')}"`
}

function printReport() {
  window.print()
}

function prettyJSON(value) {
  if (!value) return '{}'
  try {
    return JSON.stringify(typeof value === 'string' ? JSON.parse(value) : value, null, 2)
  } catch {
    return String(value)
  }
}

function percent(value) {
  return `${Number(value || 0).toFixed(2)}%`
}

function nextTickPrint() {
  return new Promise((resolve) => setTimeout(resolve, 80))
}

async function openPhotos(row) {
  photoDialog.row = row
  photoDialog.scene = config.value.photoScene || 'general'
  photoDialog.visible = true
  photos.value = await listBusinessPhotos(props.type, row.id)
}

function triggerPhoto(source) {
  if (!photoDialog.row) return
  photoTarget = { module: props.type, row: photoDialog.row, scene: photoDialog.scene }
  if (source === 'camera') {
    photoInput.value.click()
  } else {
    albumInput.value.click()
  }
}

async function handlePhoto(event) {
  const files = Array.from(event.target.files || [])
  if (files.length === 0) return
  if (!photoTarget?.row) {
    event.target.value = ''
    return
  }
  for (const file of files) {
    await uploadBusinessPhoto(photoTarget.module, photoTarget.row.id, file, photoTarget.scene)
  }
  photos.value = await listBusinessPhotos(photoTarget.module, photoTarget.row.id)
  ElMessage.success(`已上传${files.length}张照片`)
  event.target.value = ''
}

function primaryText(row) {
  return row.orderNo || row.projectNo || row.recordNo || row.receivableNo || row.statementNo || row.payableNo || row.productName || row.name || `#${row.id}`
}

function secondaryText(row) {
  return row.customerName || row.supplierName || row.deviceName || row.accountName || row.warehouse || row.code || ''
}

function statusText(row) {
  return statusLabel(row.status || row.repairStatus || '正常')
}

function amountText(row) {
  return row.totalAmount || row.amount || row.settleAmount || row.quantity || '0'
}

function dateText(row) {
  return formatDateTime(row.orderDate || row.stockTime || row.registeredAt || row.startDate || row.endDate || row.occurredAt || row.createdAt || '')
}

function currentDateTimeValue() {
  return new Date().toISOString().replace(/\.\d{3}Z$/, 'Z')
}

function defaultFieldValue(field) {
  if (dialog.editing) return ''
  if (field.type === 'date') return currentDateTimeValue()
  if (field.type === 'select') return field.options?.[0]?.value || ''
  if (field.type === 'number') return field.default ?? 0
  return ''
}
</script>

<style scoped>
.profile-page {
  display: grid;
  gap: 14px;
}

.profile-card,
.profile-log-panel {
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 8px;
  box-shadow: var(--shadow);
}

.profile-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px;
}

.profile-avatar {
  display: grid;
  place-items: center;
  width: 56px;
  height: 56px;
  flex: 0 0 auto;
  color: #fff;
  background: var(--primary);
  border-radius: 50%;
  font-size: 22px;
  font-weight: 800;
}

.profile-card h3 {
  margin: 0 0 6px;
  font-size: 20px;
}

.profile-card p {
  margin: 2px 0;
  color: var(--muted);
}

.profile-log-panel {
  padding: 14px;
}

.profile-pagination {
  margin-top: 12px;
}

.report-panel {
  display: grid;
  gap: 14px;
}

.report-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.report-card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 12px;
}

.report-card,
.report-block {
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 8px;
  box-shadow: var(--shadow);
}

.report-card {
  display: grid;
  gap: 6px;
  padding: 14px;
}

.report-card span,
.report-card small {
  color: var(--muted);
}

.report-card strong {
  font-size: 20px;
  color: var(--text);
}

.report-analysis-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 12px;
}

.report-block {
  padding: 14px;
}

.report-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 9px 0;
  border-bottom: 1px solid var(--line);
}

.report-row:last-child {
  border-bottom: 0;
}

.report-row span {
  min-width: 0;
  color: var(--text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.report-row strong {
  flex: 0 0 auto;
  color: var(--text);
  font-size: 13px;
}

.photo-toolbar {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}

.photo-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 12px;
}

.photo-card {
  display: grid;
  gap: 6px;
  color: inherit;
  text-decoration: none;
}

.photo-card img {
  width: 100%;
  aspect-ratio: 4 / 3;
  object-fit: cover;
  border-radius: 8px;
  border: 1px solid var(--el-border-color);
}

.photo-card span {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.purchase-source-picker {
  display: grid;
  gap: 10px;
  min-width: 0;
}

.source-table {
  width: 100%;
}

.sales-detail {
  display: grid;
  gap: 16px;
}

.detail-section {
  display: grid;
  gap: 10px;
}

.detail-section h3 {
  margin: 0;
  font-size: 16px;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px 12px;
  padding: 12px;
  background: var(--panel-soft);
  border: 1px solid var(--line);
  border-radius: 8px;
}

.detail-grid span {
  min-width: 0;
  overflow-wrap: anywhere;
}

.json-preview {
  max-height: 360px;
  overflow: auto;
  margin: 0;
  padding: 12px;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--panel-soft);
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  font-size: 12px;
  line-height: 1.5;
}

.inventory-tabs {
  min-width: 0;
}

.tab-filter {
  display: grid;
  grid-template-columns: minmax(180px, 1fr) minmax(160px, 220px) minmax(240px, 320px);
  gap: 10px;
  margin-bottom: 12px;
  align-items: center;
}

.statement-filter {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 12px;
}

.statement-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.statement-actions span {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.settings-panel {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px;
  margin-bottom: 16px;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--panel);
}

.settings-panel h3 {
  margin: 0 0 6px;
  font-size: 16px;
}

.settings-panel p {
  margin: 0;
  color: var(--text-muted);
  line-height: 1.5;
}

.inline-editor {
  display: grid;
  gap: 8px;
  width: 100%;
  min-width: 0;
}

.inline-editor :deep(.el-input-number) {
  width: 100%;
}

.repair-line {
  display: grid;
  gap: 8px;
  padding: 10px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--panel-soft);
}

.repair-line-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.repair-line-head strong {
  font-size: 13px;
  color: var(--text);
}

.repair-line-grid {
  display: grid;
  grid-template-columns: minmax(180px, 1.2fr) minmax(200px, 1.4fr) minmax(110px, .6fr) minmax(110px, .6fr);
  gap: 10px;
  align-items: end;
}

.repair-line-grid.outsource-grid {
  grid-template-columns: minmax(180px, 1fr) minmax(160px, 1fr) minmax(120px, .6fr);
}

.repair-line-grid label {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.repair-line-grid label > span {
  font-size: 12px;
  color: var(--text-muted);
}

@media (max-width: 767px) {
  .detail-grid {
    grid-template-columns: 1fr;
  }

  .repair-line-grid,
  .repair-line-grid.outsource-grid {
    grid-template-columns: 1fr;
  }

  .tab-filter,
  .statement-filter {
    grid-template-columns: 1fr;
  }

  .statement-actions {
    align-items: flex-start;
    flex-direction: column;
  }

  .settings-panel {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
