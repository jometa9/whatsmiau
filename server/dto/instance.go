package dto

import "github.com/verbeux-ai/whatsmiau/models"

type CreateInstanceRequest struct {
	ID string `json:"id" validate:"required"`
}

type CreateInstanceResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

type ListInstancesRequest struct {
	InstanceName string `query:"instanceName"`
	ID           string `query:"id"`
}

type ListInstancesResponse struct {
	*models.Instance

	OwnerJID     string `json:"ownerJid,omitempty"`
	InstanceName string `json:"instanceName,omitempty"`
}

type ConnectInstanceRequest struct {
	ID string `param:"id" validate:"required"`
}

type ConnectInstanceResponse struct {
	Success    bool    `json:"success"`
	QR         *string `json:"qr,omitempty"`
	Connected  bool    `json:"connected"`
	RemoteJID  *string `json:"remoteJID,omitempty"`
}

type StatusInstanceRequest struct {
	ID string `param:"id" validate:"required"`
}

type StatusInstanceResponse struct {
	ID       string                                        `json:"id,omitempty"`
	Status   string                                        `json:"state,omitempty"`
	Instance *StatusInstanceResponseEvolutionCompatibility `json:"instance,omitempty"`
}

type StatusInstanceResponseEvolutionCompatibility struct {
	InstanceName string `json:"instanceName,omitempty"`
	State        string `json:"state,omitempty"`
}

type DeleteInstanceRequest struct {
	ID string `param:"id" validate:"required"`
}

type DeleteInstanceResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

type LogoutInstanceRequest struct {
	ID string `param:"id" validate:"required"`
}

type LogoutInstanceResponse struct {
	Message string `json:"message,omitempty"`
}
