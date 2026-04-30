package service

import (
	"context"
	"fayhub/internal/model"
	"fayhub/pkg/config"
	"fayhub/pkg/utils"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ServiceTestSuite struct {
	suite.Suite
	db  *gorm.DB
	ctx context.Context
}

func (s *ServiceTestSuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err)

	s.db = db
	utils.SetGlobalDB(db)

	err = model.RegisterTenantIsolationCallbacks(db)
	s.Require().NoError(err)

	ctx := utils.SkipTenantIsolation(context.Background())
	s.ctx = ctx

	err = db.WithContext(ctx).AutoMigrate(
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

	err = db.WithContext(ctx).Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_tenant_username ON users(tenant_id, username)").Error
	s.Require().NoError(err)

	config.GlobalConfig = &config.Config{
		Security: config.SecurityConfig{
			MaxLoginAttempts: 5,
			LockDurationMin:  15,
		},
	}
}

func (s *ServiceTestSuite) TearDownSuite() {
	if s.db != nil {
		sqlDB, _ := s.db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

func (s *ServiceTestSuite) TestAuthService_Login_Success() {
	svc := &AuthService{}

	hashedPassword, err := hashPassword("Test1234!")
	s.Require().NoError(err)

	user := &model.User{
		TenantModel: model.TenantModel{TenantID: 0},
		Username:    "logintest",
		Password:    hashedPassword,
		Status:      1,
		Role:        "super_admin",
	}
	s.Require().NoError(s.db.WithContext(s.ctx).Create(user).Error)

	resp, err := svc.Login(s.ctx, LoginRequest{
		Username: "logintest",
		Password: "Test1234!",
	})

	s.Assert().NoError(err)
	s.Assert().NotNil(resp)
	s.Assert().NotEmpty(resp.Token)
	s.Assert().Equal("logintest", resp.Username)
}

func (s *ServiceTestSuite) TestAuthService_Login_WrongPassword() {
	svc := &AuthService{}

	hashedPassword, err := hashPassword("Correct1!")
	s.Require().NoError(err)

	user := &model.User{
		TenantModel: model.TenantModel{TenantID: 0},
		Username:    "wrongpwtest",
		Password:    hashedPassword,
		Status:      1,
		Role:        "super_admin",
	}
	s.Require().NoError(s.db.WithContext(s.ctx).Create(user).Error)

	resp, err := svc.Login(s.ctx, LoginRequest{
		Username: "wrongpwtest",
		Password: "wrongpassword",
	})

	s.Assert().Error(err)
	s.Assert().Nil(resp)
	s.Assert().Contains(err.Error(), "用户名或密码错误")
}

func (s *ServiceTestSuite) TestAuthService_Login_UserNotFound() {
	svc := &AuthService{}

	resp, err := svc.Login(s.ctx, LoginRequest{
		Username: "nonexistent",
		Password: "Test1234!",
	})

	s.Assert().Error(err)
	s.Assert().Nil(resp)
	s.Assert().Equal("用户名或密码错误", err.Error())
}

func (s *ServiceTestSuite) TestAuthService_Login_DisabledUser() {
	svc := &AuthService{}

	hashedPassword, err := hashPassword("Test1234!")
	s.Require().NoError(err)

	user := &model.User{
		TenantModel: model.TenantModel{TenantID: 0},
		Username:    "disableduser",
		Password:    hashedPassword,
		Status:      1,
		Role:        "tenant_user",
	}
	s.Require().NoError(s.db.WithContext(s.ctx).Create(user).Error)

	s.Require().NoError(s.db.WithContext(s.ctx).Model(user).Update("status", 0).Error)

	resp, err := svc.Login(s.ctx, LoginRequest{
		Username: "disableduser",
		Password: "Test1234!",
	})

	s.Assert().Error(err)
	s.Assert().Nil(resp)
	s.Assert().Equal("用户已被禁用", err.Error())
}

func (s *ServiceTestSuite) TestAuthService_Login_AccountLockAfterMaxFailures() {
	svc := &AuthService{}

	hashedPassword, err := hashPassword("Correct1!")
	s.Require().NoError(err)

	user := &model.User{
		TenantModel:    model.TenantModel{TenantID: 0},
		Username:       "locktest",
		Password:       hashedPassword,
		Status:         1,
		Role:           "super_admin",
		LoginFailCount: 0,
		LockedUntil:    0,
	}
	s.Require().NoError(s.db.WithContext(s.ctx).Create(user).Error)

	for i := 0; i < 4; i++ {
		_, err := svc.Login(s.ctx, LoginRequest{
			Username: "locktest",
			Password: "wrongpassword",
		})
		s.Assert().Error(err)
		s.Assert().Contains(err.Error(), "用户名或密码错误")
	}

	_, err = svc.Login(s.ctx, LoginRequest{
		Username: "locktest",
		Password: "wrongpassword",
	})
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "账户已锁定")

	_, err = svc.Login(s.ctx, LoginRequest{
		Username: "locktest",
		Password: "Correct1!",
	})
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "账户已锁定")
}

func (s *ServiceTestSuite) TestAuthService_Login_SuccessResetsFailCount() {
	svc := &AuthService{}

	hashedPassword, err := hashPassword("Correct1!")
	s.Require().NoError(err)

	user := &model.User{
		TenantModel:    model.TenantModel{TenantID: 0},
		Username:       "resetfailtest",
		Password:       hashedPassword,
		Status:         1,
		Role:           "super_admin",
		LoginFailCount: 3,
		LockedUntil:    0,
	}
	s.Require().NoError(s.db.WithContext(s.ctx).Create(user).Error)

	resp, err := svc.Login(s.ctx, LoginRequest{
		Username: "resetfailtest",
		Password: "Correct1!",
	})
	s.Assert().NoError(err)
	s.Assert().NotNil(resp)

	var updated model.User
	s.db.WithContext(s.ctx).Where("username = ?", "resetfailtest").First(&updated)
	s.Assert().Equal(0, updated.LoginFailCount)
	s.Assert().Equal(int64(0), updated.LockedUntil)
}

func (s *ServiceTestSuite) TestAuthService_Login_LockExpiredAllowsLogin() {
	svc := &AuthService{}

	hashedPassword, err := hashPassword("Correct1!")
	s.Require().NoError(err)

	pastLock := time.Now().Unix() - 60
	user := &model.User{
		TenantModel:    model.TenantModel{TenantID: 0},
		Username:       "expiredlocktest",
		Password:       hashedPassword,
		Status:         1,
		Role:           "super_admin",
		LoginFailCount: 5,
		LockedUntil:    pastLock,
	}
	s.Require().NoError(s.db.WithContext(s.ctx).Create(user).Error)

	resp, err := svc.Login(s.ctx, LoginRequest{
		Username: "expiredlocktest",
		Password: "Correct1!",
	})
	s.Assert().NoError(err)
	s.Assert().NotNil(resp)
}

func (s *ServiceTestSuite) TestAuthService_Login_LockNotExpiredBlocksLogin() {
	svc := &AuthService{}

	hashedPassword, err := hashPassword("Correct1!")
	s.Require().NoError(err)

	futureLock := time.Now().Unix() + 900
	user := &model.User{
		TenantModel:    model.TenantModel{TenantID: 0},
		Username:       "activelocktest",
		Password:       hashedPassword,
		Status:         1,
		Role:           "super_admin",
		LoginFailCount: 0,
		LockedUntil:    futureLock,
	}
	s.Require().NoError(s.db.WithContext(s.ctx).Create(user).Error)

	_, err = svc.Login(s.ctx, LoginRequest{
		Username: "activelocktest",
		Password: "Correct1!",
	})
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "账户已锁定")
}

