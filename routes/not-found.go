package routes

import (
	"fmt"
	"math/rand/v2"

	"github.com/gofiber/fiber/v3"
)

var (
	unmatchedRoutesErrorPrefixes = []string{
		"Route not found.",
		"This path doesn't exist.",
		"404.",
		"Invalid request.",
		"Unmatched route.",
		"URL invalid.",
		"Route missing.",
		"Unknown path.",
		"Path not found.",
		"Resource missing.",
		"Destination unknown.",
	}

	unmatchedRoutesErrorMessages = []string{
		"You are going to die alone.",
		"Just like your imaginary partner.",
		"Keep making mistakes like this and everyone will leave you.",
		"This is why you spend Friday nights alone.",
		"You are wasting finite seconds of a shrinking lifespan on this URL.",
		"Your existence is an error handled by no one.",
		"Screaming into the abyss won't make this page appear.",
		"This why even non judgmental people judge you.",
		"This is why your friends have a group chat without you.",
		"People don't like you; they only tolerate you.",
		"You are the backup plan for everyone you love.",
		"Go outside. Nobody is waiting for you there either.",
		"This is the reason why Epstein rejected you.",
		"This why you could never participate in Diddy parties, either as a guest or an entertainment.",
		"This is why Stephen Hawkings chose some midgets over you.",
	}
)

func notFoundHandler(c fiber.Ctx) error {
	prefix := unmatchedRoutesErrorPrefixes[rand.IntN(len(unmatchedRoutesErrorPrefixes))]
	message := unmatchedRoutesErrorMessages[rand.IntN(len(unmatchedRoutesErrorMessages))]
	response := fmt.Sprintf("%s:%s", prefix, message)
	return c.Status(fiber.StatusNotFound).SendString(response)
}
