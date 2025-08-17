/** @type {import('next').NextConfig} */
const nextConfig = {
  compress: true,
  poweredByHeader: false,
  generateEtags: true,

  images: {
    domains: [
      'image-service.bsospace.com',
      'lh3.googleusercontent.com',
      'cdn.discordapp.com',
      'avatars.githubusercontent.com',
    ],
    formats: ['image/webp', 'image/avif'],
    deviceSizes: [640, 750, 828, 1080, 1200, 1920, 2048, 3840],
    imageSizes: [16, 32, 48, 64, 96, 128, 256, 384],
    minimumCacheTTL: 60,
  },

  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          { key: 'X-Content-Type-Options', value: 'nosniff' },
          { key: 'X-Frame-Options', value: 'DENY' },
          { key: 'X-XSS-Protection', value: '1; mode=block' },
          { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
          // Help browsers allow cross-origin images/fonts to be embedded safely
          { key: 'Cross-Origin-Resource-Policy', value: 'cross-origin' },
        ],
      },
      {
        source: '/api/(.*)',
        headers: [
          { key: 'Cache-Control', value: 'no-store, max-age=0' },
        ],
      },
    ]
  },

  async redirects() {
    return [
      { source: '/', destination: '/home', permanent: true },
      {
        source: '/:path*',
        has: [{ type: 'host', value: 'www.blog.bsospace.com' }],
        destination: 'https://blog.bsospace.com/:path*',
        permanent: true,
      },
    ]
  },

  async rewrites() {
    return [
      { source: '/sitemap.xml', destination: '/api/sitemap' },
      { source: '/robots.txt', destination: '/api/robots' },
    ]
  },

  experimental: {
    optimizePackageImports: ['@/components/ui'],
  },
}

export default nextConfig
