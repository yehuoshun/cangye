// ==UserScript==
// @name         语雀工作台加速器
// @namespace    http://tampermonkey.net/
// @version      0.1
// @description  用缓存的分类数据替换语雀工作台慢速知识库列表
// @author       yehuoshun
// @match        https://www.yuque.com/dashboard*
// @grant        none
// @icon         https://www.yuque.com/favicon.ico
// ==/UserScript==

(function () {
  'use strict'

  const JSON_URL = 'https://cdn.jsdelivr.net/gh/yehuoshun/cangye@main/yuque-repos.json'
  const LOCAL_KEY = 'yuque_repos_cache'
  const CACHE_TTL = 24 * 60 * 60 * 1000 // 1天

  // ── 样式注入 ──
  const styles = `
.yuque-turbo { font-family: -apple-system, 'PingFang SC', 'Segoe UI', sans-serif; height: 100%; display: flex; }
.yuque-turbo-sidebar { width: 200px; min-width:200px; background: #f5f5f5; border-right: 1px solid #e8e8e8; overflow-y: auto; padding: 12px 0; }
.yuque-turbo-sidebar .cat-item { padding: 8px 16px; cursor: pointer; display: flex; align-items: center; gap: 8px; font-size: 13px; color: #333; border-radius: 0; transition: background .1s; }
.yuque-turbo-sidebar .cat-item:hover { background: #e0e0e0; }
.yuque-turbo-sidebar .cat-item.active { background: #d8d8eb; color: #2a4a7a; font-weight: 600; }
.yuque-turbo-sidebar .cat-count { margin-left: auto; font-size: 11px; color: #999; background: #eee; padding: 0 6px; border-radius: 8px; }
.yuque-turbo-main { flex: 1; display: flex; flex-direction: column; overflow: hidden; padding: 16px 24px; background: #fff; }
.yuque-turbo-toolbar { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
.yuque-turbo-toolbar input { flex:1; max-width:320px; padding: 8px 12px; border:1px solid #ddd; border-radius:6px; font-size:13px; outline:none; }
.yuque-turbo-toolbar input:focus { border-color: #7a7ab8; }
.yuque-turbo-toolbar .turbo-btn { padding:6px 14px; border:1px solid #ddd; border-radius:6px; background:#fff; cursor:pointer; font-size:12px; color:#666; }
.yuque-turbo-toolbar .turbo-btn:hover { background:#f0f0f0; }
.yuque-turbo-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 10px; overflow-y: auto; flex:1; }
.yuque-turbo-card { padding: 14px 16px; border: 1px solid #eee; border-radius: 8px; cursor: pointer; transition: all .15s; background: #fafafa; }
.yuque-turbo-card:hover { border-color: #7a7ab8; background: #f0f0ff; }
.yuque-turbo-card .name { font-size: 14px; font-weight: 600; color: #333; }
.yuque-turbo-card .meta { font-size: 12px; color: #999; margin-top: 4px; display: flex; gap: 12px; }
.yuque-turbo-card .cat-badge { display:inline-block; font-size: 11px; padding: 1px 8px; border-radius: 10px; background:#e8e8f0; color:#666; margin-top:6px; }
.yuque-turbo-loading { display:flex; align-items:center; justify-content:center; height:200px; color:#999; font-size:14px; }
.yuque-turbo-empty { text-align:center; padding:48px; color:#999; }
  `

  // ── 加载数据 ──
  async function loadRepos () {
    // 尝试 localStorage 缓存
    const cached = localStorage.getItem(LOCAL_KEY)
    if (cached) {
      try {
        const parsed = JSON.parse(cached)
        if (Date.now() - parsed.ts < CACHE_TTL && parsed.data?.length > 0) {
          return parsed.data
        }
      } catch {}
    }
    try {
      const resp = await fetch(JSON_URL, { cache: 'no-cache' })
      if (!resp.ok) throw new Error('HTTP ' + resp.status)
      const json = await resp.json()
      const data = json.repos || []
      localStorage.setItem(LOCAL_KEY, JSON.stringify({ ts: Date.now(), data }))
      return data
    } catch (e) {
      console.warn('[yuque-turbo] fetch failed:', e)
      // 缓存兜底
      if (cached) return JSON.parse(cached).data
      return []
    }
  }

  // ── 渲染 ──
  function render (repos) {
    // 清理原页面
    const app = document.querySelector('.app-container, #app, .dashboard-layout, [class*="dashboard"]')
    let container = document.getElementById('yuque-turbo-root')
    if (container) {
      container.innerHTML = ''
    } else {
      container = document.createElement('div')
      container.id = 'yuque-turbo-root'
      container.style.cssText = 'position:fixed;inset:0;z-index:9999;background:#fff;overflow:hidden'
      document.body.appendChild(container)
    }

    // 注入样式
    const styleEl = document.createElement('style')
    styleEl.textContent = styles
    container.appendChild(styleEl)

    // 分类统计
    const cats = {}
    repos.forEach(r => {
      const c = r.category || '其他'
      if (!cats[c]) cats[c] = []
      cats[c].push(r)
    })
    const catNames = Object.keys(cats).sort((a, b) => cats[b].length - cats[a].length)

    let currentCat = location.hash?.slice(1) || catNames[0] || ''
    let filteredRepos = cats[currentCat] || repos

    // 筛选函数
    function filterByCat (cat) {
      currentCat = cat
      location.hash = '#' + cat
      filteredRepos = cat === '__all__' ? repos : (cats[cat] || repos)
      renderList(filteredRepos)
      document.querySelectorAll('.cat-item').forEach(el => {
        el.classList.toggle('active', el.dataset.cat === cat)
      })
    }

    // 搜索
    function search (q) {
      const list = currentCat === '__all__' ? repos : (cats[currentCat] || repos)
      if (!q.trim()) return renderList(list)
      const lq = q.toLowerCase()
      const found = list.filter(r =>
        r.name.toLowerCase().includes(lq)
      )
      renderList(found)
    }

    // 主结构
    const html = `
    <div class="yuque-turbo">
      <div class="yuque-turbo-sidebar">
        <div class="cat-item" data-cat="__all__">📋 全部 (${repos.length})</div>
        ${catNames.map(c => `<div class="cat-item" data-cat="${c}">📁 ${c} <span class="cat-count">${cats[c].length}</span></div>`).join('')}
      </div>
      <div class="yuque-turbo-main">
        <div class="yuque-turbo-toolbar">
          <input id="yuque-turbo-search" placeholder="搜索知识库..." />
          <button class="turbo-btn" id="yuque-turbo-refresh">🔄 刷新</button>
          <span style="font-size:12px;color:#999">共 ${repos.length} 个知识库</span>
        </div>
        <div id="yuque-turbo-list" class="yuque-turbo-grid"></div>
      </div>
    </div>
    `

    container.insertAdjacentHTML('beforeend', html)
    document.querySelector('#yuque-turbo-search').addEventListener('input', e => search(e.target.value))
    document.querySelector('#yuque-turbo-refresh').addEventListener('click', () => {
      localStorage.removeItem(LOCAL_KEY)
      location.reload()
    })
    document.querySelectorAll('.cat-item').forEach(el => {
      el.addEventListener('click', () => filterByCat(el.dataset.cat))
    })

    // 初始选中
    const activeEl = document.querySelector(`.cat-item[data-cat="${currentCat}"]`) || document.querySelector('.cat-item')
    if (activeEl) {
      activeEl.classList.add('active')
      filterByCat(currentCat || '__all__')
    }
  }

  function renderList (list) {
    const el = document.getElementById('yuque-turbo-list')
    if (!el) return
    if (list.length === 0) {
      el.innerHTML = '<div class="yuque-turbo-empty">暂无知识库</div>'
      return
    }
    el.innerHTML = list.map(r => `
      <div class="yuque-turbo-card" data-ns="${r.namespace}">
        <div class="name">${r.name}</div>
        <div class="meta">
          <span>📄 ${r.items_count || 0} 项</span>
          <span>🕐 ${(r.updated_at || '').slice(0, 10)}</span>
        </div>
        ${r.category ? `<div class="cat-badge">${r.category}</div>` : ''}
      </div>
    `).join('')
    el.querySelectorAll('.yuque-turbo-card').forEach(card => {
      card.addEventListener('click', () => {
        const ns = card.dataset.ns
        if (ns) window.location.href = `https://www.yuque.com/${ns}`
      })
    })
  }

  // ── 启动 ──
  async function init () {
    // 等待页面加载完再替换
    await new Promise(resolve => {
      if (document.readyState === 'complete') return resolve()
      window.addEventListener('load', resolve)
    })
    // 等原始 UI 渲染（但 200ms 后就劫持）
    await new Promise(r => setTimeout(r, 200))
    const repos = await loadRepos()
    if (repos.length === 0) {
      console.warn('[yuque-turbo] 没有加载到知识库数据')
      return
    }
    render(repos)
  }

  init()
})()
