import { Node, mergeAttributes } from "@tiptap/core"
import type { CommandProps } from "@tiptap/core"
import { ReactNodeViewRenderer } from "@tiptap/react"
import { LinkPreviewNodeView } from "./LinkPreviewNodeView"

declare module "@tiptap/core" {
  interface Commands<ReturnType> {
    linkPreview: {
      insertLinkPreview: (attrs: LinkPreviewAttrs) => ReturnType
      updateLinkPreviewById: (id: string, attrs: Partial<LinkPreviewAttrs>) => ReturnType
    }
  }
}

export interface LinkPreviewAttrs {
  id?: string
  href: string
  title?: string
  description?: string
  image?: string
}

export const LinkPreviewNode = Node.create({
  name: "linkPreview",
  group: "block",
  atom: true,
  selectable: true,
  draggable: false,

  addAttributes() {
    return {
      id: { default: null },
      href: { default: "" },
      title: { default: null },
      description: { default: null },
      image: { default: null },
    }
  },

  parseHTML() {
    return [
      {
        tag: 'div[data-type="link-preview"]',
      },
    ]
  },

  renderHTML({ HTMLAttributes }) {
    return [
      "div",
      mergeAttributes(HTMLAttributes, { "data-type": "link-preview" }),
    ]
  },

  addCommands() {
    return {
      insertLinkPreview:
        (attrs: LinkPreviewAttrs) => ({ chain }) => {
          return chain()
            .insertContent({ type: this.name, attrs })
            .run()
        },
      updateLinkPreviewById:
        (id: string, attrs: Partial<LinkPreviewAttrs>) => ({ tr, state, dispatch }: CommandProps) => {
          let found = false
          state.doc.descendants((node: any, pos: number) => {
            if (node.type.name === "linkPreview" && node.attrs.id === id) {
              found = true
              const newAttrs = { ...node.attrs, ...attrs }
              tr.setNodeMarkup(pos, node.type, newAttrs)
              return false
            }
            return true
          })
          if (found && dispatch) {
            dispatch(tr)
          }
          return found
        },
    }
  },

  addNodeView() {
    return ReactNodeViewRenderer(LinkPreviewNodeView)
  },
})

export default LinkPreviewNode


