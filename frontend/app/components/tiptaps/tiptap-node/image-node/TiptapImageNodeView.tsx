import React, { useState, useRef } from 'react';
import { NodeViewWrapper, NodeViewProps } from '@tiptap/react';
import Image from 'next/image';

const Modal = ({ src, alt, open, onClose }: { src: string; alt?: string; open: boolean; onClose: () => void }) => {
    if (!open) return null;
    return (
        <div
            className="fixed inset-0 z-60 flex items-center justify-center bg-black/80"
            style={{ top: 60, bottom: 0 }}
            onClick={onClose}
        >
            <Image
                src={src}
                alt={alt || 'Image preview'}
                className="max-h-[90vh] max-w-[90vw] rounded shadow-lg border-2 border-white"
                onClick={e => e.stopPropagation()}
            />
            <button
                className="absolute top-4 right-4 text-white text-3xl font-bold bg-black/60 rounded-full px-3 py-1 hover:bg-black/80"
                onClick={onClose}
                aria-label="Close preview"
            >
                x
            </button>
        </div>
    );
};

function getSizeFromSrc(src: string): { width?: number; height?: number } {
    const match = src.match(/[?&]size=(\d+)x(\d+)/);
    if (match) {
        return { width: parseInt(match[1], 10), height: parseInt(match[2], 10) };
    }
    return {};
}

function setSizeInSrc(src: string, width: number, height: number): string {
    const base = src.replace(/[?&]size=\d+x\d+/, '');
    const hasQuery = base.includes('?');
    const cleanBase = base.replace(/[?&]$/, '');
    return cleanBase + (hasQuery ? '&' : '?') + `size=${width}x${height}`;
}

export const TiptapImageNodeView: React.FC<NodeViewProps> = ({ node, selected, updateAttributes, editor }) => {
    const { src, alt, title } = node.attrs;
    const [open, setOpen] = useState(false);
    const imgRef = useRef<HTMLImageElement>(null);
    const dragging = useRef(false);
    const startX = useRef(0);
    const startWidth = useRef(0);
    const [imgSize, setImgSize] = useState(() => getSizeFromSrc(src));

    // When src changes (e.g. undo/redo), update local imgSize
    React.useEffect(() => {
        setImgSize(getSizeFromSrc(src));
    }, [src]);

    // Drag handle logic
    const onMouseDown = (e: React.MouseEvent) => {
        e.preventDefault();
        dragging.current = true;
        startX.current = e.clientX;
        if (imgRef.current) {
            startWidth.current = imgRef.current.offsetWidth;
        }
        document.addEventListener('mousemove', onMouseMove);
        document.addEventListener('mouseup', onMouseUp);
    };

    const onMouseMove = (e: MouseEvent) => {
        if (!dragging.current || !imgRef.current) return;
        const deltaX = e.clientX - startX.current;
        let newWidth = Math.max(40, startWidth.current + deltaX);
        // Round to nearest 10px
        newWidth = Math.round(newWidth / 10) * 10;
        // Keep aspect ratio if possible
        const aspect = imgRef.current.naturalWidth && imgRef.current.naturalHeight
            ? imgRef.current.naturalWidth / imgRef.current.naturalHeight
            : 16 / 9;
        let newHeight = Math.round(newWidth / aspect / 10) * 10;
        setImgSize({ width: newWidth, height: newHeight });
        // Update src with new size
        if (updateAttributes) {
            updateAttributes({ src: setSizeInSrc(src, newWidth, newHeight) });
        }
    };

    const onMouseUp = () => {
        dragging.current = false;
        document.removeEventListener('mousemove', onMouseMove);
        document.removeEventListener('mouseup', onMouseUp);
    };

    // Render size from src query
    const { width, height } = imgSize;
    const isEditable = editor?.isEditable ?? false;

    return (
        <NodeViewWrapper as="span" className="tiptap-image-node-view " style={{ position: 'relative', display: 'inline-block' }}>
            <Image
                ref={imgRef}
                src={src}
                alt={alt || title || ''}
                className={`editor-image rounded shadow ${selected ? 'ring-2 ring-orange-400' : ''}`}
                style={{ display: 'block', margin: '1.5rem auto', zIndex: !isEditable ? 60 : undefined }}
                width={width}
                height={height}
                onClick={isEditable ? undefined : () => setOpen(true)}
                loading="lazy"
                draggable={false}
            />
            {/* Resize handle: show only when selected and in edit mode */}
            {isEditable && selected && (
                <div
                    style={{
                        position: 'absolute',
                        top: '50%',
                        right: -10,
                        transform: 'translateY(-50%)',
                        width: 16,
                        height: 32,
                        cursor: 'ew-resize',
                        zIndex: 99,
                        background: 'rgba(255,255,255,0.7)',
                        borderRadius: 4,
                        border: '1px solid #fb923c',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        boxShadow: '0 1px 4px rgba(0,0,0,0.08)',
                    }}
                    onMouseDown={onMouseDown}
                >
                    <svg width="12" height="24" viewBox="0 0 12 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <rect x="5" y="4" width="2" height="16" rx="1" fill="#fb923c" />
                    </svg>
                </div>
            )}
            {/* Show modal only in view mode */}
            {!isEditable && <Modal src={src} alt={alt || title || ''} open={open} onClose={() => setOpen(false)} />}
        </NodeViewWrapper>
    );
};

export default TiptapImageNodeView; 