package adminhttp

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"time"

	resellerdomain "github.com/dujiao-next/internal/modules/reseller/domain"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/dujiao-next/internal/platform/http/response"
	"github.com/gin-gonic/gin"
)

type InviteCodeStore interface {
	ListInviteCodes(page, pageSize int, status string) ([]resellerdomain.AgentInviteCode, int64, error)
	CreateInviteCode(item *resellerdomain.AgentInviteCode) error
	UpdateInviteCodeMemo(id uint, memo string) error
	DeleteInviteCode(id uint) error
	ToggleInviteCode(id uint) error
}

type updateInviteCodeMemoRequest struct {
	Memo string `json:"memo"`
}

// UpdateInviteCodeMemo 更新邀请码备注，不允许通过此接口修改邀请码本身。
func (h *AdminInviteCodeHandler) UpdateInviteCodeMemo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		ginutil.RespondError(c, response.CodeBadRequest, "error.invalid_id", nil)
		return
	}
	var req updateInviteCodeMemoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginutil.RespondBindError(c, err)
		return
	}
	if len([]rune(req.Memo)) > 255 {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", errors.New("memo must be at most 255 characters"))
		return
	}
	if err := h.store.UpdateInviteCodeMemo(uint(id), strings.TrimSpace(req.Memo)); err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.operation_failed", err)
		return
	}
	response.Success(c, gin.H{"id": id, "memo": strings.TrimSpace(req.Memo)})
}

type AdminInviteCodeHandler struct {
	store InviteCodeStore
}

func NewAdminInviteCodeHandler(store InviteCodeStore) *AdminInviteCodeHandler {
	if store == nil {
		panic("admin invite code handler: store is nil")
	}
	return &AdminInviteCodeHandler{store: store}
}

// ListInviteCodes 查询代理邀请码列表。
func (h *AdminInviteCodeHandler) ListInviteCodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := strings.TrimSpace(c.Query("status"))

	items, total, err := h.store.ListInviteCodes(page, pageSize, status)
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.query_failed", err)
		return
	}

	totalPage := int(total) / pageSize
	if int(total)%pageSize != 0 || totalPage == 0 {
		totalPage++
	}

	response.Success(c, gin.H{
		"items": items,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": totalPage,
		},
	})
}

type generateInviteCodesRequest struct {
	Code      string     `json:"code"`
	Prefix    string     `json:"prefix"`
	Count     int        `json:"count"`
	Memo      string     `json:"memo"`
	MaxUses   *int       `json:"max_uses"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// GenerateInviteCodes 生成代理邀请码。
func (h *AdminInviteCodeHandler) GenerateInviteCodes(c *gin.Context) {
	var req generateInviteCodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginutil.RespondBindError(c, err)
		return
	}

	if req.Count <= 0 {
		req.Count = 1
	}
	if req.Count > 100 {
		req.Count = 100
	}
	if req.ExpiresAt != nil && !req.ExpiresAt.After(time.Now()) {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", errors.New("expires_at must be in the future"))
		return
	}
	maxUses := 1
	if req.MaxUses != nil {
		maxUses = *req.MaxUses
		if maxUses < 0 {
			maxUses = 1
		}
	}

	prefix := strings.ToUpper(strings.TrimSpace(req.Prefix))
	if prefix == "" && req.Code == "" {
		prefix = "AGENT-"
	}

	var createdCodes []resellerdomain.AgentInviteCode

	if req.Code != "" {
		// 单个自定义邀请码
		code := strings.ToUpper(strings.TrimSpace(req.Code))
		item := resellerdomain.AgentInviteCode{
			Code:      code,
			Memo:      strings.TrimSpace(req.Memo),
			MaxUses:   maxUses,
			UsedCount: 0,
			Status:    "active",
			ExpiresAt: req.ExpiresAt,
		}
		if err := h.store.CreateInviteCode(&item); err != nil {
			ginutil.RespondError(c, response.CodeBadRequest, "error.code_exists_or_invalid", err)
			return
		}
		createdCodes = append(createdCodes, item)
	} else {
		// 批量随机生成：格式为 XXXX-XXXX-XXXX-XXXX (大写字母与数字组成)
		failed := 0
		for i := 0; i < req.Count; i++ {
			created := false
			for attempt := 0; attempt < 5 && !created; attempt++ {
				code, err := generateFormattedInviteCode()
				if err != nil {
					break
				}
				item := resellerdomain.AgentInviteCode{
					Code:      code,
					Memo:      strings.TrimSpace(req.Memo),
					MaxUses:   maxUses,
					UsedCount: 0,
					Status:    "active",
					ExpiresAt: req.ExpiresAt,
				}
				if err := h.store.CreateInviteCode(&item); err == nil {
					createdCodes = append(createdCodes, item)
					created = true
				}
			}
			if !created {
				failed++
			}
		}
		response.Success(c, gin.H{"items": createdCodes, "count": len(createdCodes), "requested_count": req.Count, "failed_count": failed})
		return
	}

	response.Success(c, gin.H{
		"items": createdCodes,
		"count": len(createdCodes),
	})
}

// DeleteInviteCode 删除邀请码。
func (h *AdminInviteCodeHandler) DeleteInviteCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		ginutil.RespondError(c, response.CodeBadRequest, "error.invalid_id", nil)
		return
	}

	if err := h.store.DeleteInviteCode(uint(id)); err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.delete_failed", err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

// ToggleInviteCode 切换邀请码状态。
func (h *AdminInviteCodeHandler) ToggleInviteCode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		ginutil.RespondError(c, response.CodeBadRequest, "error.invalid_id", nil)
		return
	}

	if err := h.store.ToggleInviteCode(uint(id)); err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.operation_failed", err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

// generateFormattedInviteCode 生成格式为 XXXX-XXXX-XXXX-XXXX 的随机邀请码（大写字母与数字）
func generateFormattedInviteCode() (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	max := big.NewInt(int64(len(charset)))
	parts := make([]string, 4)
	for i := 0; i < 4; i++ {
		var segment [4]byte
		for j := 0; j < 4; j++ {
			n, err := rand.Int(rand.Reader, max)
			if err != nil {
				return "", err
			}
			segment[j] = charset[n.Int64()]
		}
		parts[i] = string(segment[:])
	}
	return strings.Join(parts, "-"), nil
}
