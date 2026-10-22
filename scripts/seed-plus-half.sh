#!/usr/bin/env bash
set -euo pipefail

DB_PATH="${1:-/opt/dujiao-next/db/dujiao.db}"

if [[ ! -f "$DB_PATH" ]]; then
  echo "Database file not found: $DB_PATH" >&2
  exit 1
fi

sqlite3 "$DB_PATH" <<'SQL'
BEGIN TRANSACTION;

-- 1. 插入 chatgpt-plus-half 商品（如果不存在）
INSERT OR IGNORE INTO `products` (
  `category_id`, `slug`, `seo_meta_json`, `title_json`, `description_json`, `content_json`,
  `instructions_json`, `price_amount`, `cost_price_amount`, `wholesale_prices`, `images`,
  `tags`, `purchase_type`, `min_purchase_quantity`, `max_purchase_quantity`,
  `stock_display_mode`, `fulfillment_type`, `manual_form_schema_json`, `manual_stock_total`,
  `manual_stock_locked`, `manual_stock_sold`, `payment_channel_ids`, `is_affiliate_enabled`,
  `is_mapped`, `is_active`, `sort_order`, `created_at`, `updated_at`, `reseller_purchase_price_amount`
) VALUES (
  1,
  'chatgpt-plus-half',
  '{"description":{"en-US":"","zh-CN":"","zh-TW":""},"keywords":{"en-US":"","zh-CN":"","zh-TW":""}}',
  '{"en-US":"ChatGPT Plus (Promotion Eligibility)","zh-CN":"ChatGPT Plus 半价资格套餐","zh-TW":"ChatGPT Plus 半價資格套餐"}',
  '{"en-US":"Special promotion price for eligible ChatGPT accounts","zh-CN":"通过半价资格预检后下单，付款后自动提交充值","zh-TW":"通過半價資格預檢後下單，付款後自動提交儲值"}',
  '{"en-US":"","zh-CN":"","zh-TW":""}',
  '{"en-US":"","zh-CN":"","zh-TW":""}',
  66.00,
  1.01,
  '[]',
  '["/uploads/common/2026/10/7dcccc7f-7680-4d02-8f24-18c2a5d47292.png","/uploads/product/2026/10/d6b29846-e579-4e04-9dc4-e14fae3f29b4.jpeg"]',
  '[]',
  'guest',
  0,
  0,
  'status',
  'upstream',
  '{"fields":[{"key":"session","label":{"en-US":"ChatGPT Session","zh-CN":"ChatGPT Session","zh-TW":"ChatGPT Session"},"placeholder":{"en-US":"Submit the complete ChatGPT Session JSON before payment","zh-CN":"请在付款前提交完整的 ChatGPT Session JSON","zh-TW":"請在付款前提交完整的 ChatGPT Session JSON"},"required":true,"type":"textarea"}]}',
  1000,
  0,
  0,
  '',
  0,
  1,
  1,
  99,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP,
  62.00
);

-- 2. 插入 plus_half SKU
INSERT OR IGNORE INTO `product_skus` (
  `product_id`, `sku_code`, `spec_values_json`, `price_amount`, `cost_price_amount`,
  `manual_stock_total`, `manual_stock_locked`, `manual_stock_sold`, `is_active`, `sort_order`,
  `created_at`, `updated_at`, `reseller_purchase_price_amount`
)
SELECT 
  id,
  'plus_half',
  '{"name":"ChatGPT Plus 半价","payment_regions":[{"country":"PH","currency":"PHP"}],"plan":"plus_half","service_fee_usd_minor":15}',
  66.00,
  1.01,
  1000,
  0,
  0,
  1,
  0,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP,
  62.00
FROM `products` WHERE `slug` = 'chatgpt-plus-half';

-- 3. 插入 product_mappings
INSERT OR IGNORE INTO `product_mappings` (
  `connection_id`, `local_product_id`, `upstream_product_id`, `upstream_fulfillment_type`,
  `upstream_status`, `is_active`, `created_at`, `updated_at`
)
SELECT
  1,
  id,
  2011412075,
  'auto',
  'active',
  1,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
FROM `products` WHERE `slug` = 'chatgpt-plus-half';

-- 4. 插入 sku_mappings
INSERT OR IGNORE INTO `sku_mappings` (
  `product_mapping_id`, `local_sku_id`, `upstream_sku_id`, `upstream_price`,
  `upstream_stock`, `upstream_is_active`, `created_at`, `updated_at`
)
SELECT
  pm.id,
  ps.id,
  2011412075,
  0.15,
  -1,
  1,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
FROM `products` p
JOIN `product_skus` ps ON ps.product_id = p.id AND ps.sku_code = 'plus_half'
JOIN `product_mappings` pm ON pm.local_product_id = p.id
WHERE p.slug = 'chatgpt-plus-half'
AND NOT EXISTS (
  SELECT 1 FROM `sku_mappings` WHERE `local_sku_id` = ps.id
);

COMMIT;
SQL

echo "ChatGPT Plus Half-Price product configuration seeded successfully."
