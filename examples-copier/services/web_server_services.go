package services

import (
	"context"
	. "github.com/mongodb/code-example-tooling/code-copier/types"
)

// **** SERVICES **** //

// AuthService handles GitHub authentication
type AuthService struct {
	installationAccessToken string
}

// ReadService handles reading data from GitHub
type ReadService struct {
	ctx         context.Context
	authService *AuthService
}

// WriteService handles writing to GitHub repositories
type WriteService struct {
	filesToUpload    map[UploadKey]UploadFileContent
	filesToDeprecate map[string]Configs
	authService      *AuthService
}

// WebhookService orchestrates handling webhook events
type WebhookService struct {
	authService  *AuthService
	readService  *ReadService
	writeService *WriteService
}

// **** SERVICE CONSTRUCTORS **** //

func NewAuthService() *AuthService {
	return &AuthService{}
}

func NewReadService(authService *AuthService) *ReadService {
	return &ReadService{
		ctx:         context.Background(),
		authService: authService,
	}
}

func NewWriteService(authService *AuthService) *WriteService {
	return &WriteService{
		filesToUpload:    make(map[UploadKey]UploadFileContent),
		filesToDeprecate: make(map[string]Configs),
		authService:      authService,
	}
}

func NewWebhookService(authService *AuthService, readService *ReadService, writeService *WriteService) *WebhookService {
	return &WebhookService{
		authService:  authService,
		readService:  readService,
		writeService: writeService,
	}
}
