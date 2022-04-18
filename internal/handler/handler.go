package handler

import (
	"go.uber.org/zap"
)

type Handler struct {
	Logger *zap.Logger
}

func NewHandler(logger *zap.Logger) *Handler {
	return &Handler{
		Logger: logger,
	}
}
