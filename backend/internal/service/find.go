package service

import (
	"sort"
	"time"

	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/repository"
)

// FindService provides merged account+transfer list with pagination.
type FindService struct {
	accountRepo  *repository.AccountRepo
	transferRepo *repository.TransferRepo
	pageSize     int
}

func NewFindService(accountRepo *repository.AccountRepo, transferRepo *repository.TransferRepo, pageSize int) *FindService {
	if pageSize <= 0 {
		pageSize = 15
	}
	return &FindService{accountRepo: accountRepo, transferRepo: transferRepo, pageSize: pageSize}
}

// FindResult is the response shape for find type=all: data, page, pagemax, count, optional sums.
type FindResult struct {
	Data        []repository.FindRow `json:"data"`
	Page        int                  `json:"page"`
	PageMax     int                  `json:"pagemax"`
	Count       int                  `json:"count"`
	SumInMoney  float64              `json:"SumInMoney,omitempty"`
	SumOutMoney float64              `json:"SumOutMoney,omitempty"`
	IsTransfer  bool                 `json:"isTransfer,omitempty"`
}

// FindFilter from API data: fid, zhifu (0=all,1=收入,2=支出,3=转账), acclassid (or "inTransfer"/"outTransfer"), starttime/endtime (YYYY-MM-DD), acremark.
type FindFilter struct {
	Fid       int64
	Zhifu     int    // 0=all, 1=收入, 2=支出, 3=转账
	Acclassid int64  // or use AcclassidStr for "inTransfer"/"outTransfer"
	StartTime int64
	EndTime   int64
	Acremark  string
}

// FindTransferAccountData returns merged account and transfer rows for the user, sorted by time DESC, paginated.
func (s *FindService) FindTransferAccountData(uid int64, page int) (*FindResult, error) {
	if page < 1 {
		page = 1
	}
	acCount, err := s.accountRepo.CountByUser(uid)
	if err != nil {
		return nil, err
	}
	trCount, err := s.transferRepo.CountByUser(uid)
	if err != nil {
		return nil, err
	}
	total := acCount + trCount
	pageMax := 1
	if total > 0 && s.pageSize > 0 {
		pageMax = (total-1)/s.pageSize + 1
	}
	if page > pageMax {
		page = pageMax
	}

	// Fetch enough from each to cover the requested page when merged and sorted.
	need := page * s.pageSize
	acList, err := s.accountRepo.ListByUserWithClassFunds(uid, need, 0)
	if err != nil {
		return nil, err
	}
	trList, err := s.transferRepo.ListByUserWithFunds(uid, need, 0)
	if err != nil {
		return nil, err
	}
	merged := make([]repository.FindRow, 0, len(acList)+len(trList))
	merged = append(merged, acList...)
	merged = append(merged, trList...)
	sort.Slice(merged, func(i, j int) bool {
		if merged[i].Time != merged[j].Time {
			return merged[i].Time > merged[j].Time
		}
		return merged[i].ID > merged[j].ID
	})
	start := (page - 1) * s.pageSize
	end := start + s.pageSize
	if start > len(merged) {
		start = len(merged)
	}
	if end > len(merged) {
		end = len(merged)
	}
	pageData := merged[start:end]
	sumIn, _ := s.accountRepo.SumByTimeAndType(uid, 1, time.Time{}, time.Time{})
	sumOut, _ := s.accountRepo.SumByTimeAndType(uid, 2, time.Time{}, time.Time{})
	return &FindResult{
		Data: pageData, Page: page, PageMax: pageMax, Count: total,
		SumInMoney: sumIn, SumOutMoney: sumOut, IsTransfer: trCount > 0,
	}, nil
}

// FindTransferAccountDataFiltered returns filtered merged list and sums.
func (s *FindService) FindTransferAccountDataFiltered(uid int64, page int, f FindFilter) (*FindResult, error) {
	if page < 1 {
		page = 1
	}
	acFilter := repository.AccountFilter{
		Fid: f.Fid, Zhifu: f.Zhifu, Acclassid: f.Acclassid,
		StartTime: f.StartTime, EndTime: f.EndTime, Acremark: f.Acremark,
	}
	trFilter := repository.TransferFilter{
		Fid: f.Fid, StartTime: f.StartTime, EndTime: f.EndTime, Mark: f.Acremark,
	}
	// acclassid "inTransfer"/"outTransfer" maps to transfer direction
	if f.Zhifu == 3 {
		if f.Acclassid != 0 {
			// AcclassidStr would be "inTransfer" or "outTransfer"; we use a convention: pass via Acclassid 1=in 2=out when zhifu=3
			// For simplicity we use Acclassid: 0=all, 1=inTransfer, 2=outTransfer (only when zhifu=3)
			trFilter.Direction = int(f.Acclassid)
		}
	}
	var acList []repository.FindRow
	var trList []repository.FindRow
	var acCount, trCount int
	if f.Zhifu == 3 {
		// only transfer
		trCount, _ = s.transferRepo.CountByUserFiltered(uid, trFilter)
		need := page * s.pageSize
		trList, _ = s.transferRepo.ListByUserFiltered(uid, trFilter, need, 0)
	} else if f.Zhifu == 1 || f.Zhifu == 2 {
		// only account
		acCount, _ = s.accountRepo.CountByUserFiltered(uid, acFilter)
		need := page * s.pageSize
		acList, _ = s.accountRepo.ListByUserFiltered(uid, acFilter, need, 0)
	} else {
		acCount, _ = s.accountRepo.CountByUserFiltered(uid, acFilter)
		trCount, _ = s.transferRepo.CountByUserFiltered(uid, trFilter)
		need := page * s.pageSize
		acList, _ = s.accountRepo.ListByUserFiltered(uid, acFilter, need, 0)
		trList, _ = s.transferRepo.ListByUserFiltered(uid, trFilter, need, 0)
	}
	total := acCount + trCount
	pageMax := 1
	if total > 0 && s.pageSize > 0 {
		pageMax = (total-1)/s.pageSize + 1
	}
	if page > pageMax {
		page = pageMax
	}
	merged := make([]repository.FindRow, 0, len(acList)+len(trList))
	merged = append(merged, acList...)
	merged = append(merged, trList...)
	sort.Slice(merged, func(i, j int) bool {
		if merged[i].Time != merged[j].Time {
			return merged[i].Time > merged[j].Time
		}
		return merged[i].ID > merged[j].ID
	})
	start := (page - 1) * s.pageSize
	end := start + s.pageSize
	if start > len(merged) {
		start = len(merged)
	}
	if end > len(merged) {
		end = len(merged)
	}
	pageData := merged[start:end]
	var sumIn, sumOut float64
	if f.Zhifu == 3 {
		sumIn, sumOut, _ = s.transferRepo.SumByUserFiltered(uid, trFilter)
	} else {
		sumIn, sumOut, _ = s.accountRepo.SumByUserFiltered(uid, acFilter)
	}
	return &FindResult{
		Data: pageData, Page: page, PageMax: pageMax, Count: total,
		SumInMoney: sumIn, SumOutMoney: sumOut, IsTransfer: trCount > 0 || f.Zhifu == 3,
	}, nil
}
