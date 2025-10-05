"use client"
import * as React from "react"
import { EditorContent, EditorContext, JSONContent, useEditor } from "@tiptap/react"

// --- Tiptap Core Extensions ---
import { StarterKit } from "@tiptap/starter-kit"
import { Image } from "@tiptap/extension-image"
import { TaskItem } from "@tiptap/extension-task-item"
import { TaskList } from "@tiptap/extension-task-list"
import { TextAlign } from "@tiptap/extension-text-align"
import { Typography } from "@tiptap/extension-typography"
import { Highlight } from "@tiptap/extension-highlight"
import { Subscript } from "@tiptap/extension-subscript"
import { Superscript } from "@tiptap/extension-superscript"
import { Underline } from "@tiptap/extension-underline"
import CodeBlockLowlight from "@tiptap/extension-code-block-lowlight"
// Note: lowlight typings (as of @types/lowlight 2.x) may not match the installed lowlight version (currently v2.x).
// This import uses a namespace to avoid type errors; check compatibility if upgrading lowlight or its types.
import * as lowlightLib from "lowlight"

// --- Custom Extensions ---
import { Link } from "@/app/components/tiptaps/tiptap-extension/link-extension"
import { Selection } from "@/app/components/tiptaps/tiptap-extension/selection-extension"
import { TrailingNode } from "@/app/components/tiptaps/tiptap-extension/trailing-node-extension"

// --- UI Primitives ---
import { Button } from "@/app/components/tiptaps/tiptap-ui-primitive/button"
import { Spacer } from "@/app/components/tiptaps/tiptap-ui-primitive/spacer"
import {
  Toolbar,
  ToolbarGroup,
  ToolbarSeparator,
} from "@/app/components/tiptaps/tiptap-ui-primitive/toolbar"

// --- Tiptap Node ---
import { ImageUploadNode } from "@/app/components/tiptaps/tiptap-node/image-upload-node/image-upload-node-extension"
import { TiptapImageNodeView } from "@/app/components/tiptaps/tiptap-node/image-node/TiptapImageNodeView";
import { Image as TiptapImage, ImageOptions } from "@tiptap/extension-image";
import { ReactNodeViewRenderer } from "@tiptap/react";
import CodeBlockNode from "@/app/components/tiptaps/tiptap-node/code-block-node/code-block-node";
import { LinkPreviewNode } from "@/app/components/tiptaps/tiptap-node/link-preview-node/link-preview-node-extension";

// --- Tiptap UI ---
import { HeadingDropdownMenu } from "@/app/components/tiptaps/tiptap-ui/heading-dropdown-menu"
import { ImageUploadButton } from "@/app/components/tiptaps/tiptap-ui/image-upload-button"
import { ListDropdownMenu } from "@/app/components/tiptaps/tiptap-ui/list-dropdown-menu"
import { NodeButton } from "@/app/components/tiptaps/tiptap-ui/node-button"
import {
  HighlightPopover,
  HighlightContent,
  HighlighterButton,
} from "@/app/components/tiptaps/tiptap-ui/highlight-popover"
import {
  LinkPopover,
  LinkContent,
  LinkButton,
} from "@/app/components/tiptaps/tiptap-ui/link-popover"
import { MarkButton } from "@/app/components/tiptaps/tiptap-ui/mark-button"
import { TextAlignButton } from "@/app/components/tiptaps/tiptap-ui/text-align-button"
import { UndoRedoButton } from "@/app/components/tiptaps/tiptap-ui/undo-redo-button"

// --- Icons ---
import { ArrowLeftIcon } from "@/app/components/tiptaps/tiptap-icons/arrow-left-icon"
import { HighlighterIcon } from "@/app/components/tiptaps/tiptap-icons/highlighter-icon"
import { LinkIcon } from "@/app/components/tiptaps/tiptap-icons/link-icon"

// --- Hooks ---
import { useMobile } from "@/app/hooks/use-mobile"
import { useWindowSize } from "@/app/hooks/use-window-size"
import { useCursorVisibility } from "@/app/hooks/use-cursor-visibility"

// --- Lib ---
import { handleImageUpload, MAX_FILE_SIZE } from "@/lib/tiptap-utils"

// import content from "@/app/components/tiptap-templates/simple/data/content.json"
import { Input } from "@/components/ui/input"


// Types
type MobileViewType = "main" | "highlighter" | "link"

interface PostMetadata {
  title: string
  description: string
  tags: string[]
  category: string
  publishDate: Date
  author: string
  slug: string
}

