import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const css = readFileSync(new URL('../src/styles.css', import.meta.url), 'utf8');

test('后台声明品牌、行动和状态令牌', () => {
  assert.match(css, /--brand:\s*#183B43/);
  assert.match(css, /--action:\s*#F08A5D/);
  assert.match(css, /--success:\s*#3C9B79/);
});

test('主要操作使用暖橙，绿色保留给成功状态', () => {
  assert.match(css, /\.primary-button, \.auth-submit[^}]*background:\s*var\(--action\)/);
  assert.match(css, /\.state-serving[^}]*var\(--success\)/);
  assert.match(css, /\.state-pending_dispatch[^}]*var\(--warning\)/);
});
