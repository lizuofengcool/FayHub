package service

import (
	"context"
	"fayhub/internal/model"
	"fayhub/pkg/utils"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type RBACTestSuite struct {
	suite.Suite
	db  *gorm.DB
	ctx context.Context
}

func (s *RBACTestSuite) SetupSuite() {
	db, err := openTestDB()
	s.Require().NoError(err)

	s.db = db
	utils.SetGlobalDB(db)

	ctx := utils.SkipTenantIsolation(context.Background())
	s.ctx = ctx

	err = db.WithContext(ctx).AutoMigrate(
		&model.Role{},
		&model.User{},
		&model.UserRole{},
		&model.RoleMenu{},
		&model.RoleAPI{},
		&model.Menu{},
		&model.API{},
	)
	s.Require().NoError(err)
}

func (s *RBACTestSuite) TearDownSuite() {
	if s.db != nil {
		sqlDB, _ := s.db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

func (s *RBACTestSuite) TestCreateRole_Success() {
	svc := &RBACService{}

	role, err := svc.CreateRole(s.ctx, CreateRoleRequest{
		Name:        "test_role_create",
		Description: "测试创建角色",
		Type:        2,
	})

	s.Assert().NoError(err)
	s.Assert().NotNil(role)
	s.Assert().Equal("test_role_create", role.Name)
	s.Assert().Equal(2, role.Type)
	s.Assert().Equal(1, role.Status)
}

func (s *RBACTestSuite) TestCreateRole_DuplicateName() {
	svc := &RBACService{}

	_, err := svc.CreateRole(s.ctx, CreateRoleRequest{
		Name: "dup_role_name",
		Type: 2,
	})
	s.Require().NoError(err)

	role2, err := svc.CreateRole(s.ctx, CreateRoleRequest{
		Name: "dup_role_name",
		Type: 2,
	})

	s.Assert().Error(err)
	s.Assert().Nil(role2)
	s.Assert().Contains(err.Error(), "角色名称已存在")
}

func (s *RBACTestSuite) TestGetRoleByID_Success() {
	svc := &RBACService{}

	created, err := svc.CreateRole(s.ctx, CreateRoleRequest{
		Name: "get_role_test",
		Type: 2,
	})
	s.Require().NoError(err)

	role, err := svc.GetRoleByID(s.ctx, created.ID)

	s.Assert().NoError(err)
	s.Assert().NotNil(role)
	s.Assert().Equal("get_role_test", role.Name)
}

func (s *RBACTestSuite) TestGetRoleByID_NotFound() {
	svc := &RBACService{}

	role, err := svc.GetRoleByID(s.ctx, 99999)

	s.Assert().Error(err)
	s.Assert().Nil(role)
	s.Assert().Contains(err.Error(), "角色不存在")
}

func (s *RBACTestSuite) TestUpdateRole_Success() {
	svc := &RBACService{}

	created, err := svc.CreateRole(s.ctx, CreateRoleRequest{
		Name:        "update_role_before",
		Description: "before",
		Type:        2,
	})
	s.Require().NoError(err)

	updated, err := svc.UpdateRole(s.ctx, created.ID, UpdateRoleRequest{
		Name:        "update_role_after",
		Description: "after",
		Status:      1,
	})

	s.Assert().NoError(err)
	s.Assert().NotNil(updated)
	s.Assert().Equal("update_role_after", updated.Name)
	s.Assert().Equal("after", updated.Description)
}

func (s *RBACTestSuite) TestUpdateRole_DuplicateName() {
	svc := &RBACService{}

	_, err := svc.CreateRole(s.ctx, CreateRoleRequest{
		Name: "existing_role_name",
		Type: 2,
	})
	s.Require().NoError(err)

	created2, err := svc.CreateRole(s.ctx, CreateRoleRequest{
		Name: "another_role_name",
		Type: 2,
	})
	s.Require().NoError(err)

	updated, err := svc.UpdateRole(s.ctx, created2.ID, UpdateRoleRequest{
		Name: "existing_role_name",
	})

	s.Assert().Error(err)
	s.Assert().Nil(updated)
	s.Assert().Contains(err.Error(), "角色名称已存在")
}

func (s *RBACTestSuite) TestDeleteRole_Success() {
	svc := &RBACService{}

	created, err := svc.CreateRole(s.ctx, CreateRoleRequest{
		Name: "delete_role_test",
		Type: 2,
	})
	s.Require().NoError(err)

	err = svc.DeleteRole(s.ctx, created.ID)
	s.Assert().NoError(err)

	_, err = svc.GetRoleByID(s.ctx, created.ID)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "角色不存在")
}

func (s *RBACTestSuite) TestDeleteRole_SuperAdminProtected() {
	svc := &RBACService{}

	role := &model.Role{
		Name:   "super_admin",
		Type:   1,
		Status: 1,
	}
	s.Require().NoError(s.db.WithContext(s.ctx).Create(role).Error)

	err := svc.DeleteRole(s.ctx, role.ID)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "超级管理员角色不可删除")
}

func (s *RBACTestSuite) TestDeleteRole_CleansUpAssociations() {
	svc := &RBACService{}

	created, err := svc.CreateRole(s.ctx, CreateRoleRequest{
		Name: "cascade_delete_role",
		Type: 2,
	})
	s.Require().NoError(err)

	user := &model.User{
		TenantModel: model.TenantModel{TenantID: 0},
		Username:    "cascade_user",
		Password:    "hashed",
		Status:      1,
		Role:        "tenant_user",
	}
	s.Require().NoError(s.db.WithContext(s.ctx).Create(user).Error)

	s.db.WithContext(s.ctx).Create(&model.UserRole{UserID: user.ID, RoleID: created.ID})
	s.db.WithContext(s.ctx).Create(&model.RoleMenu{RoleID: created.ID, MenuID: 1})
	s.db.WithContext(s.ctx).Create(&model.RoleAPI{RoleID: created.ID, APIID: 1})

	err = svc.DeleteRole(s.ctx, created.ID)
	s.Assert().NoError(err)

	var userRoleCount int64
	s.db.WithContext(s.ctx).Model(&model.UserRole{}).Where("role_id = ?", created.ID).Count(&userRoleCount)
	s.Assert().Equal(int64(0), userRoleCount)

	var roleMenuCount int64
	s.db.WithContext(s.ctx).Model(&model.RoleMenu{}).Where("role_id = ?", created.ID).Count(&roleMenuCount)
	s.Assert().Equal(int64(0), roleMenuCount)

	var roleAPICount int64
	s.db.WithContext(s.ctx).Model(&model.RoleAPI{}).Where("role_id = ?", created.ID).Count(&roleAPICount)
	s.Assert().Equal(int64(0), roleAPICount)
}

func (s *RBACTestSuite) TestAssignRoleToUser_Success() {
	svc := &RBACService{}

	role, err := svc.CreateRole(s.ctx, CreateRoleRequest{
		Name: "assign_test_role",
		Type: 2,
	})
	s.Require().NoError(err)

	user := &model.User{
		TenantModel: model.TenantModel{TenantID: 0},
		Username:    "assign_role_user",
		Password:    "hashed",
		Status:      1,
		Role:        "tenant_user",
	}
	s.Require().NoError(s.db.WithContext(s.ctx).Create(user).Error)

	err = svc.AssignRoleToUser(s.ctx, user.ID, role.ID)
	s.Assert().NoError(err)
}

func (s *RBACTestSuite) TestAssignRoleToUser_AlreadyBound() {
	svc := &RBACService{}

	role, err := svc.CreateRole(s.ctx, CreateRoleRequest{
		Name: "already_bound_role",
		Type: 2,
	})
	s.Require().NoError(err)

	user := &model.User{
		TenantModel: model.TenantModel{TenantID: 0},
		Username:    "already_bound_user",
		Password:    "hashed",
		Status:      1,
		Role:        "tenant_user",
	}
	s.Require().NoError(s.db.WithContext(s.ctx).Create(user).Error)

	err = svc.AssignRoleToUser(s.ctx, user.ID, role.ID)
	s.Require().NoError(err)

	err = svc.AssignRoleToUser(s.ctx, user.ID, role.ID)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "角色已绑定")
}

