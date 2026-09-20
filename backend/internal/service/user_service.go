package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/blueship581/cyfreshfood/internal/constants"
	"github.com/blueship581/cyfreshfood/internal/model"
	"github.com/blueship581/cyfreshfood/internal/repository"
	"github.com/blueship581/cyfreshfood/internal/util"
	"golang.org/x/crypto/bcrypt"
)

// UserService 用户服务：注册、登录、资料维护。
type UserService struct {
	repo       *repository.UserRepository
	jwtSecret  string
	jwtExpire  int
	log        *slog.Logger
}

// NewUserService 构造用户服务。
func NewUserService(repo *repository.UserRepository, jwtSecret string, jwtExpire int, log *slog.Logger) *UserService {
	return &UserService{repo: repo, jwtSecret: jwtSecret, jwtExpire: jwtExpire, log: log}
}

// Register 注册用户，返回用户与 JWT。
func (s *UserService) Register(ctx context.Context, phone, password, name string) (*model.User, string, error) {
	s.log.InfoContext(ctx, constants.LOG_USER_REGISTER_START, "phone", phone)
	count, err := s.repo.CountByPhone(phone)
	if err != nil {
		return nil, "", util.LogError(s.log, ctx, constants.LOG_USER_REGISTER_START, fmt.Errorf("count user by phone: %w", err))
	}
	if count > 0 {
		return nil, "", util.ConflictError(constants.MsgDuplicatePhone, errors.New("user.phone duplicated"))
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", util.LogError(s.log, ctx, constants.LOG_USER_REGISTER_START, fmt.Errorf("hash password: %w", err))
	}
	user := &model.User{Phone: phone, PasswordHash: string(hash), Name: name, Role: constants.RoleMember}
	if err := s.repo.Create(user); err != nil {
		return nil, "", util.LogError(s.log, ctx, constants.LOG_USER_REGISTER_START, fmt.Errorf("create user: %w", err))
	}
	token, err := util.GenerateToken(s.jwtSecret, s.jwtExpire, user.ID, user.Phone, user.Role)
	if err != nil {
		return nil, "", util.LogError(s.log, ctx, constants.LOG_USER_REGISTER_START, fmt.Errorf("generate token: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_USER_REGISTER_SUCCESS, "user_id", user.ID)
	return user, token, nil
}

// Login 登录，校验手机号密码并返回 JWT。
func (s *UserService) Login(ctx context.Context, phone, password string) (*model.User, string, error) {
	user, err := s.repo.FindByPhone(phone)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, "", util.UnauthorizedError(constants.MsgLoginFailed, errors.New("user.phone not found"))
		}
		return nil, "", util.LogError(s.log, ctx, constants.LOG_USER_LOGIN_FAILED, fmt.Errorf("find user: %w", err))
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, "", util.UnauthorizedError(constants.MsgLoginFailed, errors.New("user.password mismatch"))
	}
	token, err := util.GenerateToken(s.jwtSecret, s.jwtExpire, user.ID, user.Phone, user.Role)
	if err != nil {
		return nil, "", util.LogError(s.log, ctx, constants.LOG_USER_LOGIN_FAILED, fmt.Errorf("generate token: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_USER_LOGIN_SUCCESS, "user_id", user.ID)
	return user, token, nil
}

// GetByID 查询用户。
func (s *UserService) GetByID(ctx context.Context, id uint) (*model.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NotFoundError(constants.MsgUserNotFound, errors.New("user not found"))
		}
		return nil, util.LogError(s.log, ctx, constants.LOG_USER_LOGIN_START, fmt.Errorf("find user by id: %w", err))
	}
	return user, nil
}

// UpdateProfile 更新资料。
func (s *UserService) UpdateProfile(ctx context.Context, id uint, name, avatar string) (*model.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.NotFoundError(constants.MsgUserNotFound, err)
	}
	if name != "" {
		user.Name = name
	}
	if avatar != "" {
		user.Avatar = avatar
	}
	if err := s.repo.Update(user); err != nil {
		return nil, util.LogError(s.log, ctx, constants.LOG_USER_PROFILE_UPDATED, fmt.Errorf("update user: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_USER_PROFILE_UPDATED, "user_id", id)
	return user, nil
}
