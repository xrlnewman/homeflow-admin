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
  assert.deepEqual(recommendations, { source: 'api', data: [{ id: 'tech-1', name: '陈师傅' }] });
});
