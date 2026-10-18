package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/dujiao-next/internal/constants"
	mappingcontract "github.com/dujiao-next/internal/modules/catalog/mapping/contract"
	mappingdomain "github.com/dujiao-next/internal/modules/catalog/mapping/domain"
)

// 文件组织约定:
//   service.go       — 端口/错误/装配 + 基础查询与 CRUD
//   import.go        — 单品导入 + 上游元数据列表 + 图片下载
//   batch_import.go  — 批量导入与上游分类自动创建
//   sync.go          — 同步流程（单品 / 全量库存 / 下单前兜底）
//   markup.go        — 加价重算
//   pricing.go       — 汇率与加价换算

type Options struct {
	Mappings     mappingcontract.MappingRepository
	SKUMappings  mappingcontract.SKUMappingRepository
	Products     mappingcontract.ProductRepository
	SKUs         mappingcontract.SKURepository
	Categories   mappingcontract.CategoryRepository
	Connections  mappingcontract.ConnectionProvider
	Media        mappingcontract.MediaRecorder
	Transactions mappingcontract.UnitOfWork
}

// Service 承载本地商品与上游站点之间的映射用例。
type Service struct {
	mappings        mappingcontract.MappingRepository
	skuMappings     mappingcontract.SKUMappingRepository
	products        mappingcontract.ProductRepository
	skus            mappingcontract.SKURepository
	categories      mappingcontract.CategoryRepository
	connections     mappingcontract.ConnectionProvider
	media           mappingcontract.MediaRecorder
	transactions    mappingcontract.UnitOfWork
	categoryCreator mappingcontract.CategoryCreator
	settings        mappingcontract.SettingsProvider
}

func NewService(options Options) (*Service, error) {
	if options.Media == nil {
		return nil, mappingcontract.ErrMediaRecorderRequired
	}
	return &Service{
		mappings:     options.Mappings,
		skuMappings:  options.SKUMappings,
		products:     options.Products,
		skus:         options.SKUs,
		categories:   options.Categories,
		connections:  options.Connections,
		media:        options.Media,
		transactions: options.Transactions,
	}, nil
}

// SetCategoryCreator 注入分类创建端口（装配时调用）。
func (s *Service) SetCategoryCreator(creator mappingcontract.CategoryCreator) {
	s.categoryCreator = creator
}

// SetSettings 注入设置端口（用于读取上游同步动态配置）。
func (s *Service) SetSettings(settings mappingcontract.SettingsProvider) {
	s.settings = settings
}

// GetByID 获取映射详情
func (s *Service) GetByID(id uint) (*mappingdomain.Mapping, error) {
	return s.mappings.GetByID(id)
}

// List 列表查询映射
func (s *Service) List(filter mappingcontract.ListFilter) ([]mappingdomain.Mapping, int64, error) {
	return s.mappings.List(filter)
}

// SetActive 启用/禁用映射
func (s *Service) SetActive(id uint, active bool) error {
	mapping, err := s.mappings.GetByID(id)
	if err != nil {
		return err
	}
	if mapping == nil {
		return mappingcontract.ErrMappingNotFound
	}
	if mapping.IsActive == active {
		return nil
	}
	if !active && mapping.LocalProductID > 0 {
		product, productErr := s.products.GetByID(strconv.FormatUint(uint64(mapping.LocalProductID), 10))
		if productErr != nil {
			return productErr
		}
		if product != nil && product.FulfillmentType == constants.FulfillmentTypeUpstream {
			// Stop new sales before disabling the mapping; otherwise paid orders
			// could no longer be routed to their upstream SKU.
			product.FulfillmentType = constants.FulfillmentTypeManual
			product.IsMapped = false
			product.IsActive = false
			if err := s.products.Update(product); err != nil {
				return err
			}
		}
	}
	mapping.IsActive = active
	return s.mappings.Update(mapping)
}