interface MainToolbarContentProps {
  onHighlighterClick: () => void
  onLinkClick: () => void
  onPublishClick: () => void
  isMobile: boolean
}

interface MobileToolbarContentProps {
  type: "highlighter" | "link"
  onBack: () => void
}


// Enhanced MainToolbarContent with Publish button
const MainToolbarContent = React.memo<MainToolbarContentProps>(({
  onHighlighterClick,
  onLinkClick,
  onPublishClick,
  isMobile,
}) => {
  return (
    <>
      {/* History Controls */}
      <ToolbarGroup aria-label="History controls">
        <UndoRedoButton action="undo" />
        <UndoRedoButton action="redo" />
      </ToolbarGroup>

      <ToolbarSeparator />

      {/* Content Structure */}
      <ToolbarGroup aria-label="Content structure">
        <HeadingDropdownMenu levels={[1, 2, 3, 4]} />
        <ListDropdownMenu types={["bulletList", "orderedList", "taskList"]} />
        <NodeButton type="codeBlock" />
        <NodeButton type="blockquote" />
      </ToolbarGroup>

      <ToolbarSeparator />

      {/* Text Formatting */}
      <ToolbarGroup aria-label="Text formatting">
        <MarkButton type="bold" />
        <MarkButton type="italic" />
        <MarkButton type="strike" />
        <MarkButton type="code" />
        <MarkButton type="underline" />
        {!isMobile ? (
          <HighlightPopover />
        ) : (
          <HighlighterButton onClick={onHighlighterClick} />
        )}
        {!isMobile ? <LinkPopover /> : <LinkButton onClick={onLinkClick} />}
      </ToolbarGroup>

      <ToolbarSeparator />

      {/* Advanced Typography */}
      {/* <ToolbarGroup aria-label="Advanced typography">
        <MarkButton type="superscript" />
        <MarkButton type="subscript" />
      </ToolbarGroup> */}

      <ToolbarSeparator />

      {/* Text Alignment */}
      <ToolbarGroup aria-label="Text alignment">
        <TextAlignButton align="left" />
        <TextAlignButton align="center" />
        <TextAlignButton align="right" />
        <TextAlignButton align="justify" />
      </ToolbarGroup>

      <ToolbarSeparator />

      {/* Media */}
      <ToolbarGroup aria-label="Media">
        <ImageUploadButton text="Add Image" />
      </ToolbarGroup>

      <Spacer />

      {isMobile && <ToolbarSeparator />}
    </>
  )
})

MainToolbarContent.displayName = "MainToolbarContent"

// Enhanced MobileToolbarContent with better navigation
const MobileToolbarContent = React.memo<MobileToolbarContentProps>(({
  type,
  onBack,
}) => (
  <>
    <ToolbarGroup>
      <Button
        data-style="ghost"
        onClick={onBack}
        aria-label={`Go back from ${type} menu`}
      >
        <ArrowLeftIcon className="tiptap-button-icon" />
        {type === "highlighter" ? (
          <HighlighterIcon className="tiptap-button-icon" />
        ) : (
          <LinkIcon className="tiptap-button-icon" />
        )}
      </Button>
    </ToolbarGroup>

    <ToolbarSeparator />

    {type === "highlighter" ? <HighlightContent /> : <LinkContent />}
  </>
))

MobileToolbarContent.displayName = "MobileToolbarContent"

interface SimpleEditorProps {
  onContentChange: (content: JSONContent) => void;
  initialContent?: JSONContent;
}

