const DEFAULT_API_BASE_URL = '';

const demoOrders = Object.freeze([
  { id: 'HF-20260716-018', customer: '林女士', service: '深度保洁 · 三室两厅', address: '天河区 · 珠江新城', time: '今天 14:00—16:00', state: 'pending_dispatch', amount: '¥ 368', avatar: '林' },
  { id: 'HF-20260716-017', customer: '周先生', service: '空调深度清洗 × 2', address: '越秀区 · 东风东路', time: '今天 13:30—15:00', state: 'assigned', amount: '¥ 298', avatar: '周', tech: '陈师傅' },
  { id: 'HF-20260716-016', customer: '何女士', service: '厨房局部维修', address: '海珠区 · 江南西', time: '今天 11:00—12:00', state: 'serving', amount: '¥ 189', avatar: '何', tech: '杨师傅' },
  { id: 'HF-20260715-091', customer: '苏先生', service: '全屋整理收纳', address: '番禺区 · 万博', time: '昨天 16:00—18:00', state: 'completed', amount: '¥ 520', avatar: '苏', tech: '李师傅' },
]);

export const demoSnapshot = Object.freeze({
  dashboard: Object.freeze({ orders: 12, completed: 36, pending: 3, completionRate: 0.947, revenue: 8642 }),
  orders: demoOrders,
  recommendations: Object.freeze([
    { id: 'tech-chen', name: '陈师傅', desc: '保洁认证 · 距离 1.2 km · 当前 1 单', score: '98%' },
    { id: 'tech-wang', name: '王师傅', desc: '保洁认证 · 距离 2.4 km · 当前空闲', score: '94%' },
    { id: 'tech-li', name: '李师傅', desc: '整理收纳 · 距离 3.1 km · 当前 2 单', score: '86%' },
  ]),
});

function envApiBaseUrl() {
  return typeof import.meta !== 'undefined' && import.meta.env
    ? import.meta.env.VITE_API_BASE_URL || DEFAULT_API_BASE_URL
    : DEFAULT_API_BASE_URL;
}

function joinUrl(baseUrl, path) {
  return `${baseUrl.replace(/\/$/, '')}${path}`;
}

function readEnvelope(body, response) {
  if (!response.ok || !body || body.code !== 0) {
    throw new Error(body?.message || `API 请求失败（${response.status}）`);
  }
  return body.data;
}

export function createApiClient({
  baseUrl = envApiBaseUrl(),
  token: initialToken = '',
  fetchImpl = globalThis.fetch,
} = {}) {
  let token = initialToken;

  async function request(path, options = {}) {
    if (!baseUrl) {
      throw new Error('VITE_API_BASE_URL 未配置');
    }
    if (typeof fetchImpl !== 'function') {
      throw new Error('当前环境不支持 fetch');
    }
    const headers = {
      Accept: 'application/json',
      ...(options.body ? { 'Content-Type': 'application/json' } : {}),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(options.headers || {}),
    };
    const response = await fetchImpl(joinUrl(baseUrl, `/api/v1${path}`), { ...options, headers });
    let body;
    try {
      body = await response.json();
    } catch {
      throw new Error('API 返回格式错误');
    }
    return readEnvelope(body, response);
  }

  async function withDemoFallback(fetchData, fallback) {
    try {
      return { source: 'api', data: await fetchData() };
    } catch {
      return { source: 'demo', data: fallback };
    }
  }

  async function login(credentials) {
    const data = await request('/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    });
    token = data.accessToken;
    return { source: 'api', data };
  }

  async function dashboardSummary() {
    return withDemoFallback(() => request('/dashboard/summary'), demoSnapshot.dashboard);
  }

  async function adminOrders() {
    return withDemoFallback(
      () => request('/admin/orders?page=1&pageSize=20'),
      { list: demoSnapshot.orders, total: demoSnapshot.orders.length, page: 1, pageSize: demoSnapshot.orders.length },
    );
  }

  async function technicianRecommendations() {
    return withDemoFallback(() => request('/admin/dispatch/recommendations'), demoSnapshot.recommendations);
  }

  async function snapshot() {
    const [dashboard, orders, recommendations] = await Promise.all([
      dashboardSummary(),
      adminOrders(),
      technicianRecommendations(),
    ]);
    const source = [dashboard, orders, recommendations].every((item) => item.source === 'api') ? 'api' : 'demo';
    return {
      source,
      data: {
        dashboard: dashboard.data,
        orders: orders.data?.list || [],
        recommendations: recommendations.data || [],
      },
    };
  }

  return { login, dashboardSummary, adminOrders, technicianRecommendations, snapshot };
}

