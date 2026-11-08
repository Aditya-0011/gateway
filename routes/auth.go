package routes

import (
	"slices"

	"gateway/db"
	calls "gateway/grpc/calls/auth"
	"gateway/grpc/clients"
	"gateway/middlewares"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/gofiber/fiber/v3"
)

type chain []fiber.Handler

func (ch chain) then(handlers ...fiber.Handler) chain {
	return append(slices.Clone(ch), handlers...)
}

func get(r fiber.Router, path string, ch chain) { register(r, fiber.MethodGet, path, ch) }

func post(r fiber.Router, path string, ch chain) { register(r, fiber.MethodPost, path, ch) }

func register(r fiber.Router, method, path string, ch chain) {
	rest := make([]any, len(ch)-1)
	for i, h := range ch[1:] {
		rest[i] = h
	}
	r.Add([]string{method}, path, ch[0], rest...)
}

func authRoutes(r fiber.Router, store *db.Store, cl *clients.Clients, cookie middlewares.SessionCookie,
	authn *middlewares.Authenticator, credentialLimit fiber.Handler, v protovalidate.Validator) {
	h := calls.New(store, cl.Auth, cookie)

	limited := chain{credentialLimit}
	pending := chain{authn.AllowPendingPassword(), middlewares.RequireSession()}
	user := chain{authn.Required(), middlewares.RequireSession()}
	admin := user.then(middlewares.RequireAdmin())

	post(r, "/login", limited.then(middlewares.Validate[*auth.LoginRequest](v), h.Login))
	post(r, "/reset-password", limited.then(middlewares.Validate[*auth.ResetPasswordRequest](v), h.ResetPassword))

	get(r, "/session", pending.then(h.Session))
	post(r, "/logout", pending.then(h.Logout))
	post(r, "/first-login", limited.then(pending...).then(middlewares.Validate[*auth.CompleteFirstLoginRequest](v), h.FirstLogin))

	post(r, "/change-password", limited.then(user...).then(middlewares.Validate[*auth.ChangePasswordRequest](v), h.ChangePassword))
	get(r, "/recovery-codes", user.then(middlewares.Validate[*auth.UserRequest](v), h.RecoveryCodeStatus))
	post(r, "/recovery-codes", limited.then(user...).then(middlewares.Validate[*auth.PasswordConfirmRequest](v), h.GenerateRecoveryCodes))
	post(r, "/recovery-codes/delete", limited.then(user...).then(middlewares.Validate[*auth.PasswordConfirmRequest](v), h.DeleteRecoveryCodes))
	get(r, "/sessions", user.then(h.ListSessions))
	post(r, "/sessions/revoke", user.then(h.RevokeSession))
	post(r, "/sessions/revoke-others", user.then(h.RevokeOtherSessions))
	get(r, "/projects", user.then(middlewares.Validate[*auth.UserRequest](v), h.Projects))

	get(r, "/key", user.then(middlewares.Validate[*auth.UserRequest](v), h.GetKey))
	post(r, "/rotate-key", limited.then(user...).then(middlewares.Validate[*auth.UserRequest](v), h.RotateKey))
	post(r, "/key/revoke", user.then(middlewares.Validate[*auth.UserRequest](v), h.RevokeKey))
	post(r, "/key/origins", user.then(middlewares.Validate[*auth.SetKeyOriginsRequest](v), h.SetKeyOrigins))

	get(r, "/admin/users", admin.then(middlewares.Validate[*auth.UserRequest](v), h.ListUsers))
	post(r, "/admin/users/add", admin.then(middlewares.Validate[*auth.CreateUserRequest](v), h.CreateUser))
	post(r, "/admin/users/reset-password", admin.then(middlewares.Validate[*auth.TargetUserRequest](v), h.ResetUserPassword))
	post(r, "/admin/users/status", admin.then(middlewares.Validate[*auth.SetUserStatusRequest](v), h.SetUserStatus))
	post(r, "/admin/users/projects", admin.then(middlewares.Validate[*auth.SetUserProjectsRequest](v), h.SetUserProjects))
	post(r, "/admin/users/delete", admin.then(middlewares.Validate[*auth.TargetUserRequest](v), h.DeleteUser))

	get(r, "/admin/projects", admin.then(middlewares.Validate[*auth.UserRequest](v), h.ListProjects))
	post(r, "/admin/projects/add", admin.then(middlewares.Validate[*auth.SaveProjectRequest](v), h.CreateProject))
	post(r, "/admin/projects/edit", admin.then(middlewares.Validate[*auth.SaveProjectRequest](v), h.UpdateProject))
	post(r, "/admin/projects/delete", admin.then(middlewares.Validate[*auth.DeleteProjectRequest](v), h.DeleteProject))
}
