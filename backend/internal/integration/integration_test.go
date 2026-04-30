package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fayhub/internal/controller"
	"fayhub/internal/middleware"
	"fayhub/internal/model"
	"fayhub/pkg/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type IntegrationTestSuite struct {
	suite.Suite
	db         *gorm.DB
	router     *gin.Engine
	adminCtx   context.Context
	adminToken string
	tenantID   uint
}

func (s *IntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err, "数据库初始化失败")

	s.db = db
	utils.SetGlobalDB(db)

	err = model.RegisterTenantIsolationCallbacks(db)
	s.Require().NoError(err, "租户隔离回调注册失败")

	s.adminCtx = utils.SkipTenantIsolation(context.Background())

	err = db.WithContext(s.adminCtx).AutoMigrate(
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
	s.Require().NoError(err, "数据库迁移失败")

	err = db.WithContext(s.adminCtx).Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_tenant_username ON users(tenant_id, username)").Error
	s.Require().NoError(err, "唯一索引创建失败")

	utils.InitJWTConfig("test-jwt-secret-for-integration", 24, "fayhub-test")

	s.setupTestData()
	s.setupRouter()
}

func (s *IntegrationTestSuite) setupTestData() {
	tenant := &model.Tenant{
		Name:   "集成测试租户",
		Domain: "test.fayhub.com",
		Status: 1,
	}
	s.Require().NoError(s.db.WithContext(s.adminCtx).Create(tenant).Error, "创建测试租户失败")
	s.tenantID = tenant.ID

	hashedPassword, err := hashPassword("admin123")
	s.Require().NoError(err)

	adminUser := &model.User{
		TenantModel: model.TenantModel{TenantID: tenant.ID},
		Username:    "admin",
		Password:    hashedPassword,
		Status:      1,
		Role:        "super_admin",
		Email:       "admin@test.com",
		RealName:    "管理员",
	}
	s.Require().NoError(s.db.WithContext(s.adminCtx).Create(adminUser).Error, "创建管理员用户失败")

	superAdminRole := &model.Role{
		TenantModel: model.TenantModel{TenantID: tenant.ID},
		Name:        "super_admin",
		Description: "超级管理员",
		Type:        1,
	}
	s.Require().NoError(s.db.WithContext(s.adminCtx).Create(superAdminRole).Error, "创建超级管理员角色失败")

	userRole := &model.UserRole{
		TenantModel: model.TenantModel{TenantID: tenant.ID},
		UserID:      adminUser.ID,
		RoleID:      superAdminRole.ID,
	}
	s.Require().NoError(s.db.WithContext(s.adminCtx).Create(userRole).Error, "分配角色失败")

	token, err := utils.GenerateToken(adminUser.ID, adminUser.Username, adminUser.Role, adminUser.TenantID)
	s.Require().NoError(err, "生成管理员Token失败")
	s.adminToken = token
}

