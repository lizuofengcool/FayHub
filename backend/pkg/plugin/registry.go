package plugin

import (
	"fmt"
	"log"
	"sync"
)

type Registry struct {
	mu     sync.RWMutex
	routes map[string][]RouteRegistration
	apis   map[string][]APIRegistration
	menus  map[string][]MenuRegistration
}

func NewRegistry() *Registry {
	return &Registry{
		routes: make(map[string][]RouteRegistration),
		apis:   make(map[string][]APIRegistration),
		menus:  make(map[string][]MenuRegistration),
	}
}

func (r *Registry) RegisterRoutes(tenantID int64, pluginID string, routes []RouteRegistration) {
	if len(routes) == 0 {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := pluginKey(tenantID, pluginID)
	r.routes[key] = routes

	for _, route := range routes {
		log.Printf("[Registry] 注册路由: tenant=%d, plugin=%s, %s %s -> %s",
			tenantID, pluginID, route.Method, route.Path, route.Handler)
	}
}

func (r *Registry) UnregisterRoutes(tenantID int64, pluginID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := pluginKey(tenantID, pluginID)
	if routes, exists := r.routes[key]; exists {
		for _, route := range routes {
			log.Printf("[Registry] 注销路由: tenant=%d, plugin=%s, %s %s",
				tenantID, pluginID, route.Method, route.Path)
		}
		delete(r.routes, key)
	}
}

func (r *Registry) RegisterAPIs(tenantID int64, pluginID string, apis []APIRegistration) {
	if len(apis) == 0 {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := pluginKey(tenantID, pluginID)
	r.apis[key] = apis

	for _, api := range apis {
		log.Printf("[Registry] 注册API: tenant=%d, plugin=%s, %s %s (group=%s)",
			tenantID, pluginID, api.Method, api.Path, api.Group)
	}
}

func (r *Registry) UnregisterAPIs(tenantID int64, pluginID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := pluginKey(tenantID, pluginID)
	if apis, exists := r.apis[key]; exists {
		for _, api := range apis {
			log.Printf("[Registry] 注销API: tenant=%d, plugin=%s, %s %s",
				tenantID, pluginID, api.Method, api.Path)
		}
		delete(r.apis, key)
	}
}

func (r *Registry) GetRoutes(tenantID int64, pluginID string) []RouteRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := pluginKey(tenantID, pluginID)
	return r.routes[key]
}

func (r *Registry) GetAPIs(tenantID int64, pluginID string) []APIRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := pluginKey(tenantID, pluginID)
	return r.apis[key]
}

func (r *Registry) GetAllRoutes() map[string][]RouteRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]RouteRegistration)
	for k, v := range r.routes {
		result[k] = v
	}
	return result
}

func (r *Registry) GetAllAPIs() map[string][]APIRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]APIRegistration)
	for k, v := range r.apis {
		result[k] = v
	}
	return result
}

func (r *Registry) GetRoutesForTenant(tenantID int64) []RouteRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var allRoutes []RouteRegistration
	prefix := fmt.Sprintf("t%d_", tenantID)
	for key, routes := range r.routes {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			allRoutes = append(allRoutes, routes...)
		}
	}
	return allRoutes
}

func (r *Registry) GetAPIsForTenant(tenantID int64) []APIRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var allAPIs []APIRegistration
	prefix := fmt.Sprintf("t%d_", tenantID)
	for key, apis := range r.apis {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			allAPIs = append(allAPIs, apis...)
		}
	}
	return allAPIs
}

func (r *Registry) RegisterMenus(tenantID int64, pluginID string, menus []MenuRegistration) {
	if len(menus) == 0 {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := pluginKey(tenantID, pluginID)
	r.menus[key] = menus

	for _, menu := range menus {
		log.Printf("[Registry] 注册菜单: tenant=%d, plugin=%s, %s -> %s",
			tenantID, pluginID, menu.Title, menu.Path)
	}
}

func (r *Registry) UnregisterMenus(tenantID int64, pluginID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := pluginKey(tenantID, pluginID)
	if menus, exists := r.menus[key]; exists {
		for _, menu := range menus {
			log.Printf("[Registry] 注销菜单: tenant=%d, plugin=%s, %s",
				tenantID, pluginID, menu.Title)
		}
		delete(r.menus, key)
	}
}

func (r *Registry) GetMenusForTenant(tenantID int64) []MenuRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var allMenus []MenuRegistration
	prefix := fmt.Sprintf("t%d_", tenantID)
	for key, menus := range r.menus {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			allMenus = append(allMenus, menus...)
		}
	}
	return allMenus
}

func (r *Registry) GetMenus(tenantID int64, pluginID string) []MenuRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := pluginKey(tenantID, pluginID)
	return r.menus[key]
}