func (s *ServiceTestSuite) TestUserService_Create_Success() {
	svc := &UserService{}

	tenantDB := utils.GetDB(s.ctx)

	user, err := svc.Create(s.ctx, CreateUserRequest{
		Username: "newuser",
		Password: "Password123!",
		Email:    "newuser@test.com",
		RealName: "新用户",
	})

	s.Assert().NoError(err)
	s.Assert().NotNil(user)
	s.Assert().Equal("newuser", user.Username)
	s.Assert().Equal("newuser@test.com", user.Email)

	var found model.User
	tenantDB.Where("username = ?", "newuser").First(&found)
	s.Assert().Equal("新用户", found.RealName)
}

func (s *ServiceTestSuite) TestUserService_Create_DuplicateUsername() {
	svc := &UserService{}

	_, err := svc.Create(s.ctx, CreateUserRequest{
		Username: "dupuser",
		Password: "Password123!",
	})
	s.Assert().NoError(err)

	user2, err := svc.Create(s.ctx, CreateUserRequest{
		Username: "dupuser",
		Password: "Password456!",
	})

	s.Assert().Error(err)
	s.Assert().Nil(user2)
	s.Assert().Equal("用户名已存在", err.Error())
}

func (s *ServiceTestSuite) TestUserService_GetByID_Success() {
	svc := &UserService{}

	created, err := svc.Create(s.ctx, CreateUserRequest{
		Username: "getbyiduser",
		Password: "Password123!",
	})
	s.Require().NoError(err)

	user, err := svc.GetByID(s.ctx, created.ID)

	s.Assert().NoError(err)
	s.Assert().NotNil(user)
	s.Assert().Equal("getbyiduser", user.Username)
}

