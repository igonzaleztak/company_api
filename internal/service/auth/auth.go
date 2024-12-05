package auth

import (
	"context"
	"fmt"
	"time"
	apierrors "xm_test/internal/api_errors"
	"xm_test/internal/crypto"
	"xm_test/internal/db"
	"xm_test/internal/db/models"
	"xm_test/internal/token"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type auth struct {
	logger *zap.SugaredLogger
	db     db.DatabaseAdapter
}

// NewAuthService returns a new auth service instance
func NewAuthResolver(logger *zap.SugaredLogger, db db.DatabaseAdapter) *auth {
	return &auth{
		logger: logger,
		db:     db,
	}
}

// Register registers a new user
func (s *auth) Register(email string, password string) error {
	s.logger.Infof("registering user with email '%s'", email)

	s.logger.Debugf("creating new user with email '%s'", email)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user := &models.UserModel{
		ID:          uuid.New(),
		Email:       email,
		EncPassword: crypto.Md5Hash(password),
	}
	if err := s.db.CreateUser(ctx, user); err != nil {
		return err
	}

	s.logger.Debugf("user with email '%s' created", email)
	s.logger.Infof("user with email '%s' registered", email)
	return nil
}

// Login logs in a user
func (s *auth) Login(email string, password string) (*string, error) {
	s.logger.Infof("logging in user with email '%s'", email)

	// get user from db
	s.logger.Debugf("retrieving user with email '%s' from database", email)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	s.logger.Debugf("user with email '%s' retrieved from database", email)

	// check password
	s.logger.Debugf("checking password for user with email '%s'", email)
	if user.EncPassword != crypto.Md5Hash(password) {
		err := apierrors.ErrInvalidCredentials
		return nil, err
	}
	s.logger.Debugf("password for user with email '%s' is correct", email)

	// generate token
	s.logger.Debugf("generating token for user with email '%s'", email)
	token, _, err := token.GenerateToken(user.ID.String(), email)
	if err != nil {
		msg := fmt.Errorf("error generating token for user with email '%s'", email)
		return nil, msg
	}
	s.logger.Debugf("token generated for user with email '%s'", email)
	s.logger.Infof("user with email '%s' logged in", email)
	return &token, nil
}
