package service

import (
	"time"

	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/repository"
)

type StatisticService struct {
	accountRepo *repository.AccountRepo
}

func NewStatisticService(accountRepo *repository.AccountRepo) *StatisticService {
	return &StatisticService{accountRepo: accountRepo}
}

// AccountStatisticProcess returns the same structure as PHP AccountStatisticProcess (map of string keys to float64).
func (s *StatisticService) AccountStatisticProcess(uid int64) (map[string]interface{}, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	todayEnd := today.Add(24*time.Hour - time.Second)

	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	lastDay := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, time.Local)
	monthEnd := lastDay

	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.Local)
	yearEnd := time.Date(now.Year(), 12, 31, 23, 59, 59, 0, time.Local)

	out := make(map[string]interface{})
	out["TodayDate"] = today.Format("2006-01-02")

	sum := func(typ int, start, end time.Time) float64 {
		v, _ := s.accountRepo.SumByTimeAndType(uid, typ, start, end)
		return v
	}

	out["TodayInMoney"] = sum(1, today, todayEnd)
	out["TodayOutMoney"] = sum(2, today, todayEnd)
	out["MonthInMoney"] = sum(1, monthStart, monthEnd)
	out["MonthOutMoney"] = sum(2, monthStart, monthEnd)
	out["YearInMoney"] = sum(1, yearStart, yearEnd)
	out["YearOutMoney"] = sum(2, yearStart, yearEnd)

	recent7Start := today.AddDate(0, 0, -7)
	out["Recent7DayInMoney"] = sum(1, recent7Start, todayEnd)
	out["Recent7DayOutMoney"] = sum(2, recent7Start, todayEnd)
	recent30Start := today.AddDate(0, 0, -30)
	out["Recent30DayInMoney"] = sum(1, recent30Start, todayEnd)
	out["Recent30DayOutMoney"] = sum(2, recent30Start, todayEnd)
	recent60Start := today.AddDate(0, 0, -60)
	out["Recent60DayInMoney"] = sum(1, recent60Start, todayEnd)
	out["Recent60DayOutMoney"] = sum(2, recent60Start, todayEnd)
	recent90Start := today.AddDate(0, 0, -90)
	out["Recent90DayInMoney"] = sum(1, recent90Start, todayEnd)
	out["Recent90DayOutMoney"] = sum(2, recent90Start, todayEnd)
	recent180Start := today.AddDate(0, 0, -180)
	out["Recent180DayInMoney"] = sum(1, recent180Start, todayEnd)
	out["Recent180DayOutMoney"] = sum(2, recent180Start, todayEnd)
	recent365Start := today.AddDate(0, 0, -365)
	out["Recent365DayInMoney"] = sum(1, recent365Start, todayEnd)
	out["Recent365DayOutMoney"] = sum(2, recent365Start, todayEnd)

	yesterday := today.AddDate(0, 0, -1)
	yesterdayEnd := yesterday.Add(24*time.Hour - time.Second)
	out["LastTodayInMoney"] = sum(1, yesterday, yesterdayEnd)
	out["LastTodayOutMoney"] = sum(2, yesterday, yesterdayEnd)

	lastMonthStart := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.Local)
	lastMonthEnd := time.Date(now.Year(), now.Month(), 0, 23, 59, 59, 0, time.Local)
	out["LastMonthInMoney"] = sum(1, lastMonthStart, lastMonthEnd)
	out["LastMonthOutMoney"] = sum(2, lastMonthStart, lastMonthEnd)

	lastYearStart := time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, time.Local)
	lastYearEnd := time.Date(now.Year()-1, 12, 31, 23, 59, 59, 0, time.Local)
	out["LastYearInMoney"] = sum(1, lastYearStart, lastYearEnd)
	out["LastYearOutMoney"] = sum(2, lastYearStart, lastYearEnd)

	out["SumInMoney"], _ = s.accountRepo.SumByTimeAndType(uid, 1, time.Time{}, time.Time{})
	out["SumOutMoney"], _ = s.accountRepo.SumByTimeAndType(uid, 2, time.Time{}, time.Time{})

	return out, nil
}
