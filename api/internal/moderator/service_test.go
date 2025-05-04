package moderator

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apperror "github.com/zulfikarrosadi/code_roast/internal/app-error"
	"github.com/zulfikarrosadi/code_roast/internal/user"
	"github.com/zulfikarrosadi/code_roast/pkg/schema"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) addRoles(ctx context.Context, data updateRoleRequest) (userAndRole, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(userAndRole), args.Error(1)
}

func (m *mockRepository) removeRoles(ctx context.Context, data updateRoleRequest) (userAndRole, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(userAndRole), args.Error(1)
}

func TestServiceImpl_addRoles(t *testing.T) {
	tests := []struct {
		name          string
		request       updateRoleRequest
		response      schema.Response[UpdatePermissionResponse]
		repoResponse  userAndRole
		repoError     error
		expectSuccess bool
		expectStatus  string
		expectCode    int
	}{
		{
			name: "Successfull add roles",
			request: updateRoleRequest{
				UserId: "user-id",
				RoleId: []int{user.ROLE_ID_TAKE_DOWN_POST, user.ROLE_ID_APPROVE_POST},
			},
			response: schema.Response[UpdatePermissionResponse]{
				Status: "success",
				Code:   http.StatusCreated,
				Data: UpdatePermissionResponse{
					User: updateRoleResponse{
						UserId: "user-id",
						Roles: []roles{
							{Id: user.ROLE_ID_TAKE_DOWN_POST, Name: "take_down_post"},
							{Id: user.ROLE_ID_APPROVE_POST, Name: "approve_post"},
						},
					},
				},
			},
			repoError:     nil,
			repoResponse:  userAndRole{},
			expectSuccess: true,
			expectStatus:  "success",
			expectCode:    http.StatusCreated,
		},
		{
			name:    "validation error",
			request: updateRoleRequest{},
			response: schema.Response[UpdatePermissionResponse]{
				Status: "fail",
				Code:   http.StatusBadRequest,
				Error: schema.Error{
					Message: "validation error",
					Details: map[string]string{
						"user_id": "User id is required",
						"role_id": "Role id is required",
					},
				},
			},
			repoResponse:  userAndRole{},
			repoError:     nil,
			expectSuccess: false,
			expectStatus:  "fail",
			expectCode:    http.StatusBadRequest,
		},
		{
			name: "Repository error_duplicate constraint",
			request: updateRoleRequest{
				UserId: "user-id",
				RoleId: []int{user.ROLE_ID_TAKE_DOWN_POST, user.ROLE_ID_APPROVE_POST},
			},
			response: schema.Response[UpdatePermissionResponse]{
				Status: "fail",
				Code:   http.StatusBadRequest,
				Error: schema.Error{
					Message: "This user already have that role",
				},
			},
			repoResponse:  userAndRole{},
			repoError:     apperror.New(http.StatusBadRequest, "This user already have that role", nil),
			expectSuccess: false,
			expectStatus:  "fail",
			expectCode:    http.StatusBadRequest,
		},
		{
			name: "Repository error",
			request: updateRoleRequest{
				UserId: "user-id",
				RoleId: []int{0000000, 99999},
			},
			response: schema.Response[UpdatePermissionResponse]{
				Status: "fail",
				Code:   http.StatusInternalServerError,
				Error: schema.Error{
					Message: "something went wrong, please try again later",
				},
			},
			repoResponse:  userAndRole{},
			repoError:     errors.New("repository error"),
			expectSuccess: false,
			expectStatus:  "fail",
			expectCode:    http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockRepository)
			service := serviceImpl{repository: mockRepo, v: validator.New()}

			if tt.name == "Repository error" {
				mockRepo.On("addRoles", context.Background(), updateRoleRequest{
					UserId: "user-id",
					RoleId: []int{0000000, 99999},
				}).Return(userAndRole{}, tt.repoError)
			} else {
				mockRepo.On("addRoles", context.Background(), updateRoleRequest{
					UserId: "user-id",
					RoleId: []int{user.ROLE_ID_TAKE_DOWN_POST, user.ROLE_ID_APPROVE_POST},
				}).Return(userAndRole{
					userId: "user-id",
					roles: []roles{
						{Id: user.ROLE_ID_TAKE_DOWN_POST, Name: "take_down_post"},
						{Id: user.ROLE_ID_APPROVE_POST, Name: "approve_post"},
					},
				}, tt.repoError)
			}

			resp, err := service.addRoles(context.Background(), tt.request)
			if tt.expectSuccess {
				require.NoError(t, err)
				assert.Equal(t, tt.expectStatus, resp.Status)
				assert.Equal(t, tt.expectCode, resp.Code)
				assert.Equal(t, tt.response.Data.User.UserId, resp.Data.User.UserId)
				assert.Equal(t, tt.response.Data.User.Roles, resp.Data.User.Roles)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expectStatus, resp.Status)
				assert.Equal(t, tt.expectCode, resp.Code)
			}
		})
	}
}
