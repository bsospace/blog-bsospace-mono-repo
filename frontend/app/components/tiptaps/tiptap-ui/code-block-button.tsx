import React from "react";
import { useCurrentEditor } from "@tiptap/react";
import { Button } from "@/app/components/tiptaps/tiptap-ui-primitive/button";
import { CodeIcon } from "@/app/components/tiptaps/tiptap-icons/code-icon";

export const CodeBlockButton: React.FC = () => {
  const { editor } = useCurrentEditor();

  // Don't render if editor is not ready
  if (!editor || editor.isDestroyed) {
    return null;
  }

  const toggleCodeBlock = () => {
    try {
      editor.chain().focus().toggleCodeBlock().run();
    } catch (error) {
      console.error('Error toggling code block:', error);
    }
  };

  const isActive = editor.isActive("codeBlock");

  return (
    <Button
      type="button"
      data-style="ghost"
      data-active-state={isActive ? "on" : "off"}
      onClick={toggleCodeBlock}
      aria-label="Toggle code block"
      title="Code Block (Ctrl+Shift+C)"
    >
      <CodeIcon className="tiptap-button-icon" />
    </Button>
  );
};
