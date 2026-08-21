const TOKEN_KEY = "cp_token";

function token() { return localStorage.getItem(TOKEN_KEY) || ""; }
function setToken(t) { localStorage.setItem(TOKEN_KEY, t); }
function clearToken() { localStorage.removeItem(TOKEN_KEY); }

async function api(path, opts = {}) {
  const headers = Object.assign({ "Content-Type": "application/json" }, opts.headers || {});
  const t = token();
  if (t) headers.Authorization = "Bearer " + t;
  const res = await fetch(path, Object.assign({}, opts, { headers }));
  if (res.status === 204) return null;
  const text = await res.text();
  let data = null;
  try { data = text ? JSON.parse(text) : null; } catch { data = { message: text }; }
  if (!res.ok) {
    const msg = (data && (data.message || data.code)) || ("HTTP " + res.status);
    throw new Error(msg);
  }
  return data;
}

function requireLogin() {
  if (!token()) { location.href = "/login"; return false; }
  return true;
}

async function currentMe() { return api("/api/auth/me"); }

function logout() {
  api("/api/auth/logout", { method: "POST" }).catch(() => {});
  clearToken();
  location.href = "/login";
}

function $(id) { return document.getElementById(id); }

function yuan(cents) {
  if (cents == null) return "0.00";
  return (Number(cents) / 100).toFixed(2);
}

function fillUserBar(me) {
  const el = $("userbar");
  if (!el || !me) return;
  const u = me.user || me;
  el.innerHTML = `<span>${u.display_name || u.username} · ${u.role} · 已捐 ¥${yuan(u.total_donated_cents)}</span>
    <a href="/app">广场</a> <a href="/me">我的</a>
    ${u.role === "org" || u.role === "admin" ? '<a href="/org">机构</a>' : ""}
    ${u.role === "admin" ? '<a href="/admin">后台</a>' : ""}
    <button class="btn btn-ghost" onclick="logout()">退出</button>`;
}

function catLabel(id) {
  const m = {education:"教育助学",medical:"医疗救助",disaster:"救灾应急",poverty:"扶贫济困",environment:"环境保护",animal:"动物保护",community:"社区公益",other:"其他"};
  return m[id] || id;
}

function statusLabel(id) {
  const m = {
    draft:"草稿", pending_review:"待审", published:"募捐中", closed:"募捐结束", completed:"已结项", cancelled:"已取消",
    pending:"待确认", confirmed:"已确认", rejected:"已拒绝", refunded:"已退款", cancelled:"已取消",
    income:"收入", matching:"匹配", expense:"支出", refund:"退款", adjust:"调账"
  };
  return m[id] || id;
}

window.CP = { api, token, setToken, clearToken, requireLogin, currentMe, logout, $, fillUserBar, catLabel, statusLabel, yuan };
