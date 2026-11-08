package schema

import (
	"slices"

	"github.com/Aditya-0011/common/contracts/go/auth"
)

type Identity struct {
	UserID             int      `json:"user_id"`
	Email              string   `json:"email"`
	Role               int      `json:"role"`
	MustChangePassword bool     `json:"must_change_password"`
	Projects           []string `json:"projects"`
}

func IdentityFrom(id *auth.Identity) Identity {
	projects := id.GetProjects()
	if projects == nil {
		projects = []string{}
	}
	return Identity{
		UserID:             int(id.GetUserId()),
		Email:              id.GetEmail(),
		Role:               int(id.GetRole()),
		MustChangePassword: id.GetMustChangePassword(),
		Projects:           projects,
	}
}

func (i Identity) IsAdmin() bool { return i.Role == int(auth.Role_ROLE_ADMIN) }

func (i Identity) CanAccess(project string) bool {
	return i.IsAdmin() || slices.Contains(i.Projects, project)
}

type APIKeyIdentity struct {
	Identity
	Origins []string `json:"origins"`
}

type AuthInfo struct {
	Identity
	SessionID string
	APIKey    bool
}

func (a *AuthInfo) IsSession() bool { return a.SessionID != "" }

type SessionResponse struct {
	Identity
	SessionID string `json:"session_id"`
}

type SessionInfo struct {
	ID         string `json:"id"`
	Current    bool   `json:"current"`
	CreatedAt  string `json:"created_at"`
	LastSeenAt string `json:"last_seen_at"`
	IP         string `json:"ip"`
	UserAgent  string `json:"user_agent"`
}

type SessionsResponse struct {
	Sessions []SessionInfo `json:"sessions"`
}

type RevokeSessionRequest struct {
	SessionID string `json:"session_id"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
