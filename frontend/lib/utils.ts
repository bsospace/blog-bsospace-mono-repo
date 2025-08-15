import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"
import { nanoid } from 'nanoid';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}


export function getnerarteIdFromUrl(url: string) {
  const regex = /\/nerarte\/([a-zA-Z0-9]+)/;
  const match = url.match(regex);
  return match ? match[1] : null;
}

export function getnerateId() {
  const first = nanoid(8)
  return first
}

// Fingerprint v2 — stronger & privacy-aware
// - ใช้ SHA-256 ผ่าน Web Crypto
// - รวมหลายแหล่ง entropy: UA, screen, timezone, canvas, WebGL, audio, etc.
// - มี salt (เลือกส่ง) ไว้หมุนค่าตามนโยบาย (เช่น รายวัน/รายโดเมน/รายผู้ใช้)
// - มี fallback ถ้าไม่รองรับบาง API

type Options = {
  salt?: string;            // หมุนรหัสได้ เช่น "my-app@2025-08-15"
  includeAudio?: boolean;   // ปิดได้หากกังวลเรื่อง privacy/consent
  includeWebGL?: boolean;
  includeCanvas?: boolean;
};

export async function generateFingerprint(opts: Options = {}): Promise<string> {
  const {
    salt = "",
    includeAudio = true,
    includeWebGL = true,
    includeCanvas = true,
  } = opts;

  // --- helpers ---
  const toBase64Url = (buf: ArrayBuffer) => {
    const b = String.fromCharCode(...new Uint8Array(buf));
    const base64 = btoa(b);
    return base64.replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
  };

  async function sha256(input: string): Promise<string> {
    if (crypto?.subtle?.digest) {
      const enc = new TextEncoder().encode(input);
      const hashBuf = await crypto.subtle.digest("SHA-256", enc);
      return toBase64Url(hashBuf); // สั้น แน่น อ่านง่าย
    }
    // very old fallback (ไม่แนะนำ แต่ยังดีกว่าเดิม)
    let h = 0;
    for (let i = 0; i < input.length; i++) {
      h = (h << 5) - h + input.charCodeAt(i);
      h |= 0;
    }
    return Math.abs(h).toString(36);
  }

  function safe<T>(fn: () => T, fallback: T): T {
    try { return fn(); } catch { return fallback; }
  }

  // --- core signals (เสถียร + ไม่รุกรานเกินไป) ---
  const signals: Record<string, unknown> = {
    // Navigator & platform
    ua: safe(() => navigator.userAgent ?? "", ""),
    platform: safe(() => navigator.platform ?? "", ""),
    lang: safe(() => navigator.language ?? "", ""),
    languages: safe(() => (navigator.languages ?? []).join(","), ""),
    hc: safe(() => (navigator as any).hardwareConcurrency ?? "", ""),
    dm: safe(() => (navigator as any).deviceMemory ?? "", ""),
    maxTouch: safe(
      () => (navigator as any).maxTouchPoints ?? (("ontouchstart" in window) ? 1 : 0),
      0
    ),
    // Screen
    screen: safe(() => `${screen.width}x${screen.height}@${window.devicePixelRatio || 1}`, ""),
    colorDepth: safe(() => screen.colorDepth ?? 0, 0),
    // Timezone
    tzOffset: new Date().getTimezoneOffset(),
    tzIntl: safe(() => Intl.DateTimeFormat().resolvedOptions().timeZone ?? "", ""),
    // Permissions (แค่ชื่อ ไม่เรียกขอสิทธิ์)
    // หมายเหตุ: บาง browser ไม่รองรับ
    permissions: await (async () => {
      const names = ["geolocation", "notifications", "push", "camera", "microphone", "clipboard-read", "clipboard-write"];
      const results: string[] = [];
      if (!navigator.permissions?.query) return "";
      for (const n of names) {
        try {
          const s = await navigator.permissions.query({ name: n as PermissionName });
          results.push(`${n}:${s.state}`);
        } catch {
          results.push(`${n}:unknown`);
        }
      }
      return results.join("|");
    })(),
  };

  // --- Canvas fingerprint (เบา ๆ พอประมาณ) ---
  if (includeCanvas) {
    try {
      const canvas = document.createElement("canvas");
      canvas.width = 240; canvas.height = 60;
      const ctx = canvas.getContext("2d");
      if (ctx) {
        ctx.textBaseline = "alphabetic";
        ctx.fillStyle = "#f60";
        ctx.fillRect(0, 0, 240, 60);
        ctx.fillStyle = "#069";
        ctx.font = "16px 'Segoe UI', Arial";
        ctx.fillText("FP2: canvas 🍀", 10, 20);
        ctx.strokeStyle = "#ff0";
        ctx.arc(120, 30, 15, 0, Math.PI * 2);
        ctx.stroke();

        const data = canvas.toDataURL();
        // เก็บเฉพาะส่วน base64 ช่วงต้น ๆ ลดขนาด
        const idx = data.indexOf(",") + 1;
        signals.canvas = data.substring(idx, idx + 80);
      } else {
        signals.canvas = "no-ctx";
      }
    } catch {
      signals.canvas = "err";
    }
  }

  // --- WebGL vendor/renderer (ถ้าเปิดเผย; บาง browser จะบัง) ---
  if (includeWebGL) {
    try {
      const canvas = document.createElement("canvas");
      const gl = canvas.getContext("webgl") || canvas.getContext("experimental-webgl") as WebGLRenderingContext;
      if (gl) {
        const dbg = gl.getExtension("WEBGL_debug_renderer_info");
        const vendor = dbg ? gl.getParameter((dbg as any).UNMASKED_VENDOR_WEBGL) : gl.getParameter(gl.VENDOR);
        const renderer = dbg ? gl.getParameter((dbg as any).UNMASKED_RENDERER_WEBGL) : gl.getParameter(gl.RENDERER);
        signals.webgl = `${vendor}|${renderer}`;
      } else {
        signals.webgl = "no-webgl";
      }
    } catch {
      signals.webgl = "err";
    }
  }

  // --- Lightweight Audio fingerprint (ไม่เล่นเสียงจริง; OfflineAudioContext) ---
  if (includeAudio) {
    try {
      // บางบราวเซอร์/โหมดต้อง user gesture; ถ้าใช้ไม่ได้ให้ข้าม
      const OfflineCtx = (window as any).OfflineAudioContext || (window as any).webkitOfflineAudioContext;
      if (OfflineCtx) {
        const ctx = new OfflineCtx(1, 44100, 44100);
        const osc = ctx.createOscillator();
        const comp = ctx.createDynamicsCompressor();
        osc.type = "triangle";
        osc.frequency.value = 1000;
        osc.connect(comp);
        comp.connect(ctx.destination);
        osc.start(0);
        const rendered = await ctx.startRendering();
        // สุ่ม sample เล็ก ๆ มาสร้าง signature
        const ch = rendered.getChannelData(0);
        let acc = 0;
        for (let i = 0; i < ch.length; i += 441) { // ทุก ~0.01s
          acc += Math.round((ch[i] || 0) * 1e6);
        }
        signals.audio = `a:${acc}`;
      } else {
        signals.audio = "no-offline-audio";
      }
    } catch {
      signals.audio = "err";
    }
  }

  // --- รวม & ทำ normalization ---
  const orderedKeys = Object.keys(signals).sort();
  const payload = orderedKeys.map(k => `${k}=${String(signals[k])}`).join("&");

  // --- เพิ่ม salt (ถ้าอยากหมุนรหัสเป็นรายวัน/รายโดเมน/รายผู้ใช้) ---
  const material = `${payload}||salt:${salt}`;

  // --- สร้างแฮช SHA-256 ---
  const fp = await sha256(material);

  return fp; // base64url ของ SHA-256 ทั้งก้อน ~43 ตัวอักษร
}



export const formatDate = (dateString: string) => {
  if (!dateString || dateString === "0001-01-01T00:00:00Z") return 'Not set';
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  });
};