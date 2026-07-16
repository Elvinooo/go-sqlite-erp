package system

type SettingRequest struct {
	GroupName    string `json:"groupName" binding:"required,max=64"`
	SettingKey   string `json:"settingKey" binding:"required,max=128"`
	SettingValue string `json:"settingValue"`
	ValueType    string `json:"valueType" binding:"required,max=32"`
	IsPublic     bool   `json:"isPublic"`
	Remark       string `json:"remark"`
}

type MerchantInfoRequest struct {
	CompanyName  string `json:"companyName"`
	ContactName  string `json:"contactName"`
	ContactPhone string `json:"contactPhone"`
}
