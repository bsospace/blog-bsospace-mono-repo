"use client"
import * as React from "react"
import { EditorContent, useEditor, JSONContent } from "@tiptap/react"
import dynamic from 'next/dynamic'

// Core Extensions
import { StarterKit } from "@tiptap/starter-kit"
import { Image as TiptapImage } from "@tiptap/extension-image"
import { TaskItem } from "@tiptap/extension-task-item"
import { TaskList } from "@tiptap/extension-task-list"
import { TextAlign } from "@tiptap/extension-text-align"
import { Typography } from "@tiptap/extension-typography"
import { Highlight } from "@tiptap/extension-highlight"
import { Subscript } from "@tiptap/extension-subscript"
import { Superscript } from "@tiptap/extension-superscript"
import { Underline } from "@tiptap/extension-underline"
import CodeBlockLowlight from "@tiptap/extension-code-block-lowlight"
import * as lowlightLib from "lowlight"
import { TiptapImageNodeView } from "@/app/components/tiptap-node/image-node/TiptapImageNodeView";
import { ReactNodeViewRenderer } from "@tiptap/react";
import CodeBlockNode from "@/app/components/tiptap-node/code-block-node/code-block-node";

// Custom Extensions
import { Link } from "@/app/components/tiptap-extension/link-extension"
import { Selection } from "@/app/components/tiptap-extension/selection-extension"
import { TrailingNode } from "@/app/components/tiptap-extension/trailing-node-extension"
import { LinkPreviewNode } from "@/app/components/tiptap-node/link-preview-node/link-preview-node-extension"
import Loading from "../../Loading"

interface PreviewEditorProps {
  content: JSONContent
}

export function PreviewEditor({ content }: PreviewEditorProps) {
  const lowlight = React.useMemo(() => (lowlightLib as any).createLowlight((lowlightLib as any).common), [])
  const editor = useEditor({
    editable: false,
    immediatelyRender: false, // แก้ไข SSR error
    extensions: [
      StarterKit.configure({ codeBlock: false }),
      TextAlign.configure({ types: ["heading", "paragraph"] }),
      Underline,
      TaskList,
      TaskItem,
      Highlight,
      Superscript,
      Subscript,
      Typography,
      TiptapImage.extend({
        addNodeView() {
          return ReactNodeViewRenderer(TiptapImageNodeView);
        },
      }),
      CodeBlockLowlight.extend({
        addNodeView() {
          return ReactNodeViewRenderer(CodeBlockNode);
        },
      }).configure({ lowlight }),
      Selection,
      TrailingNode,
      LinkPreviewNode,
      Link.configure({
        openOnClick: true,
        HTMLAttributes: {
          class: 'underline text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 cursor-pointer',
          rel: 'noopener noreferrer',
        },
      }),
    ],
    content,
  })

  if (!editor) {
    return (
      <div className="text-center text-gray-500 dark:text-gray-400 py-12">
        Loading preview...
      </div>
    )
  }

  return (
    <div className="w-full h-full flex flex-col items-center justify-center">
      <div className="transition-all rounded-md duration-300 max-w-screen-xl w-full ease-out sticky top-16 bg-white dark:bg-gray-900">
        <EditorContent editor={editor} />
      </div>
    </div>
  )
}

// Export dynamic version to prevent SSR issues
export const DynamicPreviewEditor = dynamic(() => Promise.resolve(PreviewEditor), {
  ssr: false,
  loading: () => (
    <div className="text-center text-gray-500 dark:text-gray-400 py-12">
      <Loading />
    </div>
  ),
})
