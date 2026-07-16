package controller

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	businessapp "erp/internal/application/business"
	"erp/internal/shared/query"
	"erp/internal/shared/response"
	"erp/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BusinessController struct {
	service *businessapp.Service
}

const maxPhotoUploadSize = 5 << 20

var allowedPhotoMIMEs = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".pdf":  "application/pdf",
}

func NewBusinessController(service *businessapp.Service) *BusinessController {
	return &BusinessController{service: service}
}

func (ctl *BusinessController) Modules(c *gin.Context) {
	response.OK(c, ctl.service.Modules())
}

func (ctl *BusinessController) Meta(c *gin.Context) {
	meta, err := ctl.service.Meta(c.Param("module"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, meta)
}

func (ctl *BusinessController) List(c *gin.Context) {
	q := query.FromGin(c)
	list, total, err := ctl.service.List(c.Request.Context(), c.Param("module"), utils.GetTenantID(c), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Page(c, list, total, q.Page, q.PageSize)
}

func (ctl *BusinessController) Get(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	item, err := ctl.service.Get(c.Request.Context(), c.Param("module"), utils.GetTenantID(c), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, item)
}

func (ctl *BusinessController) Create(c *gin.Context) {
	var req businessapp.ModuleRequest
	if !bind(c, &req) {
		return
	}
	item, err := ctl.service.Create(c.Request.Context(), c.Param("module"), utils.GetTenantID(c), utils.GetUserID(c), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Created(c, item)
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

func (ctl *BusinessController) Update(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req businessapp.ModuleRequest
	if !bind(c, &req) {
		return
	}
	item, err := ctl.service.Update(c.Request.Context(), c.Param("module"), utils.GetTenantID(c), id, utils.GetUserID(c), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, item)
}

func (ctl *BusinessController) Delete(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req businessapp.DeleteRequest
	if c.Request.ContentLength != 0 {
		if !bind(c, &req) {
			return
		}
	}
	if err := ctl.service.Delete(c.Request.Context(), c.Param("module"), utils.GetTenantID(c), id, utils.GetUserID(c), req); err != nil {
		_ = c.Error(err)
		return
	}
	if c.Param("module") == "sales" || c.Param("module") == "purchase" {
		c.Set("skipOperationAudit", true)
	}
	response.OK(c, gin.H{"id": id})
}

func (ctl *BusinessController) Action(c *gin.Context) {
	var req businessapp.ModuleActionRequest
	if !bind(c, &req) {
		return
	}
	result, err := ctl.service.Action(c.Request.Context(), c.Param("module"), utils.GetTenantID(c), utils.GetUserID(c), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, result)
}

func (ctl *BusinessController) Print(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.String(400, "打印单据不存在")
		return
	}
	module, title := printModule(c.Param("module"), c.Query("type"))
	if module == "" {
		c.String(404, "暂不支持该打印模板")
		return
	}
	item, err := ctl.service.Get(c.Request.Context(), module, utils.GetTenantID(c), id)
	if err != nil {
		c.String(404, "打印单据不存在")
		return
	}
	data := map[string]any{
		"title": title,
		"paper": firstText(c.DefaultQuery("paper", "241mm 140mm"), "241mm 140mm"),
		"item":  item,
		"now":   time.Now().Format("2006-01-02 15:04"),
	}
	tpl, err := template.New("print").Funcs(template.FuncMap{
		"v":     mapValue,
		"money": moneyText,
		"date":  dateText,
	}).Parse(printHTML)
	if err != nil {
		c.String(500, "打印模板加载失败")
		return
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		c.String(500, "打印模板渲染失败")
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, buf.String())
}

func (ctl *BusinessController) ListPhotos(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	list, err := ctl.service.ListPhotos(c.Request.Context(), c.Param("module"), utils.GetTenantID(c), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, list)
}

func (ctl *BusinessController) UploadPhoto(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(err)
		return
	}
	if err := validatePhotoUpload(fileHeader); err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	module := c.Param("module")
	scene := c.DefaultPostForm("scene", "general")
	if scene == "" {
		scene = c.DefaultQuery("scene", "general")
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext == "" {
		ext = ".jpg"
	}
	datePath := time.Now().Format("20060102")
	dir := filepath.Join("data", "uploads", module, datePath)
	fileName := fmt.Sprintf("%s%s", uuid.NewString(), ext)
	dst := filepath.Join(dir, fileName)
	if err := ensureDir(dir); err != nil {
		_ = c.Error(err)
		return
	}
	if err := c.SaveUploadedFile(fileHeader, dst); err != nil {
		_ = c.Error(err)
		return
	}
	urlPath := "/" + filepath.ToSlash(filepath.Join("uploads", module, datePath, fileName))
	item, err := ctl.service.CreatePhoto(c.Request.Context(), module, utils.GetTenantID(c), utils.GetUserID(c), id, map[string]any{
		"scene":       scene,
		"fileName":    fileHeader.Filename,
		"filePath":    dst,
		"fileUrl":     urlPath,
		"contentType": fileHeader.Header.Get("Content-Type"),
		"fileSize":    fileHeader.Size,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Created(c, item)
}

func validatePhotoUpload(fileHeader *multipart.FileHeader) error {
	if fileHeader.Size <= 0 {
		return fmt.Errorf("上传文件不能为空")
	}
	if fileHeader.Size > maxPhotoUploadSize {
		return fmt.Errorf("上传文件不能超过 5MB")
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	expectedType, ok := allowedPhotoMIMEs[ext]
	if !ok {
		return fmt.Errorf("仅支持 JPG、PNG、GIF、PDF 文件")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("上传文件读取失败")
	}
	defer file.Close()
	buf := make([]byte, 512)
	n, readErr := file.Read(buf)
	if readErr != nil && readErr != io.EOF {
		return fmt.Errorf("上传文件读取失败")
	}
	contentType := http.DetectContentType(buf[:n])
	if contentType != expectedType {
		return fmt.Errorf("上传文件类型与扩展名不匹配")
	}
	return nil
}

func printModule(module string, typ string) (string, string) {
	switch module {
	case "sales":
		switch typ {
		case "delivery":
			return "sales", "送货单"
		case "quote":
			return "sales", "报价单"
		default:
			return "sales", "销售单"
		}
	case "delivery":
		return "sales", "送货单"
	case "quote":
		return "sales", "报价单"
	case "purchase":
		return "purchase", "采购单"
	case "repair":
		switch typ {
		case "quote", "quote-print":
			return "repair", "维修报价单"
		case "settlement", "settlement-print":
			return "repair", "维修结算单"
		default:
			return "repair", "维修单"
		}
	case "customer-statement", "customer-statements":
		return "customer-statements", "客户对账单"
	case "payables":
		return "payables", "付款单"
	case "receipt":
		return "finance", "收款单"
	case "statement":
		return "receivables", "客户对账单"
	case "receivables":
		return "receivables", "欠款通知单"
	default:
		return "", ""
	}
}

func mapValue(row map[string]any, names ...string) any {
	for _, name := range names {
		if value, ok := row[name]; ok && fmt.Sprint(value) != "" {
			return value
		}
	}
	return ""
}

func moneyText(value any) string {
	if value == nil || fmt.Sprint(value) == "" {
		return "0.00"
	}
	text := fmt.Sprint(value)
	parsed, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return text
	}
	return fmt.Sprintf("%.2f", parsed)
}

func dateText(value any) string {
	text := fmt.Sprint(value)
	if len(text) >= 10 {
		return strings.ReplaceAll(text[:10], "T", " ")
	}
	return text
}

func firstText(items ...string) string {
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			return item
		}
	}
	return ""
}

const printHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.title}}</title>
  <style>
    :root { color: #111; font-family: "Microsoft YaHei", "PingFang SC", Arial, sans-serif; }
    * { box-sizing: border-box; }
    body { margin: 0; background: #f3f4f6; }
    .page { width: 241mm; min-height: 140mm; margin: 12px auto; padding: 8mm 10mm; background: #fff; }
    .header { display: grid; grid-template-columns: 1fr auto 1fr; align-items: end; border-bottom: 2px solid #111; padding-bottom: 5mm; }
    .brand { font-size: 12px; color: #444; }
    h1 { margin: 0; text-align: center; font-size: 22px; font-weight: 700; letter-spacing: 0; }
    .print-time { text-align: right; font-size: 12px; color: #444; }
    .meta { display: grid; grid-template-columns: repeat(4, 1fr); gap: 3mm 5mm; margin: 5mm 0; font-size: 12px; }
    .meta div { min-height: 6mm; border-bottom: 1px solid #999; padding-bottom: 1.5mm; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    table { width: 100%; border-collapse: collapse; margin-top: 3mm; font-size: 12px; table-layout: fixed; }
    th, td { border: 1px solid #555; padding: 2.5mm 2mm; text-align: left; vertical-align: middle; word-break: break-word; }
    th { background: #f2f2f2; font-weight: 700; text-align: center; }
    .col-name { width: 38%; }
    .col-qty { width: 12%; text-align: right; }
    .col-price, .col-amount { width: 15%; text-align: right; }
    .col-remark { width: 20%; }
    .summary { display: grid; grid-template-columns: repeat(3, 1fr); gap: 0; margin-top: 4mm; border: 1px solid #555; border-left: 0; font-size: 12px; }
    .summary div { border-left: 1px solid #555; padding: 2.5mm; }
    .summary strong { font-size: 14px; }
    .note { margin-top: 4mm; font-size: 11px; color: #555; }
    .sign { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12mm; margin-top: 10mm; font-size: 12px; }
    .sign div { border-top: 1px solid #111; padding-top: 2mm; text-align: center; }
    @media print {
      body { background: #fff; }
      .page { margin: 0; width: auto; min-height: auto; padding: 0; }
      @page { size: {{.paper}}; margin: 8mm; }
    }
    @media screen and (max-width: 767px) {
      .page { width: auto; min-height: auto; margin: 0; padding: 18px; }
      .header, .meta, .summary, .sign { grid-template-columns: 1fr; }
      .print-time { text-align: left; }
      h1 { text-align: left; margin: 8px 0; }
    }
  </style>
</head>
<body>
  {{$item := .item}}
  <main class="page">
    <section class="header">
      <div class="brand">ERP Pro</div>
      <h1>{{.title}}</h1>
      <div class="print-time">打印时间：{{.now}}</div>
    </section>
    <section class="meta">
      <div>单号：{{v $item "orderNo" "recordNo" "receivableNo" "payableNo" "statementNo" "projectNo"}}</div>
      <div>客户：{{if eq .title "采购单"}}{{v $item "merchantCompanyName"}}{{else}}{{v $item "customerName" "targetName"}}{{end}}</div>
      <div>供应商：{{if or (eq .title "销售单") (eq .title "送货单") (eq .title "报价单")}}{{v $item "merchantCompanyName"}}{{else}}{{v $item "supplierName"}}{{end}}</div>
      <div>日期：{{date (v $item "orderDate" "registeredAt" "occurredAt" "invoiceDate" "billDate" "startDate")}}</div>
      <div>状态：{{v $item "status" "repairStatus"}}</div>
      <div>联系人：{{if or (eq .title "销售单") (eq .title "送货单") (eq .title "报价单") (eq .title "采购单")}}{{v $item "merchantContactName"}}{{else}}{{v $item "contactName"}}{{end}}</div>
      <div>电话：{{if or (eq .title "销售单") (eq .title "送货单") (eq .title "报价单") (eq .title "采购单")}}{{v $item "merchantContactPhone"}}{{else}}{{v $item "contactPhone" "phone"}}{{end}}</div>
      <div>备注：{{v $item "remark"}}</div>
    </section>
    <table>
      <thead>
        <tr>
          <th class="col-name">商品名</th>
          <th class="col-qty">数量</th>
          <th class="col-price">价格</th>
          <th class="col-amount">金额</th>
          <th class="col-remark">备注</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>{{v $item "productName" "deviceName" "businessType" "sourceNo"}}</td>
          <td class="col-qty">{{v $item "quantity"}}</td>
          <td class="col-price">{{money (v $item "price" "amount")}}</td>
          <td class="col-amount">{{money (v $item "totalAmount" "amount")}}</td>
          <td>{{v $item "faultDesc" "businessNo"}}</td>
        </tr>
      </tbody>
    </table>
    <section class="summary">
      <div>合计金额：<strong>{{money (v $item "totalAmount" "amount")}}</strong></div>
      <div>已收/已付：<strong>{{money (v $item "receivedAmount" "paidAmount")}}</strong></div>
      <div>未收/未付：<strong>{{money (v $item "receivableAmount" "payableAmount" "balanceAmount" "unpaidAmount")}}</strong></div>
    </section>
    <p class="note">请核对商品、数量、价格和金额；本单据适配中一刀打印纸。</p>
    <section class="sign">
      <div>经办人</div>
      <div>客户确认</div>
      <div>公司盖章</div>
    </section>
  </main>
</body>
</html>`
