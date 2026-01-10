package controllers

import (
	"encoding/base64"
	"math/rand/v2"
	"net/http"

	"github.com/verbeux-ai/whatsmiau/env"
	"github.com/verbeux-ai/whatsmiau/lib/whatsmiau"
	"github.com/verbeux-ai/whatsmiau/models"
	"go.mau.fi/whatsmeow/types"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/skip2/go-qrcode"
	"github.com/verbeux-ai/whatsmiau/interfaces"
	"github.com/verbeux-ai/whatsmiau/server/dto"
	"github.com/verbeux-ai/whatsmiau/utils"
	"go.uber.org/zap"
)

type Instance struct {
	repo      interfaces.InstanceRepository
	whatsmiau *whatsmiau.Whatsmiau
}

func NewInstances(repository interfaces.InstanceRepository, whatsmiau *whatsmiau.Whatsmiau) *Instance {
	return &Instance{
		repo:      repository,
		whatsmiau: whatsmiau,
	}
}

func (s *Instance) Create(ctx echo.Context) error {
	var request dto.CreateInstanceRequest
	if err := ctx.Bind(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.CreateInstanceResponse{
			Success: false,
			Error:   "failed to bind request body",
			Message: err.Error(),
		})
	}

	if err := validator.New().Struct(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.CreateInstanceResponse{
			Success: false,
			Error:   "invalid request body",
			Message: err.Error(),
		})
	}

	// Crear instancia con valores por defecto
	instance := &models.Instance{
		ID:        request.ID,
		RemoteJID: "",
	}

	// Asignar proxy si está disponible en el env
	if len(env.Env.ProxyAddresses) > 0 {
		rd := rand.IntN(len(env.Env.ProxyAddresses))
		proxyUrl := env.Env.ProxyAddresses[rd]

		proxy, err := parseProxyURL(proxyUrl)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, dto.CreateInstanceResponse{
				Success: false,
				Error:   "invalid proxy url on env",
				Message: err.Error(),
			})
		}
		instance.InstanceProxy = *proxy
	}

	c := ctx.Request().Context()
	if err := s.repo.Create(c, instance); err != nil {
		zap.L().Error("failed to create instance", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, dto.CreateInstanceResponse{
			Success: false,
			Error:   "failed to create instance",
			Message: err.Error(),
		})
	}

	return ctx.JSON(http.StatusCreated, dto.CreateInstanceResponse{
		Success: true,
	})
}

func (s *Instance) List(ctx echo.Context) error {
	c := ctx.Request().Context()
	var request dto.ListInstancesRequest
	if err := ctx.Bind(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusUnprocessableEntity, err, "failed to bind request body")
	}
	if request.InstanceName == "" {
		request.InstanceName = request.ID
	}

	result, err := s.repo.List(c, request.InstanceName)
	if err != nil {
		zap.L().Error("failed to list instances", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to list instances")
	}

	var response []dto.ListInstancesResponse
	for _, instance := range result {
		jid, err := types.ParseJID(instance.RemoteJID)
		if err != nil {
			zap.L().Error("failed to parse jid", zap.Error(err))
		}

		response = append(response, dto.ListInstancesResponse{
			Instance:     &instance,
			OwnerJID:     jid.ToNonAD().String(),
			InstanceName: instance.ID,
		})
	}

	if len(response) == 0 {
		return ctx.JSON(http.StatusOK, []string{})
	}

	return ctx.JSON(http.StatusOK, response)
}

