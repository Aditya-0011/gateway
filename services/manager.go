package services

import (
	"fmt"
	calls "gateway/grpc/calls/manager"
	"gateway/middlewares"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/gofiber/fiber/v3"
)

func managerService(router fiber.Router, userClient manager.UserServiceClient, portfolioClient manager.PortfolioServiceClient) {
	validator, err := protovalidate.New()

	if err != nil {
		panic(fmt.Errorf("failed to initialize validator: %w", err))
	}

	userCalls := calls.PortfolioUserCalls(userClient, validator)
	portfolioCalls := calls.PortfolioCalls(portfolioClient, validator)

	api := router.Group("/portfolio")
	{
		api.Group("/user")
		{
			api.Get("/details", middlewares.Validate[manager.SimpleRequest](validator), userCalls.GetUserDetails)
			api.Post("/edit", middlewares.RequireSession(), middlewares.Validate[manager.EditUserDetailsRequest](validator), userCalls.EditUserDetails)
		}

		api.Group("/message")
		{
			api.Get("/list", portfolioCalls.GetMessages)
			api.Post("/add", middlewares.RequireSession(), middlewares.Validate[manager.AddMessageRequest](validator), portfolioCalls.AddMessage)
			api.Post("/delete", middlewares.RequireSession(), middlewares.Validate[manager.DeleteMessageRequest](validator), portfolioCalls.DeleteMessage)
		}

		api.Group("/technology")
		{
			api.Get("/list", portfolioCalls.GetTechnologies)
			api.Post("/add", middlewares.RequireSession(), middlewares.Validate[manager.TechnologyCreateRequest](validator), portfolioCalls.CreateTechnology)
			api.Post("/edit", middlewares.RequireSession(), middlewares.Validate[manager.TechnologyUpdateRequest](validator), portfolioCalls.UpdateTechnology)
			api.Post("/delete", middlewares.RequireSession(), middlewares.Validate[manager.DeleteRequest](validator), portfolioCalls.DeleteTechnology)
		}

		api.Group("/project")
		{
			api.Get("/list", portfolioCalls.GetProjects)
			api.Post("/add", middlewares.RequireSession(), middlewares.Validate[manager.ProjectCreateRequest](validator), portfolioCalls.CreateProject)
			api.Post("/edit", middlewares.RequireSession(), middlewares.Validate[manager.ProjectUpdateRequest](validator), portfolioCalls.UpdateProject)
			api.Post("/delete", middlewares.RequireSession(), middlewares.Validate[manager.DeleteRequest](validator), portfolioCalls.DeleteProject)
		}

		api.Group("/experience")
		{
			api.Get("/list", portfolioCalls.GetExperiences)
			api.Post("/add", middlewares.RequireSession(), middlewares.Validate[manager.ExperienceCreateRequest](validator), portfolioCalls.CreateExperience)
			api.Post("/edit", middlewares.RequireSession(), middlewares.Validate[manager.ExperienceUpdateRequest](validator), portfolioCalls.UpdateExperience)
			api.Post("/delete", middlewares.RequireSession(), middlewares.Validate[manager.DeleteRequest](validator), portfolioCalls.DeleteExperience)
		}
	}
}
