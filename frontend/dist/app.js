'use strict';

const STRINGS = {
  en: {
    scanning: 'Checking your computer…',
    scanningSub: (done, total) => `${done} of ${total} places checked`,
    readyLabel: 'Selected for cleaning',
    selectedSub: (items, files) => `${items} ${items === 1 ? 'item' : 'items'} · ${files} ${files === 1 ? 'file' : 'files'}`,
    nothingSelected: 'Select what you want to clean below.',
    promise: 'Nothing is deleted until you confirm.',
    rescan: 'Scan again',
    clean: 'Clean…',
    cleaning: 'Cleaning…',
    cleaningSub: (done, total) => `${done} of ${total} done`,
    freedLabel: 'Freed',
    doneSub: (files) => `${files} ${files === 1 ? 'file' : 'files'} deleted.`,
    inUseNote: (n) => `${n} ${n === 1 ? 'file was' : 'files were'} in use by other programs and left alone.`,
    failedNote: (n) => `${n} ${n === 1 ? 'file' : 'files'} could not be deleted.`,
    error: 'Something went wrong',
    categories: {
      system: 'System', logs: 'Logs and crash data', packages: 'Package caches',
      graphics: 'Graphics', browser: 'Browsers', developer: 'Developer tools',
    },
    files: (n) => `${n} ${n === 1 ? 'file' : 'files'}`,
    needsAdmin: 'Needs administrator',
    needsRoot: 'Needs root',
    closeBrowser: 'Close the browser first',
    closeProgram: (names) => `Close ${names} first`,
    failed: 'Error',
    clean0: 'Already clean',
    lowerBound: 'At least this much. Some folders could only be read with administrator rights.',
    deleted: (n) => `${n} deleted`,
    inUse: (n) => `${n} in use`,
    notDeleted: (n) => `${n} not deleted`,
    skipped: 'Skipped',
    adminBanner: (n) => `${n} ${n === 1 ? 'item needs' : 'items need'} administrator rights.`,
    restartAdmin: 'Restart as administrator',
    rootBanner: 'System items need root. To clean them, run',
    confirmTitle: (files) => `Delete ${files} ${files === '1' ? 'file' : 'files'}?`,
    confirmBody: (bytes, trash) => `${bytes} will be deleted permanently. The files are not moved to the ${trash}.`,
    trash: { windows: 'Recycle Bin', linux: 'Trash' },
    confirmCosts: 'Worth knowing:',
    cancel: 'Cancel',
    deletePermanently: 'Delete permanently',
    standardUser: 'Standard user',
    administrator: 'Administrator',
    root: 'root',
    footer: 'Works offline · No telemetry · Open source, MIT license',
    source: 'Source code',
    noBridge: 'This page is part of the sweepkit desktop app and does not work on its own.',
    nothingFound: 'Nothing to clean was found on this computer.',
  },
  tr: {
    scanning: 'Bilgisayarınız kontrol ediliyor…',
    scanningSub: (done, total) => `${total} konumdan ${done} tanesi kontrol edildi`,
    readyLabel: 'Temizlik için seçilen',
    selectedSub: (items, files) => `${items} öğe · ${files} dosya`,
    nothingSelected: 'Aşağıdan temizlemek istediklerinizi seçin.',
    promise: 'Siz onaylamadan hiçbir şey silinmez.',
    rescan: 'Yeniden tara',
    clean: 'Temizle…',
    cleaning: 'Temizleniyor…',
    cleaningSub: (done, total) => `${total} öğeden ${done} tanesi bitti`,
    freedLabel: 'Açılan alan',
    doneSub: (files) => `${files} dosya silindi.`,
    inUseNote: (n) => `${n} dosya başka programlar tarafından kullanıldığı için silinmedi.`,
    failedNote: (n) => `${n} dosya silinemedi.`,
    error: 'Bir şeyler ters gitti',
    categories: {
      system: 'Sistem', logs: 'Günlükler ve çökme verileri', packages: 'Paket önbellekleri',
      graphics: 'Grafik', browser: 'Tarayıcılar', developer: 'Geliştirici araçları',
    },
    files: (n) => `${n} dosya`,
    needsAdmin: 'Yönetici gerekli',
    needsRoot: 'Root gerekli',
    closeBrowser: 'Önce tarayıcıyı kapatın',
    closeProgram: (names) => `Önce ${names} kapatın`,
    failed: 'Hata',
    clean0: 'Zaten temiz',
    lowerBound: 'En az bu kadar. Bazı klasörler ancak yönetici haklarıyla okunabiliyor.',
    deleted: (n) => `${n} silindi`,
    inUse: (n) => `${n} kullanımda`,
    notDeleted: (n) => `${n} silinemedi`,
    skipped: 'Atlandı',
    adminBanner: (n) => `${n} öğe yönetici hakları gerektiriyor.`,
    restartAdmin: 'Yönetici olarak yeniden başlat',
    rootBanner: 'Sistem öğeleri root gerektirir. Temizlemek için şunu çalıştırın:',
    confirmTitle: (files) => `${files} dosya silinsin mi?`,
    confirmBody: (bytes, trash) => `${bytes} kalıcı olarak silinecek. Dosyalar ${trash} taşınmaz.`,
    trash: { windows: "Geri Dönüşüm Kutusu'na", linux: "Çöp Kutusu'na" },
    confirmCosts: 'Bilmeniz gerekenler:',
    cancel: 'Vazgeç',
    deletePermanently: 'Kalıcı olarak sil',
    standardUser: 'Standart kullanıcı',
    administrator: 'Yönetici',
    root: 'root',
    footer: 'Çevrimdışı çalışır · Telemetri yok · Açık kaynak, MIT lisansı',
    source: 'Kaynak kodu',
    noBridge: 'Bu sayfa sweepkit masaüstü uygulamasının bir parçasıdır ve tek başına çalışmaz.',
    nothingFound: 'Bu bilgisayarda temizlenecek bir şey bulunamadı.',
  },
};

