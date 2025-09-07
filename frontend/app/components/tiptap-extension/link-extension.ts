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
            // If the editor is not editable (e.g., preview) or openOnClick is enabled,
            // do not intercept clicks so that the default link behavior can occur.
            if (!editor.isEditable || (this as any).options?.openOnClick) {
              return false
            }

            const { schema, doc, tr } = view.state
            
            // Check if position is valid
            if (pos < 0 || pos > doc.content.size) {
              return false
            }

            let range: ReturnType<typeof getMarkRange> | undefined

            try {
              if (schema.marks.link) {
                range = getMarkRange(doc.resolve(pos), schema.marks.link)
              }
            } catch (error) {
              // If resolve fails, return false
              return false
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

            try {
              const $start = doc.resolve(start)
              const $end = doc.resolve(end)
              const transaction = tr.setSelection(new TextSelection($start, $end))

              view.dispatch(transaction)
              // Prevent default navigation in editable mode when selecting a link.
              return true
            } catch (error) {
              // If resolve fails, return false
              return false
            }
          },
        },
      }),
    ]
  },
})

export default Link
