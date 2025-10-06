import * as React from "react"
import type { NodeViewProps } from "@tiptap/react"
import { NodeViewWrapper } from "@tiptap/react"
import { X } from "lucide-react"
import Image from "next/image"

// Add CSS for mobile responsiveness
const mobileStyles = `
  @media (max-width: 768px) {
    [data-type="link-preview"] {
      width: 100% !important;
    }
  }
`

// Inject styles
if (typeof document !== 'undefined') {
  const styleElement = document.createElement('style')
  styleElement.textContent = mobileStyles
  if (!document.head.querySelector('style[data-link-preview-mobile]')) {
    styleElement.setAttribute('data-link-preview-mobile', 'true')
    document.head.appendChild(styleElement)
  }
}

// Alignment Icons
const AlignLeftIcon: React.FC = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
    <path d="M3 3h18v2H3V3zm0 4h12v2H3V7zm0 4h18v2H3v-2zm0 4h12v2H3v-2z"/>
  </svg>
)

const AlignCenterIcon: React.FC = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
    <path d="M3 3h18v2H3V3zm2 4h14v2H5V7zm-2 4h18v2H3v-2zm2 4h14v2H5v-2z"/>
  </svg>
)

const AlignRightIcon: React.FC = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
    <path d="M3 3h18v2H3V3zm6 4h12v2H9V7zm-6 4h18v2H3v-2zm6 4h12v2H9v-2z"/>
  </svg>
)

export function LinkPreviewNodeView(props: NodeViewProps) {
  const { node, updateAttributes, editor } = props
  const href: string = node.attrs.href
  const title: string | null = node.attrs.title
  const description: string | null = node.attrs.description
  const image: string | null = node.attrs.image
  const width: number = node.attrs.width || 100
  const align: 'left' | 'center' | 'right' = node.attrs.align || 'left'

  let hostname = ""
  try { hostname = new URL(href).hostname.replace("www.", "") } catch {}

  const handleWidthChange = (newWidth: number) => {
    updateAttributes({ width: newWidth })
  }

  const handleAlignChange = (newAlign: 'left' | 'center' | 'right') => {
    updateAttributes({ align: newAlign })
  }

  const handleRemove = () => {
    editor.chain()
      .focus()
      .deleteRange({ from: props.getPos(), to: props.getPos() + node.nodeSize })
      .run()
  }

  // Check if editor is in edit mode (not preview mode)
  const isEditable = editor.isEditable

  return (
    <NodeViewWrapper 
      className={`group not-prose rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden bg-white dark:bg-gray-800 shadow-sm hover:shadow-md transition-all duration-300 hover:-translate-y-0.5 w-full md:w-auto my-3`}
      data-type="link-preview"
      style={{ 
        width: `${width}%`,
        maxWidth: '100%',
        marginLeft: align === 'center' ? 'auto' : align === 'right' ? 'auto' : '0',
        marginRight: align === 'center' ? 'auto' : align === 'left' ? 'auto' : '0'
      }}
    >
      {image ? (
        <div className="relative h-44 overflow-hidden bg-gray-100 dark:bg-gray-900">
          <Image src={image} alt="preview" fill className="object-cover transition-transform duration-300 group-hover:scale-105" unoptimized />
          {isEditable && (
            <>
              <button
                onClick={() => handleRemove()}
                className="absolute top-2 right-2 p-1 bg-white rounded-full shadow-md hover:shadow-lg transition  cursor-pointer z-50"
                title="Remove image"
              >
                <X className="w-4 h-4 text-gray-600" />
              </button>
            </>
          )}
          <div className="absolute inset-0 bg-gradient-to-t from-black/30 via-transparent to-transparent" />
        </div>
      ) : null}
      <div className="p-4 space-y-2">
        <div className="flex items-center gap-2">
          <span className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">{hostname || href}</span>
        </div>
        <a href={href} target="_blank" rel="noopener noreferrer" className="block">
          <h3 className="font-semibold text-gray-900 dark:text-gray-100 text-base leading-snug hover:text-orange-500 dark:hover:text-orange-400 transition-colors line-clamp-2">
            {title || href}
          </h3>
        </a>
        {description ? (
          <p className="text-sm text-gray-600 dark:text-gray-300 leading-relaxed line-clamp-2">{description}</p>
        ) : null}
      </div>
      
      {/* Resize Controls - Only show in edit mode */}
      {isEditable && (
        <div className="p-2 bg-gray-50 flex justify-between items-center dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 space-y-2">
          {/* Width Controls */}
          <div className="flex items-center">
            <input
              id="link-preview-width-slider"
              type="range"
              min={10}
              max={100}
              value={width}
              onChange={e => handleWidthChange(Number(e.target.value))}
              className="w-32"
            />
            <input
              type="number"
              min={10}
              max={100}
              value={width}
              onChange={e => handleWidthChange(Number(e.target.value))}
              className="w-14 px-1 py-0.5 border rounded text-xs bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            />
            <span className="text-xs text-gray-500 dark:text-gray-400">%</span>
          </div>
          
          {/* Alignment Controls */}
          <div className="flex items-center gap-2 pb-2">
            <div className="flex gap-1">
              <button
                onClick={() => handleAlignChange('left')}
                className={`px-2 py-1 text-xs rounded flex items-center justify-center ${
                  align === 'left' 
                    ? 'bg-orange-500 text-white' 
                    : 'bg-gray-200 dark:bg-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-500'
                }`}
                title="จัดซ้าย"
              >
                <AlignLeftIcon />
              </button>
              <button
                onClick={() => handleAlignChange('center')}
                className={`px-2 py-1 text-xs rounded flex items-center justify-center ${
                  align === 'center' 
                    ? 'bg-orange-500 text-white' 
                    : 'bg-gray-200 dark:bg-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-500'
                }`}
                title="จัดกึ่งกลาง"
              >
                <AlignCenterIcon />
              </button>
              <button
                onClick={() => handleAlignChange('right')}
                className={`px-2 py-1 text-xs rounded flex items-center justify-center ${
                  align === 'right' 
                    ? 'bg-orange-500 text-white' 
                    : 'bg-gray-200 dark:bg-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-500'
                }`}
                title="จัดขวา"
              >
                <AlignRightIcon />
              </button>
            </div>
          </div>
        </div>
      )}
    </NodeViewWrapper>
  )
}

export default LinkPreviewNodeView


