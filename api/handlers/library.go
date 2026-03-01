package handlers

import (
	"strings"

	"Scholarweave/internal/models"
	"Scholarweave/internal/services"

	"github.com/gofiber/fiber/v3"
)

type LibraryHandler struct {
	libraryService *services.LibraryService
}

func NewLibraryHandler(libraryService *services.LibraryService) *LibraryHandler {
	return &LibraryHandler{libraryService: libraryService}
}

func getUserIDFromContext(c fiber.Ctx) (string, error) {
	userID, _ := c.Locals("user_id").(string)
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	return userID, nil
}

func (h *LibraryHandler) ListFavorites(c fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	favorites := h.libraryService.ListFavorites(userID)
	return c.JSON(fiber.Map{"favorites": favorites})
}

func (h *LibraryHandler) AddFavorite(c fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	var request models.AddFavoriteRequest
	if bindErr := c.Bind().Body(&request); bindErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	favorite, serviceErr := h.libraryService.AddFavorite(userID, request)
	if serviceErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, serviceErr.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(favorite)
}

func (h *LibraryHandler) RemoveFavorite(c fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	paperID := c.Params("paperId")
	if strings.TrimSpace(paperID) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "paperId is required")
	}

	if removeErr := h.libraryService.RemoveFavorite(userID, paperID); removeErr != nil {
		return fiber.NewError(fiber.StatusNotFound, removeErr.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *LibraryHandler) ListReadingLists(c fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	lists := h.libraryService.ListReadingLists(userID)
	return c.JSON(fiber.Map{"lists": lists})
}

func (h *LibraryHandler) CreateReadingList(c fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	var request models.CreateReadingListRequest
	if bindErr := c.Bind().Body(&request); bindErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	list, serviceErr := h.libraryService.CreateReadingList(userID, request.Name, request.Description)
	if serviceErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, serviceErr.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(list)
}

func (h *LibraryHandler) AddReadingListItem(c fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	listID := c.Params("listId")
	if strings.TrimSpace(listID) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "listId is required")
	}

	var request models.AddReadingListItemRequest
	if bindErr := c.Bind().Body(&request); bindErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	list, serviceErr := h.libraryService.AddItemToReadingList(userID, listID, request)
	if serviceErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, serviceErr.Error())
	}

	return c.JSON(list)
}

func (h *LibraryHandler) RemoveReadingListItem(c fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	listID := c.Params("listId")
	paperID := c.Params("paperId")
	if strings.TrimSpace(listID) == "" || strings.TrimSpace(paperID) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "listId and paperId are required")
	}

	list, serviceErr := h.libraryService.RemoveItemFromReadingList(userID, listID, paperID)
	if serviceErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, serviceErr.Error())
	}

	return c.JSON(list)
}

func (h *LibraryHandler) DeleteReadingList(c fiber.Ctx) error {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return err
	}

	listID := c.Params("listId")
	if strings.TrimSpace(listID) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "listId is required")
	}

	if serviceErr := h.libraryService.DeleteReadingList(userID, listID); serviceErr != nil {
		return fiber.NewError(fiber.StatusNotFound, serviceErr.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}
