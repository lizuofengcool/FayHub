package service

import (
	"context"
	"fayhub/internal/model"
	errs "fayhub/pkg/errors"
	"fayhub/pkg/sanitizer"
	"fayhub/pkg/utils"

	"gorm.io/gorm"
)

type SensitiveWordService struct{}

func (s *SensitiveWordService) List(ctx context.Context, page, pageSize int, keyword, category string, level int) ([]model.SensitiveWord, int64, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var list []model.SensitiveWord
	var total int64

	query := db.Model(&model.SensitiveWord{})
	if keyword != "" {
		query = query.Where("word LIKE ?", "%"+keyword+"%")
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if level > 0 {
		query = query.Where("level = ?", level)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询敏感词总数失败")
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, errs.NewServiceError(errs.ErrDatabase, "查询敏感词列表失败")
	}

	return list, total, nil
}

func (s *SensitiveWordService) Create(ctx context.Context, word, category string, level int) (*model.SensitiveWord, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	if word == "" {
		return nil, errs.NewServiceError(errs.ErrParamValidation, "敏感词不能为空")
	}

	var existing model.SensitiveWord
	if err := db.Where("word = ?", word).First(&existing).Error; err == nil {
		return nil, errs.NewServiceError(errs.ErrConflict, "敏感词已存在")
	}

	if level <= 0 {
		level = 1
	}

	record := &model.SensitiveWord{
		Word:     word,
		Category: category,
		Level:    level,
		Status:   1,
	}

	if err := db.Create(record).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "创建敏感词失败")
	}

	sanitizer.GetMatcher().AddWord(word)

	return record, nil
}

func (s *SensitiveWordService) Update(ctx context.Context, id int64, word, category string, level, status int) (*model.SensitiveWord, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return nil, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var record model.SensitiveWord
	if err := db.First(&record, id).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrResourceNotFound, "敏感词不存在")
	}

	oldWord := record.Word

	if word != "" {
		record.Word = word
	}
	if category != "" {
		record.Category = category
	}
	if level > 0 {
		record.Level = level
	}
	if status >= 0 {
		record.Status = status
	}

	if err := db.Save(&record).Error; err != nil {
		return nil, errs.NewServiceError(errs.ErrDatabase, "更新敏感词失败")
	}

	matcher := sanitizer.GetMatcher()
	matcher.RemoveWord(oldWord)
	if record.Status == 1 {
		matcher.AddWord(record.Word)
	}

	return &record, nil
}

func (s *SensitiveWordService) Delete(ctx context.Context, id int64) error {
	db := utils.GetDB(ctx)
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var record model.SensitiveWord
	if err := db.First(&record, id).Error; err != nil {
		return errs.NewServiceError(errs.ErrResourceNotFound, "敏感词不存在")
	}

	if err := db.Delete(&record).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "删除敏感词失败")
	}

	sanitizer.GetMatcher().RemoveWord(record.Word)

	return nil
}

func (s *SensitiveWordService) BatchCreate(ctx context.Context, words []string, category string, level int) (int, error) {
	db := utils.GetDB(ctx)
	if db == nil {
		return 0, errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	if len(words) == 0 {
		return 0, errs.NewServiceError(errs.ErrParamValidation, "词列表不能为空")
	}

	count := 0
	matcher := sanitizer.GetMatcher()

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, word := range words {
			if word == "" {
				continue
			}

			var existing model.SensitiveWord
			if err := tx.Where("word = ?", word).First(&existing).Error; err == nil {
				continue
			}

			record := &model.SensitiveWord{
				Word:     word,
				Category: category,
				Level:    level,
				Status:   1,
			}

			if err := tx.Create(record).Error; err != nil {
				return err
			}

			matcher.AddWord(word)
			count++
		}
		return nil
	})

	if err != nil {
		return 0, errs.NewServiceError(errs.ErrDatabase, "批量创建敏感词失败")
	}

	return count, nil
}

func (s *SensitiveWordService) RebuildMatcher(ctx context.Context) error {
	db := utils.GetDB(utils.SkipTenantIsolation(ctx))
	if db == nil {
		return errs.NewServiceError(errs.ErrDBNotConnected, "")
	}

	var words []model.SensitiveWord
	if err := db.Where("status = 1").Find(&words).Error; err != nil {
		return errs.NewServiceError(errs.ErrDatabase, "查询敏感词失败")
	}

	wordList := make([]string, len(words))
	for i, w := range words {
		wordList[i] = w.Word
	}

	sanitizer.GetMatcher().BuildFromWords(wordList)

	return nil
}

func (s *SensitiveWordService) Check(ctx context.Context, text string) (bool, []string, string) {
	if text == "" {
		return false, nil, text
	}

	hasSensitive := sanitizer.HasSensitiveWord(text)
	if !hasSensitive {
		return false, nil, text
	}

	words := sanitizer.FindSensitiveWords(text)
	sanitized := sanitizer.SanitizeTextPreserve(text)

	return true, words, sanitized
}
