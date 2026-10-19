package infrastructure

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/hanjo-erp/hanjo/internal/modules/rbac/domain"
)

// RBACCache defines contracts for high-performance in-memory permission evaluations.
type RBACCache interface {
	HasPermission(roleID, permCode string) bool
	GetRoleCode(roleID string) string
	GetRolePermissions(roleID string) []string
	Reload(ctx context.Context) error
}

type rbacCache struct {
	mu        sync.RWMutex
	roleRepo  domain.RoleRepository
	rolePerms map[string]map[string]bool // role_id -> set of permission codes
	roleCodes map[string]string          // role_id -> role_code
	log       *zap.Logger
}

// NewRBACCache initializes a thread-safe in-memory RBAC Cache layer.
func NewRBACCache(roleRepo domain.RoleRepository, log *zap.Logger) RBACCache {
	return &rbacCache{
		roleRepo:  roleRepo,
		rolePerms: make(map[string]map[string]bool),
		roleCodes: make(map[string]string),
		log:       log,
	}
}

// HasPermission checks if a given roleID holds the required permission code in O(1) time.
// Crucial: Roles with code 'DEVELOPER' automatically bypass all permission checks!
func (c *rbacCache) HasPermission(roleID, permCode string) bool {
	if roleID == "" || permCode == "" {
		return false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// 1. DEVELOPER super-role full bypass
	if code, exists := c.roleCodes[roleID]; exists && code == "DEVELOPER" {
		return true
	}

	// 2. Standard role permission evaluation
	perms, exists := c.rolePerms[roleID]
	if !exists {
		return false
	}

	return perms[permCode]
}

// GetRoleCode retrieves the cached role code for a roleID.
func (c *rbacCache) GetRoleCode(roleID string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.roleCodes[roleID]
}

// GetRolePermissions returns all permission codes granted to a given roleID.
func (c *rbacCache) GetRolePermissions(roleID string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	perms, exists := c.rolePerms[roleID]
	if !exists {
		return []string{}
	}

	result := make([]string, 0, len(perms))
	for code := range perms {
		result = append(result, code)
	}
	return result
}

// Reload queries the database to refresh all role-permission mappings and role codes in memory.
func (c *rbacCache) Reload(ctx context.Context) error {
	mappings, err := c.roleRepo.GetAllRolePermissionsMapping(ctx)
	if err != nil {
		c.log.Error("failed to load role permission mappings for rbac cache", zap.Error(err))
		return err
	}

	codes, err := c.roleRepo.GetAllRoleCodes(ctx)
	if err != nil {
		c.log.Error("failed to load role codes for rbac cache", zap.Error(err))
		return err
	}

	newPerms := make(map[string]map[string]bool, len(mappings))
	for roleID, permList := range mappings {
		set := make(map[string]bool, len(permList))
		for _, permCode := range permList {
			set[permCode] = true
		}
		newPerms[roleID] = set
	}

	c.mu.Lock()
	c.rolePerms = newPerms
	c.roleCodes = codes
	c.mu.Unlock()

	c.log.Info("in-memory rbac cache refreshed successfully",
		zap.Int("roles_cached", len(codes)),
		zap.Int("mapped_roles", len(newPerms)),
	)
	return nil
}
