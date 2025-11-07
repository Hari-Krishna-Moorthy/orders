package controller

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/Hari-Krishna-Moorthy/orders/internals/app/pubsub"
	"github.com/Hari-Krishna-Moorthy/orders/internals/app/types"
	"github.com/Hari-Krishna-Moorthy/orders/internals/platform/config"
)

type WSController struct{ mgr *pubsub.Manager }

var wsCtrl *WSController

func NewWSController(mgr *pubsub.Manager) *WSController {
	if wsCtrl == nil {
		wsCtrl = &WSController{mgr: mgr}
	}
	return wsCtrl
}
func GetWSController() *WSController { return wsCtrl }

type clientReq struct {
	Type      string         `json:"type"`
	Topic     string         `json:"topic"`
	ClientID  string         `json:"client_id"`
	Message   pubsub.Message `json:"message"`
	LastN     int            `json:"last_n"`
	RequestID string         `json:"request_id"`
}

type serverResp struct {
	Type      string             `json:"type"`
	RequestID string             `json:"request_id,omitempty"`
	Topic     string             `json:"topic,omitempty"`
	Message   pubsub.Message     `json:"message,omitempty"`
	Error     *types.ServerError `json:"error,omitempty"`
	TS        time.Time          `json:"ts"`
	Status    string             `json:"status,omitempty"`
}

func (h *WSController) Register(app *fiber.App) {
	cfg := config.Get()
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		c.SetReadDeadline(time.Now().Add(time.Duration(cfg.App.WSReadTimeoutSec) * time.Second))
		defer c.Close()
		for {
			var req clientReq
			if err := c.ReadJSON(&req); err != nil {
				return
			}
			switch req.Type {
			case "subscribe":
				s, err := h.mgr.Subscribe(req.Topic, req.ClientID, cfg.App.SubscriberQueueSize)
				if err != nil {
					_ = c.WriteJSON(serverResp{Type: "error", RequestID: req.RequestID, TS: time.Now(), Error: &types.ServerError{Code: "TOPIC_NOT_FOUND", Message: err.Error()}})
					continue
				}
				_ = c.WriteJSON(serverResp{Type: "ack", Status: "ok", Topic: req.Topic, RequestID: req.RequestID, TS: time.Now()})
				// optional replay
				if req.LastN > 0 {
					if msgs, _ := h.mgr.Replay(req.Topic, req.LastN); len(msgs) > 0 {
						for _, m := range msgs {
							_ = c.WriteJSON(serverResp{Type: "event", Topic: req.Topic, Message: m, TS: time.Now()})
						}
					}
				}
				// forward incoming events
				go func() {
					for {
						select {
						case m := <-s.Ch:
							_ = c.WriteJSON(serverResp{Type: "event", Topic: req.Topic, Message: m, TS: time.Now()})
						case <-s.Quit:
							return
						}
					}
				}()
			case "unsubscribe":
				_ = h.mgr.Unsubscribe(req.Topic, req.ClientID)
				_ = c.WriteJSON(serverResp{Type: "ack", Status: "ok", Topic: req.Topic, RequestID: req.RequestID, TS: time.Now()})
			case "publish":
				msg := pubsub.Message{ID: req.Message.ID, Payload: req.Message.Payload, TS: time.Now()}
				if err := h.mgr.Publish(req.Topic, msg); err != nil {
					_ = c.WriteJSON(serverResp{Type: "error", RequestID: req.RequestID, TS: time.Now(), Error: &types.ServerError{Code: "TOPIC_NOT_FOUND", Message: err.Error()}})
					continue
				}
				_ = c.WriteJSON(serverResp{Type: "ack", Status: "ok", Topic: req.Topic, RequestID: req.RequestID, TS: time.Now()})
			case "ping":
				_ = c.WriteJSON(serverResp{Type: "pong", RequestID: req.RequestID, TS: time.Now()})
			default:
				_ = c.WriteJSON(serverResp{Type: "error", TS: time.Now(), Error: &types.ServerError{Code: "BAD_REQUEST", Message: "unknown type"}})
			}
		}
	}))
}
