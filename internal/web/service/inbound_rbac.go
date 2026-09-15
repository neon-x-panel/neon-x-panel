package service

import (
	"strings"
	"time"

	"github.com/neon-x-panel/NEON-X-PANEL/v3/internal/database"
	"github.com/neon-x-panel/NEON-X-PANEL/v3/internal/database/model"
	"github.com/neon-x-panel/NEON-X-PANEL/v3/internal/util/common"
	"github.com/neon-x-panel/NEON-X-PANEL/v3/internal/xray"
)

func (s *InboundService) DisableClientsByOwnerAdminID(ownerAdminID int) (bool, int64, error) {
	return s.disableClientsByOwnerAdminID(ownerAdminID, 0)
}
func (s *InboundService) DisableClientsByDisabledOwnerAdminID(ownerAdminID int) (bool, int64, error) {
	return s.disableClientsByOwnerAdminID(ownerAdminID, ownerAdminID)
}
func (s *InboundService) disableClientsByOwnerAdminID(ownerAdminID int, disabledByOwnerAdminID int) (bool, int64, error) {
	if ownerAdminID <= 0 {
		return false, 0, common.NewError("invalid admin id")
	}
	db := database.GetDB()
	if db == nil {
		return false, 0, common.NewError("database is not initialized")
	}
	var rawEmails []string
	if err := db.Model(&model.ClientRecord{}).Where("owner_admin_id = ? AND enable = ?", ownerAdminID, true).Pluck("email", &rawEmails).Error; err != nil {
		return false, 0, err
	}
	emails := FilterVisibleClientEmails(rawEmails)
	if len(emails) == 0 {
		return false, 0, nil
	}
	emailSet := make(map[string]struct{}, len(emails))
	clean := make([]string, 0, len(emails))
	for _, e := range emails {
		e = strings.TrimSpace(e)
		if e == "" { continue }
		k := strings.ToLower(e)
		if _, dup := emailSet[k]; dup { continue }
		emailSet[k] = struct{}{}
		clean = append(clean, e)
	}
	if len(clean) == 0 {
		return false, 0, nil
	}
	// collect inbound ids for these emails
	type target struct {
		InboundID int  `gorm:"column:inbound_id"`
		Email     string
	}
	var targets []target
	if err := db.Raw(`
		SELECT inbounds.id AS inbound_id, clients.email AS email
		FROM clients
		JOIN client_inbounds ON client_inbounds.client_id = clients.id
		JOIN inbounds ON inbounds.id = client_inbounds.inbound_id
		WHERE clients.email IN ?
	`, clean).Scan(&targets).Error; err != nil {
		return false, 0, err
	}
	byInbound := make(map[int]map[string]struct{})
	for _, tt := range targets {
		if tt.Email == "" { continue }
		if byInbound[tt.InboundID] == nil {
			byInbound[tt.InboundID] = make(map[string]struct{})
		}
		byInbound[tt.InboundID][tt.Email] = struct{}{}
	}
	needRestart := false
	for inboundID, group := range byInbound {
		if _, _, mErr := s.markClientsDisabledInSettings(db, inboundID, group); mErr != nil {
			needRestart = true
		}
	}
	now := time.Now().UnixMilli()
	if err := db.Model(xray.ClientTraffic{}).Where("email IN ?", clean).Update("enable", false).Error; err != nil {
		return needRestart, int64(len(clean)), err
	}
	updates := map[string]any{"enable": false, "disabled_by_owner_admin_id": disabledByOwnerAdminID, "updated_at": now}
	if disabledByOwnerAdminID <= 0 {
		updates["disabled_by_owner_admin_id"] = 0
	}
	if err := db.Model(&model.ClientRecord{}).Where("owner_admin_id = ? AND email IN ?", ownerAdminID, clean).Updates(updates).Error; err != nil {
		return needRestart, int64(len(clean)), err
	}
	return needRestart, int64(len(clean)), nil
}