// EnableUpstreamFulfillment repairs an existing local mapping before it can
// receive paid orders. It validates the connection and every active local SKU,
// refreshes only the upstream form schema, and updates the product's mapped and
// fulfillment flags together.
func (s *Service) EnableUpstreamFulfillment(id uint) error {
	mapping, err := s.mappings.GetByID(id)
	if err != nil {
		return err
	}
	if mapping == nil {
		return mappingcontract.ErrMappingNotFound
	}
	if !mapping.IsActive || mapping.UpstreamStatus != mappingdomain.UpstreamStatusActive {
		return mappingcontract.ErrUpstreamFulfillmentNotReady
	}
	conn, err := s.connections.GetByID(mapping.ConnectionID)
	if err != nil {
		return err
	}
	if conn == nil || conn.Status != "active" {
		return mappingcontract.ErrUpstreamFulfillmentNotReady
	}
	adapter, err := s.connections.GetAdapter(conn)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	upstreamProduct, err := adapter.GetProduct(ctx, mapping.UpstreamProductID)
	if err != nil {
		return fmt.Errorf("load upstream product: %w", err)
	}
	if upstreamProduct == nil || !upstreamProduct.IsActive || len(upstreamProduct.SKUs) == 0 {
		return mappingcontract.ErrUpstreamFulfillmentNotReady
	}

	product, err := s.products.GetByID(strconv.FormatUint(uint64(mapping.LocalProductID), 10))
	if err != nil {
		return err
	}
	if product == nil {
		return mappingcontract.ErrMappingNotFound
	}
	localSKUs, err := s.skus.ListByProduct(product.ID, true)
	if err != nil {
		return err
	}
	skuMappings, err := s.skuMappings.ListByProductMapping(mapping.ID)
	if err != nil {
		return err
	}
	activeMappings := make(map[uint]bool, len(skuMappings))
	for _, skuMapping := range skuMappings {
		activeMappings[skuMapping.LocalSKUID] = skuMapping.UpstreamIsActive
	}
	if len(localSKUs) == 0 {
		return mappingcontract.ErrUpstreamFulfillmentNotReady
	}
	for _, sku := range localSKUs {
		if !activeMappings[sku.ID] {
			return mappingcontract.ErrUpstreamFulfillmentNotReady
		}
	}

	product.IsMapped = true
	product.FulfillmentType = constants.FulfillmentTypeUpstream
	if upstreamProduct.ManualFormSchema != nil {
		product.ManualFormSchemaJSON = upstreamProduct.ManualFormSchema
	}
	return s.products.Update(product)
}

// Delete 删除映射（不删除本地商品）
func (s *Service) Delete(id uint) error {
	mapping, err := s.mappings.GetByID(id)
	if err != nil {
		return err
	}
	if mapping == nil {
		return mappingcontract.ErrMappingNotFound
	}

	// 删除 SKU 映射
	if err := s.skuMappings.DeleteByProductMapping(id); err != nil {
		return err
	}

	// 还原本地商品状态：取消映射标记、交付类型改回 manual、自动下架
	if mapping.LocalProductID > 0 {
		localProduct, err := s.products.GetByID(strconv.FormatUint(uint64(mapping.LocalProductID), 10))
		if err == nil && localProduct != nil {
			localProduct.IsMapped = false
			if localProduct.FulfillmentType == constants.FulfillmentTypeUpstream {
				localProduct.FulfillmentType = constants.FulfillmentTypeManual
				localProduct.IsActive = false // 下架，防止用户下单后无法交付
			}
			_ = s.products.Update(localProduct)
		}
	}

	return s.mappings.Delete(id)
}

// GetSKUMappings 获取映射的 SKU 映射列表
func (s *Service) GetSKUMappings(mappingID uint) ([]mappingdomain.SKUMapping, error) {
	return s.skuMappings.ListByProductMapping(mappingID)
}

// GetMappedUpstreamIDs 获取指定连接下所有已映射的上游商品 ID
func (s *Service) GetMappedUpstreamIDs(connectionID uint) ([]uint, error) {
	return s.mappings.ListUpstreamIDsByConnection(connectionID)
}
