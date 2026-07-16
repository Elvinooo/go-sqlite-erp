package system

import (
	"context"
	"errors"

	systemmodel "erp/internal/domain/system/model"
	systemrepo "erp/internal/domain/system/repository"
	"erp/internal/shared/query"
	"gorm.io/gorm"
)

type Service struct {
	settings systemrepo.SettingRepository
}

func NewService(settings systemrepo.SettingRepository) *Service {
	return &Service{settings: settings}
}

func (s *Service) ListSettings(ctx context.Context, tenantID uint64, q query.PageQuery) ([]systemmodel.SystemSetting, int64, error) {
	return s.settings.List(ctx, tenantID, q)
}

func (s *Service) CreateSetting(ctx context.Context, tenantID uint64, req SettingRequest, operatorID uint64) (*systemmodel.SystemSetting, error) {
	setting := &systemmodel.SystemSetting{
		GroupName: req.GroupName, SettingKey: req.SettingKey,
		SettingValue: req.SettingValue, ValueType: req.ValueType, IsPublic: req.IsPublic,
	}
	setting.TenantID = tenantID
	setting.Remark = req.Remark
	setting.CreatedBy = &operatorID
	setting.UpdatedBy = &operatorID
	return setting, s.settings.Create(ctx, setting)
}

func (s *Service) UpdateSetting(ctx context.Context, tenantID uint64, id uint64, req SettingRequest, operatorID uint64) (*systemmodel.SystemSetting, error) {
	setting, err := s.settings.FindByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	setting.GroupName = req.GroupName
	setting.SettingKey = req.SettingKey
	setting.SettingValue = req.SettingValue
	setting.ValueType = req.ValueType
	setting.IsPublic = req.IsPublic
	setting.Remark = req.Remark
	setting.UpdatedBy = &operatorID
	return setting, s.settings.Update(ctx, setting)
}

func (s *Service) DeleteSetting(ctx context.Context, tenantID uint64, id uint64) error {
	return s.settings.Delete(ctx, tenantID, id)
}

func (s *Service) RestoreTestData(ctx context.Context, tenantID uint64, operatorID uint64) (map[string]any, error) {
	return s.settings.RestoreTestData(ctx, tenantID, operatorID)
}

func (s *Service) MerchantInfo(ctx context.Context, tenantID uint64) (map[string]any, error) {
	company, err := s.settingValue(ctx, tenantID, "merchant.company_name")
	if err != nil {
		return nil, err
	}
	contact, err := s.settingValue(ctx, tenantID, "merchant.contact_name")
	if err != nil {
		return nil, err
	}
	phone, err := s.settingValue(ctx, tenantID, "merchant.contact_phone")
	if err != nil {
		return nil, err
	}
	return map[string]any{"companyName": company, "contactName": contact, "contactPhone": phone}, nil
}

func (s *Service) SaveMerchantInfo(ctx context.Context, tenantID uint64, operatorID uint64, req MerchantInfoRequest) (map[string]any, error) {
	values := map[string]string{
		"merchant.company_name":  req.CompanyName,
		"merchant.contact_name":  req.ContactName,
		"merchant.contact_phone": req.ContactPhone,
	}
	for key, value := range values {
		if err := s.upsertSetting(ctx, tenantID, operatorID, key, value); err != nil {
			return nil, err
		}
	}
	return s.MerchantInfo(ctx, tenantID)
}

func (s *Service) settingValue(ctx context.Context, tenantID uint64, key string) (string, error) {
	setting, err := s.settings.FindByKey(ctx, tenantID, key)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return setting.SettingValue, nil
}

func (s *Service) upsertSetting(ctx context.Context, tenantID uint64, operatorID uint64, key string, value string) error {
	setting, err := s.settings.FindByKey(ctx, tenantID, key)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item := &systemmodel.SystemSetting{
			GroupName:    "merchant",
			SettingKey:   key,
			SettingValue: value,
			ValueType:    "string",
			IsPublic:     false,
		}
		item.TenantID = tenantID
		item.CreatedBy = &operatorID
		item.UpdatedBy = &operatorID
		return s.settings.Create(ctx, item)
	}
	if err != nil {
		return err
	}
	setting.SettingValue = value
	setting.UpdatedBy = &operatorID
	return s.settings.Update(ctx, setting)
}