func (s *IntegrationTestSuite) setupRouter() {
	r := gin.New()
	r.Use(gin.Recovery())

	controllerGroup := controller.ControllerGroupApp

	r.GET("/", controllerGroup.SystemController.HomePage)
	r.GET("/api/health", controllerGroup.SystemController.HealthCheck)

	systemGroup := r.Group("/api")
	systemGroup.Use(middleware.TenantMiddleware())

	authGroup := r.Group("/api/auth")
	authGroup.POST("/login", controllerGroup.AuthController.Login)
	authGroup.Use(middleware.JwtAuthMiddleware())
	authGroup.Use(middleware.TenantMiddleware())
	{
		authGroup.POST("/logout", controllerGroup.AuthController.Logout)
		authGroup.POST("/refresh", controllerGroup.AuthController.RefreshToken)
		authGroup.GET("/me", controllerGroup.AuthController.GetCurrentUser)
	}

	userGroup := r.Group("/api/users")
	userGroup.Use(middleware.JwtAuthMiddleware())
	userGroup.Use(middleware.TenantMiddleware())
	{
		userGroup.POST("", controllerGroup.UserController.CreateUser)
		userGroup.GET("", controllerGroup.UserController.GetUserList)
		userGroup.GET("/profile", controllerGroup.UserController.GetProfile)
		userGroup.PUT("/change-password", controllerGroup.UserController.ChangePassword)
		userGroup.GET("/:id", controllerGroup.UserController.GetUser)
		userGroup.PUT("/:id", controllerGroup.UserController.UpdateUser)
		userGroup.DELETE("/:id", controllerGroup.UserController.DeleteUser)
		userGroup.PUT("/:id/reset-password", middleware.SuperAdminMiddleware(), controllerGroup.UserController.ResetPassword)
	}

	tenantGroup := r.Group("/api/tenants")
	tenantGroup.Use(middleware.JwtAuthMiddleware())
	tenantGroup.Use(middleware.TenantMiddleware())
	{
		tenantGroup.POST("", controllerGroup.TenantController.CreateTenant)
		tenantGroup.GET("", controllerGroup.TenantController.GetTenantList)
		tenantGroup.GET("/:id", controllerGroup.TenantController.GetTenant)
		tenantGroup.PUT("/:id", controllerGroup.TenantController.UpdateTenant)
		tenantGroup.DELETE("/:id", controllerGroup.TenantController.DeleteTenant)
	}

	rbacGroup := r.Group("/api/rbac")
	rbacGroup.Use(middleware.JwtAuthMiddleware())
	rbacGroup.Use(middleware.TenantMiddleware())
	{
		rbacGroup.POST("/roles", middleware.SuperAdminMiddleware(), controllerGroup.RBACController.CreateRole)
		rbacGroup.GET("/roles", controllerGroup.RBACController.GetRoleList)
		rbacGroup.GET("/roles/:roleID", controllerGroup.RBACController.GetRoleByID)
		rbacGroup.PUT("/roles/:roleID", middleware.SuperAdminMiddleware(), controllerGroup.RBACController.UpdateRole)
		rbacGroup.DELETE("/roles/:roleID", middleware.SuperAdminMiddleware(), controllerGroup.RBACController.DeleteRole)
		rbacGroup.POST("/assign-role", middleware.SuperAdminMiddleware(), controllerGroup.RBACController.AssignRoleToUser)
		rbacGroup.POST("/remove-role", middleware.SuperAdminMiddleware(), controllerGroup.RBACController.RemoveRoleFromUser)
		rbacGroup.GET("/users/:userID/roles", controllerGroup.RBACController.GetUserRoles)
		rbacGroup.GET("/users/:userID/permissions", controllerGroup.RBACController.GetUserPermissions)
	}

	menuGroup := r.Group("/api/menus")
	menuGroup.Use(middleware.JwtAuthMiddleware())
	menuGroup.Use(middleware.TenantMiddleware())
	{
		menuGroup.POST("", middleware.SuperAdminMiddleware(), controllerGroup.MenuController.CreateMenu)
		menuGroup.GET("", controllerGroup.MenuController.GetMenuList)
		menuGroup.GET("/tree", controllerGroup.MenuController.GetMenuTree)
		menuGroup.GET("/:menuID", controllerGroup.MenuController.GetMenuByID)
		menuGroup.PUT("/:menuID", middleware.SuperAdminMiddleware(), controllerGroup.MenuController.UpdateMenu)
		menuGroup.DELETE("/:menuID", middleware.SuperAdminMiddleware(), controllerGroup.MenuController.DeleteMenu)
		menuGroup.POST("/assign-role", middleware.SuperAdminMiddleware(), controllerGroup.MenuController.AssignRoleMenus)
		menuGroup.GET("/role/:roleID", controllerGroup.MenuController.GetRoleMenus)
	}

	apiGroup := r.Group("/api/apis")
	apiGroup.Use(middleware.JwtAuthMiddleware())
	apiGroup.Use(middleware.TenantMiddleware())
	{
		apiGroup.POST("", middleware.SuperAdminMiddleware(), controllerGroup.APIController.CreateAPI)
		apiGroup.GET("", controllerGroup.APIController.GetAPIList)
		apiGroup.GET("/:apiID", controllerGroup.APIController.GetAPIByID)
		apiGroup.PUT("/:apiID", middleware.SuperAdminMiddleware(), controllerGroup.APIController.UpdateAPI)
		apiGroup.DELETE("/:apiID", middleware.SuperAdminMiddleware(), controllerGroup.APIController.DeleteAPI)
		apiGroup.POST("/assign-role", middleware.SuperAdminMiddleware(), controllerGroup.APIController.AssignRoleAPIs)
		apiGroup.GET("/role/:roleID", controllerGroup.APIController.GetRoleAPIs)
	}

	s.router = r
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		sqlDB, _ := s.db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

func (s *IntegrationTestSuite) makeRequest(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody *bytes.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		s.Require().NoError(err)
		reqBody = bytes.NewReader(jsonData)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, path, reqBody)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

func (s *IntegrationTestSuite) parseResponse(w *httptest.ResponseRecorder) map[string]interface{} {
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err, "响应JSON解析失败")
	return resp
}

