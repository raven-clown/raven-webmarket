package rbac

const (
	RoleAdmin    = "admin"
	RoleDevAdmin = "dev_admin"
)

var roleRank = map[string]int{
	RoleAdmin:    1,
	RoleDevAdmin: 2,
}

var AllPermissions = []string{
	"cms", "products", "packages", "promotions", "milestones", "redeem",
	"users", "kpi", "audit", "activity", "purchases",
	"monitoring_view", "monitoring_edit", "autoscale", "security",
	"reset", "reset_monthly", "cache", "admin_accounts",
}

var DefaultAdminPermissions = []string{
	"cms", "products", "packages", "promotions", "milestones", "redeem",
	"users", "kpi", "audit", "activity", "purchases", "monitoring_view", "reset_monthly",
}

var AdminPermissions = map[string]string{
	"cms":             RoleAdmin,
	"products":        RoleAdmin,
	"packages":        RoleAdmin,
	"promotions":      RoleAdmin,
	"milestones":      RoleAdmin,
	"redeem":          RoleAdmin,
	"users":           RoleAdmin,
	"kpi":             RoleAdmin,
	"audit":           RoleAdmin,
	"activity":        RoleAdmin,
	"purchases":       RoleAdmin,
	"monitoring_view": RoleAdmin,
	"monitoring_edit": RoleDevAdmin,
	"autoscale":       RoleDevAdmin,
	"security":        RoleDevAdmin,
	"reset":           RoleDevAdmin,
	"reset_monthly":   RoleAdmin,
	"cache":           RoleDevAdmin,
	"admin_accounts":  RoleDevAdmin,
}

func HasMinRole(role, minRole string) bool {
	return roleRank[role] >= roleRank[minRole]
}

func CanAccessDevOnly(role string) bool {
	return role == RoleDevAdmin
}

func Can(role, permission string) bool {
	min, ok := AdminPermissions[permission]
	if !ok {
		return CanAccessDevOnly(role)
	}
	return HasMinRole(role, min)
}

func CanWithPermissions(role string, permissions []string, permission string) bool {
	if role == RoleDevAdmin {
		return true
	}
	if len(permissions) > 0 {
		for _, p := range permissions {
			if p == permission || p == "*" {
				return true
			}
		}
		return false
	}
	return Can(role, permission)
}
