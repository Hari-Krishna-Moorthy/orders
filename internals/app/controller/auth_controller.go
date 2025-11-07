package controller

import (
	"net/http"

	"github.com/Hari-Krishna-Moorthy/orders/internals/app/service"
	"github.com/Hari-Krishna-Moorthy/orders/internals/app/types"
	"github.com/gofiber/fiber/v2"
)

type AuthController struct{ svc *service.AuthService }

var authCtrl *AuthController

func NewAuthController(svc *service.AuthService) *AuthController {
	if authCtrl == nil {
		authCtrl = &AuthController{svc: svc}
	}
	return authCtrl
}
func GetAuthController() *AuthController { return authCtrl }

func (h *AuthController) SignUp(c *fiber.Ctx) error {
	var req types.SignUpRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	resp, err := h.svc.SignUp(types.SignUpInput{Email: req.Email, Password: req.Password, Name: req.Name})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(resp)
}

func (h *AuthController) SignIn(c *fiber.Ctx) error {
	var req types.SignInRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	resp, err := h.svc.SignIn(types.SignInInput{Email: req.Email, Password: req.Password})
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(resp)
}

func (h *AuthController) SignOut(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(types.GenericOK{Status: "signed_out"})
}