// Enhanced SimpleEditor with Publish Modal
export function SimpleEditor(
  { onContentChange, initialContent }: SimpleEditorProps
) {
  const lowlight = React.useMemo(() => (lowlightLib as any).createLowlight((lowlightLib as any).common), [])
  const isMobile = useMobile()
  const windowSize = useWindowSize()
  const [mobileView, setMobileView] = React.useState<MobileViewType>("main")
  const [isLoading, setIsLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)
  const [showPublishModal, setShowPublishModal] = React.useState(false)
  const toolbarRef = React.useRef<HTMLDivElement>(null)

  // Enhanced editor configuration with better error handling
  const editor = useEditor({
    onUpdate: ({ editor }) => {
      const json = editor.getJSON();
      onContentChange?.(json);
    },
    immediatelyRender: false,
    editorProps: {
      attributes: {
        autocomplete: "off",
        autocorrect: "off",
        autocapitalize: "off",
        "aria-label": "Rich text editor. Use keyboard shortcuts or toolbar buttons to format text.",
        role: "textbox",
        "aria-multiline": "true",
        "aria-describedby": "editor-help",
      },
      handlePaste: (view, event) => {
        let clipboardText = ""
        const items = event.clipboardData?.items
        // for check items in clipboard has value
        if (!items) return false

        // loop for check values from clipboard are text or image
        for (let i = 0; i < items.length; i++) {
          const item = items[i]
          if (item.type === 'text/plain') {
            item.getAsString((text) => {
              clipboardText = text?.trim() || "";
            });
          }
          else if (item.type.startsWith('image/')) {
            const file = item.getAsFile()
            if (!file) return false

            const src = URL.createObjectURL(file) // local preview url

            event.preventDefault()

            editor?.chain()
              .focus()
              .insertContent({
                type: 'image',
                attrs: { src },
              })
              .run()
            return true
          }
        }

        const isLikelyUrl = (value: string) => {
          try {
            if (/^[a-zA-Z][a-zA-Z0-9+.-]*:/.test(value)) return true
            if (/^([\w-]+\.)+[\w-]{2,}(\/.*)?$/i.test(value)) return true
            return false
          } catch {
            return false
          }
        }

        if (!isLikelyUrl(clipboardText)) return false

        const normalizeUrl = (raw: string): string => {
          if (/^[a-zA-Z][a-zA-Z0-9+.-]*:/.test(raw)) return raw
          return `https://${raw}`
        }

        const href = normalizeUrl(clipboardText)

        const { from, to, empty } = view.state.selection
        const chain = editor?.chain().focus().extendMarkRange('link')
        if (!chain) return false

        event.preventDefault()

        if (empty) {
          // Insert the pasted URL as a link immediately
          const tx = editor?.chain()
            .focus()
            .insertContent({
              type: 'text',
              text: clipboardText,
              marks: [{ type: 'link', attrs: { href } }],
            })
          tx?.run()

          // Then unfurl in the background and add a preview card
          fetch('/api/unfurl', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ url: href }),
          })
            .then((res) => (res.ok ? res.json() : null))
            .then((data) => {
              if (data) {
                const title = data.title || clipboardText
                const description = data.description || ''
                const image = data.image
                let hostname = ''
                try { hostname = new URL(href).hostname } catch { }
                const favicon = hostname ? 'https://icons.duckduckgo.com/ip3/' + hostname + '.ico' : ''

                // Replace the just-inserted URL text (last inserted) with a linkPreview node
                const id = Math.random().toString(36).slice(2)
                const { state } = editor!
                const pos = state.selection.from
                editor?.chain()
                  .focus()
                  .deleteRange({ from: pos - clipboardText.length, to: pos })
                  .insertContent({
                    type: 'linkPreview',
                    attrs: {
                      id,
                      href,
                      title,
                      description,
                      image,
                      width: 100,
                      align: 'left'
                    },
                  })
                  .run()
              }
            })
            .catch(() => {
              // ignore errors; link text is already inserted
            })
        } else {
          chain
            .setTextSelection({ from, to })
            .setLink({ href })
            .run()
        }

        return true
      },
      handleKeyDown: (view, event) => {
        // Enhanced keyboard shortcuts
        if (event.ctrlKey || event.metaKey) {
          switch (event.key) {
            case 'z':
              if (event.shiftKey) {
                editor?.chain().focus().redo().run()
                return true
              }
              editor?.chain().focus().undo().run()
              return true
            case 's':
              event.preventDefault()
              // Add auto-save functionality here
              return true
            case 'p':
              if (event.shiftKey) {
                event.preventDefault()
                setShowPublishModal(true)
                return true
              }
              break
          }
        }
        return false
      },
    },
    extensions: [
      StarterKit.configure({
        history: {
          depth: 500,
        },
        codeBlock: false,
      }),
      TextAlign.configure({
        types: ["heading", "paragraph"],
        alignments: ["left", "center", "right", "justify"],
      }),
      Underline,
      TaskList.configure({
        HTMLAttributes: {
          class: 'task-list',
        },
      }),
      TaskItem.configure({
        nested: true,
        HTMLAttributes: {
          class: 'task-item',
        },
      }),
      Highlight.configure({
        multicolor: true,
        HTMLAttributes: {
          class: 'highlight',
        },
      }),
      TiptapImage.extend({
        addNodeView() {
          return ReactNodeViewRenderer(TiptapImageNodeView);
        },
      }).configure({
        HTMLAttributes: {
          class: 'editor-image',
          loading: 'lazy',
        },
      }),
      CodeBlockLowlight.extend({
        addNodeView() {
          return ReactNodeViewRenderer(CodeBlockNode);
        },
      }).configure({ lowlight }),
      Typography,
      Superscript,
      Subscript,
      Selection,
      ImageUploadNode.configure({
        accept: "image/*",
        maxSize: MAX_FILE_SIZE,
        limit: 5, // Increased limit
        upload: async (file: File) => {
          try {
            setError(null)
            const result = await handleImageUpload(file)
            return result
          } catch (uploadError) {
            const errorMessage = uploadError instanceof Error ?
              uploadError.message :
              'Failed to upload image'
            setError(errorMessage)
            throw uploadError
          }
        },
        onError: (error) => {
          console.error("Upload failed:", error)
          setError(error instanceof Error ? error.message : 'Upload failed')
        },
      }),
      TrailingNode,
      LinkPreviewNode,
      Link.configure({
        openOnClick: true,
        autolink: true,
        linkOnPaste: true,
        protocols: ['http', 'https', 'mailto'],
        HTMLAttributes: {
          class: 'underline text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 cursor-pointer',
          rel: 'noopener noreferrer',
          target: '_blank',
        },
      }),
    ],
    content: initialContent,
    onCreate: () => {
      setIsLoading(false)
    },
  })

  const bodyRect = useCursorVisibility({
    editor,
    overlayHeight: toolbarRef.current?.getBoundingClientRect().height ?? 0,
  })

  // Auto-reset mobile view when switching to desktop
  React.useEffect(() => {
    if (!isMobile && mobileView !== "main") {
      setMobileView("main")
    }
  }, [isMobile, mobileView])

  // Keyboard shortcuts info
  const keyboardShortcuts = React.useMemo(() => [
    { key: 'Ctrl+B', action: 'Bold' },
    { key: 'Ctrl+I', action: 'Italic' },
    { key: 'Ctrl+U', action: 'Underline' },
    { key: 'Ctrl+Z', action: 'Undo' },
    { key: 'Ctrl+Shift+Z', action: 'Redo' },
    { key: 'Ctrl+K', action: 'Add Link' },
  ], [])

  // Handle mobile view changes with animation
  const handleMobileViewChange = React.useCallback((view: MobileViewType) => {
    setMobileView(view)
  }, [])

  // Error dismissal
  const dismissError = React.useCallback(() => {
    setError(null)
  }, [])


  const handleShowContent = () => {
    if (editor) {
      const html = editor.getHTML()
      const json = editor.getJSON()
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-96 rounded-lg">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-2"></div>
          <p className="text-gray-600 dark:text-gray-400">Loading editor...</p>
        </div>
      </div>
    )
  }

  return (
    <EditorContext.Provider value={{ editor }}>
      {/* Enhanced Toolbar */}
      <Toolbar
        ref={toolbarRef}
        className="transition-all duration-300 max-w-screen-xl w-full ease-out sticky top-16 z-10 bg-white dark:bg-gray-900  dark:border-gray-700 shadow-sm"
        style={
          isMobile
            ? {
              bottom: `calc(100% - ${windowSize.height - bodyRect.y}px)`,
              transform: mobileView !== "main" ? "translateY(-2px)" : "translateY(0)",
            }
            : {}
        }
      >
        {mobileView === "main" ? (
          <MainToolbarContent
            onHighlighterClick={() => handleMobileViewChange("highlighter")}
            onLinkClick={() => handleMobileViewChange("link")}
            onPublishClick={() => setShowPublishModal(true)}
            isMobile={isMobile}
          />
        ) : (
          <MobileToolbarContent
            type={mobileView === "highlighter" ? "highlighter" : "link"}
            onBack={() => handleMobileViewChange("main")}
          />
        )}
      </Toolbar>

      <div className="w-full min-h-[50vh] max-w-screen-xl focus:outline-none bg-white dark:bg-gray-900">
        {/* Enhanced Editor Content */}
        <EditorContent
          editor={editor}
          role="presentation"
          className="w-full border h-full min-h-[50vh]   rounded-b-md select-text transition-all duration-200 ease-out focus:outline-none focus:ring-2 dark:focus:outline-none bg-transparent dark:bg-transparent"
        />
      </div>
    </EditorContext.Provider>
  )
}
