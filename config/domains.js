/**
 * FayHub 域名集中化配置管理
 * 版本: 2.0 - 按照原始计划重新配置
 * 说明: 严格按照原始计划配置域名映射
 */

// 开发环境域名配置
const devDomains = {
  // 管理后台域名 - 映射到 localhost:3000
  ADMIN: 'https://admin.fayhub.com',
  ADMIN_TARGET: 'http://localhost:3000',
  
  // 插件市场域名 - 映射到 localhost:3002/market
  WWW: 'https://www.fayhub.com',
  WWW_TARGET: 'http://localhost:3002/market',
  
  // 开发者中心域名 - 映射到 localhost:3003
  DEV: 'https://dev.fayhub.com',
  DEV_TARGET: 'http://localhost:3003',
  
  // API 服务域名 - 映射到 localhost:80
  API: 'https://api.fayhub.com',
  API_TARGET: 'http://localhost:80',
  
  // SSO 服务域名 - 映射到 localhost:80
  SSO: 'https://sso.fayhub.com',
  SSO_TARGET: 'http://localhost:80',
  
  // 反向代理配置
  get proxyConfig() {
    return {
      [this.ADMIN]: this.ADMIN_TARGET,
      [this.WWW]: this.WWW_TARGET,
      [this.DEV]: this.DEV_TARGET,
      [this.API]: this.API_TARGET,
      [this.SSO]: this.SSO_TARGET
    };
  }
};

// 生产环境域名配置
const prodDomains = {
  ADMIN: 'https://admin.fayhub.com',
  API: 'https://api.fayhub.com',
  WWW: 'https://www.fayhub.com',
  DEV: 'https://dev.fayhub.com',
  SSO: 'https://sso.fayhub.com',
  
  // 生产环境使用默认端口（80/443）
  get adminUrl() { return this.ADMIN; },
  get apiUrl() { return this.API; },
  get marketUrl() { return this.WWW; },
  get wwwUrl() { return this.WWW; },
  get devUrl() { return this.DEV; },
  get ssoUrl() { return this.SSO; }
};

// 根据环境选择配置
const isProduction = process.env.NODE_ENV === 'production';
const domains = isProduction ? prodDomains : devDomains;

// 导出配置
module.exports = {
  domains,
  devDomains,
  prodDomains,
  
  // 辅助函数
  getDomainConfig(env = process.env.NODE_ENV) {
    return env === 'production' ? prodDomains : devDomains;
  },
  
  // 验证域名配置
  validateConfig() {
    const required = ['ADMIN', 'API', 'WWW', 'DEV', 'SSO'];
    const missing = required.filter(key => !domains[key]);
    if (missing.length > 0) {
      throw new Error(`域名配置缺失: ${missing.join(', ')}`);
    }
    return true;
  }
};

// 自动验证配置
if (typeof module !== 'undefined' && module.exports) {
  try {
    module.exports.validateConfig();
    console.log('✅ 域名配置验证通过');
  } catch (error) {
    console.error('❌ 域名配置验证失败:', error.message);
  }
}