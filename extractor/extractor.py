from flask import Flask, request, jsonify
import os
import fitz  # PyMuPDF
import pytesseract
from pdf2image import convert_from_bytes
from bs4 import BeautifulSoup
from langchain_community.document_loaders import WebBaseLoader
from langchain_text_splitters import RecursiveCharacterTextSplitter

# Playwright (ใช้เฉพาะกรณีระบุ js:true หรือ fallback อัตโนมัติ)
from playwright.sync_api import sync_playwright

app = Flask(__name__)

# =========================
# Helpers
# =========================
def html_to_clean_text(html: str) -> str:
    """Clean HTML -> plain text, removing boilerplate parts."""
    soup = BeautifulSoup(html, "html.parser")

    # remove noisy sections
    for sel in ["nav", "header", "footer", "aside", "script", "style", "noscript"]:
        for tag in soup.select(sel):
            tag.decompose()

    text = soup.get_text(separator="\n")
    lines = [line.strip() for line in text.splitlines() if line.strip()]
    return "\n".join(lines)

def normalize_text(s: str) -> str:
    return "\n".join(line.strip() for line in (s or "").splitlines() if line.strip())

def render_js_page(url: str, wait_selector: str | None = None, timeout_ms: int = 20000) -> tuple[str, str | None]:
    """Render a page with Playwright (Chromium headless), return (html, title)."""
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        try:
            context = browser.new_context()
            page = context.new_page()
            page.goto(url, timeout=timeout_ms, wait_until="networkidle")
            # handle lazy-load
            page.evaluate("""() => new Promise(r => {
                let i = setInterval(() => { window.scrollBy(0, document.body.scrollHeight); }, 250);
                setTimeout(() => { clearInterval(i); r(); }, 1200);
            })""")
            if wait_selector:
                try:
                    page.wait_for_selector(wait_selector, timeout=timeout_ms)
                except Exception:
                    # ignore if not found within timeout
                    pass
            html = page.content()
            title = None
            try:
                title = page.title()
            except Exception:
                pass
            return html, title
        finally:
            browser.close()

# =========================
# Health
# =========================
@app.get("/health")
@app.get("/healthz")
def health():
    return jsonify({"status": "ok"}), 200

# =========================
# PDF -> Text (with OCR fallback)
# =========================
@app.post("/extract-text")
def extract_text():
    file = request.files.get('file')
    if not file:
        return jsonify({"error": "No file provided"}), 400

    try:
        data = file.read()
        pdf = fitz.open(stream=data, filetype="pdf")
        full_text = ""
        langs = os.getenv("TESS_LANGS", "tha+eng")

        for page in pdf:
            text = page.get_text()
            if text and text.strip():
                full_text += text
            else:
                # OCR only the page lacking text
                images = convert_from_bytes(
                    data, dpi=300,
                    first_page=page.number + 1,
                    last_page=page.number + 1
                )
                for image in images:
                    ocr_text = pytesseract.image_to_string(image, lang=langs)
                    full_text += ocr_text

        return jsonify({"text": full_text})
    except Exception as e:
        return jsonify({"error": str(e)}), 500

# =========================
# URL -> Documents (clean text + optional split, JS-render optional/fallback)
# =========================
@app.post("/web-to-doc")
def web_to_doc():
    """
    JSON:
    {
      "url": "https://example.com",               // หรือ
      "urls": ["https://a.com","https://b.com"],
      "split": false,                              // true = แบ่ง chunk
      "chunk_size": 1200,
      "chunk_overlap": 200,
      "js": false,                                 // true = บังคับเรนเดอร์ด้วย Playwright
      "wait_selector": "article, main, .prose",    // (optional) ตัวบ่งชี้ว่า content ขึ้นแล้ว
      "fallback_threshold": 500                    // (optional) ถ้า text สั้นกว่านี้จะ fallback ไป JS อัตโนมัติ
    }
    """
    payload = request.get_json(silent=True) or {}
    url = payload.get("url")
    urls = payload.get("urls") or ([url] if url else [])
    if not urls:
        return jsonify({"error": "Missing 'url' or 'urls'"}), 400

    do_split = bool(payload.get("split", False))
    chunk_size = int(payload.get("chunk_size", 1200))
    chunk_overlap = int(payload.get("chunk_overlap", 200))
    force_js = bool(payload.get("js", False))
    wait_selector = payload.get("wait_selector")
    fallback_threshold = int(payload.get("fallback_threshold", 500))

    try:
        texts_with_meta = []

        for u in urls:
            meta = {"source": u}
            text = ""
            title = None

            if force_js:
                html, title = render_js_page(u, wait_selector=wait_selector)
                text = normalize_text(html_to_clean_text(html))
            else:
                # 1) ลองโหลดแบบธรรมดาก่อน (HTML ดิบ)
                loader = WebBaseLoader([u])
                docs = loader.load()
                raw = (docs[0].page_content or "") if docs else ""
                text = normalize_text(
                    html_to_clean_text(raw) if ("<" in raw and ">" in raw) else raw
                )
                # 2) ถ้าสั้นผิดปกติ → fallback ไป JS-render
                if len(text) < fallback_threshold:
                    html, title = render_js_page(u, wait_selector=wait_selector)
                    text = normalize_text(html_to_clean_text(html))

            if title and "title" not in meta:
                meta["title"] = title

            if text:
                texts_with_meta.append((text, meta))

        # สร้าง documents
        documents = []
        if do_split:
            splitter = RecursiveCharacterTextSplitter(
                chunk_size=chunk_size,
                chunk_overlap=chunk_overlap,
                separators=["\n\n", "\n", " ", ""]
            )
            for text, meta in texts_with_meta:
                chunks = splitter.split_text(text)
                for i, c in enumerate(chunks):
                    documents.append({"page_content": c, "metadata": {**meta, "chunk_index": i}})
        else:
            for text, meta in texts_with_meta:
                documents.append({"page_content": text, "metadata": meta})

        return jsonify({"count": len(documents), "documents": documents})
    except Exception as e:
        return jsonify({"error": str(e)}), 500

# =========================
# Raw HTML -> Documents (useful if you render via Playwright elsewhere)
# =========================
@app.post("/web-to-doc-html")
def web_to_doc_html():
    """
    JSON:
    {
      "html": "<!doctype html> ...",
      "meta": {"source": "...", "title": "..."},
      "split": false,
      "chunk_size": 1200,
      "chunk_overlap": 200
    }
    """
    payload = request.get_json(silent=True) or {}
    html = payload.get("html", "")
    meta = payload.get("meta", {}) or {}
    do_split = bool(payload.get("split", False))
    chunk_size = int(payload.get("chunk_size", 1200))
    chunk_overlap = int(payload.get("chunk_overlap", 200))

    if not html.strip():
        return jsonify({"error": "Missing 'html'"}), 400

    try:
        text = normalize_text(html_to_clean_text(html))

        out = []
        if do_split:
            splitter = RecursiveCharacterTextSplitter(
                chunk_size=chunk_size,
                chunk_overlap=chunk_overlap,
                separators=["\n\n", "\n", " ", ""]
            )
            chunks = splitter.split_text(text)
            for i, c in enumerate(chunks):
                out.append({
                    "page_content": c,
                    "metadata": {**meta, "chunk_index": i}
                })
        else:
            out.append({
                "page_content": text,
                "metadata": meta
            })

        return jsonify({"count": len(out), "documents": out})
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == "__main__":
    # 0.0.0.0 for docker
    app.run(host="0.0.0.0", port=5002)
