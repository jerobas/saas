package main

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"fmt"
)

//go:embed license/public.pem
var publicKeyData []byte

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

// startup é chamado quando o Wails inicia.
// Aqui você pode verificar a licença logo na abertura.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	ok, _ := getUserStatus()
	if !ok {
		fmt.Println("License is not active or has expired.")
		// Aqui você poderia emitir um evento para o frontend
		// redirecionar para a tela de ativação, por exemplo.
	}
}

// UserService lida com a parte de identidade e licença do usuário
type UserService struct{}

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

// ActivateLicense valida a string da licença usando a chave pública ed25519
func (u *UserService) ActivateLicense(licenseString string) (bool, error) {
	block, _ := pem.Decode(publicKeyData)
	if block == nil {
		return false, fmt.Errorf("invalid public key")
	}

	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, err
	}

	pub, ok := pubAny.(ed25519.PublicKey)
	if !ok {
		return false, fmt.Errorf("not an ed25519 public key")
	}

	// Assume que ValidateLicenseWithKey está definido em outro arquivo (ex: license_utils.go)
	payload, err := ValidateLicenseWithKey(licenseString, pub)
	if err != nil {
		return false, err
	}

	// Salva e ativa localmente
	if err := saveUserData(payload.U, payload.E); err != nil {
		return false, err
	}
	if err := updateUserData(payload.U, true, payload.X); err != nil {
		return false, err
	}

	return true, nil
}
