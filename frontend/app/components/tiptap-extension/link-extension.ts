import TiptapLink from "@tiptap/extension-link"
import type { EditorView } from "@tiptap/pm/view"
import { getMarkRange } from "@tiptap/react"
import { Plugin, TextSelection } from "@tiptap/pm/state"

export const Link = TiptapLink.extend({
  inclusive: false,

  parseHTML() {
    return [
      {
        tag: 'a[href]:not([data-type="button"]):not([href *= "javascript:" i])',
      },
    ]
  },

  addProseMirrorPlugins() {
    const { editor } = this

    return [
      ...(this.parent?.() || []),
      new Plugin({
        props: {
          handleKeyDown: (_: EditorView, event: KeyboardEvent) => {
            const { selection } = editor.state

            if (event.key === "Escape" && selection.empty !== true) {
              editor.commands.focus(selection.to, { scrollIntoView: false })
            }

            return false
          },
          handleClick(view, pos) {
            const { schema, doc } = view.state
            let range: ReturnType<typeof getMarkRange> | undefined

            if (schema.marks.link) {
              range = getMarkRange(doc.resolve(pos), schema.marks.link)
            }

            if (!range) {
              return false
            }

            const { from, to } = range
            const start = Math.min(from, to)
            const end = Math.max(from, to)

            if (pos < start || pos > end) {
              return false
            }

            // Get the link mark attributes to extract href
            const $pos = doc.resolve(pos)
            const linkMark = $pos.marks().find(mark => mark.type === schema.marks.link)
            
            if (linkMark && linkMark.attrs.href) {
              const href = linkMark.attrs.href
              
              // Check if it's an external link
              const isExternal = href.startsWith('http://') || href.startsWith('https://') || href.startsWith('//')
              
              if (isExternal) {
                // Open external links in new tab
                window.open(href, '_blank', 'noopener,noreferrer')
              } else {
                // Navigate to internal links in same tab
                window.location.href = href
              }
              
              return true
            }

            return false
          },
        },
      }),
    ]
  },
})

export default Link
