import * as React from "react"
import type { NodeViewProps } from "@tiptap/react"
import { NodeViewWrapper } from "@tiptap/react"

export function LinkPreviewNodeView(props: NodeViewProps) {
  const { node } = props
  const href: string = node.attrs.href
  const title: string | null = node.attrs.title
  const description: string | null = node.attrs.description
  const image: string | null = node.attrs.image

  let hostname = ""
  try { hostname = new URL(href).hostname.replace("www.", "") } catch {}

  return (
    <NodeViewWrapper className="group not-prose my-3 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden bg-white dark:bg-[#0B0F16] shadow-sm hover:shadow-md transition-all duration-300 hover:-translate-y-0.5" data-type="link-preview">
      {image ? (
        <div className="relative h-44 overflow-hidden bg-gray-100 dark:bg-gray-900">
          <img src={image} alt="preview" className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-105" />
          <div className="absolute inset-0 bg-gradient-to-t from-black/30 via-transparent to-transparent" />
        </div>
      ) : null}
      <div className="p-4 space-y-2">
        <div className="flex items-center gap-2">
          <span className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">{hostname || href}</span>
        </div>
        <a href={href} target="_blank" rel="noopener noreferrer" className="block">
          <h3 className="font-semibold text-gray-900 dark:text-gray-100 text-base leading-snug hover:text-blue-600 dark:hover:text-blue-400 transition-colors line-clamp-2">
            {title || href}
          </h3>
        </a>
        {description ? (
          <p className="text-sm text-gray-600 dark:text-gray-300 leading-relaxed line-clamp-2">{description}</p>
        ) : null}
      </div>
    </NodeViewWrapper>
  )
}

export default LinkPreviewNodeView