func (s *RBACTestSuite) TestAssignRoleToUser_UserNotFound() {
	svc := &RBACService{}

	role, err := svc.CreateRole(s.ctx, CreateRoleRequest{
		Name: "user_notfound_role",
		Type: 2,
	})
	s.Require().NoError(err)

	err = svc.AssignRoleToUser(s.ctx, 99999, role.ID)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "用户不存在")
}

func (s *RBACTestSuite) TestAssignRoleToUser_RoleNotFound() {
	svc := &RBACService{}

	user := &model.User{
		TenantModel: model.TenantModel{TenantID: 0},
		Username:    "role_notfound_user",
		Password:    "hashed",
		Status:      1,
		Role:        "tenant_user",
	}
	s.Require().NoError(s.db.WithContext(s.ctx).Create(user).Error)

	err := svc.AssignRoleToUser(s.ctx, user.ID, 99999)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "角色不存在")
}

func (s *RBACTestSuite) TestRemoveRoleFromUser_Success() {
	svc := &RBACService{}

	role, err := svc.CreateRole(s.ctx, CreateRoleRequest{
		Name: "remove_role_test",
		Type: 2,
	})
	s.Require().NoError(err)

	user := &model.User{
		TenantModel: model.TenantModel{TenantID: 0},
		Username:    "remove_role_user",
		Password:    "hashed",
		Status:      1,
		Role:        "tenant_user",
	}
	s.Require().NoError(s.db.WithContext(s.ctx).Create(user).Error)

	err = svc.AssignRoleToUser(s.ctx, user.ID, role.ID)
	s.Require().NoError(err)

	err = svc.RemoveRoleFromUser(s.ctx, user.ID, role.ID)
	s.Assert().NoError(err)
}

func (s *RBACTestSuite) TestRemoveRoleFromUser_NotBound() {
	svc := &RBACService{}

	err := svc.RemoveRoleFromUser(s.ctx, 99999, 99999)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "角色未绑定")
}

func (s *RBACTestSuite) TestGetRoleList_Pagination() {
	svc := &RBACService{}

	for i := 0; i < 5; i++ {
		_, err := svc.CreateRole(s.ctx, CreateRoleRequest{
			Name: "list_role_" + string(rune('A'+i)),
			Type: 2,
		})
		s.Require().NoError(err)
	}

	resp, err := svc.GetRoleList(s.ctx, GetRoleListRequest{
		Page:     1,
		PageSize: 3,
	})

	s.Assert().NoError(err)
	s.Assert().NotNil(resp)
	s.Assert().True(resp.Total >= 5)
}

func TestRBACTestSuite(t *testing.T) {
	suite.Run(t, new(RBACTestSuite))
}
