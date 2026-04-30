package service

import (
	"context"
	"fayhub/internal/model"
	"fayhub/pkg/utils"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TenantIsolationTestSuite struct {
	suite.Suite
	db       *gorm.DB
	tenantA  *model.Tenant
	tenantB  *model.Tenant
	ctxA     context.Context
	ctxB     context.Context
	adminCtx context.Context
}

func (s *TenantIsolationTestSuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err)

	s.db = db
	utils.SetGlobalDB(db)

	err = model.RegisterTenantIsolationCallbacks(db)
	s.Require().NoError(err)

	adminCtx := utils.SkipTenantIsolation(context.Background())
	s.adminCtx = adminCtx

	err = db.WithContext(adminCtx).AutoMigrate(
		&model.Tenant{},
		&model.User{},
		&model.Role{},
		&model.Menu{},
		&model.API{},
		&model.RoleMenu{},
		&model.RoleAPI{},
		&model.UserRole{},
		&model.TenantRole{},
		&model.TokenBlacklistEntry{},
	)
	s.Require().NoError(err)

	err = db.WithContext(adminCtx).Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_tenant_username ON users(tenant_id, username)").Error
	s.Require().NoError(err)

	s.tenantA = &model.Tenant{Name: "租户A", Domain: "a.fayhub.com", Status: 1}
	s.tenantB = &model.Tenant{Name: "租户B", Domain: "b.fayhub.com", Status: 1}
	s.Require().NoError(db.WithContext(adminCtx).Create(s.tenantA).Error)
	s.Require().NoError(db.WithContext(adminCtx).Create(s.tenantB).Error)

	s.ctxA = utils.WithTenantID(context.Background(), s.tenantA.ID)
	s.ctxB = utils.WithTenantID(context.Background(), s.tenantB.ID)
}

