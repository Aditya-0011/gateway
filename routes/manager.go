package routes

import (
	calls "gateway/grpc/calls/manager"
	"gateway/middlewares"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/gofiber/fiber/v3"
)

func managerRouter(router fiber.Router, authMiddleware fiber.Handler, userClient manager.UserServiceClient, portfolioClient manager.PortfolioServiceClient, validator protovalidate.Validator) {

	portfolioUserCall := calls.PortfolioUserCalls(userClient, validator)
	portfolioCall := calls.PortfolioCalls(portfolioClient, validator)

	userApi := router.Group("/user", authMiddleware)
	{
		userApi.Get("/details", middlewares.Validate[manager.SimpleRequest](validator), portfolioUserCall.GetUserDetails)
		userApi.Post("/edit", middlewares.RequireSession(), middlewares.Validate[manager.EditUserDetailsRequest](validator), portfolioUserCall.EditUserDetails)
	}

	messageApi := router.Group("/message", authMiddleware)
	{
		messageApi.Get("/list", middlewares.RequireSession(), middlewares.Validate[manager.SimpleRequest](validator), portfolioCall.GetMessages)
		messageApi.Post("/add", middlewares.Validate[manager.AddMessageRequest](validator), portfolioCall.AddMessage)
		messageApi.Post("/delete", middlewares.RequireSession(), middlewares.Validate[manager.DeleteMessageRequest](validator), portfolioCall.DeleteMessage)
	}

	technologyApi := router.Group("/technology", authMiddleware)
	{
		technologyApi.Get("/list", middlewares.Validate[manager.SimpleRequest](validator), portfolioCall.GetTechnologies)
		technologyApi.Post("/add", middlewares.RequireSession(), middlewares.Validate[manager.TechnologyCreateRequest](validator), portfolioCall.CreateTechnology)
		technologyApi.Post("/edit", middlewares.RequireSession(), middlewares.Validate[manager.TechnologyUpdateRequest](validator), portfolioCall.UpdateTechnology)
		technologyApi.Post("/delete", middlewares.RequireSession(), middlewares.Validate[manager.DeleteRequest](validator), portfolioCall.DeleteTechnology)
	}

	projectApi := router.Group("/project", authMiddleware)
	{
		projectApi.Get("/list", middlewares.Validate[manager.SimpleRequest](validator), portfolioCall.GetProjects)
		projectApi.Post("/add", middlewares.RequireSession(), middlewares.Validate[manager.ProjectCreateRequest](validator), portfolioCall.CreateProject)
		projectApi.Post("/edit", middlewares.RequireSession(), middlewares.Validate[manager.ProjectUpdateRequest](validator), portfolioCall.UpdateProject)
		projectApi.Post("/delete", middlewares.RequireSession(), middlewares.Validate[manager.DeleteRequest](validator), portfolioCall.DeleteProject)
	}

	experienceApi := router.Group("/experience", authMiddleware)
	{
		experienceApi.Get("/list", middlewares.Validate[manager.SimpleRequest](validator), portfolioCall.GetExperiences)
		experienceApi.Post("/add", middlewares.RequireSession(), middlewares.Validate[manager.ExperienceCreateRequest](validator), portfolioCall.CreateExperience)
		experienceApi.Post("/edit", middlewares.RequireSession(), middlewares.Validate[manager.ExperienceUpdateRequest](validator), portfolioCall.UpdateExperience)
		experienceApi.Post("/delete", middlewares.RequireSession(), middlewares.Validate[manager.DeleteRequest](validator), portfolioCall.DeleteExperience)
	}
}
