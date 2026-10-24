import type { EtherScribeAPI, JobResult, SectionCatalog } from "../preload/preload";
declare global { interface Window { etherscribe: EtherScribeAPI } }
const api = window.etherscribe;
const el = <T extends HTMLElement = HTMLElement>(id: string) => document.getElementById(id) as T;
const labels: Record<string,string> = {device_info:"機器概要",system_settings:"システム設定",network:"ネットワーク",vlans:"VLAN",lag:"LAG",stp:"STP",objects:"オブジェクト",policies:"ポリシー",nat:"NAT",vpn:"VPN",security_profiles:"セキュリティ",ha:"HA",logging:"ログ・監視"};
let catalog: SectionCatalog;
let mode: "convert" | "compare" = "convert";
let busy = false, ready = false;
const paths = {convert:"",before:"",after:""};
const format = el<HTMLSelectElement>("format");
const preset = el<HTMLSelectElement>("preset");
const run = el<HTMLButtonElement>("run");
const preview = el<HTMLIFrameElement>("preview");
function updateRun() { run.disabled = busy || !ready || (mode === "convert" ? !paths.convert || !selected().length : !paths.before || !paths.after); }
function selected() { return Array.from(document.querySelectorAll<HTMLInputElement>("#sections input:checked"), x => x.value); }
function count() { el("sectionCount").textContent = `${selected().length} 項目`; updateRun(); }
function resetResult() { preview.hidden = true; preview.removeAttribute("srcdoc"); el("deviceSummary").hidden = true; el("metrics").hidden = true; el("metrics").replaceChildren(); el("outputPath").textContent = ""; }
function message(text: string, error = false) { const result = el("result"); result.className = `result-message${error ? " error" : ""}`; result.textContent = text; }
function switchMode(next: typeof mode) {
 if (busy) return;
 mode = next;
 const compare = mode === "compare";
 el("convertInput").hidden = compare; el("compareInput").hidden = !compare; el("convertOptions").hidden = compare;
 el("tabConvert").classList.toggle("active", !compare); el("tabCompare").classList.toggle("active", compare);
 el("tabConvert").setAttribute("aria-pressed", String(!compare)); el("tabCompare").setAttribute("aria-pressed", String(compare));
 format.querySelector<HTMLOptionElement>('option[value="excel"]')!.disabled = compare;
 if (compare && format.value === "excel") format.value = "html";
 el("pageTitle").textContent = compare ? "設定差分" : "パラメータシート";
 el("pageDescription").textContent = compare ? "設定の変更点を、ひとつのレポートに。" : "機器の設定を、読みやすいドキュメントへ。";
 run.textContent = compare ? "設定差分を生成して保存" : "パラメータシートを生成して保存";
 resetResult(); message(compare ? "変更前と変更後の設定ファイルを選択してください。" : "設定ファイルと出力内容を選択してください。"); el("resultBadge").textContent = "準備"; updateRun();
}
el("tabConvert").onclick = () => switchMode("convert"); el("tabCompare").onclick = () => switchMode("compare");
for (const [key,id] of [["convert","Convert"],["before","Before"],["after","After"]] as const) {
 el(`pick${id}`).onclick = async () => {
  try { const p = await api.openFile(); if (p) { paths[key] = p; el(key === "convert" ? "convertPath" : `${key}Path`).textContent = p; resetResult(); message("設定を選択しました。出力内容を確認して生成してください。"); el("resultBadge").textContent = "準備"; updateRun(); } }
  catch { message("ファイル選択に失敗しました。もう一度選択してください。", true); }
 };
}
preset.onchange = () => {
 const p = catalog.presets.find(p => p.key === preset.value)!;
 document.querySelectorAll<HTMLInputElement>("#sections input").forEach(x => { x.checked = p.sections.includes(x.value); });
 el("presetDescription").textContent = p.description; count();
};
run.onclick = async () => {
 if (run.disabled) return;
 busy = true; resetResult(); message("保存先を選択すると解析を開始します…"); el("resultBadge").textContent = "処理中";
 document.querySelectorAll<HTMLButtonElement | HTMLSelectElement | HTMLInputElement>("button, select, input").forEach(x => { x.disabled = true; });
 try {
  const res: JobResult = mode === "convert" ? await api.convert({path:paths.convert,format:format.value,preset:preset.value,sections:selected()}) : await api.compare({before:paths.before,after:paths.after,format:format.value});
  if (res.cancelled) { message("保存をキャンセルしました。"); el("resultBadge").textContent = "取消"; return; }
  message(mode === "convert" ? "パラメータシートを保存しました。" : "差分レポートを保存しました。"); el("resultBadge").textContent = "完了"; el("outputPath").textContent = res.outputPath ?? "";
  const summary = res.summary ?? {};
  el("deviceSummary").textContent = [summary.hostname,summary.model,summary.osVersion].filter(Boolean).join(" · "); el("deviceSummary").hidden = !el("deviceSummary").textContent;
  const names: Record<string,string> = {interfaces:"インターフェース",policies:"ポリシー",routes:"ルート",coverage:"未対応・解析範囲",changes:"変更件数",vlans:"VLAN"};
  for (const [key,label] of Object.entries(names)) { if (typeof summary[key] !== "number") continue; const box = document.createElement("div"); box.className = "metric"; const n = document.createElement("strong"); n.textContent = String(summary[key]); box.append(n,document.createTextNode(label)); el("metrics").append(box); }
  el("metrics").hidden = false;
  if (format.value === "html" && res.outputPath) {
   try { preview.srcdoc = await api.previewHtml(res.outputPath); preview.hidden = false; }
   catch { message("保存は完了しましたが、プレビューを読み込めませんでした。保存先のファイルを開いてください。", true); }
  }
 } catch { message("処理に失敗しました。対応形式の設定ファイルと、保存先の書き込み権限を確認してください。",true); el("resultBadge").textContent = "エラー"; }
 finally { busy = false; document.querySelectorAll<HTMLButtonElement | HTMLSelectElement | HTMLInputElement>("button, select, input").forEach(x => { x.disabled = false; }); format.querySelector<HTMLOptionElement>('option[value="excel"]')!.disabled = mode === "compare"; updateRun(); }
};
async function initialize() {
 try {
  const [info, choices] = await Promise.all([api.version(),api.sections()]); catalog = choices;
  el("engineStatus").textContent = `● エンジン接続済み · v${info.version}`;
  preset.replaceChildren();
  const order = ["ops","overview","connectivity","policy","all","switch-overview","switch-l2","switch-l3","switch-ops","switch-all"];
  choices.presets.sort((a,b) => order.indexOf(a.key)-order.indexOf(b.key)).forEach(p => preset.add(new Option(p.label,p.key)));
  choices.sections.forEach(key => { const label = document.createElement("label"); const input = document.createElement("input"); input.type = "checkbox"; input.value = key; input.onchange = count; label.append(input,document.createTextNode(labels[key] ?? key)); el("sections").append(label); });
  preset.disabled = false; preset.value = "ops"; preset.dispatchEvent(new Event("change")); ready = true; updateRun();
 } catch { el("engineStatus").textContent = "エンジンに接続できません。アプリを再起動してください。"; message("エンジンの起動に失敗しました。", true); }
}
api.onEngineError(() => { ready = false; el("engineStatus").textContent = "エンジン接続エラー"; updateRun(); });
void initialize();
