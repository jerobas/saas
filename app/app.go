package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	_"embed"
)

//go:embed license/public.pem
var publicKeyData []byte

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// UserService struct
// Encapsula as funções relacionadas a usuários
type UserService struct{}

// NewUserService cria uma nova instância de UserService
func NewUserService() *UserService {
	return &UserService{}
}

func (u *UserService) SaveUserData(id string, email string) error {
	return saveUserData(id, email)
}

func (u *UserService) UpdateUserData(id string, active bool, expirationDate string) error {
	return updateUserData(id, active, expirationDate)
}

func (u *UserService) GetUserStatus() (bool, error) {
	return getUserStatus()
}

func (u *UserService) ActivateLicense(license string) (bool, error) {
	if len(publicKeyData) == 0 {
		return false, fmt.Errorf("Chave pública embutida não encontrada ou inválida")
	}

	block, _ := pem.Decode(publicKeyData)
	if block == nil {
		return false, fmt.Errorf("Chave pública embutida inválida")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("Erro ao analisar a chave pública embutida: %v", err)
	}

	// Validate the license using the public key
	data, err := ValidateLicenseWithKey(license, pub)
	if err != nil {
		return false, err
	}

	err = updateUserData(data.U, err == nil, data.X)
	if err != nil {
		return false, err
	}

	return true, nil
}