func (s *Instance) Connect(ctx echo.Context) error {
	c := ctx.Request().Context()
	var request dto.ConnectInstanceRequest
	if err := ctx.Bind(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusUnprocessableEntity, err, "failed to bind request body")
	}

	result, err := s.repo.List(c, request.ID)
	if err != nil {
		zap.L().Error("failed to list instances", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to list instances")
	}

	if len(result) == 0 {
		return utils.HTTPFail(ctx, http.StatusNotFound, err, "instance not found")
	}

	// Verificar el estado actual de la instancia
	status, err := s.whatsmiau.Status(request.ID)
	if err != nil {
		zap.L().Error("failed to get status instance", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to get status instance")
	}

	// Si ya está conectada, retornar que ya está conectada
	if status == whatsmiau.Connected {
		return ctx.JSON(http.StatusOK, dto.ConnectInstanceResponse{
			Message:   "instance already connected",
			Connected: true,
		})
	}

	// Verificar si hay un QR code en cache
	if qrCode, ok := s.whatsmiau.GetQRCode(request.ID); ok && qrCode != "" {
		png, err := qrcode.Encode(qrCode, qrcode.Medium, 512)
		if err != nil {
			zap.L().Error("failed to encode qrcode", zap.Error(err))
			return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to encode qrcode")
		}
		return ctx.JSON(http.StatusOK, dto.ConnectInstanceResponse{
			Message:   "QR code available",
			Connected: false,
			Base64:    "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
		})
	}

	// Si el observer está corriendo pero aún no hay QR, intentar conectar y esperar un poco
	if s.whatsmiau.IsObserverRunning(request.ID) {
		// El observer ya está corriendo, intentar obtener el QR
		qrCode, err := s.whatsmiau.Connect(c, request.ID)
		if err != nil {
			zap.L().Error("failed to connect instance", zap.Error(err))
			return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to connect instance")
		}
		if qrCode != "" {
			png, err := qrcode.Encode(qrCode, qrcode.Medium, 512)
			if err != nil {
				zap.L().Error("failed to encode qrcode", zap.Error(err))
				return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to encode qrcode")
			}
			return ctx.JSON(http.StatusOK, dto.ConnectInstanceResponse{
				Message:   "QR code available",
				Connected: false,
				Base64:    "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
			})
		}
	}

	// Iniciar la conexión y obtener el QR
	qrCode, err := s.whatsmiau.Connect(c, request.ID)
	if err != nil {
		zap.L().Error("failed to connect instance", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to connect instance")
	}

	// Si hay QR code, retornarlo en base64
	if qrCode != "" {
		png, err := qrcode.Encode(qrCode, qrcode.Medium, 512)
		if err != nil {
			zap.L().Error("failed to encode qrcode", zap.Error(err))
			return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to encode qrcode")
		}
		return ctx.JSON(http.StatusOK, dto.ConnectInstanceResponse{
			Message:   "QR code available",
			Connected: false,
			Base64:    "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
		})
	}

	// Si no hay QR code, significa que ya está conectada (el método Connect retorna "" cuando ya está conectada)
	return ctx.JSON(http.StatusOK, dto.ConnectInstanceResponse{
		Message:   "instance already connected",
		Connected: true,
	})
}

func (s *Instance) ConnectQRBuffer(ctx echo.Context) error {
	c := ctx.Request().Context()
	var request dto.ConnectInstanceRequest
	if err := ctx.Bind(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusUnprocessableEntity, err, "failed to bind request body")
	}

	result, err := s.repo.List(c, request.ID)
	if err != nil {
		zap.L().Error("failed to list instances", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to list instances")
	}

	if len(result) == 0 {
		return utils.HTTPFail(ctx, http.StatusNotFound, err, "instance not found")
	}

	qrCode, err := s.whatsmiau.Connect(c, request.ID)
	if err != nil {
		zap.L().Error("failed to connect instance", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to connect instance")
	}
	if qrCode != "" {
		png, err := qrcode.Encode(qrCode, qrcode.Medium, 256)
		if err != nil {
			zap.L().Error("failed to encode qrcode", zap.Error(err))
			return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to encode qrcode")
		}
		return ctx.Blob(http.StatusOK, "image/png", png)
	}

	return ctx.NoContent(http.StatusOK)
}

func (s *Instance) Status(ctx echo.Context) error {
	c := ctx.Request().Context()
	var request dto.ConnectInstanceRequest
	if err := ctx.Bind(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusUnprocessableEntity, err, "failed to bind request body")
	}

	result, err := s.repo.List(c, request.ID)
	if err != nil {
		zap.L().Error("failed to list instances", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to list instances")
	}

	if len(result) == 0 {
		return utils.HTTPFail(ctx, http.StatusNotFound, err, "instance not found")
	}

	status, err := s.whatsmiau.Status(request.ID)
	if err != nil {
		zap.L().Error("failed to get status instance", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to get status instance")
	}

	return ctx.JSON(http.StatusOK, dto.StatusInstanceResponse{
		ID:     request.ID,
		Status: string(status),
		Instance: &dto.StatusInstanceResponseEvolutionCompatibility{
			InstanceName: request.ID,
			State:        string(status),
		},
	})
}

func (s *Instance) Logout(ctx echo.Context) error {
	c := ctx.Request().Context()
	var request dto.DeleteInstanceRequest
	if err := ctx.Bind(&request); err != nil {
		return utils.HTTPFail(ctx, http.StatusUnprocessableEntity, err, "failed to bind request body")
	}

	result, err := s.repo.List(c, request.ID)
	if err != nil {
		zap.L().Error("failed to list instances", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to list instances")
	}

	if len(result) == 0 {
		return utils.HTTPFail(ctx, http.StatusNotFound, err, "instance not found")
	}

	if err := s.whatsmiau.Logout(c, request.ID); err != nil {
		zap.L().Error("failed to logout instance", zap.Error(err))
		return utils.HTTPFail(ctx, http.StatusInternalServerError, err, "failed to logout instance")
	}

	return ctx.JSON(http.StatusOK, dto.DeleteInstanceResponse{
		Message: "instance logout successfully",
	})
}

func (s *Instance) Delete(ctx echo.Context) error {
	c := ctx.Request().Context()
	var request dto.DeleteInstanceRequest
	if err := ctx.Bind(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.DeleteInstanceResponse{
			Success: false,
			Error:   "failed to bind request body",
			Message: err.Error(),
		})
	}

	result, err := s.repo.List(c, request.ID)
	if err != nil {
		zap.L().Error("failed to list instances", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, dto.DeleteInstanceResponse{
			Success: false,
			Error:   "failed to list instances",
			Message: err.Error(),
		})
	}

	if len(result) == 0 {
		return ctx.JSON(http.StatusOK, dto.DeleteInstanceResponse{
			Success: false,
			Error:   "instance doesn't exists",
			Message: "instance not found",
		})
	}

	if err := s.whatsmiau.Logout(ctx.Request().Context(), request.ID); err != nil {
		zap.L().Error("failed to disconnect instance", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, dto.DeleteInstanceResponse{
			Success: false,
			Error:   "failed to logout instance",
			Message: err.Error(),
		})
	}

	if err := s.repo.Delete(c, request.ID); err != nil {
		zap.L().Error("failed to delete instance", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, dto.DeleteInstanceResponse{
			Success: false,
			Error:   "failed to delete instance",
			Message: err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, dto.DeleteInstanceResponse{
		Success: true,
	})
}
