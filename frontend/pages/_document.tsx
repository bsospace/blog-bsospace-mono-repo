import { Html, Head, Main, NextScript } from 'next/document'

export default function Document() {
  return (
    <Html lang="en">
        <Head>
          {/* Favicon using logo.svg */}
          <link rel="icon" type="image/svg+xml" href="/logo.svg" />
          {/* Optional PNG fallback for older browsers */}
          <link rel="alternate icon" href="/favicon-32x32.png" sizes="32x32" />
        </Head>
      <body>
        <Main />
        <NextScript />
      </body>
    </Html>
  )
}
