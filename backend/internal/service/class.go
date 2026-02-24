package service

import (
	"strconv"

	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/config"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/repository"
)

type ClassService struct {
	cfg  *config.Config
	repo *repository.ClassRepo
}

func NewClassService(cfg *config.Config, repo *repository.ClassRepo) *ClassService {
	return &ClassService{cfg: cfg, repo: repo}
}

// GetClassData returns map[classid]classname for API. type: 0 all, 1 in, 2 out.
func (s *ClassService) GetClassData(uid int64, typ int) (map[string]string, error) {
	list, err := s.repo.ListByUID(uid, typ)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for _, c := range list {
		out[strconv.FormatInt(c.ClassID, 10)] = c.ClassName
	}
	return out, nil
}

// GetClassAllData returns full class list for API (slice of maps).
func (s *ClassService) GetClassAllData(uid int64, typ int) (interface{}, error) {
	list, err := s.repo.ListByUID(uid, typ)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(list))
	for _, c := range list {
		out = append(out, map[string]interface{}{
			"classid":   c.ClassID,
			"classname": c.ClassName,
			"classtype": c.ClassType,
			"ufid":      c.Ufid,
			"sort":      c.Sort,
		})
	}
	return out, nil
}

func (s *ClassService) AddNewClass(classname string, classtype int, uid int64) (interface{}, error) {
	if len(classname) > s.cfg.Limits.MaxClassName {
		return []interface{}{false, "分类名太长"}, nil
	}
	id, err := s.repo.Create(classname, classtype, uid, 255)
	if err != nil {
		return []interface{}{false, "添加失败"}, nil
	}
	return []interface{}{true, id}, nil
}

func (s *ClassService) EditClassName(classID int64, name string, classtype int, uid int64) (interface{}, error) {
	if classtype != 1 && classtype != 2 {
		return []interface{}{false, "类别无效"}, nil
	}
	if err := s.repo.UpdateNameAndType(classID, uid, name, classtype); err != nil {
		return []interface{}{false, "更新失败"}, nil
	}
	return []interface{}{true, "更新成功"}, nil
}

func (s *ClassService) DelClass(classID, uid int64) (interface{}, error) {
	if err := s.repo.Delete(classID, uid); err != nil {
		return []interface{}{false, "删除失败"}, nil
	}
	return []interface{}{true, "删除成功"}, nil
}
