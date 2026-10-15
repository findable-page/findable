const NEEDS_CLOSE = new Set(["script", "style", "svg", "title", "h1", "h2"]);
const NAME = /^[a-zA-Z][\w:-]*/;
const ESCAPED = "&lt;";

function tagEnd(html: string, from: number): number {
  let i = from;
  let previous = "";
  while (i < html.length) {
    const ch = html[i] as string;
    if (ch === ">") return i;
    if ((ch === '"' || ch === "'") && previous === "=") {
      const closing = html.indexOf(ch, i + 1);
      if (closing !== -1) i = closing;
    }
    if (ch !== " " && ch !== "\t" && ch !== "\n" && ch !== "\r") previous = ch;
    i++;
  }
  return -1;
}

// Turns markup that could make the regex-based analysers backtrack (tags that never end, raw-text elements that
// never close, stray "<" inside a tag) into escaped text, in one linear pass.
export function sanitizeHtml(html: string): string {
  const lower = html.toLowerCase();
  const lastGt = html.lastIndexOf(">");
  const closeAt = new Map<string, number>();
  const out: string[] = [];
  let copied = 0;
  let i = 0;
  const escapeAt = (index: number) => {
    out.push(html.slice(copied, index), ESCAPED);
    copied = index + 1;
  };
  for (;;) {
    const lt = html.indexOf("<", i);
    if (lt === -1) break;
    if (html.startsWith("<!--", lt)) {
      const end = html.indexOf("-->", lt + 4);
      if (end === -1) {
        for (let j = lt; j !== -1; j = html.indexOf("<", j + 1)) escapeAt(j);
        break;
      }
      i = end + 3;
      continue;
    }
    const second = html[lt + 1];
    if (second === "?" || (second === "!" && html.slice(lt + 2, lt + 9).toLowerCase() === "doctype")) {
      if (lt > lastGt) {
        escapeAt(lt);
        i = lt + 1;
      } else {
        i = html.indexOf(">", lt) + 1;
      }
      continue;
    }
    const closing = second === "/";
    const name = NAME.exec(html.slice(lt + (closing ? 2 : 1), lt + 40))?.[0];
    if (!name) {
      escapeAt(lt);
      i = lt + 1;
      continue;
    }
    const end = tagEnd(html, lt + 1 + name.length + (closing ? 1 : 0));
    if (end === -1) {
      for (let j = lt; j !== -1; j = html.indexOf("<", j + 1)) escapeAt(j);
      break;
    }
    const lowerName = name.toLowerCase();
    const selfClosing = html[end - 1] === "/";
    if (!closing && !selfClosing && NEEDS_CLOSE.has(lowerName)) {
      let close = closeAt.get(lowerName);
      if (close === undefined || (close !== -1 && close < end)) {
        close = lower.indexOf(`</${lowerName}`, end);
        closeAt.set(lowerName, close);
      }
      if (close === -1) {
        escapeAt(lt);
        i = lt + 1;
        continue;
      }
    }
    for (let j = html.indexOf("<", lt + 1); j !== -1 && j < end; j = html.indexOf("<", j + 1)) escapeAt(j);
    i = end + 1;
    if (!closing && !selfClosing && (lowerName === "script" || lowerName === "style")) {
      const close = closeAt.get(lowerName);
      if (close !== undefined && close > end) i = close;
    }
  }
  out.push(html.slice(copied));
  return out.join("");
}
