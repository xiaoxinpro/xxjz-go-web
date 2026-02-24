package service

import (
	"strconv"
	"time"

	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/repository"
)

type ChartService struct {
	accountRepo *repository.AccountRepo
	classRepo   *repository.ClassRepo
}

func NewChartService(accountRepo *repository.AccountRepo, classRepo *repository.ClassRepo) *ChartService {
	return &ChartService{accountRepo: accountRepo, classRepo: classRepo}
}

// YearData returns getYearData-compatible map for the given year: Year, InMoney, OutMoney (1..12), InSumMoney, OutSumMoney, InSumClassMoney, OutSumClassMoney, SurplusMoney, SurplusSumMoney, InClassMoney, OutClassMoney.
func (s *ChartService) YearData(uid int64, year int) (map[string]interface{}, error) {
	if year < 2000 {
		return map[string]interface{}{"Year": "FALSE"}, nil
	}
	inClasses, _ := s.classRepo.ListByUID(uid, 1)
	outClasses, _ := s.classRepo.ListByUID(uid, 2)

	inMoney := make(map[string]float64)
	outMoney := make(map[string]float64)
	surplusMoney := make(map[string]float64)
	surplusSumMoney := make(map[string]float64)
	inClassMoney := make(map[string]map[string]float64)
	outClassMoney := make(map[string]map[string]float64)
	inSumClassMoney := make(map[string]float64)
	outSumClassMoney := make(map[string]float64)

	var inSum, outSum float64

	for m := 1; m <= 12; m++ {
		monthStart := time.Date(year, time.Month(m), 1, 0, 0, 0, 0, time.Local)
		lastDay := time.Date(year, time.Month(m+1), 0, 23, 59, 59, 0, time.Local)
		key := strconv.Itoa(m)

		inM, _ := s.accountRepo.SumByTimeAndType(uid, 1, monthStart, lastDay)
		outM, _ := s.accountRepo.SumByTimeAndType(uid, 2, monthStart, lastDay)
		inMoney[key] = inM
		outMoney[key] = outM
		surplusMoney[key] = inM - outM
		inSum += inM
		outSum += outM
		surplusSumMoney[key] = inSum - outSum

		inClassMoney[key] = make(map[string]float64)
		for _, c := range inClasses {
			v, _ := s.accountRepo.SumByTimeAndTypeAndClass(uid, 1, c.ClassID, monthStart, lastDay)
			inClassMoney[key][c.ClassName] = v
			inSumClassMoney[c.ClassName] += v
		}
		outClassMoney[key] = make(map[string]float64)
		for _, c := range outClasses {
			v, _ := s.accountRepo.SumByTimeAndTypeAndClass(uid, 2, c.ClassID, monthStart, lastDay)
			outClassMoney[key][c.ClassName] = v
			outSumClassMoney[c.ClassName] += v
		}
	}

	// InClassMoney/OutClassMoney in PHP are className -> [1..12]; we built [1..12] -> className. Transpose.
	inClassByName := make(map[string]map[string]float64)
	for _, c := range inClasses {
		inClassByName[c.ClassName] = make(map[string]float64)
		for k := 1; k <= 12; k++ {
			inClassByName[c.ClassName][strconv.Itoa(k)] = inClassMoney[strconv.Itoa(k)][c.ClassName]
		}
	}
	outClassByName := make(map[string]map[string]float64)
	for _, c := range outClasses {
		outClassByName[c.ClassName] = make(map[string]float64)
		for k := 1; k <= 12; k++ {
			outClassByName[c.ClassName][strconv.Itoa(k)] = outClassMoney[strconv.Itoa(k)][c.ClassName]
		}
	}

	return map[string]interface{}{
		"Year":             year,
		"InMoney":          inMoney,
		"OutMoney":         outMoney,
		"SurplusMoney":     surplusMoney,
		"InClassMoney":     inClassByName,
		"OutClassMoney":    outClassByName,
		"InSumMoney":       inSum,
		"OutSumMoney":      outSum,
		"InSumClassMoney":  inSumClassMoney,
		"OutSumClassMoney": outSumClassMoney,
		"SurplusSumMoney":  surplusSumMoney,
	}, nil
}
