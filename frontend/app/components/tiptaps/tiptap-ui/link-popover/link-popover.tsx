import * as React from "react"
import { isNodeSelection, type Editor } from "@tiptap/react"

// --- Hooks ---
import { useTiptapEditor } from "@/app/hooks/use-tiptap-editor"

// --- Icons ---
import { CornerDownLeftIcon } from "@/app/components/tiptaps/tiptap-icons/corner-down-left-icon"
import { ExternalLinkIcon } from "@/app/components/tiptaps/tiptap-icons/external-link-icon"
import { LinkIcon } from "@/app/components/tiptaps/tiptap-icons/link-icon"
import { TrashIcon } from "@/app/components/tiptaps/tiptap-icons/trash-icon"

// --- Lib ---
import { isMarkInSchema } from "@/lib/tiptap-utils"

// --- UI Primitives ---
import type { ButtonProps } from "@/app/components/tiptaps/tiptap-ui-primitive/button"
import { Button } from "@/app/components/tiptaps/tiptap-ui-primitive/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/app/components/tiptaps/tiptap-ui-primitive/popover"
import { Separator } from "@/app/components/tiptaps/tiptap-ui-primitive/separator"


export interface LinkHandlerProps {
  editor: Editor | null
  onSetLink?: () => void
  onLinkActive?: () => void
}

export interface LinkMainProps {
  url: string
  setUrl: React.Dispatch<React.SetStateAction<string>>
  setLink: () => void
  removeLink: () => void
  isActive: boolean
}

export const useLinkHandler = (props: LinkHandlerProps) => {
  const { editor, onSetLink, onLinkActive } = props
  const [url, setUrl] = React.useState<string>("")

  React.useEffect(() => {
    if (!editor) return

    // Get URL immediately on mount
    const { href } = editor.getAttributes("link")

    if (editor.isActive("link") && !url) {
      setUrl(href || "")
      onLinkActive?.()
    }
  }, [editor, onLinkActive, url])

  React.useEffect(() => {
    if (!editor) return

    const updateLinkState = () => {
      const { href } = editor.getAttributes("link")
      setUrl(href || "")

      if (editor.isActive("link") && !url) {
        onLinkActive?.()
      }
    }

    editor.on("selectionUpdate", updateLinkState)
    return () => {
      editor.off("selectionUpdate", updateLinkState)
    }
  }, [editor, onLinkActive, url])

  const setLink = React.useCallback(() => {
    if (!url || !editor) return

    const normalizeUrl = (raw: string): string => {
      try {
        // If it already has a protocol, leave as-is
        if (/^[a-zA-Z][a-zA-Z0-9+.-]*:/.test(raw)) return raw
        // Add https:// by default for bare domains
        return `https://${raw}`
      } catch {
        return raw
      }
    }

    const href = normalizeUrl(url.trim())

    const { from, to, empty } = editor.state.selection
    const chain = editor.chain().focus().extendMarkRange("link")

    if (empty) {
      // Insert the URL text, select it, then apply the link mark
      chain
        .insertContent(href)
        .setTextSelection({ from, to: from + href.length })
        .setLink({ href })
        .run()
    } else {
      // Apply link mark to the selected text
      chain
        .setLink({ href })
        .run()
    }

    onSetLink?.()
  }, [editor, onSetLink, url])

  const removeLink = React.useCallback(() => {
    if (!editor) return
    editor
      .chain()
      .focus()
      .unsetMark("link", { extendEmptyMarkRange: true })
      .setMeta("preventAutolink", true)
      .run()
    setUrl("")
  }, [editor])

  return {
    url,
    setUrl,
    setLink,
    removeLink,
    isActive: editor?.isActive("link") || false,
  }
}

export const LinkButton = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, children, ...props }, ref) => {
    return (
      <Button
        type="button"
        className={className}
        data-style="ghost"
        role="button"
        tabIndex={-1}
        aria-label="Link"
        tooltip="Link"
        ref={ref}
        {...props}
      >
        {children || <LinkIcon className="tiptap-button-icon" />}
      </Button>
    )
  }
)

export const LinkContent: React.FC<{
  editor?: Editor | null
}> = ({ editor: providedEditor }) => {
  const editor = useTiptapEditor(providedEditor)

  const linkHandler = useLinkHandler({
    editor: editor,
  })

  return <LinkMain {...linkHandler} />
}