func (s *TenantIsolationTestSuite) TearDownSuite() {
	if s.db != nil {
		sqlDB, _ := s.db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

func (s *TenantIsolationTestSuite) TestUserIsolation_CreateInTenantA_NotVisibleInTenantB() {
	svc := &UserService{}

	userA, err := svc.Create(s.ctxA, CreateUserRequest{
		Username: "tenant_a_user",
		Password: "Password123!",
		Email:    "a@tenant.com",
	})
	s.Assert().NoError(err)
	s.Assert().NotNil(userA)

	tenantDB_B := utils.GetDB(s.ctxB)
	var found model.User
	err = tenantDB_B.Where("username = ?", "tenant_a_user").First(&found).Error
	s.Assert().Error(err, "租户B不应该能看到租户A的用户")
}

func (s *TenantIsolationTestSuite) TestUserIsolation_CreateSameUsernameInDifferentTenants() {
	svc := &UserService{}

	userA, err := svc.Create(s.ctxA, CreateUserRequest{
		Username: "shared_name",
		Password: "Password123!",
	})
	s.Assert().NoError(err)
	s.Assert().NotNil(userA)

	userB, err := svc.Create(s.ctxB, CreateUserRequest{
		Username: "shared_name",
		Password: "Password456!",
	})
	s.Assert().NoError(err, "不同租户应该可以创建同名用户")
	s.Assert().NotNil(userB)

	s.Assert().NotEqual(userA.ID, userB.ID, "不同租户的同名用户应该是不同记录")
}

func (s *TenantIsolationTestSuite) TestUserIsolation_GetByID_CrossTenantBlocked() {
	svc := &UserService{}

	userA, err := svc.Create(s.ctxA, CreateUserRequest{
		Username: "cross_tenant_test",
		Password: "Password123!",
	})
	s.Require().NoError(err)

	_, err = svc.GetByID(s.ctxB, userA.ID)
	s.Assert().Error(err, "租户B不应该能通过ID查询租户A的用户")
	s.Assert().Equal("用户不存在", err.Error())
}

func (s *TenantIsolationTestSuite) TestUserIsolation_Update_CrossTenantBlocked() {
	svc := &UserService{}

	userA, err := svc.Create(s.ctxA, CreateUserRequest{
		Username: "update_cross_test",
		Password: "Password123!",
		Email:    "original@test.com",
	})
	s.Require().NoError(err)

	_, err = svc.Update(s.ctxB, userA.ID, UpdateUserRequest{
		Email: "hacked@test.com",
	})
	s.Assert().Error(err, "租户B不应该能修改租户A的用户")
}

func (s *TenantIsolationTestSuite) TestUserIsolation_Delete_CrossTenantBlocked() {
	svc := &UserService{}

	userA, err := svc.Create(s.ctxA, CreateUserRequest{
		Username: "delete_cross_test",
		Password: "Password123!",
	})
	s.Require().NoError(err)

	err = svc.Delete(s.ctxB, userA.ID)
	s.Assert().Error(err, "租户B不应该能删除租户A的用户")
}

func (s *TenantIsolationTestSuite) TestUserIsolation_ListOnlyShowsOwnTenant() {
	svc := &UserService{}

	_, err := svc.Create(s.ctxA, CreateUserRequest{
		Username: "a_list_user",
		Password: "Password123!",
	})
	s.Require().NoError(err)

	_, err = svc.Create(s.ctxB, CreateUserRequest{
		Username: "b_list_user",
		Password: "Password123!",
	})
	s.Require().NoError(err)

	respA, err := svc.GetList(s.ctxA, UserListRequest{Page: 1, PageSize: 100})
	s.Assert().NoError(err)

	for _, user := range respA.List {
		s.Assert().Equal(s.tenantA.ID, user.TenantID, "租户A的用户列表不应该包含其他租户的用户")
	}

	respB, err := svc.GetList(s.ctxB, UserListRequest{Page: 1, PageSize: 100})
	s.Assert().NoError(err)

	for _, user := range respB.List {
		s.Assert().Equal(s.tenantB.ID, user.TenantID, "租户B的用户列表不应该包含其他租户的用户")
	}
}

func (s *TenantIsolationTestSuite) TestUserIsolation_ChangePassword_CrossTenantBlocked() {
	svc := &UserService{}

	userA, err := svc.Create(s.ctxA, CreateUserRequest{
		Username: "pw_cross_test",
		Password: "Oldpass123!",
	})
	s.Require().NoError(err)

	err = svc.ChangePassword(s.ctxB, userA.ID, "Oldpass123!", "Hackedpass1!")
	s.Assert().Error(err, "租户B不应该能修改租户A用户的密码")
}

func (s *TenantIsolationTestSuite) TestUserIsolation_ResetPassword_CrossTenantBlocked() {
	svc := &UserService{}

	userA, err := svc.Create(s.ctxA, CreateUserRequest{
		Username: "reset_cross_test",
		Password: "Oldpass123!",
	})
	s.Require().NoError(err)

	err = svc.ResetPassword(s.ctxB, userA.ID, "Hackedpass1!")
	s.Assert().Error(err, "租户B不应该能重置租户A用户的密码")
}

func (s *TenantIsolationTestSuite) TestRoleIsolation_CreateInTenantA_NotVisibleInTenantB() {
	svc := &RBACService{}

	roleA, err := svc.CreateRole(s.ctxA, CreateRoleRequest{
		Name:        "tenant_a_role",
		Description: "租户A专属角色",
		Type:        2,
	})
	s.Assert().NoError(err)
	s.Assert().NotNil(roleA)

	tenantDB_B := utils.GetDB(s.ctxB)
	var found model.Role
	err = tenantDB_B.Where("name = ?", "tenant_a_role").First(&found).Error
	s.Assert().Error(err, "租户B不应该能看到租户A的角色")
}

func (s *TenantIsolationTestSuite) TestRoleIsolation_GetByID_CrossTenantBlocked() {
	svc := &RBACService{}

	roleA, err := svc.CreateRole(s.ctxA, CreateRoleRequest{
		Name:        "cross_role_test",
		Description: "跨租户角色测试",
		Type:        2,
	})
	s.Require().NoError(err)

	_, err = svc.GetRoleByID(s.ctxB, roleA.ID)
	s.Assert().Error(err, "租户B不应该能查询租户A的角色")
}

func (s *TenantIsolationTestSuite) TestRoleIsolation_Update_CrossTenantBlocked() {
	svc := &RBACService{}

	roleA, err := svc.CreateRole(s.ctxA, CreateRoleRequest{
		Name:        "update_role_cross",
		Description: "原始描述",
		Type:        2,
	})
	s.Require().NoError(err)

	_, err = svc.UpdateRole(s.ctxB, roleA.ID, UpdateRoleRequest{
		Description: "被篡改",
	})
	s.Assert().Error(err, "租户B不应该能修改租户A的角色")
}

func (s *TenantIsolationTestSuite) TestRoleIsolation_Delete_CrossTenantBlocked() {
	svc := &RBACService{}

	roleA, err := svc.CreateRole(s.ctxA, CreateRoleRequest{
		Name:        "delete_role_cross",
		Description: "待删除角色",
		Type:        2,
	})
	s.Require().NoError(err)

	err = svc.DeleteRole(s.ctxB, roleA.ID)
	s.Assert().Error(err, "租户B不应该能删除租户A的角色")
}

func (s *TenantIsolationTestSuite) TestSkipTenantIsolation_AdminCanSeeAll() {
	svc := &UserService{}

	_, err := svc.Create(s.ctxA, CreateUserRequest{
		Username: "admin_visible_a",
		Password: "Password123!",
	})
	s.Require().NoError(err)

	_, err = svc.Create(s.ctxB, CreateUserRequest{
		Username: "admin_visible_b",
		Password: "Password123!",
	})
	s.Require().NoError(err)

	adminDB := utils.GetDB(s.adminCtx)
	var count int64
	adminDB.Model(&model.User{}).Count(&count)
	s.Assert().True(count >= 2, "管理员跳过租户隔离后应该能看到所有用户")
}

func TestTenantIsolationTestSuite(t *testing.T) {
	suite.Run(t, new(TenantIsolationTestSuite))
}
