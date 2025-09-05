import { NextResponse } from "next/server"

type UnfurlResult = {
  url: string
  title?: string
  description?: string
  image?: string
}

function extractMeta(html: string, regex: RegExp): string | undefined {
  const match = html.match(regex)
  return match?.[1]?.trim()
}

export async function POST(request: Request) {
  try {
    const { url } = await request.json()
    if (!url || typeof url !== "string") {
      return NextResponse.json({ error: "Invalid url" }, { status: 400 })
    }

    // Basic URL validation
    let target: URL
    try {
      target = new URL(url)
    } catch {
      // Try adding https if missing protocol
      try {
        target = new URL(`https://${url}`)
      } catch {
        return NextResponse.json({ error: "Invalid URL format" }, { status: 400 })
      }
    }

    const controller = new AbortController()
    const timeout = setTimeout(() => controller.abort(), 8000)

    const res = await fetch(target.toString(), {
      method: "GET",
      headers: {
        // Use a common UA to avoid some servers blocking requests
        "User-Agent":
          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124 Safari/537.36",
        Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
      },
      redirect: "follow",
      signal: controller.signal,
      cache: "no-store",
    })
    clearTimeout(timeout)

    if (!res.ok) {
      return NextResponse.json({ error: `Failed to fetch: ${res.status}` }, { status: 502 })
    }

    const html = await res.text()

    // Extract common OG tags and fallbacks
    const title =
      extractMeta(html, /<meta[^>]+property=["']og:title["'][^>]*content=["']([^"']+)["'][^>]*>/i) ||
      extractMeta(html, /<title[^>]*>([^<]*)<\/title>/i) ||
      extractMeta(html, /<meta[^>]+name=["']twitter:title["'][^>]*content=["']([^"']+)["'][^>]*>/i)

    const description =
      extractMeta(html, /<meta[^>]+property=["']og:description["'][^>]*content=["']([^"']+)["'][^>]*>/i) ||
      extractMeta(html, /<meta[^>]+name=["']description["'][^>]*content=["']([^"']+)["'][^>]*>/i) ||
      extractMeta(html, /<meta[^>]+name=["']twitter:description["'][^>]*content=["']([^"']+)["'][^>]*>/i)

    let image =
      extractMeta(html, /<meta[^>]+property=["']og:image["'][^>]*content=["']([^"']+)["'][^>]*>/i) ||
      extractMeta(html, /<meta[^>]+name=["']twitter:image["'][^>]*content=["']([^"']+)["'][^>]*>/i)

    // Resolve relative image URLs
    if (image) {
      try {
        const imgUrl = new URL(image, target)
        image = imgUrl.toString()
      } catch {
        // keep original if cannot resolve
      }
    }

    const result: UnfurlResult = {
      url: target.toString(),
      title: title || undefined,
      description: description || undefined,
      image: image || undefined,
    }

    return NextResponse.json(result, { status: 200 })
  } catch (err) {
    return NextResponse.json({ error: "Unexpected error" }, { status: 500 })
  }
}


