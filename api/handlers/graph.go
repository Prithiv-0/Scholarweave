package handlers

import (
	"strconv"
	"strings"

	"Scholarweave/internal/services"

	"github.com/gofiber/fiber/v3"
)

type GraphHandler struct {
	graphService *services.GraphService
}

type SearchGraphRequest struct {
	PaperIDs []string `json:"paper_ids"`
}

func NewGraphHandler(graphService *services.GraphService) *GraphHandler {
	return &GraphHandler{graphService: graphService}
}

func (h *GraphHandler) GetPaperGraph(c fiber.Ctx) error {
	id := strings.TrimSpace(c.Params("id"))
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Paper ID is required")
	}

	depth := 1
	rawDepth := strings.TrimSpace(c.Query("depth"))
	if rawDepth != "" {
		parsedDepth, err := strconv.Atoi(rawDepth)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "depth must be a number")
		}
		if parsedDepth < 1 || parsedDepth > 2 {
			return fiber.NewError(fiber.StatusBadRequest, "depth must be between 1 and 2")
		}
		depth = parsedDepth
	}

	graph, err := h.graphService.BuildPaperGraph(c, id, depth)
	if err != nil {
		return mapGraphServiceError(err, "paper")
	}

	return c.JSON(graph)
}

func (h *GraphHandler) GetAuthorGraph(c fiber.Ctx) error {
	id := strings.TrimSpace(c.Params("id"))
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Author ID is required")
	}

	graph, err := h.graphService.BuildAuthorGraph(c, id)
	if err != nil {
		return mapGraphServiceError(err, "author")
	}

	return c.JSON(graph)
}

func (h *GraphHandler) GetConceptGraph(c fiber.Ctx) error {
	id := strings.TrimSpace(c.Params("id"))
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Concept ID is required")
	}

	graph, err := h.graphService.BuildConceptGraph(c, id)
	if err != nil {
		return mapGraphServiceError(err, "concept")
	}

	return c.JSON(graph)
}

func (h *GraphHandler) GetSearchGraph(c fiber.Ctx) error {
	var request SearchGraphRequest
	if err := c.Bind().Body(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if len(request.PaperIDs) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "paper_ids is required")
	}

	graph, err := h.graphService.BuildSearchGraph(c, request.PaperIDs)
	if err != nil {
		return mapGraphServiceError(err, "search")
	}

	return c.JSON(graph)
}

func mapGraphServiceError(err error, resource string) error {
	message := strings.ToLower(strings.TrimSpace(err.Error()))

	if strings.Contains(message, "is required") {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if strings.Contains(message, "returned 404") || strings.Contains(message, "no works found") {
		resourceName := "Resource"
		switch resource {
		case "paper":
			resourceName = "Paper"
		case "author":
			resourceName = "Author"
		case "concept":
			resourceName = "Concept"
		case "search":
			resourceName = "Search"
		}
		return fiber.NewError(fiber.StatusNotFound, resourceName+" graph source not found")
	}

	if resource == "search" && strings.Contains(message, "failed to build graph") {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return fiber.NewError(fiber.StatusInternalServerError, "Failed to build "+resource+" graph")
}
