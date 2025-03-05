package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apperror "github.com/zulfikarrosadi/code_roast/internal/app-error"
	"github.com/zulfikarrosadi/code_roast/internal/user"
	"golang.org/x/crypto/bcrypt"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) register(ctx context.Context, newUser user.User, auth authentication) (publicUserData, error) {
	args := m.Called(ctx, newUser, auth)
	return args.Get(0).(publicUserData), args.Error(1)
}

func (m *mockRepository) loginByEmail(ctx context.Context, email string, auth authentication) (user.User, error) {
	args := m.Called(ctx, email, auth)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *mockRepository) findRefreshToken(ctx context.Context, token string) (publicUserData, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(publicUserData), args.Error(1)
}

func TestServiceImpl_register(t *testing.T) {
	tests := []struct {
		name          string
		request       registrationRequest
		repoResponse  publicUserData
		repoError     error
		expectSuccess bool
		expectStatus  string
		expectCode    int
	}{
		{
			name: "Successful registration",
			request: registrationRequest{
				Fullname:             "Test User",
				Email:                "test@example.com",
				Password:             "password123",
				PasswordConfirmation: "password123",
				Agent:                "Mozilla",
				RemoteIp:             "127.0.0.1",
			},
			repoResponse: publicUserData{
				id:       "generated-id",
				email:    "test@example.com",
				fullname: "Test User",
				roles:    []user.Roles{},
			},
			repoError:     nil,
			expectSuccess: true,
			expectStatus:  "success",
			expectCode:    http.StatusCreated,
		},
		{
			name: "Validation error",
			request: registrationRequest{
				Fullname:             "",
				Email:                "invalid-email",
				Password:             "",
				PasswordConfirmation: "",
			},
			repoResponse:  publicUserData{},
			repoError:     nil,
			expectSuccess: false,
			expectStatus:  "fail",
			expectCode:    http.StatusBadRequest,
		},
		{
			name: "Repository error",
			request: registrationRequest{
				Fullname:             "Test User",
				Email:                "test@example.com",
				Password:             "password123",
				PasswordConfirmation: "password123",
				Agent:                "Mozilla",
				RemoteIp:             "127.0.0.1",
			},
			repoResponse:  publicUserData{},
			repoError:     errors.New("database error"),
			expectSuccess: false,
			expectStatus:  "fail",
			expectCode:    http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			validator := validator.New()
			service := &ServiceImpl{Repository: mockRepo, v: validator}

			mockRepo.On("register", mock.Anything, mock.Anything, mock.Anything).Return(tt.repoResponse, tt.repoError)

			resp, err := service.register(context.Background(), tt.request)

			if tt.expectSuccess {
				require.NoError(t, err)
				assert.Equal(t, tt.expectStatus, resp.Status)
				assert.Equal(t, tt.expectCode, resp.Code)
				assert.Equal(t, tt.repoResponse.id, resp.Data.User.ID)
				assert.Equal(t, tt.repoResponse.email, resp.Data.User.Email)
				assert.Equal(t, tt.repoResponse.fullname, resp.Data.User.Fullname)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expectStatus, resp.Status)
				assert.Equal(t, tt.expectCode, resp.Code)
			}
		})
	}
}

func TestServiceImpl_loginByEmail(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), 10)
	tests := []struct {
		name          string
		request       loginRequest
		repoResponse  user.User
		repoError     error
		expectSuccess bool
		expectStatus  string
		expectCode    int
	}{
		{
			name: "Successful login",
			request: loginRequest{
				Email:    "test@example.com",
				Password: "password123",
				authentication: authentication{
					agent:     "Mozilla",
					remoteIP:  "127.0.0.1",
					lastLogin: time.Now().Unix(),
				},
			},
			repoResponse: user.User{
				Id:       "user-id",
				Fullname: "Test User",
				Email:    "test@example.com",
				Password: string(hashedPassword),
			},
			repoError:     nil,
			expectSuccess: true,
			expectStatus:  "success",
			expectCode:    http.StatusOK,
		},
		{
			name: "Password compare failed",
			request: loginRequest{
				Email:    "test@example.com",
				Password: "wrong password",
				authentication: authentication{
					remoteIP:  "127.0.0.1",
					lastLogin: time.Now().Unix(),
				},
			},
			repoResponse:  user.User{},
			repoError:     nil,
			expectStatus:  "fail",
			expectSuccess: false,
			expectCode:    http.StatusBadRequest,
		},
		{
			name: "User not found",
			request: loginRequest{
				Email:    "non-existent user",
				Password: "wrong password",
				authentication: authentication{
					remoteIP:  "127.0.0.1",
					lastLogin: time.Now().Unix(),
				},
			},
			repoResponse:  user.User{},
			repoError:     apperror.New(http.StatusBadRequest, "email or password is incorrect", nil),
			expectStatus:  "fail",
			expectSuccess: false,
			expectCode:    http.StatusBadRequest,
		},
		{
			name:          "Validation error",
			request:       loginRequest{},
			repoResponse:  user.User{},
			repoError:     nil,
			expectSuccess: false,
			expectStatus:  "fail",
			expectCode:    http.StatusBadRequest,
		},
		{
			name: "Repo error_failed to insert auth data",
			request: loginRequest{
				Email:          "test@example.com",
				Password:       "password123",
				authentication: authentication{},
			},
			repoResponse:  user.User{},
			repoError:     fmt.Errorf("repository: insert new user auth credentials failed %w", nil),
			expectSuccess: false,
			expectStatus:  "fail",
			expectCode:    http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			validator := validator.New()
			service := &ServiceImpl{Repository: mockRepo, v: validator}

			if tt.name == "User not found" {
				mockRepo.On("loginByEmail", context.Background(), "non-existent user", mock.Anything).Return(tt.repoResponse, tt.repoError)
			} else {
				mockRepo.On("loginByEmail", context.Background(), "test@example.com", mock.Anything).Return(tt.repoResponse, tt.repoError)
			}

			resp, err := service.login(context.Background(), tt.request)
			if tt.expectSuccess {
				require.NoError(t, err)
				assert.Equal(t, tt.expectStatus, resp.Status)
				assert.Equal(t, tt.expectCode, resp.Code)
				assert.Equal(t, tt.repoResponse.Id, resp.Data.User.ID)
				assert.Equal(t, tt.repoResponse.Email, resp.Data.User.Email)
				assert.Equal(t, tt.repoResponse.Fullname, resp.Data.User.Fullname)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expectStatus, resp.Status)
				assert.Equal(t, tt.expectCode, resp.Code)
			}
		})
	}
}
