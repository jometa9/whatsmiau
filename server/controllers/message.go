package controllers

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/verbeux-ai/whatsmiau/interfaces"
	"github.com/verbeux-ai/whatsmiau/lib/whatsmiau"
	"github.com/verbeux-ai/whatsmiau/server/dto"
	"go.uber.org/zap"
)

type Message struct {
	repo      interfaces.InstanceRepository
	whatsmiau *whatsmiau.Whatsmiau
}

func NewMessages(repository interfaces.InstanceRepository, whatsmiau *whatsmiau.Whatsmiau) *Message {
	return &Message{
		repo:      repository,
		whatsmiau: whatsmiau,
	}
}

func (s *Message) SendText(ctx echo.Context) error {
	var request dto.SendTextRequest
	if err := ctx.Bind(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.SendTextResponse{
			Success: false,
			Error:   "failed to bind request body",
			Message: err.Error(),
		})
	}

	if err := validator.New().Struct(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.SendTextResponse{
			Success: false,
			Error:   "invalid request body",
			Message: err.Error(),
		})
	}

	jid, err := numberToJid(request.Number)
	if err != nil {
		zap.L().Error("error converting number to jid", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, dto.SendTextResponse{
			Success: false,
			Error:   "invalid number format",
			Message: err.Error(),
		})
	}

	sendText := &whatsmiau.SendText{
		Text:       request.Text,
		InstanceID: request.InstanceID,
		RemoteJID:  jid,
	}

	c := ctx.Request().Context()
	_, err = s.whatsmiau.SendText(c, sendText)
	if err != nil {
		zap.L().Error("Whatsmiau.SendText failed", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, dto.SendTextResponse{
			Success: false,
			Error:   "failed to send text",
			Message: err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, dto.SendTextResponse{
		Success: true,
	})
}

func (s *Message) SendAudio(ctx echo.Context) error {
	var request dto.SendAudioRequest
	if err := ctx.Bind(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.SendAudioResponse{
			Success: false,
			Error:   "failed to bind request body",
			Message: err.Error(),
		})
	}

	if err := validator.New().Struct(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.SendAudioResponse{
			Success: false,
			Error:   "invalid request body",
			Message: err.Error(),
		})
	}

	jid, err := numberToJid(request.Number)
	if err != nil {
		zap.L().Error("error converting number to jid", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, dto.SendAudioResponse{
			Success: false,
			Error:   "invalid number format",
			Message: err.Error(),
		})
	}

	sendAudio := &whatsmiau.SendAudioRequest{
		AudioURL:   request.Audio,
		InstanceID: request.InstanceID,
		RemoteJID:  jid,
	}

	c := ctx.Request().Context()
	_, err = s.whatsmiau.SendAudio(c, sendAudio)
	if err != nil {
		zap.L().Error("Whatsmiau.SendAudioRequest failed", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, dto.SendAudioResponse{
			Success: false,
			Error:   "failed to send audio",
			Message: err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, dto.SendAudioResponse{
		Success: true,
	})
}

// For evolution compatibility
func (s *Message) SendMedia(ctx echo.Context) error {
	var request dto.SendMediaRequest
	if err := ctx.Bind(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.SendDocumentResponse{
			Success: false,
			Error:   "failed to bind request body",
			Message: err.Error(),
		})
	}

	if err := validator.New().Struct(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.SendDocumentResponse{
			Success: false,
			Error:   "invalid request body",
			Message: err.Error(),
		})
	}

	// Detectar mimetype automáticamente desde la URL
	detectedMimetype, err := detectMimetypeFromURL(request.SendDocumentRequest.Media)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.SendDocumentResponse{
			Success: false,
			Error:   "failed to detect mimetype from URL",
			Message: err.Error(),
		})
	}

	switch request.Mediatype {
	case "image":
		return s.sendImage(ctx, request.SendDocumentRequest, detectedMimetype)
	}

	return s.sendDocument(ctx, request.SendDocumentRequest, detectedMimetype)
}

func (s *Message) SendDocument(ctx echo.Context) error {
	var request dto.SendDocumentRequest
	if err := ctx.Bind(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.SendDocumentResponse{
			Success: false,
			Error:   "failed to bind request body",
			Message: err.Error(),
		})
	}

	if err := validator.New().Struct(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.SendDocumentResponse{
			Success: false,
			Error:   "invalid request body",
			Message: err.Error(),
		})
	}

	// Detectar mimetype automáticamente desde la URL
	detectedMimetype, err := detectMimetypeFromURL(request.Media)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.SendDocumentResponse{
			Success: false,
			Error:   "failed to detect mimetype from URL",
			Message: err.Error(),
		})
	}

	// Extraer extensión de la URL para determinar si es imagen o documento
	ext := strings.ToLower(filepath.Ext(request.Media))
	if isImageExtension(ext) {
		return s.sendImage(ctx, request, detectedMimetype)
	}

	return s.sendDocument(ctx, request, detectedMimetype)
}

func (s *Message) sendDocument(ctx echo.Context, request dto.SendDocumentRequest, mimetype string) error {
	jid, err := numberToJid(request.Number)
	if err != nil {
		zap.L().Error("error converting number to jid", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, dto.SendDocumentResponse{
			Success: false,
			Error:   "invalid number format",
			Message: err.Error(),
		})
	}

	// Extraer nombre de archivo de la URL
	fileName := filepath.Base(request.Media)
	if fileName == "" || fileName == "." {
		fileName = "document"
	}

	sendData := &whatsmiau.SendDocumentRequest{
		InstanceID: request.InstanceID,
		MediaURL:   request.Media,
		Caption:    request.Caption,
		FileName:   fileName,
		RemoteJID:  jid,
		Mimetype:   mimetype,
	}

	c := ctx.Request().Context()
	_, err = s.whatsmiau.SendDocument(c, sendData)
	if err != nil {
		zap.L().Error("Whatsmiau.SendDocument failed", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, dto.SendDocumentResponse{
			Success: false,
			Error:   "failed to send document",
			Message: err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, dto.SendDocumentResponse{
		Success: true,
	})
}

func (s *Message) sendImage(ctx echo.Context, request dto.SendDocumentRequest, mimetype string) error {
	jid, err := numberToJid(request.Number)
	if err != nil {
		zap.L().Error("error converting number to jid", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, dto.SendDocumentResponse{
			Success: false,
			Error:   "invalid number format",
			Message: err.Error(),
		})
	}

	sendData := &whatsmiau.SendImageRequest{
		InstanceID: request.InstanceID,
		MediaURL:   request.Media,
		Caption:    request.Caption,
		RemoteJID:  jid,
		Mimetype:   mimetype,
	}

	c := ctx.Request().Context()
	_, err = s.whatsmiau.SendImage(c, sendData)
	if err != nil {
		zap.L().Error("Whatsmiau.SendImage failed", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, dto.SendDocumentResponse{
			Success: false,
			Error:   "failed to send image",
			Message: err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, dto.SendDocumentResponse{
		Success: true,
	})
}
