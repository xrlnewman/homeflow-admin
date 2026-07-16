export const orderStateLabels = Object.freeze({
  pending_confirmation: '待确认',
  pending_dispatch: '待派单',
  assigned: '已派单',
  en_route: '前往中',
  serving: '服务中',
  pending_customer_confirmation: '待客户确认',
  completed: '已完成',
  cancelled: '已取消',
});

export function getOrderStateLabel(state) {
  return orderStateLabels[state] ?? '未知状态';
}

export function isTerminalOrderState(state) {
  return state === 'completed' || state === 'cancelled';
}
