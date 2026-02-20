package service

import (
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/config"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/repository"
)

type FundsService struct {
	cfg  *config.Config
	repo *repository.FundsRepo
}

func NewFundsService(cfg *config.Config, repo *repository.FundsRepo) *FundsService {
	return &FundsService{cfg: cfg, repo: repo}
}

// GetFundsData returns list of funds for API (slice of map with fundsid, fundsname, etc).
func (s *FundsService) GetFundsData(uid int64) (interface{}, error) {
	list, err := s.repo.ListByUID(uid)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(list))
	for _, f := range list {
		out = append(out, map[string]interface{}{
			"fundsid":   f.FundsID,
			"fundsname": f.FundsName,
			"uid":       f.UID,
			"sort":      f.Sort,
		})
	}
	return out, nil
}

func (s *FundsService) GetFundsIdData(fundsID, uid int64) (interface{}, error) {
	f, err := s.repo.GetByID(fundsID, uid)
	if err != nil || f == nil {
		return nil, err
	}
	return map[string]interface{}{
		"fundsid":   f.FundsID,
		"fundsname": f.FundsName,
		"uid":       f.UID,
		"sort":      f.Sort,
	}, nil
}

func (s *FundsService) AddNewFunds(name string, money float64, uid int64) (interface{}, error) {
	if len(name) > s.cfg.Limits.MaxFundsName {
		return []interface{}{false, "资金账户名太长"}, nil
	}
	id, err := s.repo.Create(name, uid, 255)
	if err != nil {
		return []interface{}{false, "添加失败"}, nil
	}
	return []interface{}{true, id}, nil
}

func (s *FundsService) EditFundsName(fundsID int64, name string, uid int64) (interface{}, error) {
	if err := s.repo.UpdateName(fundsID, uid, name); err != nil {
		return []interface{}{false, "更新失败"}, nil
	}
	return []interface{}{true, "更新成功"}, nil
}

func (s *FundsService) DeleteFunds(oldID, uid int64, newID int64) (interface{}, error) {
	if err := s.repo.Delete(oldID, uid); err != nil {
		return []interface{}{false, "删除失败"}, nil
	}
	return []interface{}{true, "删除成功"}, nil
}
