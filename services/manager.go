package services

import (
	"fmt"
	calls "gateway/grpc/calls/manager"
	"gateway/middlewares"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/gofiber/fiber/v3"
)

func managerService(router fiber.Router, authMiddleware fiber.Handler, userClient manager.UserServiceClient, portfolioClient manager.PortfolioServiceClient) {
	validator, err := protovalidate.New()

	if err != nil {
		panic(fmt.Errorf("failed to initialize validator: %w", err))
	}

	userCalls := calls.PortfolioUserCalls(userClient, validator)
	portfolioCalls := calls.PortfolioCalls(portfolioClient, validator)

	userApi := router.Group("/user", authMiddleware)
	{
		userApi.Get("/details", middlewares.Validate[manager.SimpleRequest](validator), userCalls.GetUserDetails)
		userApi.Post("/edit", middlewares.RequireSession(), middlewares.Validate[manager.EditUserDetailsRequest](validator), userCalls.EditUserDetails)
	}

	messageApi := router.Group("/message", authMiddleware)
	{
		messageApi.Get("/list", middlewares.Validate[manager.SimpleRequest](validator), portfolioCalls.GetMessages)
		messageApi.Post("/add", middlewares.RequireSession(), middlewares.Validate[manager.AddMessageRequest](validator), portfolioCalls.AddMessage)
		messageApi.Post("/delete", middlewares.RequireSession(), middlewares.Validate[manager.DeleteMessageRequest](validator), portfolioCalls.DeleteMessage)
	}

	technologyApi := router.Group("/technology", authMiddleware)
	{
		technologyApi.Get("/list", middlewares.Validate[manager.SimpleRequest](validator), portfolioCalls.GetTechnologies)
		technologyApi.Post("/add", middlewares.RequireSession(), middlewares.Validate[manager.TechnologyCreateRequest](validator), portfolioCalls.CreateTechnology)
		technologyApi.Post("/edit", middlewares.RequireSession(), middlewares.Validate[manager.TechnologyUpdateRequest](validator), portfolioCalls.UpdateTechnology)
		technologyApi.Post("/delete", middlewares.RequireSession(), middlewares.Validate[manager.DeleteRequest](validator), portfolioCalls.DeleteTechnology)
	}

	projectApi := router.Group("/project", authMiddleware)
	{
		projectApi.Get("/list", middlewares.Validate[manager.SimpleRequest](validator), portfolioCalls.GetProjects)
		projectApi.Post("/add", middlewares.RequireSession(), middlewares.Validate[manager.ProjectCreateRequest](validator), portfolioCalls.CreateProject)
		projectApi.Post("/edit", middlewares.RequireSession(), middlewares.Validate[manager.ProjectUpdateRequest](validator), portfolioCalls.UpdateProject)
		projectApi.Post("/delete", middlewares.RequireSession(), middlewares.Validate[manager.DeleteRequest](validator), portfolioCalls.DeleteProject)
	}

	experienceApi := router.Group("/experience", authMiddleware)
	{
		experienceApi.Get("/list", middlewares.Validate[manager.SimpleRequest](validator), portfolioCalls.GetExperiences)
		experienceApi.Post("/add", middlewares.RequireSession(), middlewares.Validate[manager.ExperienceCreateRequest](validator), portfolioCalls.CreateExperience)
		experienceApi.Post("/edit", middlewares.RequireSession(), middlewares.Validate[manager.ExperienceUpdateRequest](validator), portfolioCalls.UpdateExperience)
		experienceApi.Post("/delete", middlewares.RequireSession(), middlewares.Validate[manager.DeleteRequest](validator), portfolioCalls.DeleteExperience)
	}
}
