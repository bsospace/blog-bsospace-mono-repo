import getBrowserFingerprint from 'get-browser-fingerprint';

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

    try {
        // ใช้ get-browser-fingerprint library แทน custom implementation
        const baseFingerprint = await getBrowserFingerprint({
            hardwareOnly: false, // ใช้ข้อมูลทั้งหมดเพื่อความแม่นยำ
            debug: false
        });

        // ถ้ามี salt ให้เพิ่มเข้าไปใน fingerprint
        if (salt) {
            const encoder = new TextEncoder();
            const data = encoder.encode(baseFingerprint.toString() + salt);
            const hashBuffer = await crypto.subtle.digest('SHA-256', data);
            const hashArray = Array.from(new Uint8Array(hashBuffer));
            const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
            return hashHex;
        }

        return baseFingerprint.toString();
    } catch (error) {
        console.error('Error generating fingerprint:', error);
        
        // fallback: สร้าง fingerprint แบบง่ายจากข้อมูลพื้นฐาน (เฉพาะใน browser)
        if (typeof window !== 'undefined' && typeof navigator !== 'undefined' && typeof screen !== 'undefined') {
            const fallbackData = {
                ua: navigator.userAgent || '',
                platform: navigator.platform || '',
                screen: `${screen.width}x${screen.height}`,
                timezone: new Date().getTimezoneOffset(),
                salt: salt
            };
            
            const fallbackString = JSON.stringify(fallbackData);
            const encoder = new TextEncoder();
            const data = encoder.encode(fallbackString);
            const hashBuffer = await crypto.subtle.digest('SHA-256', data);
            const hashArray = Array.from(new Uint8Array(hashBuffer));
            const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
            
            return hashHex;
        }
        
        // ถ้าไม่ใช่ browser environment ให้ return ค่าว่าง
        return 'no-fingerprint-available';
    }
}
