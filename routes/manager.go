package routes

import (
	"strconv"
	"time"

	"gateway/db"
	calls "gateway/grpc/calls/manager"
	"gateway/grpc/clients"
	"gateway/middlewares"
	"gateway/schema"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/gofiber/fiber/v3"
)

// portfolioPaths are the Manager routes a portfolio may call with its API key,
// from its server or from a browser on one of the key's origins.
var portfolioPaths = []string{
	"/user/details",
	"/technology/list",
	"/project/list",
	"/experience/list",
	"/message/add",
}

func managerRoutes(r fiber.Router, store *db.Store, cl *clients.Clients, authenticate, access fiber.Handler, v protovalidate.Validator) {
	user := calls.PortfolioUserCalls(cl.ManagerUser)
	portfolio := calls.PortfolioCalls(cl.ManagerPortfolio)

	public := chain{authenticate, access}
	owner := public.then(middlewares.RequireSession())
	messageLimit := middlewares.RateLimit{
		Name:    "message",
		Limit:   20,
		Window:  time.Hour,
		Message: "Too many messages, try again later",
		By: func(c fiber.Ctx) string {
			return "user:" + strconv.Itoa(c.Locals("auth").(*schema.AuthInfo).UserID)
		},
	}.Handler(store)

	get(r, "/user/details", public.then(middlewares.Validate[*manager.SimpleRequest](v), user.GetUserDetails))
	post(r, "/user/edit", owner.then(middlewares.Validate[*manager.EditUserDetailsRequest](v), user.EditUserDetails))

	get(r, "/message/list", owner.then(middlewares.Validate[*manager.SimpleRequest](v), portfolio.GetMessages))
	post(r, "/message/add", public.then(messageLimit, middlewares.Validate[*manager.AddMessageRequest](v), portfolio.AddMessage))
	post(r, "/message/delete", owner.then(middlewares.Validate[*manager.DeleteMessageRequest](v), portfolio.DeleteMessage))

	get(r, "/technology/list", public.then(middlewares.Validate[*manager.SimpleRequest](v), portfolio.GetTechnologies))
	post(r, "/technology/add", owner.then(middlewares.Validate[*manager.TechnologyCreateRequest](v), portfolio.CreateTechnology))
	post(r, "/technology/edit", owner.then(middlewares.Validate[*manager.TechnologyUpdateRequest](v), portfolio.UpdateTechnology))
	post(r, "/technology/delete", owner.then(middlewares.Validate[*manager.DeleteRequest](v), portfolio.DeleteTechnology))

	get(r, "/project/list", public.then(middlewares.Validate[*manager.SimpleRequest](v), portfolio.GetProjects))
	post(r, "/project/add", owner.then(middlewares.Validate[*manager.ProjectCreateRequest](v), portfolio.CreateProject))
	post(r, "/project/edit", owner.then(middlewares.Validate[*manager.ProjectUpdateRequest](v), portfolio.UpdateProject))
	post(r, "/project/delete", owner.then(middlewares.Validate[*manager.DeleteRequest](v), portfolio.DeleteProject))

	get(r, "/experience/list", public.then(middlewares.Validate[*manager.SimpleRequest](v), portfolio.GetExperiences))
	post(r, "/experience/add", owner.then(middlewares.Validate[*manager.ExperienceCreateRequest](v), portfolio.CreateExperience))
	post(r, "/experience/edit", owner.then(middlewares.Validate[*manager.ExperienceUpdateRequest](v), portfolio.UpdateExperience))
	post(r, "/experience/delete", owner.then(middlewares.Validate[*manager.DeleteRequest](v), portfolio.DeleteExperience))
}