func (s *IntegrationTestSuite) generateTokenForUser(userID uint, username, role string, tenantID uint) string {
	token, err := utils.GenerateToken(userID, username, role, tenantID)
	s.Require().NoError(err)
	return token
}

func (s *IntegrationTestSuite) createTestUser(username, password, role string, tenantID uint) *model.User {
	hashedPassword, err := hashPassword(password)
	s.Require().NoError(err)

	user := &model.User{
		TenantModel: model.TenantModel{TenantID: tenantID},
		Username:    username,
		Password:    hashedPassword,
		Status:      1,
		Role:        role,
	}
	s.Require().NoError(s.db.WithContext(s.adminCtx).Create(user).Error)
	return user
}

func (s *IntegrationTestSuite) createTestTenant(name, domain string) *model.Tenant {
	tenant := &model.Tenant{
		Name:   name,
		Domain: domain,
		Status: 1,
	}
	s.Require().NoError(s.db.WithContext(s.adminCtx).Create(tenant).Error)
	return tenant
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *IntegrationTestSuite) TestHealthCheck() {
	w := s.makeRequest("GET", "/api/health", nil, "")
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
	s.Assert().Equal("running", resp["data"].(map[string]interface{})["status"])
}

func (s *IntegrationTestSuite) TestHomePage() {
	w := s.makeRequest("GET", "/", nil, "")
	s.Assert().Equal(http.StatusOK, w.Code)
	s.Assert().Contains(w.Body.String(), "FayHub")
}

func (s *IntegrationTestSuite) TestAuth_Login_Success() {
	body := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	w := s.makeRequest("POST", "/api/auth/login", body, "")
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])

	data := resp["data"].(map[string]interface{})
	s.Assert().NotEmpty(data["token"])
	s.Assert().Equal("admin", data["username"])
}

func (s *IntegrationTestSuite) TestAuth_Login_WrongPassword() {
	body := map[string]string{
		"username": "admin",
		"password": "wrongpassword",
	}
	w := s.makeRequest("POST", "/api/auth/login", body, "")
	s.Assert().Equal(http.StatusUnauthorized, w.Code)
}

func (s *IntegrationTestSuite) TestAuth_Login_UserNotFound() {
	body := map[string]string{
		"username": "nonexistent",
		"password": "test123",
	}
	w := s.makeRequest("POST", "/api/auth/login", body, "")
	s.Assert().Equal(http.StatusUnauthorized, w.Code)
}