func (s *ServiceTestSuite) TestUserService_GetByID_NotFound() {
	svc := &UserService{}

	user, err := svc.GetByID(s.ctx, 99999)

	s.Assert().Error(err)
	s.Assert().Nil(user)
	s.Assert().Equal("用户不存在", err.Error())
}

func (s *ServiceTestSuite) TestUserService_Update_Success() {
	svc := &UserService{}

	created, err := svc.Create(s.ctx, CreateUserRequest{
		Username: "updateuser",
		Password: "Password123!",
		Email:    "old@test.com",
	})
	s.Require().NoError(err)

	updated, err := svc.Update(s.ctx, created.ID, UpdateUserRequest{
		Email:    "new@test.com",
		RealName: "更新后",
	})

	s.Assert().NoError(err)
	s.Assert().NotNil(updated)
	s.Assert().Equal("new@test.com", updated.Email)
	s.Assert().Equal("更新后", updated.RealName)
}

func (s *ServiceTestSuite) TestUserService_Delete_Success() {
	svc := &UserService{}

	created, err := svc.Create(s.ctx, CreateUserRequest{
		Username: "deleteuser",
		Password: "Password123!",
	})
	s.Require().NoError(err)

	err = svc.Delete(s.ctx, created.ID)
	s.Assert().NoError(err)

	_, err = svc.GetByID(s.ctx, created.ID)
	s.Assert().Error(err)
	s.Assert().Equal("用户不存在", err.Error())
}

func (s *ServiceTestSuite) TestUserService_ChangePassword_Success() {
	svc := &UserService{}

	created, err := svc.Create(s.ctx, CreateUserRequest{
		Username: "changepwuser",
		Password: "Oldpass123!",
	})
	s.Require().NoError(err)

	err = svc.ChangePassword(s.ctx, created.ID, "Oldpass123!", "Newpass123!")
	s.Assert().NoError(err)
}

func (s *ServiceTestSuite) TestUserService_ChangePassword_WrongOldPassword() {
	svc := &UserService{}

	created, err := svc.Create(s.ctx, CreateUserRequest{
		Username: "changepwuser2",
		Password: "Oldpass123!",
	})
	s.Require().NoError(err)

	err = svc.ChangePassword(s.ctx, created.ID, "wrongold", "Newpass123!")
	s.Assert().Error(err)
	s.Assert().Equal("原密码错误", err.Error())
}

