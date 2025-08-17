import React, { useEffect, useRef, useState } from "react";
import { NodeViewWrapper, NodeViewContent, ReactNodeViewProps } from "@tiptap/react";
import { Copy, Check, ChevronDown } from "lucide-react";

const CodeBlockNode: React.FC<ReactNodeViewProps> = ({
  node,
  updateAttributes,
  extension,
  editor,
}) => {
  const [isCopied, setIsCopied] = useState(false);
  const [isLanguageOpen, setIsLanguageOpen] = useState(false);
  const defaultLanguage = node.attrs.language || "null";
  const wrapperRef = useRef<HTMLDivElement | null>(null);
  const dropdownRef = useRef<HTMLDivElement | null>(null);
  const isEditable = editor?.isEditable ?? true;

  const handleCopy = async () => {
    try {
      const codeElement = wrapperRef.current?.querySelector('code');
      if (codeElement && navigator.clipboard) {
        await navigator.clipboard.writeText(codeElement.textContent || '');
        setIsCopied(true);
        setTimeout(() => setIsCopied(false), 2000);
      }
    } catch (err) {
      console.error('Failed to copy code:', err);
    }
  };

  const handleLanguageChange = (language: string) => {
    updateAttributes({ language });
    setIsLanguageOpen(false);
  };

  const languages = (extension as any)?.options?.lowlight?.listLanguages?.() || [];
  const commonLanguages = [
    'javascript', 'typescript', 'html', 'css', 'python', 'java', 'csharp', 'sql', 'json',
    'php', 'ruby', 'go', 'rust', 'cpp', 'c', 'csharp', 'swift', 'kotlin', 'scala'
  ];

  const filteredLanguages = languages.filter((lang: string) => 
    commonLanguages.includes(lang.toLowerCase())
  );

  // Close language dropdown on outside click or Escape key
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (!isLanguageOpen) return;
      const target = event.target as Node;
      if (
        wrapperRef.current &&
        !wrapperRef.current.contains(target)
      ) {
        setIsLanguageOpen(false);
      }
    };

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setIsLanguageOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isLanguageOpen]);

  // Ensure dropdown is closed and disabled in preview (non-editable) mode
  useEffect(() => {
    if (!isEditable && isLanguageOpen) {
      setIsLanguageOpen(false);
    }
  }, [isEditable, isLanguageOpen]);

  return (
    <NodeViewWrapper className="code-block-wrapper" ref={wrapperRef}>
      {/* Header with language selector and copy button */}
      <div className="code-block-header">
        <div className="code-block-language-selector">
          <button
            type="button"
            className="language-selector-button"
            onClick={isEditable ? () => setIsLanguageOpen(!isLanguageOpen) : undefined}
            contentEditable={false}
            aria-haspopup="listbox"
            aria-expanded={isLanguageOpen}
            aria-label="Select code language"
            disabled={!isEditable}
          >
            <span className="language-label">
              {defaultLanguage === "null" ? "Auto" : defaultLanguage}
            </span>
            <ChevronDown className="chevron-icon" />
          </button>
          
          {isEditable && isLanguageOpen && (
            <div className="language-dropdown" ref={dropdownRef} role="listbox">
              <div className="language-option" role="option" aria-selected={defaultLanguage === 'null'} onClick={() => handleLanguageChange("null")}>
                Auto
              </div>
              <div className="language-separator"></div>
              {filteredLanguages.map((lang: string) => (
                <div
                  key={lang}
                  className={`language-option ${defaultLanguage === lang ? 'selected' : ''}`}
                  role="option"
                  aria-selected={defaultLanguage === lang}
                  onClick={() => handleLanguageChange(lang)}
                >
                  {lang}
                </div>
              ))}
            </div>
          )}
        </div>

        <button
          type="button"
          className="copy-button"
          onClick={handleCopy}
          contentEditable={false}
          title={isCopied ? "Copied!" : "Copy code"}
          aria-label={isCopied ? "Copied!" : "Copy code"}
        >
          {isCopied ? (
            <Check className="copy-icon copied" />
          ) : (
            <Copy className="copy-icon" />
          )}
        </button>
      </div>

      {/* Code content */}
      <div className="code-block-content">
        <pre>
          <NodeViewContent
            as="code"
            className={`hljs ${defaultLanguage && defaultLanguage !== 'null' ? `language-${defaultLanguage}` : ''}`}
          />
        </pre>
      </div>
    </NodeViewWrapper>
  );
};

export default CodeBlockNode;
