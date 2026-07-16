<script setup>
import { computed, ref } from 'vue';
import { getOrderStateLabel } from './domain/order-state.js';

const navItems = [
  { label: '运营总览', icon: '⌂', key: 'overview' },
  { label: '订单调度', icon: '↗', key: 'orders', badge: '12' },
  { label: '服务项目', icon: '✦', key: 'services' },
  { label: '服务人员', icon: '◎', key: 'technicians' },
  { label: '客户管理', icon: '◌', key: 'customers' },
  { label: '评价中心', icon: '♡', key: 'reviews' },
];

const orders = ref([
  { id: 'HF-20260716-018', customer: '林女士', service: '深度保洁 · 三室两厅', address: '天河区 · 珠江新城', time: '今天 14:00—16:00', state: 'pending_dispatch', amount: '¥ 368', avatar: '林' },
  { id: 'HF-20260716-017', customer: '周先生', service: '空调深度清洗 × 2', address: '越秀区 · 东风东路', time: '今天 13:30—15:00', state: 'assigned', amount: '¥ 298', avatar: '周', tech: '陈师傅' },
  { id: 'HF-20260716-016', customer: '何女士', service: '厨房局部维修', address: '海珠区 · 江南西', time: '今天 11:00—12:00', state: 'serving', amount: '¥ 189', avatar: '何', tech: '杨师傅' },
  { id: 'HF-20260715-091', customer: '苏先生', service: '全屋整理收纳', address: '番禺区 · 万博', time: '昨天 16:00—18:00', state: 'completed', amount: '¥ 520', avatar: '苏', tech: '李师傅' },
]);

const activeNav = ref('overview');
const range = ref('近 7 天');
const showDispatch = ref(false);
const selectedOrder = ref(null);
const toast = ref('');

const pageTitle = computed(() => navItems.find((item) => item.key === activeNav.value)?.label ?? '运营总览');
const pendingOrders = computed(() => orders.value.filter((order) => order.state === 'pending_dispatch').length);

function openDispatch(order) {
  selectedOrder.value = order;
  showDispatch.value = true;
}

function assignTechnician(name) {
  if (selectedOrder.value) {
    selectedOrder.value.state = 'assigned';
    selectedOrder.value.tech = name;
    toast.value = `${name} 已接收 ${selectedOrder.value.id}`;
  }
  showDispatch.value = false;
  window.setTimeout(() => { toast.value = ''; }, 2800);
}