func (s *IntegrationTestSuite) TestAuth_GetCurrentUser() {
	w := s.makeRequest("GET", "/api/auth/me", nil, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestAuth_GetCurrentUser_Unauthorized() {
	w := s.makeRequest("GET", "/api/auth/me", nil, "")
	s.Assert().Equal(http.StatusUnauthorized, w.Code)
}

func (s *IntegrationTestSuite) TestAuth_RefreshToken() {
	refreshUser := s.createTestUser("refreshuser", "password123", "tenant_user", s.tenantID)
	refreshToken := s.generateTokenForUser(refreshUser.ID, refreshUser.Username, refreshUser.Role, refreshUser.TenantID)

	body := map[string]string{
		"token": refreshToken,
	}
	w := s.makeRequest("POST", "/api/auth/refresh", body, refreshToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestAuth_Logout() {
	logoutUser := s.createTestUser("logoutuser", "password123", "tenant_user", s.tenantID)
	logoutToken := s.generateTokenForUser(logoutUser.ID, logoutUser.Username, logoutUser.Role, logoutUser.TenantID)

	w := s.makeRequest("POST", "/api/auth/logout", nil, logoutToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestUser_Create_Success() {
	body := map[string]string{
		"username":  "newuser",
		"password":  "Password123!",
		"email":     "newuser@test.com",
		"real_name": "新用户",
	}
	w := s.makeRequest("POST", "/api/users", body, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])

	data := resp["data"].(map[string]interface{})
	s.Assert().Equal("newuser", data["username"])
	s.Assert().Equal("newuser@test.com", data["email"])
}

func (s *IntegrationTestSuite) TestUser_Create_MissingFields() {
	body := map[string]string{
		"username": "missingpw",
	}
	w := s.makeRequest("POST", "/api/users", body, s.adminToken)
	s.Assert().Equal(http.StatusBadRequest, w.Code)
}

func (s *IntegrationTestSuite) TestUser_GetList() {
	w := s.makeRequest("GET", "/api/users?page=1&page_size=10", nil, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])

	data := resp["data"].(map[string]interface{})
	s.Assert().NotNil(data["list"])
}

func (s *IntegrationTestSuite) TestUser_GetByID() {
	user := s.createTestUser("getuser", "password123", "tenant_user", s.tenantID)

	w := s.makeRequest("GET", "/api/users/"+uintToString(user.ID), nil, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestUser_Update() {
	user := s.createTestUser("updateuser", "password123", "tenant_user", s.tenantID)

	body := map[string]string{
		"email":     "updated@test.com",
		"real_name": "更新后",
	}
	w := s.makeRequest("PUT", "/api/users/"+uintToString(user.ID), body, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestUser_Delete() {
	user := s.createTestUser("deleteuser", "password123", "tenant_user", s.tenantID)

	w := s.makeRequest("DELETE", "/api/users/"+uintToString(user.ID), nil, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestUser_ChangePassword() {
	user := s.createTestUser("changepwuser", "Oldpass123", "tenant_user", s.tenantID)
	token := s.generateTokenForUser(user.ID, user.Username, user.Role, user.TenantID)

	body := map[string]string{
		"old_password": "Oldpass123",
		"new_password": "Newpass1234!",
	}
	w := s.makeRequest("PUT", "/api/users/change-password", body, token)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestUser_ChangePassword_WrongOldPassword() {
	user := s.createTestUser("wrongpwuser", "Oldpass123", "tenant_user", s.tenantID)
	token := s.generateTokenForUser(user.ID, user.Username, user.Role, user.TenantID)

	body := map[string]string{
		"old_password": "wrongold",
		"new_password": "Newpass1234!",
	}
	w := s.makeRequest("PUT", "/api/users/change-password", body, token)
	s.Assert().Equal(http.StatusBadRequest, w.Code)
}

func (s *IntegrationTestSuite) TestUser_UnauthorizedAccess() {
	w := s.makeRequest("GET", "/api/users", nil, "")
	s.Assert().Equal(http.StatusUnauthorized, w.Code)
}

func (s *IntegrationTestSuite) TestTenant_Create() {
	body := map[string]string{
		"name":        "新租户",
		"domain":      "new.fayhub.com",
		"description": "集成测试创建",
	}
	w := s.makeRequest("POST", "/api/tenants", body, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestTenant_GetList() {
	w := s.makeRequest("GET", "/api/tenants?page=1&page_size=10", nil, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestTenant_GetByID() {
	tenant := s.createTestTenant("查询租户", "query.fayhub.com")

	w := s.makeRequest("GET", "/api/tenants/"+uintToString(tenant.ID), nil, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestTenant_Update() {
	tenant := s.createTestTenant("更新前租户", "before.fayhub.com")

	body := map[string]string{
		"name":        "更新后租户",
		"description": "已更新",
	}
	w := s.makeRequest("PUT", "/api/tenants/"+uintToString(tenant.ID), body, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestTenant_Delete() {
	tenant := s.createTestTenant("删除租户", "delete.fayhub.com")

	w := s.makeRequest("DELETE", "/api/tenants/"+uintToString(tenant.ID), nil, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestRBAC_CreateRole() {
	body := map[string]interface{}{
		"name":        "测试角色",
		"description": "集成测试角色",
		"type":        2,
	}
	w := s.makeRequest("POST", "/api/rbac/roles", body, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestRBAC_GetRoleList() {
	w := s.makeRequest("GET", "/api/rbac/roles?page=1&page_size=10", nil, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestRBAC_CreateRole_DuplicateName() {
	body := map[string]interface{}{
		"name":        "重复角色",
		"description": "第一次创建",
		"type":        2,
	}
	w := s.makeRequest("POST", "/api/rbac/roles", body, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	body2 := map[string]interface{}{
		"name":        "重复角色",
		"description": "第二次创建",
		"type":        2,
	}
	w2 := s.makeRequest("POST", "/api/rbac/roles", body2, s.adminToken)
	s.Assert().Equal(http.StatusBadRequest, w2.Code)
}

func (s *IntegrationTestSuite) TestRBAC_AssignRoleToUser() {
	user := s.createTestUser("assignroleuser", "password123", "tenant_user", s.tenantID)

	role := &model.Role{
		TenantModel: model.TenantModel{TenantID: s.tenantID},
		Name:        "可分配角色",
		Description: "用于分配测试",
		Type:        2,
	}
	s.Require().NoError(s.db.WithContext(s.adminCtx).Create(role).Error)

	body := map[string]interface{}{
		"user_id": user.ID,
		"role_id": role.ID,
	}
	w := s.makeRequest("POST", "/api/rbac/assign-role", body, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestRBAC_GetUserRoles() {
	user := s.createTestUser("rolesuser", "password123", "tenant_user", s.tenantID)

	w := s.makeRequest("GET", "/api/rbac/users/"+uintToString(user.ID)+"/roles", nil, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestRBAC_GetUserPermissions() {
	user := s.createTestUser("permuser", "password123", "tenant_user", s.tenantID)

	w := s.makeRequest("GET", "/api/rbac/users/"+uintToString(user.ID)+"/permissions", nil, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestMenu_Create() {
	body := map[string]interface{}{
		"title":      "测试菜单",
		"path":       "/test",
		"component":  "TestView",
		"icon":       "test",
		"sort":       1,
		"parent_id":  0,
		"type":       2,
		"status":     1,
		"permission": "test:view",
	}
	w := s.makeRequest("POST", "/api/menus", body, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestMenu_GetList() {
	w := s.makeRequest("GET", "/api/menus?page=1&page_size=10", nil, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestMenu_GetTree() {
	w := s.makeRequest("GET", "/api/menus/tree", nil, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestAPI_Create() {
	body := map[string]interface{}{
		"path":        "/api/test/integration",
		"method":      "GET",
		"description": "集成测试API",
		"group":       "测试组",
		"status":      1,
	}
	w := s.makeRequest("POST", "/api/apis", body, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestAPI_GetList() {
	w := s.makeRequest("GET", "/api/apis?page=1&page_size=10", nil, s.adminToken)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func (s *IntegrationTestSuite) TestFullWorkflow_LoginToCRUD() {
	loginBody := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	w := s.makeRequest("POST", "/api/auth/login", loginBody, "")
	s.Assert().Equal(http.StatusOK, w.Code)

	loginResp := s.parseResponse(w)
	s.Assert().Equal(float64(200), loginResp["code"])
	loginData := loginResp["data"].(map[string]interface{})
	token := loginData["token"].(string)
	s.Assert().NotEmpty(token)

	createBody := map[string]string{
		"username":  "workflow_user",
		"password":  "Password123!",
		"email":     "workflow@test.com",
		"real_name": "流程测试用户",
	}
	w = s.makeRequest("POST", "/api/users", createBody, token)
	s.Assert().Equal(http.StatusOK, w.Code)

	createResp := s.parseResponse(w)
	s.Assert().Equal(float64(200), createResp["code"])
	userData := createResp["data"].(map[string]interface{})
	userID := userData["id"]

	w = s.makeRequest("GET", "/api/users/"+fmtUint(userID), nil, token)
	s.Assert().Equal(http.StatusOK, w.Code)

	updateBody := map[string]string{
		"email":     "updated@test.com",
		"real_name": "流程更新用户",
	}
	w = s.makeRequest("PUT", "/api/users/"+fmtUint(userID), updateBody, token)
	s.Assert().Equal(http.StatusOK, w.Code)

	w = s.makeRequest("DELETE", "/api/users/"+fmtUint(userID), nil, token)
	s.Assert().Equal(http.StatusOK, w.Code)
}

func (s *IntegrationTestSuite) TestMultiTenantIsolation_UserData() {
	tenantA := s.createTestTenant("隔离租户A", "iso-a.fayhub.com")
	tenantB := s.createTestTenant("隔离租户B", "iso-b.fayhub.com")

	userA := s.createTestUser("iso_user_a", "password123", "tenant_user", tenantA.ID)
	_ = s.createTestUser("iso_user_b", "password123", "tenant_user", tenantB.ID)

	tokenA := s.generateTokenForUser(userA.ID, userA.Username, userA.Role, tenantA.ID)

	w := s.makeRequest("GET", "/api/users/"+uintToString(userA.ID), nil, tokenA)
	s.Assert().Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Assert().Equal(float64(200), resp["code"])
}

func uintToString(n uint) string {
	return fmtUint(float64(n))
}

func fmtUint(v interface{}) string {
	switch val := v.(type) {
	case float64:
		return formatUint(uint(val))
	case uint:
		return formatUint(val)
	default:
		return ""
	}
}

func formatUint(n uint) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return formatUint(n/10) + string(rune('0'+n%10))
}

func TestIntegrationTestSuite(t *testing.T) {
	_ = time.Now()
	suite.Run(t, new(IntegrationTestSuite))
}
