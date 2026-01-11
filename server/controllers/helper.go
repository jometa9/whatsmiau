package controllers

import (
	"fmt"
	"mime"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/verbeux-ai/whatsmiau/models"
	"go.mau.fi/whatsmeow/types"
)

func numberToJid(number string) (*types.JID, error) {
	splitNumber := strings.Split(number, "@")
	if len(splitNumber) != 2 {
		number += "@s.whatsapp.net"
	}

	if len(splitNumber[0]) < 12 {
		return nil, fmt.Errorf("invalid jid, put country prefix")
	}

	jid, err := types.ParseJID(number)
	if err != nil {
		return nil, fmt.Errorf("invalid jid (number)")
	}

	return &jid, nil
}

func parseProxyURL(proxyURL string) (*models.InstanceProxy, error) {
	if !strings.Contains(proxyURL, "://") {
		return nil, fmt.Errorf("invalid proxy url, missing scheme: %s", proxyURL)
	}

	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy url: %w", err)
	}

	var username, password string
	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()
	}

	host, port, err := splitHostPort(u.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid host/port: %w", err)
	}

	return &models.InstanceProxy{
		ProxyHost:     host,
		ProxyPort:     port,
		ProxyProtocol: strings.ToUpper(u.Scheme),
		ProxyUsername: username,
		ProxyPassword: password,
	}, nil
}

func splitHostPort(h string) (string, string, error) {
	parts := strings.Split(h, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected host:port, got %s", h)
	}
	return parts[0], parts[1], nil
}

// detectMimetypeFromURL detecta el mimetype basándose en la extensión de la URL
// Retorna el mimetype y un error si la extensión no es válida o no está soportada
func detectMimetypeFromURL(urlStr string) (string, error) {
	// Extraer la extensión de la URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(u.Path))
	if ext == "" {
		return "", fmt.Errorf("no file extension found in URL")
	}

	// Detectar mimetype basándose en la extensión
	mimetype := mime.TypeByExtension(ext)
	if mimetype == "" {
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}

	return mimetype, nil
}

