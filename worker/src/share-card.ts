import { escapeHtml } from "./learn";
import { cardTheme, scoreQuip } from "./quips";

export interface CardScores {
  host: string;
  overall: number;
  seo: number;
  aeo: number;
  geo: number;
}

const PILLARS = [
  { key: "seo", label: "SEO", color: "#0E7C4A" },
  { key: "aeo", label: "AEO", color: "#4F2BFF" },
  { key: "geo", label: "GEO", color: "#D1480D" },
] as const;

function ringColor(score: number): string {
  if (score >= 70) return "#0E7C4A";
  if (score >= 50) return "#B76E00";
  return "#D1480D";
}

function bar(label: string, color: string, score: number): string {
  return `<div style="display:flex;align-items:center;margin-top:22px">
<div style="display:flex;width:90px;font-size:30px;font-weight:800;color:${color}">${label}</div>
<div style="display:flex;width:420px;height:26px;border-radius:13px;background:#ECEEF5">
<div style="display:flex;width:${Math.round(4.2 * score)}px;height:26px;border-radius:13px;background:${color}"></div></div>
<div style="display:flex;margin-left:22px;font-size:30px;font-weight:800">${score}</div></div>`;
}

export function shareCardHtml(s: CardScores): string {
  const host = escapeHtml(s.host.length > 32 ? `${s.host.slice(0, 31)}…` : s.host);
  const ring = ringColor(s.overall);
  const bars = PILLARS.map((p) => bar(p.label, p.color, s[p.key])).join("");
  const theme = cardTheme(s.host);
  const quip = escapeHtml(scoreQuip(s.overall, s.host));
  return `<div style="display:flex;flex-direction:column;width:1200px;height:630px;padding:56px 72px;background:${theme.background};font-family:Inter;color:#0F1222">
<div style="display:flex;font-size:34px;font-weight:800;color:${theme.accent}">findable</div>
<div style="display:flex;font-size:54px;font-weight:800;margin-top:18px">${host}</div>
<div style="display:flex;flex:1;align-items:center">
<div style="display:flex;flex-direction:column;align-items:center;justify-content:center;width:300px;height:300px;border-radius:150px;border:26px solid ${ring};background:#ffffff">
<div style="display:flex;font-size:110px;font-weight:800;line-height:1">${s.overall}</div>
<div style="display:flex;font-size:28px;color:#5B6078">/100</div></div>
<div style="display:flex;flex-direction:column;margin-left:72px">${bars}</div></div>
<div style="display:flex;align-items:center;font-size:28px;font-weight:700">
<div style="display:flex;flex-grow:1;color:#3A3F55">${quip}</div>
<div style="display:flex;color:#ffffff;background:${theme.accent};padding:10px 24px;border-radius:14px">findable.page</div></div></div>`;
}
