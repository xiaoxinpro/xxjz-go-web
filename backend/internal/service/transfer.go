package service

import (
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/config"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/repository"
)

type TransferService struct {
	cfg          *config.Config
	transferRepo *repository.TransferRepo
	fundsRepo    *repository.FundsRepo
}

func NewTransferService(cfg *config.Config, transferRepo *repository.TransferRepo, fundsRepo *repository.FundsRepo) *TransferService {
	return &TransferService{cfg: cfg, transferRepo: transferRepo, fundsRepo: fundsRepo}
}

// AddTransfer validates and inserts one transfer. time is Unix timestamp; if 0, today 00:00:00 is used.
// Returns ok, msg, tid.
func (s *TransferService) AddTransfer(uid int64, money float64, source_fid, target_fid, time int64, mark string) (ok bool, msg string, tid int64) {
	if money <= 0 {
		return false, "转账金额必须大于0", 0
	}
	if s.cfg.Money.MaxValue > 0 && money > s.cfg.Money.MaxValue {
		return false, "金额超出允许范围", 0
	}
	if source_fid == target_fid {
		return false, "转出账户与转入账户不能相同", 0
	}
	sf, _ := s.fundsRepo.GetByID(source_fid, uid)
	if sf == nil {
		return false, "转出账户不存在或无权使用", 0
	}
	tf, _ := s.fundsRepo.GetByID(target_fid, uid)
	if tf == nil {
		return false, "转入账户不存在或无权使用", 0
	}
	if len(mark) > s.cfg.Limits.MaxMarkValue {
		return false, "备注过长", 0
	}
	if time == 0 {
		// caller (handler) should set to today 00:00:00
	}
	tid, err := s.transferRepo.Insert(uid, money, source_fid, target_fid, time, mark)
	if err != nil {
		return false, "保存失败: " + err.Error(), 0
	}
	return true, "转账成功", tid
}

// DeleteTransfer deletes one transfer by tid and uid. Returns ok, msg.
func (s *TransferService) DeleteTransfer(uid, tid int64) (ok bool, msg string) {
	n, err := s.transferRepo.Delete(tid, uid)
	if err != nil {
		return false, "删除失败: " + err.Error()
	}
	if n == 0 {
		return false, "记录不存在或无权删除"
	}
	return true, "删除成功"
}