const WARN_ICON = '<svg viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M8 1.5a1 1 0 0 1 .87.5l6 10.5A1 1 0 0 1 14 14H2a1 1 0 0 1-.87-1.5l6-10.5A1 1 0 0 1 8 1.5Zm0 4a.75.75 0 0 0-.75.75v3a.75.75 0 0 0 1.5 0v-3A.75.75 0 0 0 8 5.5Zm0 6.25a.9.9 0 1 0 0-1.8.9.9 0 0 0 0 1.8Z"/></svg>';

const state = {
  lang: initialLang(),
  info: null,
  phase: 'loading', // loading | scanning | ready | cleaning | done | error
  scans: new Map(),
  cleans: new Map(),
  selected: new Set(),
  touched: new Set(),
  progress: { done: 0, total: 0 },
  error: '',
};

const $ = (id) => document.getElementById(id);
let api = null;

function S() { return STRINGS[state.lang]; }

function initialLang() {
  try {
    const saved = localStorage.getItem('lang');
    if (saved === 'en' || saved === 'tr') return saved;
  } catch { /* storage unavailable */ }
  return (navigator.language || '').toLowerCase().startsWith('tr') ? 'tr' : 'en';
}

function setLang(lang) {
  state.lang = lang;
  try { localStorage.setItem('lang', lang); } catch { /* storage unavailable */ }
  render();
}

function esc(s) {
  return String(s).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
}

function text(t) { return (t && (t[state.lang] || t.en)) || ''; }

