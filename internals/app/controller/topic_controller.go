package controller

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/Hari-Krishna-Moorthy/orders/internals/app/pubsub"
	"github.com/Hari-Krishna-Moorthy/orders/internals/app/types"
	"github.com/Hari-Krishna-Moorthy/orders/internals/platform/config"
)

type TopicController struct{ mgr *pubsub.Manager }

var topicCtrl *TopicController

func NewTopicController(mgr *pubsub.Manager) *TopicController {
	if topicCtrl == nil {
		topicCtrl = &TopicController{mgr: mgr}
	}
	return topicCtrl
}
func GetTopicController() *TopicController { return topicCtrl }

func (h *TopicController) Create(c *fiber.Ctx) error {
	var req types.CreateTopicReq
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	cfg := config.Get()
	if err := h.mgr.CreateTopic(req.Name, cfg.App.ReplayLastN); err != nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "exists"})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"status": "created", "topic": req.Name})
}

func (h *TopicController) Delete(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "name required"})
	}
	if err := h.mgr.DeleteTopic(name); err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "not_found"})
	}
	return c.JSON(fiber.Map{"status": "deleted", "topic": name})
}

func (h *TopicController) List(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"topics": h.mgr.List()})
}

func (h *TopicController) Stats(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"topics": h.mgr.Stats()})
}