func (s *InboundService) EnableClientsDisabledByOwnerAdminID(ownerAdminID int) (bool, int64, error) {
	if ownerAdminID <= 0 {
		return false, 0, common.NewError("invalid admin id")
	}
	db := database.GetDB()
	if db == nil {
		return false, 0, common.NewError("database is not initialized")
	}
	now := time.Now().UnixMilli()
	var rawEmails []string
	if err := db.Table("clients AS c").Select("c.email").
		Joins("JOIN client_traffics AS ct ON ct.email = c.email").
		Where("c.owner_admin_id = ? AND c.disabled_by_owner_admin_id = ? AND c.enable = ? AND ct.enable = ?", ownerAdminID, ownerAdminID, false, false).
		Where("(ct.total <= 0 OR COALESCE(ct.up,0)+COALESCE(ct.down,0) < ct.total)").
		Where("(ct.expiry_time <= 0 OR ct.expiry_time > ?)", now).
		Pluck("c.email", &rawEmails).Error; err != nil {
		return false, 0, err
	}
	emails := FilterVisibleClientEmails(rawEmails)
	if len(emails) == 0 {
		return false, 0, nil
	}
	clean := dedupEmails(emails)
	if len(clean) == 0 {
		return false, 0, nil
	}
	if err := db.Model(xray.ClientTraffic{}).Where("email IN ?", clean).Update("enable", true).Error; err != nil {
		return false, int64(len(clean)), err
	}
	if err := db.Model(&model.ClientRecord{}).Where("owner_admin_id = ? AND disabled_by_owner_admin_id = ? AND email IN ?", ownerAdminID, ownerAdminID, clean).Updates(map[string]any{"enable": true, "disabled_by_owner_admin_id": 0, "updated_at": now}).Error; err != nil {
		return false, int64(len(clean)), err
	}
	return true, int64(len(clean)), nil
}

func (s *InboundService) EnableEligibleClientsByOwnerAdminID(ownerAdminID int) (bool, int64, error) {
	if ownerAdminID <= 0 {
		return false, 0, common.NewError("invalid admin id")
	}
	db := database.GetDB()
	if db == nil {
		return false, 0, common.NewError("database is not initialized")
	}
	now := time.Now().UnixMilli()
	var rawEmails []string
	if err := db.Table("clients AS c").Select("c.email").
		Joins("JOIN client_traffics AS ct ON ct.email = c.email").
		Where("c.owner_admin_id = ? AND c.enable = ? AND ct.enable = ?", ownerAdminID, false, false).
		Where("(ct.total <= 0 OR COALESCE(ct.up,0)+COALESCE(ct.down,0) < ct.total)").
		Where("(ct.expiry_time <= 0 OR ct.expiry_time > ?)", now).
		Pluck("c.email", &rawEmails).Error; err != nil {
		return false, 0, err
	}
	emails := FilterVisibleClientEmails(rawEmails)
	if len(emails) == 0 {
		return false, 0, nil
	}
	clean := dedupEmails(emails)
	if len(clean) == 0 {
		return false, 0, nil
	}
	if err := db.Model(xray.ClientTraffic{}).Where("email IN ?", clean).Update("enable", true).Error; err != nil {
		return false, int64(len(clean)), err
	}
	if err := db.Model(&model.ClientRecord{}).Where("owner_admin_id = ? AND email IN ?", ownerAdminID, clean).Updates(map[string]any{"enable": true, "disabled_by_owner_admin_id": 0, "updated_at": now}).Error; err != nil {
		return false, int64(len(clean)), err
	}
	return true, int64(len(clean)), nil
}

func dedupEmails(emails []string) []string {
	m := make(map[string]struct{}, len(emails))
	out := make([]string, 0, len(emails))
	for _, e := range emails {
		e = strings.TrimSpace(e)
		if e == "" { continue }
		k := strings.ToLower(e)
		if _, ok := m[k]; ok { continue }
		m[k] = struct{}{}
		out = append(out, e)
	}
	return out
}
