import { generateHTML } from '@tiptap/core'

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

import { Link } from "@/app/components/tiptaps/tiptap-extension/link-extension"
import { Selection } from "@/app/components/tiptaps/tiptap-extension/selection-extension"
import { TrailingNode } from "@/app/components/tiptaps/tiptap-extension/trailing-node-extension"
import { ImageUploadNode } from "@/app/components/tiptaps/tiptap-node/image-upload-node/image-upload-node-extension"
import { LinkPreviewNode } from "@/app/components/tiptaps/tiptap-node/link-preview-node/link-preview-node-extension"

import { JSONContent } from "@tiptap/react"

export function generateHtmlFromContent(content: JSONContent): string {
  const ImageWithStyle = TiptapImage.extend({
    addAttributes() {
      return {
        ...this.parent?.(),
        style: {
          default: null,
          parseHTML: element => element.getAttribute('style'),
          renderHTML: attributes => {
            return attributes.style ? { style: attributes.style } : {};
          },
        },
      };
    },
  });
  return generateHTML(content, [
    StarterKit,
    TextAlign.configure({
      types: ["heading", "paragraph"],
      alignments: ["left", "center", "right", "justify"],
    }),
    Underline,
    TaskList,
    TaskItem,
    Highlight.configure({ multicolor: true }),
    ImageWithStyle,
    Typography,
    Superscript,
    Subscript,
    Selection,
    TrailingNode,
    ImageUploadNode,
    LinkPreviewNode,
    Link,
  ])
}
