package gormstore

import (
	"errors"

	"github.com/dujiao-next/internal/modules/smsactivation/domain"
	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(activation *domain.SmsActivation) error {
	return s.db.Create(activation).Error
}

func (s *Store) Update(activation *domain.SmsActivation) error {
	return s.db.Save(activation).Error
}

func (s *Store) GetLatestByOrderNo(orderNo string) (*domain.SmsActivation, error) {
	var item domain.SmsActivation
	err := s.db.Where("order_no = ?", orderNo).Order("id DESC").First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (s *Store) ListByOrderNo(orderNo string) ([]domain.SmsActivation, error) {
	var list []domain.SmsActivation
	err := s.db.Where("order_no = ?", orderNo).Order("id DESC").Find(&list).Error
	return list, err
}