func (s *ServiceTestSuite) TestUserService_ResetPassword_Success() {
	svc := &UserService{}

	created, err := svc.Create(s.ctx, CreateUserRequest{
		Username: "resetpwuser",
		Password: "Oldpass123!",
	})
	s.Require().NoError(err)

	err = svc.ResetPassword(s.ctx, created.ID, "Resetpass1!")
	s.Assert().NoError(err)
}

func (s *ServiceTestSuite) TestUserService_GetList_Pagination() {
	svc := &UserService{}

	for i := 0; i < 5; i++ {
		_, err := svc.Create(s.ctx, CreateUserRequest{
			Username: "listuser" + string(rune('A'+i)),
			Password: "Password123!",
		})
		s.Require().NoError(err)
	}

	resp, err := svc.GetList(s.ctx, UserListRequest{
		Page:     1,
		PageSize: 3,
	})

	s.Assert().NoError(err)
	s.Assert().NotNil(resp)
	s.Assert().Equal(3, len(resp.List))
	s.Assert().True(resp.Total >= 5)
}

func (s *ServiceTestSuite) TestTenantService_Create_Success() {
	svc := &TenantService{}

	tenant, err := svc.Create(s.ctx, CreateTenantRequest{
		Name:        "测试租户A",
		Domain:      "a.fayhub.com",
		Description: "单元测试创建的租户",
	})

	s.Assert().NoError(err)
	s.Assert().NotNil(tenant)
	s.Assert().Equal("测试租户A", tenant.Name)
	s.Assert().Equal("a.fayhub.com", tenant.Domain)
	s.Assert().Equal(1, tenant.Status)
}

func (s *ServiceTestSuite) TestTenantService_GetByID_Success() {
	svc := &TenantService{}

	created, err := svc.Create(s.ctx, CreateTenantRequest{
		Name:   "查询租户",
		Domain: "query.fayhub.com",
	})
	s.Require().NoError(err)

	tenant, err := svc.GetByID(s.ctx, created.ID)

	s.Assert().NoError(err)
	s.Assert().NotNil(tenant)
	s.Assert().Equal("查询租户", tenant.Name)
}

func (s *ServiceTestSuite) TestTenantService_Update_Success() {
	svc := &TenantService{}

	created, err := svc.Create(s.ctx, CreateTenantRequest{
		Name:   "更新前租户",
		Domain: "before.fayhub.com",
	})
	s.Require().NoError(err)

	updated, err := svc.Update(s.ctx, created.ID, UpdateTenantRequest{
		Name:        "更新后租户",
		Description: "已更新",
	})

	s.Assert().NoError(err)
	s.Assert().NotNil(updated)
	s.Assert().Equal("更新后租户", updated.Name)
	s.Assert().Equal("已更新", updated.Description)
}

func (s *ServiceTestSuite) TestTenantService_Delete_Success() {
	svc := &TenantService{}

	created, err := svc.Create(s.ctx, CreateTenantRequest{
		Name:   "删除租户",
		Domain: "delete.fayhub.com",
	})
	s.Require().NoError(err)

	err = svc.Delete(s.ctx, created.ID)
	s.Assert().NoError(err)

	_, err = svc.GetByID(s.ctx, created.ID)
	s.Assert().Error(err)
}

func (s *ServiceTestSuite) TestTenantService_GetList_Pagination() {
	svc := &TenantService{}

	for i := 0; i < 3; i++ {
		_, err := svc.Create(s.ctx, CreateTenantRequest{
			Name:   "列表租户" + string(rune('X'+i)),
			Domain: "list" + string(rune('x'+i)) + ".fayhub.com",
		})
		s.Require().NoError(err)
	}

	resp, err := svc.GetList(s.ctx, TenantListRequest{
		Page:     1,
		PageSize: 2,
	})

	s.Assert().NoError(err)
	s.Assert().NotNil(resp)
	s.Assert().True(resp.Total >= 3)
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func init() {
	config.GlobalConfig = &config.Config{
		Security: config.SecurityConfig{
			MaxLoginAttempts: 5,
			LockDurationMin:  15,
		},
	}
	_ = strings.Contains("", "")
	_ = time.Now()
}
