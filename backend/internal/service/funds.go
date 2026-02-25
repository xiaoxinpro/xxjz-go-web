package service

import (
	"fmt"
	"time"

	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/config"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/repository"
)

type FundsService struct {
	cfg         *config.Config
	repo        *repository.FundsRepo
	accountRepo *repository.AccountRepo
	transferRepo *repository.TransferRepo
}

func NewFundsService(cfg *config.Config, repo *repository.FundsRepo, accountRepo *repository.AccountRepo, transferRepo *repository.TransferRepo) *FundsService {
	return &FundsService{cfg: cfg, repo: repo, accountRepo: accountRepo, transferRepo: transferRepo}
}

// GetFundsData returns list of funds for API (slice of map with fundsid, fundsname, money { in, out, over, count }, etc).
func (s *FundsService) GetFundsData(uid int64) (interface{}, error) {
	list, err := s.repo.ListByUID(uid)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(list)+1)

	// 兼容旧版：在满足 IsShowDefaultFunds 条件时，加入「默认」账户（id = -1）
	if s.shouldShowDefaultFunds(uid, len(list)) {
		out = append(out, s.fundRowWithStats(uid, int64(-1), "默认", uid, 0))
	}

	for _, f := range list {
		out = append(out, s.fundRowWithStats(uid, f.FundsID, f.FundsName, f.UID, f.Sort))
	}
	return out, nil
}

// fundRowWithStats 返回单条资金账户 + 收入/支出/剩余/记录数（与旧版 GetFundsAccountSumData 对齐）
func (s *FundsService) fundRowWithStats(uid, fundsid int64, name string, rowUID int64, sort int) map[string]interface{} {
	row := map[string]interface{}{
		"fundsid":   fundsid,
		"fundsname": name,
		"uid":       rowUID,
		"sort":      sort,
	}
	if s.accountRepo == nil {
		row["money"] = map[string]interface{}{"in": 0, "out": 0, "over": 0, "count": 0}
		return row
	}
	sumIn, sumOut, count, err := s.accountRepo.FundsStats(uid, fundsid)
	if err != nil {
		row["money"] = map[string]interface{}{"in": 0, "out": 0, "over": 0, "count": 0}
		return row
	}
	over := sumIn - sumOut
	row["money"] = map[string]interface{}{"in": sumIn, "out": sumOut, "over": over, "count": count}
	return row
}

// shouldShowDefaultFunds 复刻旧版 IsShowDefaultFunds 逻辑：
// - 若无自定义资金账户：始终显示默认账户；
// - 若有自定义资金账户：仅当存在 fid = -1 的记账时才显示默认账户。
func (s *FundsService) shouldShowDefaultFunds(uid int64, fundsCount int) bool {
	// 无自定义账户时始终显示默认账户
	if fundsCount == 0 {
		return true
	}
	if s.accountRepo == nil {
		// 安全降级：若未注入 accountRepo，则保守起见不显示默认账户
		return false
	}
	has, err := s.accountRepo.HasDefaultFunds(uid)
	if err != nil {
		// 发生错误时不影响主流程，保持与旧版相近的体验：默认不显示
		return false
	}
	return has
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
	// 与旧版一致：初始金额 > 0 时插入一条「系统→该账户」的转账（source_fid=0 表示初始注入）
	if money > 0 && s.transferRepo != nil {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).Unix()
		mark := name + "账户的默认金额"
		_, _ = s.transferRepo.Insert(uid, money, 0, id, today, mark)
	}
	// 与旧版一致：若这是用户第一个自定义资金账户，将默认账户 -1 下的记账合并到新账户
	list, _ := s.repo.ListByUID(uid)
	if len(list) == 1 && s.accountRepo != nil && s.transferRepo != nil {
		_, _ = s.accountRepo.ReassignFunds(uid, -1, id)
		_, _ = s.transferRepo.ReassignFunds(uid, -1, id)
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
	// 对齐旧版 DeleteFunds 逻辑：支持将记账/转账从旧账户合并到新账户
	if oldID == newID {
		return []interface{}{false, "转移资金账户错误，无法删除资金账户!"}, nil
	}

	// 校验目标账户（newID），允许 -1（默认账户）或存在的资金账户
	if newID != -1 {
		f, err := s.repo.GetByID(newID, uid)
		if err != nil {
			return []interface{}{false, "删除失败"}, nil
		}
		if f == nil {
			return []interface{}{false, "待转移的资金账户不存在!"}, nil
		}
	}

	// 先转移记账数据中的 fid
	var movedCount int64
	if s.accountRepo != nil {
		n, err := s.accountRepo.ReassignFunds(uid, oldID, newID)
		if err != nil {
			return []interface{}{false, "删除失败"}, nil
		}
		movedCount = n
	}

	// 删除资金账户本身（oldID = -1 时不需要删除表记录）
	retDelete := int64(1)
	if oldID != -1 {
		if err := s.repo.Delete(oldID, uid); err != nil {
			return []interface{}{false, "删除失败"}, nil
		}
	}

	// 转移转账记录中的 source_fid / target_fid
	if s.transferRepo != nil {
		_, _ = s.transferRepo.ReassignFunds(uid, oldID, newID)
	}

	// 组合提示信息（与旧版风格接近）
	if movedCount == 0 && retDelete == 1 {
		return []interface{}{true, "资金账户删除成功。"}, nil
	}
	if movedCount > 0 && retDelete == 1 {
		return []interface{}{true, "记账数据转移" + formatInt(movedCount) + "条，资金账户删除成功。"}, nil
	}
	return []interface{}{false, "资金账户删除失败，请返回重试。"}, nil
}

// formatInt 简单的 int64 转字符串，避免在 service 中额外引入 strconv 依赖。
func formatInt(n int64) string {
	return fmt.Sprintf("%d", n)
}