const LinkMain: React.FC<LinkMainProps> = ({
  url,
  setUrl,
  setLink,
  removeLink,
  isActive,
}) => {
  const [isValid, setIsValid] = React.useState(true)

  const isLikelyUrl = React.useCallback((value: string) => {
    if (!value) return false
    try {
      if (/^[a-zA-Z][a-zA-Z0-9+.-]*:/.test(value)) return true
      if (/^([\w-]+\.)+[\w-]{2,}(\/.*)?$/i.test(value)) return true
      return false
    } catch {
      return false
    }
  }, [])

  React.useEffect(() => {
    if (!url) {
      setIsValid(true)
      return
    }
    setIsValid(isLikelyUrl(url))
  }, [url, isLikelyUrl])

  React.useEffect(() => {
    let mounted = true
    if (!url && typeof navigator !== 'undefined' && navigator.clipboard) {
      navigator.clipboard.readText().then((text) => {
        if (!mounted) return
        const candidate = text?.trim()
        if (candidate && isLikelyUrl(candidate)) {
          setUrl(candidate)
        }
      }).catch(() => {})
    }
    return () => { mounted = false }
  }, [url, setUrl, isLikelyUrl])

  const handleKeyDown = (event: React.KeyboardEvent) => {
    if (event.key === "Enter") {
      event.preventDefault()
      if (url && isValid) setLink()
    }
  }

  return (
    <>
      <div className="mb-2">
        <div className="text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Add link</div>
        <div className="flex items-center gap-2">
          <input
            type="url"
            placeholder="https://example.com"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            onKeyDown={handleKeyDown}
            autoComplete="off"
            autoCorrect="off"
            autoCapitalize="off"
            className={`tiptap-input tiptap-input-clamp flex-1 ${!isValid && url ? 'ring-1 ring-red-400' : ''}`}
            aria-invalid={!isValid}
          />
          <Button
            type="button"
            onClick={async () => {
              try {
                const text = await navigator.clipboard.readText()
                if (text) setUrl(text.trim())
              } catch {}
            }}
            title="Paste from clipboard"
            data-style="ghost"
          >
            Paste
          </Button>
        </div>
        {!isValid && url && (
          <div className="text-xs text-red-500 mt-1">Invalid URL</div>
        )}
      </div>

      <div className="flex items-center justify-between">
        <div className="tiptap-button-group" data-orientation="horizontal">
          <Button
            type="button"
            onClick={setLink}
            title="Apply link"
            disabled={!url || !isValid}
            data-style="ghost"
          >
            <CornerDownLeftIcon className="tiptap-button-icon" />
          </Button>
          <Button
            type="button"
            onClick={() => {
              const normalizeUrl = (raw: string): string => {
                if (/^[a-zA-Z][a-zA-Z0-9+.-]*:/.test(raw)) return raw
                return `https://${raw}`
              }
              const target = url ? normalizeUrl(url) : ''
              if (target) window.open(target, "_blank")
            }}
            title="Open in new window"
            disabled={!url || !isValid}
            data-style="ghost"
          >
            <ExternalLinkIcon className="tiptap-button-icon" />
          </Button>
        </div>

        <div className="tiptap-button-group" data-orientation="horizontal">
          <Button
            type="button"
            onClick={removeLink}
            title="Remove link"
            disabled={!url && !isActive}
            data-style="ghost"
          >
            <TrashIcon className="tiptap-button-icon" />
          </Button>
        </div>
      </div>
    </>
  )
}

export interface LinkPopoverProps extends Omit<ButtonProps, "type"> {
  /**
   * The TipTap editor instance.
   */
  editor?: Editor | null
  /**
   * Whether to hide the link popover.
   * @default false
   */
  hideWhenUnavailable?: boolean
  /**
   * Callback for when the popover opens or closes.
   */
  onOpenChange?: (isOpen: boolean) => void
  /**
   * Whether to automatically open the popover when a link is active.
   * @default true
   */
  autoOpenOnLinkActive?: boolean
}

export function LinkPopover({
  editor: providedEditor,
  hideWhenUnavailable = false,
  onOpenChange,
  autoOpenOnLinkActive = true,
  ...props
}: LinkPopoverProps) {
  const editor = useTiptapEditor(providedEditor)

  const linkInSchema = isMarkInSchema("link", editor)

  const [isOpen, setIsOpen] = React.useState(false)

  const onSetLink = () => {
    setIsOpen(false)
  }

  const onLinkActive = () => setIsOpen(autoOpenOnLinkActive)

  const linkHandler = useLinkHandler({
    editor: editor,
    onSetLink,
    onLinkActive,
  })

  const isDisabled = React.useMemo(() => {
    if (!editor) return true
    if (editor.isActive("codeBlock")) return true
    return !editor.can().setLink?.({ href: "" })
  }, [editor])

  const canSetLink = React.useMemo(() => {
    if (!editor) return false
    try {
      return editor.can().setMark("link")
    } catch {
      return false
    }
  }, [editor])

  const isActive = editor?.isActive("link") ?? false

  const handleOnOpenChange = React.useCallback(
    (nextIsOpen: boolean) => {
      setIsOpen(nextIsOpen)
      onOpenChange?.(nextIsOpen)
    },
    [onOpenChange]
  )

  const show = React.useMemo(() => {
    if (!linkInSchema || !editor) {
      return false
    }

    if (hideWhenUnavailable) {
      if (isNodeSelection(editor.state.selection) || !canSetLink) {
        return false
      }
    }

    return true
  }, [linkInSchema, hideWhenUnavailable, editor, canSetLink])

  if (!show || !editor || !editor.isEditable) {
    return null
  }

  return (
    <Popover open={isOpen} onOpenChange={handleOnOpenChange}>
      <PopoverTrigger asChild>
        <LinkButton
          disabled={isDisabled}
          data-active-state={isActive ? "on" : "off"}
          data-disabled={isDisabled}
          {...props}
        />
      </PopoverTrigger>

      <PopoverContent
        className="backdrop-blur-sm bg-white/95 dark:bg-gray-900/90 ring-1 ring-gray-200 dark:ring-gray-700 rounded-xl shadow-xl p-4 w-[360px]"
      >
        <LinkMain {...linkHandler} />
      </PopoverContent>
    </Popover>
  )
}

LinkButton.displayName = "LinkButton"
