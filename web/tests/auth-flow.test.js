import test from 'node:test';
import assert from 'node:assert/strict';
import { createApiClient, resolveAuthState } from '../src/api/client.js';

test('根据 API 配置和令牌决定登录入口状态', () => {
  assert.equal(resolveAuthState({ baseUrl: '', token: '' }), 'offline-ready');
  assert.equal(resolveAuthState({ baseUrl: 'https://api.example.test', token: '' }), 'login');
  assert.equal(resolveAuthState({ baseUrl: 'https://api.example.test', token: 'token-demo' }), 'authenticated');
});

test('退出登录调用接口并清理本地访问令牌', async () => {
  const values = new Map([['homeflow_access_token', 'token-demo']]);
  const calls = [];
  const storage = {
    getItem: (key) => values.get(key) || null,
    setItem: (key, value) => values.set(key, value),
    removeItem: (key) => values.delete(key),
  };
  const client = createApiClient({
    baseUrl: 'https://api.example.test',
    storage,
    fetchImpl: async (url, options) => {
      calls.push({ url, options });
      return new Response(JSON.stringify({ code: 0, data: {} }), { status: 200 });
    },
  });

  const result = await client.logout();

  assert.equal(result.source, 'api');
  assert.equal(values.has('homeflow_access_token'), false);
  assert.equal(calls[0].url, 'https://api.example.test/api/v1/auth/logout');
  assert.equal(calls[0].options.method, 'POST');
  assert.equal(calls[0].options.headers.Authorization, 'Bearer token-demo');
});

test('接口退出失败时仍清理本地令牌，避免下次启动使用过期会话', async () => {
  const values = new Map([['homeflow_access_token', 'expired-token']]);
  const storage = {
    getItem: (key) => values.get(key) || null,
    setItem: (key, value) => values.set(key, value),
    removeItem: (key) => values.delete(key),
  };
  const client = createApiClient({
    baseUrl: 'https://api.example.test',
    storage,
    fetchImpl: async () => { throw new Error('network down'); },
  });

  const result = await client.logout();

  assert.equal(result.source, 'demo');
  assert.equal(result.remoteError, 'network down');
  assert.equal(values.has('homeflow_access_token'), false);
});
