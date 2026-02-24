package service

import (
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/config"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/repository"
)

type AccountService struct {
	cfg         *config.Config
	accountRepo *repository.AccountRepo
	classRepo   *repository.ClassRepo
	fundsRepo   *repository.FundsRepo
}

func NewAccountService(cfg *config.Config, accountRepo *repository.AccountRepo, classRepo *repository.ClassRepo, fundsRepo *repository.FundsRepo) *AccountService {
	return &AccountService{cfg: cfg, accountRepo: accountRepo, classRepo: classRepo, fundsRepo: fundsRepo}
}

// AddAccount validates and inserts one account. actime is Unix timestamp; if 0, today 00:00:00 is used.
// Returns ok, msg, acid.
func (s *AccountService) AddAccount(uid int64, acmoney float64, acclassid, actime int64, acremark string, zhifu int64, fid int64) (ok bool, msg string, acid int64) {
	if acmoney <= 0 {
		return false, "金额必须大于0", 0
	}
	if s.cfg.Money.MaxValue > 0 && acmoney > s.cfg.Money.MaxValue {
		return false, "金额超出允许范围", 0
	}
	if zhifu != 1 && zhifu != 2 {
		return false, "收支类型无效", 0
	}
	if len(acremark) > s.cfg.Limits.MaxMarkValue {
		return false, "备注过长", 0
	}
	class, err := s.classRepo.GetByID(acclassid, uid)
	if err != nil || class == nil {
		return false, "分类不存在或无权使用", 0
	}
	if int64(class.ClassType) != zhifu {
		return false, "分类与收支类型不匹配", 0
	}
	if fid != 0 && fid != -1 {
		f, _ := s.fundsRepo.GetByID(fid, uid)
		if f == nil {
			return false, "资金账户不存在或无权使用", 0
		}
	}
	if fid == 0 {
		fid = -1
	}
	if actime == 0 {
		// use start of today in local time (handler or caller can set)
		// leave 0 and repo will need to handle; actually we should set in handler
		// for simplicity set in handler before call
	}
	acid, err = s.accountRepo.Insert(uid, acmoney, acclassid, actime, acremark, zhifu, fid)
	if err != nil {
		return false, "保存失败: " + err.Error(), 0
	}
	return true, "添加成功", acid
}

// DeleteAccount deletes one account by acid and uid. Returns ok, msg.
func (s *AccountService) DeleteAccount(uid, acid int64) (ok bool, msg string) {
	n, err := s.accountRepo.Delete(acid, uid)
	if err != nil {
		return false, "删除失败: " + err.Error()
	}
	if n == 0 {
		return false, "记录不存在或无权删除"
	}
	return true, "删除成功"
}
