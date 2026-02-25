package service

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/config"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/repository"
)

// ImageItem for API response: id, name, url, time.
type ImageItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Time int64  `json:"time"`
}

type ImageService struct {
	cfg       *config.Config
	imageRepo *repository.ImageRepo
	uploadDir string
}

func NewImageService(cfg *config.Config, imageRepo *repository.ImageRepo, uploadDir string) *ImageService {
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	return &ImageService{cfg: cfg, imageRepo: imageRepo, uploadDir: uploadDir}
}

// BuildURL returns the public URL for an image (savepath + savename).
func (s *ImageService) BuildURL(savepath, savename string) string {
	if s.cfg.Image.CacheURL != "" {
		return strings.TrimSuffix(s.cfg.Image.CacheURL, "/") + "/" + strings.TrimPrefix(savepath, "/") + savename
	}
	return "/uploads/" + strings.TrimPrefix(savepath, "/") + savename
}

// GetImages returns list of image items with URL for the account.
func (s *ImageService) GetImages(uid, acid int64) ([]ImageItem, error) {
	rows, err := s.imageRepo.ListByAcid(uid, acid)
	if err != nil {
		return nil, err
	}
	out := make([]ImageItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, ImageItem{
			ID:   r.ID,
			Name: r.Name,
			URL:  s.BuildURL(r.Savepath, r.Savename),
			Time: r.Time,
		})
	}
	return out, nil
}

// SetImageAcid binds image to account (set_image). Returns true if updated.
func (s *ImageService) SetImageAcid(uid, imageID, acid int64) (bool, error) {
	n, err := s.imageRepo.CountByAcid(uid, acid)
	if err != nil {
		return false, err
	}
	if n >= s.cfg.Image.MaxCount {
		return false, nil
	}
	affected, err := s.imageRepo.UpdateAcid(uid, imageID, acid)
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

// DeleteImage removes one image from DB. Caller is responsible for deleting the file.
func (s *ImageService) DeleteImage(uid, acid, imageID int64) (int64, error) {
	return s.imageRepo.Delete(uid, acid, imageID)
}

// CountByAcid returns number of images for the account.
func (s *ImageService) CountByAcid(uid, acid int64) (int, error) {
	return s.imageRepo.CountByAcid(uid, acid)
}

// ListByAcidForDelete returns image rows for an acid (for deleting files when deleting account).
func (s *ImageService) ListByAcidForDelete(uid, acid int64) ([]repository.ImageRow, error) {
	return s.imageRepo.ListByAcidForDelete(uid, acid)
}

// GetImagePath returns savepath and savename for an image (for file delete).
func (s *ImageService) GetImagePath(uid, imageID int64) (savepath, savename string, err error) {
	return s.imageRepo.GetByID(uid, imageID)
}

// DeleteAllByAcid removes all image rows for an account. Caller deletes files.
func (s *ImageService) DeleteAllByAcid(uid, acid int64) (int64, error) {
	return s.imageRepo.Delete(uid, acid, 0)
}

// AllowedExt returns whether ext (lowercase, no dot) is allowed.
func (s *ImageService) AllowedExt(ext string) bool {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	for _, e := range s.cfg.Image.AllowedExt {
		if strings.ToLower(e) == ext {
			return true
		}
	}
	return false
}

// AddImageFromFile saves content to uploadDir/savepath/savename and inserts a row. Returns the new ImageItem.
func (s *ImageService) AddImageFromFile(uid, acid int64, originalName string, size int64, ext string, content io.Reader) (*ImageItem, error) {
	if int(size) > s.cfg.Image.MaxSize {
		return nil, nil // caller can treat as "file too large"
	}
	n, _ := s.imageRepo.CountByAcid(uid, acid)
	if n >= s.cfg.Image.MaxCount {
		return nil, nil
	}
	year := time.Now().Format("2006")
	savepath := year + "/image" + strconv.FormatInt(uid, 10) + "/"
	savename := uuid.New().String() + "." + ext
	fullDir := filepath.Join(s.uploadDir, savepath)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return nil, err
	}
	fullPath := filepath.Join(fullDir, savename)
	f, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if _, err := io.Copy(f, content); err != nil {
		os.Remove(fullPath)
		return nil, err
	}
	t := time.Now().Unix()
	name := filepath.Base(originalName)
	if name == "" || name == "." {
		name = savename
	}
	id, err := s.imageRepo.Insert(uid, acid, name, ext, savepath, savename, int(size), "", t)
	if err != nil {
		os.Remove(fullPath)
		return nil, err
	}
	return &ImageItem{ID: id, Name: name, URL: s.BuildURL(savepath, savename), Time: t}, nil
}