function setNav(key) {
  activeNav.value = key;
  if (key !== 'overview') {
    toast.value = `${navItems.find((item) => item.key === key)?.label} 模块已加载演示数据`;
    window.setTimeout(() => { toast.value = ''; }, 2200);
  }
}
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand">
        <div class="brand-mark">H</div>
        <div>
          <strong>HomeFlow</strong>
          <span>到家云运营中心</span>
        </div>
      </div>
      <div class="workspace-switcher">
        <span class="workspace-dot"></span>
        <span class="workspace-name">广州直营店</span>
        <span class="workspace-chevron">⌄</span>
      </div>
      <p class="nav-caption">工作台</p>
      <nav class="nav-list" aria-label="主导航">
        <button v-for="item in navItems" :key="item.key" class="nav-item" :class="{ active: activeNav === item.key }" type="button" @click="setNav(item.key)">
          <span class="nav-icon">{{ item.icon }}</span>
          <span>{{ item.label }}</span>
          <span v-if="item.badge && pendingOrders" class="nav-badge">{{ pendingOrders }}</span>
        </button>
      </nav>
      <p class="nav-caption nav-caption--bottom">系统</p>
      <nav class="nav-list">
        <button class="nav-item" type="button" @click="setNav('settings')"><span class="nav-icon">⚙</span><span>系统设置</span></button>
        <button class="nav-item" type="button" @click="setNav('audit')"><span class="nav-icon">▤</span><span>操作审计</span></button>
      </nav>
      <div class="sidebar-footer">
        <div class="mini-avatar">许</div>
        <div><strong>许汝林</strong><span>超级管理员</span></div>
        <button type="button" aria-label="打开个人菜单">•••</button>
      </div>
    </aside>

    <main class="main-content">
      <header class="topbar">
        <div class="breadcrumb"><span>工作台</span><b>/</b><strong>{{ pageTitle }}</strong></div>
        <div class="topbar-actions">
          <button class="icon-button" type="button" aria-label="搜索">⌕</button>
          <button class="icon-button notification" type="button" aria-label="通知"><span>♧</span><i></i></button>
          <div class="topbar-date">2026 年 7 月 16 日 · 星期四</div>
        </div>
      </header>

      <section class="page-heading">
        <div>
          <p class="eyebrow">THURSDAY, JUL 16 · 广州</p>
          <h1>早上好，许汝林 <span>✦</span></h1>
          <p class="heading-copy">今天有 <strong>{{ pendingOrders }} 笔订单</strong> 等待派单，服务团队状态良好。</p>
        </div>
        <button class="primary-button" type="button" @click="setNav('orders')"><span>＋</span> 新建服务订单</button>
      </section>

      <section class="metric-grid" aria-label="经营数据">
        <article class="metric-card metric-card--green"><div class="metric-top"><span class="metric-label">今日成交额</span><span class="metric-trend">↗ 12.8%</span></div><strong>¥ 8,642<span>.00</span></strong><div class="metric-foot"><span>对比昨日</span><div class="mini-sparkline"><i></i><i></i><i></i><i></i><i></i><i></i><i></i><i></i></div></div></article>
        <article class="metric-card"><div class="metric-top"><span class="metric-label">完成订单</span><span class="metric-trend metric-trend--blue">↗ 8.4%</span></div><strong>36<span> 笔</span></strong><div class="metric-foot"><span>完成率 94.7%</span><div class="ring-progress"><i></i></div></div></article>
        <article class="metric-card"><div class="metric-top"><span class="metric-label">新增客户</span><span class="metric-trend metric-trend--purple">↗ 16.2%</span></div><strong>24<span> 位</span></strong><div class="metric-foot"><span>本周累计 108 位</span><div class="avatar-stack"><i>林</i><i>周</i><i>何</i><i>＋</i></div></div></article>
        <article class="metric-card metric-card--dark"><div class="metric-top"><span class="metric-label">平均评分</span><span class="metric-star">★</span></div><strong>4.92<span>/5</span></strong><div class="metric-foot"><span>近 30 天 · 218 条评价</span><span class="rating-word">优秀</span></div></article>
      </section>

      <section class="content-grid">
        <article class="panel order-panel">
          <div class="panel-heading"><div><h2>订单调度</h2><p>需要你关注的服务任务</p></div><button class="link-button" type="button" @click="setNav('orders')">查看全部 <span>→</span></button></div>
          <div class="order-tabs"><button class="tab active" type="button">全部 <span>12</span></button><button class="tab" type="button">待派单 <span>{{ pendingOrders }}</span></button><button class="tab" type="button">服务中 <span>3</span></button><button class="tab" type="button">已完成 <span>36</span></button></div>
          <div class="order-list">
            <div v-for="order in orders" :key="order.id" class="order-row">
              <div class="customer-avatar" :class="`tone-${order.avatar.charCodeAt(0) % 4}`">{{ order.avatar }}</div>
              <div class="order-main"><div class="order-title"><strong>{{ order.customer }}</strong><span class="order-id">{{ order.id }}</span></div><p>{{ order.service }}</p><div class="order-meta"><span>⌖ {{ order.address }}</span><span>◷ {{ order.time }}</span></div></div>
              <div class="order-side"><span class="state-pill" :class="`state-${order.state}`">{{ getOrderStateLabel(order.state) }}</span><strong>{{ order.amount }}</strong><button v-if="order.state === 'pending_dispatch'" class="assign-button" type="button" @click="openDispatch(order)">立即派单</button><span v-else class="tech-name">{{ order.tech }}</span></div>
            </div>
          </div>
        </article>

        <aside class="right-column">
          <article class="panel alert-panel"><div class="panel-heading"><div><h2>今日提醒</h2><p>运营事项一览</p></div><span class="date-chip">7 月 16 日</span></div><div class="alert-list"><div class="alert-item"><span class="alert-icon alert-icon--orange">!</span><div><strong>3 笔订单待派单</strong><p>最早服务时间 13:30</p></div><span class="alert-arrow">→</span></div><div class="alert-item"><span class="alert-icon alert-icon--blue">◷</span><div><strong>2 位师傅即将下班</strong><p>请确认晚间订单安排</p></div><span class="alert-arrow">→</span></div><div class="alert-item"><span class="alert-icon alert-icon--green">✓</span><div><strong>服务满意度达标</strong><p>本周平均评分 4.92</p></div><span class="alert-arrow">→</span></div></div></article>
          <article class="panel team-panel"><div class="panel-heading"><div><h2>服务团队</h2><p>今日在线 12 / 15 人</p></div><button class="more-button" type="button">•••</button></div><div class="team-progress"><div><span>今日负载</span><strong>68%</strong></div><div class="progress-track"><i style="width: 68%"></i></div></div><div class="team-members"><div v-for="member in [{name:'陈师傅',role:'保洁 · 进行中',tone:'green'},{name:'杨师傅',role:'维修 · 进行中',tone:'blue'},{name:'李师傅',role:'整理 · 空闲',tone:'purple'}]" :key="member.name" class="team-member"><div class="member-avatar" :class="`member-${member.tone}`">{{ member.name.slice(0, 1) }}</div><div><strong>{{ member.name }}</strong><span>{{ member.role }}</span></div><i class="online-dot"></i></div></div><button class="outline-button" type="button" @click="setNav('technicians')">管理服务人员 <span>→</span></button></article>
        </aside>
      </section>

      <section class="bottom-grid">
        <article class="panel chart-panel"><div class="panel-heading"><div><h2>订单趋势</h2><p>近 7 天的订单与成交额</p></div><select v-model="range" aria-label="选择时间范围"><option>近 7 天</option><option>近 30 天</option><option>本季度</option></select></div><div class="chart-legend"><span><i class="legend-dot legend-dot--green"></i>完成订单</span><span><i class="legend-dot legend-dot--light"></i>服务收入</span></div><div class="bar-chart"><div v-for="item in [{day:'周五',value:46,amount:'¥ 5.1k'},{day:'周六',value:62,amount:'¥ 6.8k'},{day:'周日',value:55,amount:'¥ 6.2k'},{day:'周一',value:74,amount:'¥ 7.5k'},{day:'周二',value:68,amount:'¥ 7.1k'},{day:'周三',value:82,amount:'¥ 8.2k'},{day:'今天',value:94,amount:'¥ 8.6k'}]" :key="item.day" class="bar-item"><span class="bar-value">{{ item.amount }}</span><div class="bar-track"><i :style="{height: `${item.value}%`}"></i></div><span class="bar-label">{{ item.day }}</span></div></div></article>
        <article class="panel quality-panel"><div class="panel-heading"><div><h2>服务质量</h2><p>本月客户反馈概览</p></div><span class="quality-score">4.92 <small>/ 5</small></span></div><div class="quality-main"><div class="quality-ring"><div><strong>98<span>%</span></strong><small>满意度</small></div></div><div class="quality-breakdown"><div><span>响应速度</span><div class="tiny-track"><i style="width: 96%"></i></div><strong>4.9</strong></div><div><span>服务态度</span><div class="tiny-track"><i style="width: 99%"></i></div><strong>5.0</strong></div><div><span>专业程度</span><div class="tiny-track"><i style="width: 94%"></i></div><strong>4.8</strong></div></div></div><button class="quality-link" type="button" @click="setNav('reviews')">查看全部评价 <span>→</span></button></article>
      </section>

      <footer class="page-footer"><span>HomeFlow 到家云 · 免费开源运营系统</span><span>数据更新时间：刚刚</span></footer>
    </main>

    <div v-if="showDispatch" class="modal-backdrop" @click.self="showDispatch = false"><section class="dispatch-modal" role="dialog" aria-modal="true" aria-labelledby="dispatch-title"><button class="modal-close" type="button" aria-label="关闭" @click="showDispatch = false">×</button><p class="eyebrow">ORDER {{ selectedOrder?.id }}</p><h2 id="dispatch-title">为订单安排服务人员</h2><p class="modal-copy">系统按技能、距离和当前负载为你推荐以下人员。</p><div class="recommend-list"><button v-for="person in [{name:'陈师傅',desc:'保洁认证 · 距离 1.2 km · 当前 1 单',score:'98%'},{name:'王师傅',desc:'保洁认证 · 距离 2.4 km · 当前空闲',score:'94%'},{name:'李师傅',desc:'整理收纳 · 距离 3.1 km · 当前 2 单',score:'86%'}]" :key="person.name" class="recommend-item" type="button" @click="assignTechnician(person.name)"><span class="recommend-avatar">{{ person.name.slice(0, 1) }}</span><span><strong>{{ person.name }}</strong><small>{{ person.desc }}</small></span><b>{{ person.score }}</b><span class="recommend-arrow">→</span></button></div></section></div>
    <div v-if="toast" class="toast">✓ {{ toast }}</div>
  </div>
</template>
