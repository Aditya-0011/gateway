package routes

import (
	calls "gateway/grpc/calls/manager"
	"gateway/middlewares"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/gofiber/fiber/v3"
)

func managerRouter(router fiber.Router, requireHost fiber.Handler, authMiddleware fiber.Handler, userClient manager.UserServiceClient, portfolioClient manager.PortfolioServiceClient, validator protovalidate.Validator) {

	portfolioUserCall := calls.PortfolioUserCalls(userClient, validator)
	portfolioCall := calls.PortfolioCalls(portfolioClient, validator)

	// User
	router.Get("/user/details", requireHost, authMiddleware, middlewares.Validate[manager.SimpleRequest](validator), portfolioUserCall.GetUserDetails)
	router.Post("/user/edit", requireHost, authMiddleware, middlewares.RequireSession(), middlewares.Validate[manager.EditUserDetailsRequest](validator), portfolioUserCall.EditUserDetails)

	// Message
	router.Get("/message/list", requireHost, authMiddleware, middlewares.RequireSession(), middlewares.Validate[manager.SimpleRequest](validator), portfolioCall.GetMessages)
	router.Post("/message/add", requireHost, authMiddleware, middlewares.Validate[manager.AddMessageRequest](validator), portfolioCall.AddMessage)
	router.Post("/message/delete", requireHost, authMiddleware, middlewares.RequireSession(), middlewares.Validate[manager.DeleteMessageRequest](validator), portfolioCall.DeleteMessage)

	// Technology
	router.Get("/technology/list", requireHost, authMiddleware, middlewares.Validate[manager.SimpleRequest](validator), portfolioCall.GetTechnologies)
	router.Post("/technology/add", requireHost, authMiddleware, middlewares.RequireSession(), middlewares.Validate[manager.TechnologyCreateRequest](validator), portfolioCall.CreateTechnology)
	router.Post("/technology/edit", requireHost, authMiddleware, middlewares.RequireSession(), middlewares.Validate[manager.TechnologyUpdateRequest](validator), portfolioCall.UpdateTechnology)
	router.Post("/technology/delete", requireHost, authMiddleware, middlewares.RequireSession(), middlewares.Validate[manager.DeleteRequest](validator), portfolioCall.DeleteTechnology)

	// Project
	router.Get("/project/list", requireHost, authMiddleware, middlewares.Validate[manager.SimpleRequest](validator), portfolioCall.GetProjects)
	router.Post("/project/add", requireHost, authMiddleware, middlewares.RequireSession(), middlewares.Validate[manager.ProjectCreateRequest](validator), portfolioCall.CreateProject)
	router.Post("/project/edit", requireHost, authMiddleware, middlewares.RequireSession(), middlewares.Validate[manager.ProjectUpdateRequest](validator), portfolioCall.UpdateProject)
	router.Post("/project/delete", requireHost, authMiddleware, middlewares.RequireSession(), middlewares.Validate[manager.DeleteRequest](validator), portfolioCall.DeleteProject)

	// Experience
	router.Get("/experience/list", requireHost, authMiddleware, middlewares.Validate[manager.SimpleRequest](validator), portfolioCall.GetExperiences)
	router.Post("/experience/add", requireHost, authMiddleware, middlewares.RequireSession(), middlewares.Validate[manager.ExperienceCreateRequest](validator), portfolioCall.CreateExperience)
	router.Post("/experience/edit", requireHost, authMiddleware, middlewares.RequireSession(), middlewares.Validate[manager.ExperienceUpdateRequest](validator), portfolioCall.UpdateExperience)
	router.Post("/experience/delete", requireHost, authMiddleware, middlewares.RequireSession(), middlewares.Validate[manager.DeleteRequest](validator), portfolioCall.DeleteExperience)
}

