package controllers

import (
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/verbeux-ai/whatsmiau/interfaces"
	"github.com/verbeux-ai/whatsmiau/lib/whatsmiau"
	"github.com/verbeux-ai/whatsmiau/server/dto"
	"go.mau.fi/whatsmeow/types"
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

	if request.Quoted != nil && len(request.Quoted.Key.Id) > 0 && len(request.Quoted.Message.Conversation) > 0 {
		sendText.QuoteMessage = request.Quoted.Message.Conversation
		sendText.QuoteMessageID = request.Quoted.Key.Id
	}

	c := ctx.Request().Context()
	if err := s.whatsmiau.ChatPresence(&whatsmiau.ChatPresenceRequest{
		InstanceID: request.InstanceID,
		RemoteJID:  jid,
		Presence:   types.ChatPresenceComposing,
	}); err != nil {
		zap.L().Error("Whatsmiau.ChatPresence", zap.Error(err))
	} else {
		time.Sleep(time.Millisecond * time.Duration(request.Delay)) // TODO: create a more robust solution
	}

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

	sendText := &whatsmiau.SendAudioRequest{
		AudioURL:   request.Audio,
		InstanceID: request.InstanceID,
		RemoteJID:  jid,
	}

	if request.Quoted != nil && len(request.Quoted.Key.Id) > 0 && len(request.Quoted.Message.Conversation) > 0 {
		sendText.QuoteMessage = request.Quoted.Message.Conversation
		sendText.QuoteMessageID = request.Quoted.Key.Id
	}

	c := ctx.Request().Context()
	if err := s.whatsmiau.ChatPresence(&whatsmiau.ChatPresenceRequest{
		InstanceID: request.InstanceID,
		RemoteJID:  jid,
		Presence:   types.ChatPresenceComposing,
		Media:      types.ChatPresenceMediaAudio,
	}); err != nil {
		zap.L().Error("Whatsmiau.ChatPresence", zap.Error(err))
	} else {
		time.Sleep(time.Millisecond * time.Duration(request.Delay)) // TODO: create a more robust solution
	}

	_, err = s.whatsmiau.SendAudio(c, sendText)
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
	switch request.Mediatype {
	case "image":
		request.SendDocumentRequest.Mimetype = "image/png"
		return s.sendImage(ctx, request.SendDocumentRequest)
	}

	return s.sendDocument(ctx, request.SendDocumentRequest)
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

	return s.sendDocument(ctx, request)
}

func (s *Message) sendDocument(ctx echo.Context, request dto.SendDocumentRequest) error {
	jid, err := numberToJid(request.Number)
	if err != nil {
		zap.L().Error("error converting number to jid", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, dto.SendDocumentResponse{
			Success: false,
			Error:   "invalid number format",
			Message: err.Error(),
		})
	}

	sendData := &whatsmiau.SendDocumentRequest{
		InstanceID: request.InstanceID,
		MediaURL:   request.Media,
		Caption:    request.Caption,
		FileName:   request.FileName,
		RemoteJID:  jid,
		Mimetype:   request.Mimetype,
	}

	c := ctx.Request().Context()
	time.Sleep(time.Millisecond * time.Duration(request.Delay)) // TODO: create a more robust solution

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

func (s *Message) SendImage(ctx echo.Context) error {
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

	return s.sendImage(ctx, request)
}

func (s *Message) sendImage(ctx echo.Context, request dto.SendDocumentRequest) error {
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
		Mimetype:   request.Mimetype,
	}

	c := ctx.Request().Context()
	time.Sleep(time.Millisecond * time.Duration(request.Delay)) // TODO: create a more robust solution

	_, err = s.whatsmiau.SendImage(c, sendData)
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