// Escapes a description and shows `commands` as code.
function rich(s) { return esc(s).replace(/`([^`]+)`/g, '<code>$1</code>'); }

function fmtInt(n) { return new Intl.NumberFormat(state.lang).format(n); }

function fmtBytes(n) {
  if (n < 1024) return `${fmtInt(n)} B`;
  const units = ['KB', 'MB', 'GB', 'TB'];
  let v = n / 1024;
  let i = 0;
  while (v >= 1024 && i < units.length - 1) { v /= 1024; i++; }
  const num = new Intl.NumberFormat(state.lang, { minimumFractionDigits: 1, maximumFractionDigits: 1 }).format(v);
  return `${num} ${units[i]}`;
}

function target(id) { return state.info.targets.find((t) => t.id === id); }

function ordered(ids) {
  return state.info.targets.map((t) => t.id).filter((id) => ids.has(id));
}

function selectable(r) { return r && r.status === 'ready' && r.files > 0; }

// A target's checkbox follows its default until the user touches it.
function applySelection(id) {
  const r = state.scans.get(id);
  const t = target(id);
  const want = state.touched.has(id) ? state.selected.has(id) : t.default;
  if (want && selectable(r)) state.selected.add(id);
  else state.selected.delete(id);
}

function totals() {
  let bytes = 0;
  let files = 0;
  for (const id of state.selected) {
    const r = state.scans.get(id);
    if (selectable(r)) { bytes += r.bytes; files += r.files; }
  }
  return { bytes, files, items: state.selected.size };
}

// Rendering

function render() {
  const s = S();
  document.documentElement.lang = state.lang;
  for (const b of document.querySelectorAll('[data-lang]')) {
    b.setAttribute('aria-pressed', String(b.dataset.lang === state.lang));
  }
  if (!state.info) return;
  $('version').textContent = state.info.version;
  const badge = $('privilege');
  badge.textContent = state.info.elevated ? (state.info.os === 'windows' ? s.administrator : s.root) : s.standardUser;
  badge.classList.toggle('elevated', state.info.elevated);
  $('promise').textContent = s.footer;
  $('source').textContent = s.source;
  renderSummary();
  renderBanner();
  renderList();
}

function renderSummary() {
  const s = S();
  const el = $('summary');
  const { done, total } = state.progress;
  const bar = (busy) => (busy && total ? `<div class="progress" style="width:${Math.round((done / total) * 100)}%"></div>` : '');
  const busy = state.phase === 'scanning' || state.phase === 'cleaning';

  if (state.phase === 'loading') {
    el.innerHTML = `<div><div class="summary-label">${esc(s.scanning)}</div><div class="summary-value muted num">—</div></div>`;
    return;
  }
  if (state.phase === 'error') {
    el.innerHTML = `<div><div class="summary-label">${esc(s.error)}</div><div class="summary-sub">${esc(state.error)}</div></div>
      <div class="actions"><button class="btn" data-action="scan">${esc(s.rescan)}</button></div>`;
    return;
  }
  if (state.phase === 'done' || state.phase === 'cleaning') {
    let freed = 0; let deleted = 0; let inUse = 0; let failed = 0;
    for (const r of state.cleans.values()) {
      freed += r.freed; deleted += r.deleted; inUse += r.inUse || 0; failed += r.failed || 0;
    }
    const notes = [];
    if (inUse) notes.push(s.inUseNote(fmtInt(inUse)));
    if (failed) notes.push(s.failedNote(fmtInt(failed)));
    const label = state.phase === 'cleaning' ? s.cleaning : s.freedLabel;
    const sub = state.phase === 'cleaning' ? s.cleaningSub(done, total) : s.doneSub(fmtInt(deleted));
    el.innerHTML = `<div>
        <div class="summary-label">${esc(label)}</div>
        <div class="summary-value num ${state.phase === 'done' ? 'done' : 'muted'}">${esc(fmtBytes(freed))}</div>
        <div class="summary-sub">${esc(sub)}</div>
        ${notes.length ? `<div class="summary-note">${esc(notes.join(' '))}</div>` : ''}
      </div>
      <div class="actions"><button class="btn primary" data-action="scan" ${busy ? 'disabled' : ''}>${esc(s.rescan)}</button></div>
      ${bar(busy)}`;
    return;
  }

  const { bytes, files, items } = totals();
  const label = state.phase === 'scanning' ? s.scanning : s.readyLabel;
  const sub = state.phase === 'scanning'
    ? s.scanningSub(done, total)
    : (items ? s.selectedSub(items, fmtInt(files)) : s.nothingSelected);
  el.innerHTML = `<div>
      <div class="summary-label">${esc(label)}</div>
      <div class="summary-value num ${items ? '' : 'muted'}">${esc(fmtBytes(bytes))}</div>
      <div class="summary-sub">${esc(sub)}</div>
      <div class="summary-note">${esc(s.promise)}</div>
    </div>
    <div class="actions">
      <button class="btn" data-action="scan" ${busy ? 'disabled' : ''}>${esc(s.rescan)}</button>
      <button class="btn primary" data-action="clean" ${busy || !files ? 'disabled' : ''}>${esc(s.clean)}</button>
    </div>
    ${bar(busy)}`;
}

function renderBanner() {
  const s = S();
  const el = $('banner');
  const waiting = [...state.scans.values()].filter((r) => r.status === 'needs_admin' && (r.files > 0 || r.partial));
  if (state.info.elevated || !waiting.length || state.phase === 'scanning') {
    el.hidden = true;
    return;
  }
  el.hidden = false;
  el.innerHTML = state.info.canElevate
    ? `<span>${esc(s.adminBanner(waiting.length))}</span><button class="btn small" data-action="elevate">${esc(s.restartAdmin)}</button>`
    : `<span>${esc(s.rootBanner)} <code>sudo sweepkit clean</code></span>`;
}

function statusPill(t, r) {
  const s = S();
  switch (r.status) {
    case 'needs_admin':
      return `<span class="pill">${esc(state.info.os === 'windows' ? s.needsAdmin : s.needsRoot)}</span>`;
    case 'blocked': {
      const label = t.category === 'browser' ? s.closeBrowser : s.closeProgram((r.blockers || []).join(', '));
      return `<span class="pill">${esc(label)}</span>`;
    }
    case 'failed':
      return `<span class="pill error" title="${esc(r.error || '')}">${esc(s.failed)}</span>`;
    case 'ready':
      return r.files === 0 && !r.partial ? `<span class="pill quiet">${esc(s.clean0)}</span>` : '';
    default:
      return '';
  }
}

function sizeCell(r) {
  const s = S();
  const c = state.cleans.get(r.id);
  if (c) {
    if (c.status !== 'done') return `<strong>${esc(s.skipped)}</strong>`;
    const bits = [s.deleted(fmtInt(c.deleted))];
    if (c.inUse) bits.push(s.inUse(fmtInt(c.inUse)));
    if (c.failed) bits.push(s.notDeleted(fmtInt(c.failed)));
    return `<strong class="freed num">−${esc(fmtBytes(c.freed))}</strong><small class="num">${esc(bits.join(' · '))}</small>`;
  }
  if (r.partial && r.files === 0) return `<strong class="num" title="${esc(s.lowerBound)}">—</strong>`;
  if (r.files === 0) return '';
  const size = fmtBytes(r.bytes) + (r.partial ? '+' : '');
  const title = r.partial ? ` title="${esc(s.lowerBound)}"` : '';
  return `<strong class="num"${title}>${esc(size)}</strong><small class="num">${esc(s.files(fmtInt(r.files)))}</small>`;
}

function rowHTML(t) {
  const r = state.scans.get(t.id);
  const locked = state.phase !== 'ready';
  const enabled = selectable(r) && !locked;
  const alreadyClean = r.status === 'ready' && r.files === 0;
  const caveat = alreadyClean ? '' : text(t.caveat);
  return `<label class="row ${selectable(r) ? '' : 'disabled'}" data-id="${esc(t.id)}">
      <input type="checkbox" data-id="${esc(t.id)}" aria-label="${esc(text(t.name))}" ${state.selected.has(t.id) ? 'checked' : ''} ${enabled ? '' : 'disabled'}>
      <div>
        <div class="row-title">${esc(text(t.name))} ${statusPill(t, r)}</div>
        <div class="row-desc">${rich(text(t.description))}</div>
        ${caveat ? `<div class="row-caveat">${WARN_ICON}<span>${rich(caveat)}</span></div>` : ''}
      </div>
      <div class="row-size">${sizeCell(r)}</div>
    </label>`;
}

function skeletonRows(n) {
  let html = '';
  for (let i = 0; i < n; i++) {
    html += `<div class="row disabled"><span></span><div><span class="skeleton" style="width:${40 + ((i * 17) % 30)}%"></span>
      <span class="skeleton" style="width:${70 + ((i * 13) % 20)}%;margin-top:8px;height:10px"></span></div>
      <div><span class="skeleton" style="width:64px"></span></div></div>`;
  }
  return `<div class="group-title">&nbsp;</div><div class="rows">${html}</div>`;
}

function renderList() {
  const s = S();
  const el = $('list');
  let html = '';
  for (const cat of state.info.categories) {
    const rows = state.info.targets.filter((t) => {
      const r = state.scans.get(t.id);
      return t.category === cat && r && r.status !== 'not_found';
    });
    if (!rows.length) continue;
    html += `<h2 class="group-title">${esc(s.categories[cat] || cat)}</h2><div class="rows">${rows.map(rowHTML).join('')}</div>`;
  }
  if (!html && (state.phase === 'scanning' || state.phase === 'loading')) html = skeletonRows(4);
  if (!html && state.phase === 'ready') html = `<div class="empty">${esc(s.nothingFound)}</div>`;
  el.innerHTML = html;
}

// Actions

async function scan() {
  const ids = state.info.targets.map((t) => t.id);
  state.phase = 'scanning';
  state.scans = new Map();
  state.cleans = new Map();
  state.progress = { done: 0, total: ids.length };
  render();
  try {
    const results = await api.Scan(ids);
    for (const r of results) {
      state.scans.set(r.id, r);
      applySelection(r.id);
    }
    state.phase = 'ready';
  } catch (e) {
    state.phase = 'error';
    state.error = String(e);
  }
  render();
}

function confirmClean() {
  const s = S();
  const ids = ordered(state.selected);
  const { bytes, files } = totals();
  $('confirm-title').textContent = s.confirmTitle(fmtInt(files));
  $('confirm-body').textContent = s.confirmBody(fmtBytes(bytes), s.trash[state.info.os] || s.trash.linux);
  $('confirm-items').innerHTML = ids.map((id) => {
    const r = state.scans.get(id);
    return `<li><span>${esc(text(target(id).name))}</span><span class="num">${esc(fmtBytes(r.bytes))}</span></li>`;
  }).join('');
  const costs = ids.map((id) => text(target(id).caveat)).filter(Boolean);
  const unique = [...new Set(costs)];
  $('confirm-costs').innerHTML = unique.length
    ? `${esc(s.confirmCosts)}<ul>${unique.map((c) => `<li>${esc(c)}</li>`).join('')}</ul>`
    : '';
  $('confirm-cancel').textContent = s.cancel;
  $('confirm-ok').textContent = s.deletePermanently;
  $('confirm').showModal();
  $('confirm-cancel').focus();
}

async function clean() {
  const ids = ordered(state.selected);
  state.phase = 'cleaning';
  state.cleans = new Map();
  state.progress = { done: 0, total: ids.length };
  render();
  try {
    const results = await api.Clean(ids);
    state.cleans = new Map(results.map((r) => [r.id, r]));
    state.selected.clear();
    state.phase = 'done';
  } catch (e) {
    state.phase = 'error';
    state.error = String(e);
  }
  render();
}

function onProgress(p) {
  if (p.phase === 'scan' && state.phase === 'scanning' && p.result) {
    state.scans.set(p.id, p.result);
    applySelection(p.id);
  }
  if ((p.phase === 'scan' && state.phase === 'scanning') || (p.phase === 'clean' && state.phase === 'cleaning')) {
    state.progress = { done: p.done, total: p.total };
    render();
  }
}

// Wiring

function bind() {
  document.addEventListener('click', (e) => {
    const langButton = e.target.closest('[data-lang]');
    if (langButton) { setLang(langButton.dataset.lang); return; }
    const action = e.target.closest('[data-action]');
    if (!action || action.disabled) return;
    if (action.dataset.action === 'scan') scan();
    if (action.dataset.action === 'clean') confirmClean();
    if (action.dataset.action === 'elevate') {
      // Declining the UAC prompt rejects the call; that is the user's choice, not an error.
      api.RestartAsAdmin().catch(() => {});
    }
  });
  $('list').addEventListener('change', (e) => {
    const box = e.target.closest('input[type="checkbox"]');
    if (!box) return;
    const id = box.dataset.id;
    state.touched.add(id);
    if (box.checked) state.selected.add(id); else state.selected.delete(id);
    // Only the totals change; redrawing the list would steal keyboard focus.
    renderSummary();
  });
  // The form closes the dialog itself; only the delete button starts cleaning.
  $('confirm-ok').addEventListener('click', () => clean());
  $('source').addEventListener('click', () => api.OpenProjectPage());
}

async function waitForBridge() {
  for (let i = 0; i < 40; i++) {
    if (window.go && window.go.main && window.go.main.App && window.runtime) return window.go.main.App;
    await new Promise((r) => setTimeout(r, 50));
  }
  return null;
}

async function init() {
  render();
  api = await waitForBridge();
  if (!api) {
    document.body.innerHTML = `<p class="fatal">${esc(S().noBridge)}</p>`;
    return;
  }
  state.info = await api.Info();
  window.runtime.EventsOn('progress', onProgress);
  bind();
  scan();
}

init();
