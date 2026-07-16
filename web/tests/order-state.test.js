import test from 'node:test';
import assert from 'node:assert/strict';
import { getOrderStateLabel, isTerminalOrderState } from '../src/domain/order-state.js';

test('订单状态显示中文业务标签', () => {
  assert.equal(getOrderStateLabel('pending_dispatch'), '待派单');
  assert.equal(getOrderStateLabel('en_route'), '前往中');
});

test('完成和取消是终态，其他状态不是终态', () => {
  assert.equal(isTerminalOrderState('completed'), true);
  assert.equal(isTerminalOrderState('cancelled'), true);
  assert.equal(isTerminalOrderState('serving'), false);
});
