package service

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/Reverseflash45/sipinjam-lab/app/model"
	"github.com/Reverseflash45/sipinjam-lab/app/repository"
	"github.com/Reverseflash45/sipinjam-lab/helper"
)

type AuthService interface {
	Register(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	Me(c *fiber.Ctx) error
}

type authService struct {
	repo      repository.UserRepository
	jwtSecret string
	jwtTTL    time.Duration
}

func NewAuthService(repo repository.UserRepository, jwtSecret string, jwtTTL time.Duration) AuthService {
	return &authService{repo: repo, jwtSecret: jwtSecret, jwtTTL: jwtTTL}
}

func (s *authService) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("format body tidak valid")
	}
	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.IdentityNo = strings.TrimSpace(req.IdentityNo)

	if _, err := s.repo.FindByEmail(c.Context(), req.Email); err == nil {
		return helper.Conflict("EMAIL_ALREADY_REGISTERED", "email sudah terdaftar")
	} else if !errors.Is(err, repository.ErrNotFound) {
		return translate(err)
	}

	if _, err := s.repo.FindByIdentityNo(c.Context(), req.IdentityNo); err == nil {
		return helper.Conflict("IDENTITY_NO_ALREADY_REGISTERED", "NIM/NIP sudah terdaftar")
	} else if !errors.Is(err, repository.ErrNotFound) {
		return translate(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return helper.Internal(err)
	}

	user := &model.User{
		Name:         strings.TrimSpace(req.Name),
		Email:        req.Email,
		IdentityNo:   req.IdentityNo,
		PasswordHash: string(hash),
		Role:         req.Role,
	}

	created, err := s.repo.Create(c.Context(), user)
	if err != nil {
		return translate(err)
	}

	token, err := s.generateToken(created)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Created(c, "REGISTER_SUCCESS", fiber.Map{
		"token": token,
		"user":  created,
	})
}

func (s *authService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("format body tidak valid")
	}
	if fields := helper.ValidateStruct(req); fields != nil {
		return helper.Validation(fields)
	}

	user, err := s.repo.FindByIdentityNo(c.Context(), strings.TrimSpace(req.IdentityNo))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Unauthorized("NIM/NIP atau password salah")
		}
		return translate(err)
	}

	if !user.IsActive {
		return helper.Forbidden("akun tidak aktif")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return helper.Unauthorized("NIM/NIP atau password salah")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "LOGIN_SUCCESS", fiber.Map{
		"token": token,
		"user":  user,
	})
}

func (s *authService) Me(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int)
	if !ok {
		return helper.Unauthorized("sesi tidak valid")
	}

	user, err := s.repo.FindByID(c.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Unauthorized("pengguna tidak ditemukan")
		}
		return translate(err)
	}

	return helper.Success(c, fiber.StatusOK, "OK", user)
}

func (s *authService) generateToken(u *model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":         u.ID,
		"role":        u.Role,
		"identity_no": u.IdentityNo,
		"iat":         time.Now().Unix(),
		"exp":         time.Now().Add(s.jwtTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}