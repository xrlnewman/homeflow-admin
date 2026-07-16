import test from 'node:test';
import assert from 'node:assert/strict';
import { createApiClient, demoSnapshot } from '../src/api/client.js';

test('API 客户端携带 Bearer Token 并解包运营概览', async () => {
  const calls = [];
  const client = createApiClient({
    baseUrl: 'https://api.example.test',
    token: 'token-demo',
    fetchImpl: async (url, options) => {
      calls.push({ url, options });
      return new Response(JSON.stringify({ code: 0, message: 'ok', data: { orders: 9, pending: 2 } }), {
        status: 200,
        headers: { 'content-type': 'application/json' },
      });
    },
  });

  const result = await client.dashboardSummary();

  assert.deepEqual(result, { source: 'api', data: { orders: 9, pending: 2 } });
  assert.equal(calls[0].url, 'https://api.example.test/api/v1/dashboard/summary');
  assert.equal(calls[0].options.headers.Authorization, 'Bearer token-demo');
});

test('接口不可用时返回演示数据，页面不会白屏', async () => {
  const client = createApiClient({
    baseUrl: 'https://api.example.test',
    fetchImpl: async () => { throw new Error('network down'); },
  });

  const result = await client.snapshot();

  assert.equal(result.source, 'demo');
  assert.deepEqual(result.data.dashboard, demoSnapshot.dashboard);
  assert.deepEqual(result.data.orders, demoSnapshot.orders);
  assert.deepEqual(result.data.recommendations, demoSnapshot.recommendations);
});

test('管理员订单和派单推荐接口按统一 envelope 返回业务数据', async () => {
  const client = createApiClient({
    baseUrl: 'https://api.example.test/',
    token: 'token-demo',
    fetchImpl: async (url) => {
      if (url.includes('/admin/orders')) {
        return new Response(JSON.stringify({ code: 0, data: { list: [{ id: 'HF-1', state: 'pending_dispatch' }], total: 1 } }), { status: 200 });
      }
      return new Response(JSON.stringify({ code: 0, data: [{ id: 'tech-1', name: '陈师傅' }] }), { status: 200 });
    },
  });

  const [orders, recommendations] = await Promise.all([client.adminOrders(), client.technicianRecommendations()]);

  assert.deepEqual(orders, { source: 'api', data: { list: [{ id: 'HF-1', state: 'pending_dispatch' }], total: 1 } });
  assert.equal(recommendations.source, 'api');
  assert.equal(recommendations.data[0].id, 'tech-1');
  assert.equal(recommendations.data[0].name, '陈师傅');
});

test('登录成功后保存访问令牌，并供后续接口复用', async () => {
  const values = new Map();
  const storage = {
    getItem: (key) => values.get(key) || null,
    setItem: (key, value) => values.set(key, value),
  };
  const calls = [];
  const client = createApiClient({
    baseUrl: 'https://api.example.test',
    storage,
    fetchImpl: async (url, options) => {
      calls.push({ url, options });
      const body = url.endsWith('/auth/login')
        ? { code: 0, data: { accessToken: 'token-from-login' } }
        : { code: 0, data: { orders: 1 } };
      return new Response(JSON.stringify(body), { status: 200 });
    },
  });

  await client.login({ phone: '13900000000', password: 'demo123456' });
  await client.dashboardSummary();

  assert.equal(values.get('homeflow_access_token'), 'token-from-login');
  assert.equal(calls[1].options.headers.Authorization, 'Bearer token-from-login');
});

test('派单推荐缺少姓名时补齐可渲染的师傅信息', async () => {
  const client = createApiClient({
    baseUrl: 'https://api.example.test',
    fetchImpl: async () => new Response(JSON.stringify({
      code: 0,
      data: [{ id: 'tech-demo', skills: ['cleaning'], areas: ['north'], shiftAvailable: true, load: 1 }],
    }), { status: 200 }),
  });

  const result = await client.technicianRecommendations();

  assert.equal(result.source, 'api');
  assert.equal(result.data[0].name, '师傅 tech-demo');
  assert.match(result.data[0].desc, /保洁/);
});
