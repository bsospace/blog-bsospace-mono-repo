'use client';
/* eslint-disable @next/next/no-img-element */
import React, { useState, useRef, useEffect } from 'react';
import { Send, Bot, User, MessageCircle, X, Minimize2, Maximize2, Minimize } from 'lucide-react';
import { Textarea } from '@/components/ui/textarea';
import { useAuth } from '../contexts/auth-context';
import envConfig from '../configs/env-config';
import { Post } from '../interfaces';
import { useRouter } from 'next/navigation'
import {
  BrowserAIError,
  containsPromptInjection,
  createBrowserAISession,
  getBrowserAIAvailability,
  needsWebSearch,
  streamBrowserAI,
} from '../../lib/browser-ai';
import type { BrowserAIAvailability, BrowserAISession } from '../../lib/browser-ai';
// import { countTokens, isTokenLimitExceeded } from '../utils/token';

const WORD_LIMIT = 100;

interface WebSearchResult {
  title: string;
  url: string;
  snippet: string;
  source: string;
  type: string;
  publishedAt: string;
  thumbnail: string;
}

interface WebSearchResponse {
  query: string;
  total: number;
  results: WebSearchResult[];
}

interface AIProps {
  isOpen?: boolean;
  isFullOpen?: boolean;
  messages?: { id: number; type: 'user' | 'bot'; content: string; timestamp: Date }[];
  onClose?: () => void;
  Post: Post;
  onSendMessage?: (message: string) => void;
  onToggle?: () => void;
  onInputChange?: (text: string) => void;
  inputText?: string;
  isTyping?: boolean;
  setIsTyping?: (typing: boolean) => void;
}

const toSafeHttpUrl = (value: string, allowRelative = false): string | null => {
  const normalizedValue = value.trim();
  if (!normalizedValue) return null;
  try {
    const url = allowRelative
      ? new URL(normalizedValue, window.location.origin)
      : new URL(normalizedValue);
    return url.protocol === 'http:' || url.protocol === 'https:' ? url.toString() : null;
  } catch {
    return null;
  }
};

// Function to parse search results JSON
const parseSearchResults = (jsonObj: any): React.ReactNode => {
  if (!jsonObj.query || !jsonObj.results) return null;

  return (
    <div className="my-3 space-y-2">
      <div className="text-sm font-medium text-muted-foreground">
      </div>
      <div className="space-y-2">
        {jsonObj.results.slice(0, 5).map((result: any, idx: number) => {
          const safeUrl = toSafeHttpUrl(result.url);
          const source = result.source ?? (() => {
            try { return safeUrl ? new URL(safeUrl).hostname : ''; } catch { return ''; }
          })();
          const CardInner = (
            <div className="flex items-start gap-2">
              <div className="flex-1 min-w-0">
                <div className="text-sm font-medium text-orange-600 dark:text-orange-400 hover:underline line-clamp-2">
                  {result.title}
                </div>
                {result.snippet && (
                  <div className="text-xs text-muted-foreground mt-1 line-clamp-2">
                    {result.snippet}
                  </div>
                )}
                {source && (
                  <div className="text-xs text-muted-foreground mt-1">{source}</div>
                )}
              </div>
            </div>
          );
          return safeUrl ? (
            <a
              key={idx}
              href={safeUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="block p-3 rounded-lg border border-border hover:bg-muted/50 transition-colors"
            >
              {CardInner}
            </a>
          ) : (
            <div key={idx} className="block p-3 rounded-lg border border-border bg-muted/30">
              {CardInner}
            </div>
          );
        })}
      </div>
    </div>
  );
};

// Markdown parser function
const parseMarkdown = (text: string): React.ReactNode => {
  if (!text) return null;
  // Remove common assistant header markers like `assistant<|end_header_id|>:` (allow optional spaces and case-insensitive)
  text = text.replace(/assistant\s*<\|end_header_id\|>\s*:?\s*/gi, '');

  try {
    const parsed = JSON.parse(text);
    if (parsed.intro && parsed.text) {

      // ถ้ามีคำว่า "assistant" ให้ตัดออก
      const textWithoutAssistant = (parsed.text || '').replace(/assistant\s*<\|end_header_id\|>\s*:?\s*/gi, '').trim();
      const jsonObj = JSON.parse(textWithoutAssistant);
      if (jsonObj.query && jsonObj.results) {
        return parseSearchResults(jsonObj);
      }
    }
  } catch {
    // continue normal markdown parsing
  }

  // Logic for handling inline JSON in markdown
  const jsonMatch = text.match(/\{\"query\"[\s\S]*?\}(?=\s|$)/);
  const intro = text.replace(/\{\"intro\"[\s\S]*?\}(?=\s|$)/, '');

  if (jsonMatch && intro) {
    try {
      const jsonObj = JSON.parse(jsonMatch[0]);
      if (jsonObj.query && jsonObj.results) {
        const beforeJson = text.substring(0, jsonMatch.index);
        const afterJson = text.substring((jsonMatch.index || 0) + jsonMatch[0].length);

        return (
          <>
            {beforeJson && parseMarkdown(beforeJson)}
            {parseSearchResults(jsonObj)}
            {afterJson && parseMarkdown(afterJson)}
          </>
        );
      }
    } catch {
      // ignore invalid json
    }
  }

  // parse code blocks normally
  const codeBlockRegex = /```([\s\S]*?)```/g;
  const parts: (string | { type: 'codeblock'; content: string })[] = [];
  let lastIndex = 0;
  let match;

  while ((match = codeBlockRegex.exec(text)) !== null) {
    if (match.index > lastIndex) parts.push(text.slice(lastIndex, match.index));
    parts.push({ type: 'codeblock', content: match[1] });
    lastIndex = match.index + match[0].length;
  }
  if (lastIndex < text.length) parts.push(text.slice(lastIndex));

  return parts.map((part, index) =>
    typeof part === 'object' && part.type === 'codeblock' ? (
      <pre key={index} className="bg-gray-800 text-gray-100 p-3 rounded-md my-2 overflow-x-auto">
        <code>{part.content}</code>
      </pre>
    ) : (
      parseInlineMarkdown(part as string, index)
    )
  );
};

const parseInlineMarkdown = (text: string, keyPrefix: number): React.ReactNode => {
  const elements: React.ReactNode[] = [];
  let currentIndex = 0;

  // Define regex patterns - order matters (longer patterns first)
  const patterns = [
    { regex: /\*\*\*(.+?)\*\*\*/g, component: 'strongem', priority: 3 }, // ***bold+italic***
    { regex: /\*\*(.+?)\*\*/g, component: 'strong', priority: 2 }, // **bold**
    { regex: /\*(.+?)\*/g, component: 'em', priority: 1 }, // *italic*
    { regex: /`(.+?)`/g, component: 'code', priority: 2 }, // `code`
    { regex: /~~(.+?)~~/g, component: 'del', priority: 2 }, // ~~strikethrough~~
    { regex: /\[(.+?)\]\((.+?)\)/g, component: 'link', priority: 3 }, // [text](url)
  ];

  // Find all matches
  const matches: Array<{
    start: number;
    end: number;
    content: string;
    component: string;
    url?: string;
    priority: number;
  }> = [];

  patterns.forEach((pattern) => {
    let match;
    const regex = new RegExp(pattern.regex.source, 'g');

    while ((match = regex.exec(text)) !== null) {
      matches.push({
        start: match.index,
        end: match.index + match[0].length,
        content: match[1],
        component: pattern.component,
        url: match[2],
        priority: pattern.priority
      });
    }
  });

  // Sort by start position, then priority
  matches.sort((a, b) => {
    if (a.start !== b.start) return a.start - b.start;
    return b.priority - a.priority;
  });

  // Remove overlapping matches
  const validMatches = matches.filter((match, index) => {
    return !matches.slice(0, index).some(prevMatch =>
      match.start >= prevMatch.start && match.start < prevMatch.end
    );
  });

  // Build result
  validMatches.forEach((match, index) => {
    // Add text before match
    if (match.start > currentIndex) {
      elements.push(text.slice(currentIndex, match.start));
    }

    // Add formatted element
    const key = `${keyPrefix}-${index}`;
    switch (match.component) {
      case 'strongem':
        elements.push(<strong key={key} className="font-bold"><em className="italic">{match.content}</em></strong>);
        break;
      case 'strong':
        elements.push(<strong key={key} className="font-bold">{match.content}</strong>);
        break;
      case 'em':
        elements.push(<em key={key} className="italic">{match.content}</em>);
        break;
      case 'code':
        elements.push(
          <code key={key} className="bg-gray-200 dark:bg-gray-700 px-1 py-0.5 rounded text-sm font-mono">
            {match.content}
          </code>
        );
        break;
      case 'del':
        elements.push(<del key={key} className="line-through">{match.content}</del>);
        break;
      case 'link':
        {
          const safeUrl = match.url ? toSafeHttpUrl(match.url, true) : null;
          elements.push(
            safeUrl ? (
              <a key={key} href={safeUrl} target="_blank" rel="noopener noreferrer"
                className="text-blue-500 hover:text-blue-700 underline">
                {match.content}
              </a>
            ) : match.content
          );
        }
        break;
    }

    currentIndex = match.end;
  });

  // Add remaining text
  if (currentIndex < text.length) {
    elements.push(text.slice(currentIndex));
  }

  // Handle line breaks - do this only once at the end
  const result: React.ReactNode[] = [];
  elements.forEach((element, elemIndex) => {
    if (typeof element === 'string') {
      const lines = element.split('\n');
      lines.forEach((line, lineIndex) => {
        result.push(
          <React.Fragment key={`${keyPrefix}-frag-${elemIndex}-${lineIndex}`}>
            {line}
          </React.Fragment>
        );
        if (lineIndex < lines.length - 1) {
          result.push(<br key={`${keyPrefix}-br-${elemIndex}-${lineIndex}`} />);
        }
      });
    } else {
      result.push(element);
    }
  });

  return result;
};

const BlogAIChat: React.FC<AIProps> = ({
  isOpen: initialIsOpen,
  isFullOpen: initialIsFullOpen,
  messages: initialMessages,
  onClose,
  Post,
  onSendMessage,
  onToggle,
  onInputChange,
  inputText: initialInputText,
  isTyping: initialIsTyping,
  setIsTyping: externalSetIsTyping
}) => {
  const [isOpen, setIsOpen] = useState(initialIsOpen || false);
  const [isFullscreen, setIsFullscreen] = useState(initialIsFullOpen || false);
  const [messages, setMessages] = useState<any[]>([]);
  const [inputText, setInputText] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const messageIdRef = useRef(1);
  const browserAISessionRef = useRef<BrowserAISession | null>(null);
  const browserAISessionPromiseRef = useRef<Promise<BrowserAISession> | null>(null);
  const [browserAIStatus, setBrowserAIStatus] = useState<BrowserAIAvailability | 'checking' | 'error'>('checking');
  const [browserAIError, setBrowserAIError] = useState('');
  const { user } = useAuth();
  const router = useRouter();
  const [wordCount, setWordCount] = useState(0);
  const [wordLimitExceeded, setWordLimitExceeded] = useState(false);
  const [wordError, setWordError] = useState('');

  // Update states when props change (for URL-based control)
  useEffect(() => {
    if (initialIsOpen !== undefined) {
      setIsOpen(initialIsOpen);
    }
    if (initialIsFullOpen !== undefined) {
      setIsFullscreen(initialIsFullOpen);
    }
  }, [initialIsOpen, initialIsFullOpen]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  useEffect(() => {
    if (!isOpen || !Post?.id) return;
    browserAISessionRef.current?.destroy();
    browserAISessionRef.current = null;
    browserAISessionPromiseRef.current = null;
    messageIdRef.current = 1;
    setBrowserAIError('');
    setBrowserAIStatus('checking');
    setMessages([
      {
        id: 1,
        type: 'bot',
        content: 'สวัสดีครับ! ผมคือ **AI Assistant บนเครื่องของคุณ** ผมจะตอบจากเนื้อหาบทความก่อนโดยไม่ต้องเข้าสู่ระบบครับ\n\nถ้าบทความมีข้อมูลไม่พอ ผู้ใช้ที่เข้าสู่ระบบจะค้นข้อมูลจากเว็บต่อได้',
        timestamp: new Date()
      }
    ]);
    void getBrowserAIAvailability().then(setBrowserAIStatus);
  }, [isOpen, Post?.id]);

  useEffect(() => () => {
    browserAISessionRef.current?.destroy();
    browserAISessionRef.current = null;
  }, []);

  useEffect(() => {
    if (isOpen) {
      // Lock scroll
      document.body.style.overflow = 'hidden';
    } else {
      // Unlock scroll
      document.body.style.overflow = '';
    }
    // Cleanup เผื่อ component ถูก unmount
    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  // Update word count on input change
  useEffect(() => {
    const count = inputText.trim().split(/\s+/).filter(Boolean).length;
    setWordCount(count);
    const exceeded = count > WORD_LIMIT;
    setWordLimitExceeded(exceeded);
    if (exceeded) {
      setWordError('คำถามยาวเกินไป');
    } else {
      setWordError('');
    }
  }, [inputText]);

  const fetchExternalWebContext = async (question: string): Promise<{
    context: string;
    results: WebSearchResult[];
  }> => {
    const controller = new AbortController();
    const timeout = window.setTimeout(() => controller.abort(), 20000);
    let response: Response;
    try {
      response = await fetch(`${envConfig.apiBaseUrl}/ai/${Post.id}/search`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ query: question }),
        signal: controller.signal,
      });
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') {
        throw new BrowserAIError('ค้นข้อมูลจากเว็บนานเกินไป กรุณาลองใหม่อีกครั้งครับ', 'failed');
      }
      throw new BrowserAIError('ค้นข้อมูลจากเว็บไม่สำเร็จ กรุณาลองใหม่อีกครั้งครับ', 'failed');
    } finally {
      window.clearTimeout(timeout);
    }

    if (!response.ok) {
      if (response.status === 401) {
        throw new BrowserAIError('เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่ก่อนค้นข้อมูลจากเว็บครับ', 'failed');
      }
      if (response.status === 429) {
        const retryAfter = Number(response.headers.get('Retry-After'));
        const waitMessage = Number.isFinite(retryAfter) ? ` กรุณารอ ${Math.ceil(retryAfter)} วินาที` : '';
        throw new BrowserAIError(`ค้นข้อมูลจากเว็บบ่อยเกินไปครับ${waitMessage}`, 'failed');
      }
      throw new BrowserAIError('ค้นข้อมูลจากเว็บไม่สำเร็จ กรุณาลองใหม่อีกครั้งครับ', 'failed');
    }

    const payload = await response.json() as WebSearchResponse;
    const results = Array.isArray(payload.results)
      ? payload.results.slice(0, 5).map((result) => ({
        title: String(result.title || '').slice(0, 240),
        url: toSafeHttpUrl(String(result.url || '')) || '',
        snippet: String(result.snippet || '').slice(0, 600),
        source: String(result.source || '').slice(0, 120),
        type: String(result.type || 'web'),
        publishedAt: String(result.publishedAt || '').slice(0, 80),
        thumbnail: '',
      })).filter((result) => result.title && result.url)
      : [];

    const context = results.map((result, index) => [
      `SOURCE ${index + 1}`,
      `TITLE: ${result.title}`,
      `URL: ${result.url}`,
      `SNIPPET: ${result.snippet}`,
    ].join('\n')).join('\n\n');

    return { context, results };
  };

  const ensureBrowserAISession = async (): Promise<BrowserAISession> => {
    if (browserAISessionRef.current) return browserAISessionRef.current;
    if (browserAISessionPromiseRef.current) return browserAISessionPromiseRef.current;

    const sessionPromise = createBrowserAISession(Post.content, (progress) => {
      setBrowserAIStatus(progress >= 100 ? 'available' : 'downloading');
    })
      .then((session) => {
        browserAISessionRef.current = session;
        setBrowserAIStatus('available');
        setBrowserAIError('');
        return session;
      })
      .catch((error: unknown) => {
        const message = error instanceof BrowserAIError
          ? error.message
          : 'Gemini Nano could not start in this browser.';
        setBrowserAIStatus(error instanceof BrowserAIError && error.code === 'unavailable' ? 'unavailable' : 'error');
        setBrowserAIError(message);
        throw error;
      });

    browserAISessionPromiseRef.current = sessionPromise;
    try {
      return await sessionPromise;
    } finally {
      browserAISessionPromiseRef.current = null;
    }
  };

  const handleOpen = () => {
    setIsOpen(true);
    void ensureBrowserAISession().catch(() => undefined);
  };

  const streamIntoMessage = async (
    session: BrowserAISession,
    question: string,
    webContext: string,
    botMessageId: number,
  ): Promise<string> => {
    let response = '';
    for await (const chunk of streamBrowserAI(session, question, webContext)) {
      response += chunk;
      setMessages((prev) => prev.map(msg =>
        msg.id === botMessageId
          ? { ...msg, content: response }
          : msg
      ));
    }
    return response;
  };

  const handleSendMessage = async () => {
    const question = inputText.trim();
    if (!question) return;
    if (wordLimitExceeded) {
      setWordError('คำถามยาวเกินไป');
      return;
    }
    setWordError('');

    const userMessage = {
      id: ++messageIdRef.current,
      type: 'user' as const,
      content: question,
      timestamp: new Date(),
    };
    const botMessageId = ++messageIdRef.current;

    setMessages((prev) => [...prev, userMessage, {
      id: botMessageId,
      type: 'bot' as const,
      content: '',
      searchResults: [],
      searchQuery: question,
      timestamp: new Date(),
    }]);
    setInputText('');
    setIsTyping(true);

    try {
      if (containsPromptInjection(question)) {
        throw new BrowserAIError(
          'ผมตอบได้เฉพาะคำถามเกี่ยวกับบทความ และไม่สามารถเปลี่ยนคำสั่งหรือเปิดเผย prompt ภายในได้ครับ',
          'blocked',
        );
      }

      const session = await ensureBrowserAISession();
      let botMessage = await streamIntoMessage(session, question, '', botMessageId);

      if (needsWebSearch(botMessage)) {
        if (!user) {
          setMessages((prev) => prev.map(msg =>
            msg.id === botMessageId
              ? { ...msg, content: 'บทความนี้ไม่มีข้อมูลเพียงพอสำหรับคำถามนี้ครับ หากต้องการค้นข้อมูลจากเว็บ กรุณาเข้าสู่ระบบก่อน' }
              : msg
          ));
          return;
        }

        setMessages((prev) => prev.map(msg =>
          msg.id === botMessageId
            ? { ...msg, content: 'กำลังค้นข้อมูลเพิ่มเติมจากเว็บครับ...' }
            : msg
        ));

        const search = await fetchExternalWebContext(question);
        setMessages((prev) => prev.map(msg =>
          msg.id === botMessageId
            ? { ...msg, content: '', searchResults: search.results }
            : msg
        ));

        if (!search.results.length) {
          setMessages((prev) => prev.map(msg =>
            msg.id === botMessageId
              ? { ...msg, content: 'ค้นข้อมูลจากเว็บแล้ว แต่ไม่พบแหล่งข้อมูลที่เหมาะสมครับ' }
              : msg
          ));
          return;
        }

        botMessage = await streamIntoMessage(session, question, search.context, botMessageId);
        if (needsWebSearch(botMessage)) {
          setMessages((prev) => prev.map(msg =>
            msg.id === botMessageId
              ? { ...msg, content: 'ยังไม่พบข้อมูลที่เชื่อถือได้เพียงพอสำหรับคำถามนี้ครับ' }
              : msg
          ));
        }
      }
    } catch (error: unknown) {
      console.error('Browser AI error:', error);
      const message = error instanceof BrowserAIError
        ? error.message
        : browserAIError || 'ขออภัยครับ Gemini Nano ใช้งานไม่ได้บน browser หรืออุปกรณ์นี้';
      setMessages((prev) => prev.map(msg =>
        msg.id === botMessageId
          ? { ...msg, content: message }
          : msg
      ));
    } finally {
      setIsTyping(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  const formatTime = (date: Date) => {
    return date.toLocaleTimeString('th-TH', {
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const toggleFullscreen = () => {
    setIsFullscreen(!isFullscreen);

    // Update URL based on fullscreen state
    if (typeof window !== 'undefined') {
      const url = new URL(window.location.href);
      if (!isFullscreen) {
        url.searchParams.set('chat_full', 'true');
        url.searchParams.delete('chat');
      } else {
        url.searchParams.set('chat', 'true');
        url.searchParams.delete('chat_full');
      }
      router.push(url.pathname + url.search);
    }
  };

  const handleMinimize = () => {
    if (isFullscreen) {
      setIsFullscreen(false);
    } else {
      setIsOpen(false);
    }
  };

  const handleClose = () => {
    setIsOpen(false);
    // Reset fullscreen when closing
    if (isFullscreen) {
      setIsFullscreen(false);
    }

    // Update URL to remove chat parameters
    if (typeof window !== 'undefined') {
      const url = new URL(window.location.href);
      url.searchParams.delete('chat');
      url.searchParams.delete('chat_full');
      router.push(url.pathname + url.search);
    }
  };

  function getUserAvatar() {
    if (user?.avatar) {
      return <img src={user.avatar} alt="User Avatar" className="h-6 w-6 rounded-full" />;
    }
    return <User className="h-3 w-3" />;
  }

  // Fixed positioning logic
  const getContainerClasses = () => {
    if (isFullscreen) {
      return 'fixed inset-0 z-50';
    }
    return 'fixed z-50 bottom-8 right-4';
  };

  const getChatWindowClasses = () => {
    if (isFullscreen) {
      return 'bg-background border border-border rounded-lg shadow-lg flex flex-col w-full h-full max-w-none max-h-none';
    }
    return 'bg-background border border-border rounded-lg shadow-lg flex flex-col w-80 h-96';
  };

  const browserAIInputDisabled = browserAIStatus === 'unavailable';
  const browserAIStatusText = browserAIStatus === 'available'
    ? 'ทำงานบนเครื่องของคุณ'
    : browserAIStatus === 'downloadable'
      ? 'กำลังเตรียมดาวน์โหลดโมเดล'
      : browserAIStatus === 'downloading'
        ? 'กำลังดาวน์โหลดโมเดล'
        : browserAIStatus === 'unavailable'
          ? 'อุปกรณ์นี้ไม่รองรับ'
          : browserAIStatus === 'error'
            ? 'เริ่มใช้งานไม่ได้'
            : 'กำลังตรวจสอบความพร้อม';

  return (
    <div className={getContainerClasses()}>
      {/* Chat Toggle Button */}
      {!isOpen && (
        <button
          onClick={handleOpen}
          className="inline-flex items-center justify-center rounded-full w-12 h-12 bg-primary text-primary-foreground shadow-lg hover:shadow-xl hover:bg-primary/90 transition-all duration-200 relative"
        >
          <MessageCircle className="h-5 w-5" />
        </button>
      )}

      {/* Chat Window */}
      {isOpen && (
        <div className={getChatWindowClasses()}>
          {/* Header */}
          <div className="flex items-center justify-between px-3 py-2 border-b border-border">
            <div className="flex items-center space-x-2">
              <span className="flex items-center justify-center w-6 h-6 rounded-full bg-muted">
                <Bot className="w-3.5 h-3.5 text-muted-foreground" />
              </span>
              <div>
                <span className="flex items-center gap-1 text-sm">
                  <span className="font-medium">Blog AI Assistant</span>
                  <div className="relative" title={browserAIStatusText}>
                    <span className={`w-2 h-2 rounded-full inline-block ${browserAIStatus === 'available' ? 'bg-green-500' : 'bg-amber-500'}`}></span>
                    {browserAIStatus === 'available' && (
                      <span className="w-2 h-2 bg-green-500 rounded-full inline-block animate-ping absolute right-0 top-[7px]"></span>
                    )}
                  </div>
                </span>
                <span className="text-[10px] text-muted-foreground">{browserAIStatusText}</span>
              </div>
            </div>

            <div className="flex gap-0.5">
              <button
                onClick={toggleFullscreen}
                className="inline-flex items-center justify-center rounded-md w-8 h-8 hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                title={isFullscreen ? "ย่อหน้าต่าง" : "ขยายเต็มจอ"}
              >
                {isFullscreen ? <Minimize className="h-4 w-4" /> : <Maximize2 className="h-4 w-4" />}
              </button>
              <button
                onClick={handleClose}
                className="inline-flex items-center justify-center rounded-md w-8 h-8 hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                title="ปิด"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          </div>

          {/* Messages */}
          <div
            className={`flex-1 overflow-y-auto p-4 space-y-3 ${isFullscreen ? 'max-w-4xl mx-auto w-full' : ''}`}
            style={{
              maxHeight: isFullscreen ? 'calc(100vh - 120px)' : '22rem',
              minHeight: '6rem',
              maxWidth: isFullscreen ? '56rem' : '22rem', // 896px หรือ 352px
              width: '100%',
              overflowY: 'auto',
              overflowX: 'hidden',
              scrollbarWidth: 'thin',
              scrollbarColor: '#cbd5e1 #f1f5f9',
            }}
          >
            {messages.map((message) => {
              if (!message.content) return null;

              return (
                <div
                  key={message.id}
                  className={`flex ${message.type === 'user' ? 'justify-end' : 'justify-start'}`}
                >
                  <div
                    className={`flex max-w-[85%] ${isFullscreen ? 'max-w-2xl' : 'max-w-[85%]'
                      } ${message.type === 'user' ? 'flex-row-reverse' : 'flex-row'}`}
                  >
                    <div
                      className={`flex-shrink-0 ${message.type === 'user' ? 'ml-2' : 'mr-2'}`}
                    >
                      <div
                        className={`w-6 h-6 rounded-full flex items-center justify-center ${message.type === 'user'
                          ? 'bg-primary text-primary-foreground'
                          : 'bg-muted text-muted-foreground'
                          }`}
                      >
                        {message.type === 'user' ? getUserAvatar() : <Bot className="h-3 w-3" />}
                      </div>
                    </div>
                    <div
                      className={`rounded-lg px-3 py-2 ${message.type === 'user'
                        ? 'bg-primary text-primary-foreground'
                        : 'bg-muted text-foreground'
                        }`}
                    >
                      <div
                        className={`text-sm ${message.type === 'user' ? 'text-white' : ''}`}
                      >
                        {parseMarkdown(message.content)}
                        {message.type === 'bot' && message.searchResults?.length > 0 && (
                          parseSearchResults({ query: message.searchQuery, results: message.searchResults })
                        )}
                      </div>
                      <p
                        className={`text-xs mt-1 ${message.type === 'user'
                          ? 'text-primary-foreground/70'
                          : 'text-muted-foreground'
                          }`}
                      >
                        {formatTime(message.timestamp)}
                      </p>
                    </div>
                  </div>
                </div>
              );
            })}


            {/* Typing Indicator - แสดงเฉพาะเมื่อกำลัง typing และไม่มี content ใน bot message ล่าสุด */}
            {isTyping && (
              <div className="flex justify-start">
                <div className="flex mr-2">
                  <div className="w-6 h-6 rounded-full flex items-center justify-center bg-muted text-muted-foreground">
                    <Bot className="h-3 w-3" />
                  </div>
                </div>
                <div className="bg-muted rounded-lg px-3 py-2">
                  <div className="flex space-x-1">
                    <div className="w-1.5 h-1.5 bg-muted-foreground rounded-full animate-bounce"></div>
                    <div className="w-1.5 h-1.5 bg-muted-foreground rounded-full animate-bounce" style={{ animationDelay: '0.1s' }}></div>
                    <div className="w-1.5 h-1.5 bg-muted-foreground rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
                  </div>
                </div>
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>

          {/* Input */}
          <div className={`p-4 border-t border-border ${isFullscreen ? 'max-w-4xl mx-auto w-full' : ''}`}
          >
            <div className="flex space-x-2">
              <Textarea
                value={inputText}
                onChange={(e) => setInputText(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder={browserAIInputDisabled ? browserAIError || "อุปกรณ์นี้ยังไม่รองรับ Gemini Nano" : "พิมพ์คำถามเกี่ยวกับบทความ..."}
                className={`flex-1 min-h-9 px-3 py-2 text-sm ${isFullscreen ? 'max-h-32' : 'max-h-20'}`}
                rows={1}
                disabled={browserAIInputDisabled}
              />
              <button
                onClick={handleSendMessage}
                disabled={!inputText.trim() || isTyping || browserAIInputDisabled || wordLimitExceeded}
                className="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-9 px-3"
              >
                <Send className="h-4 w-4" />
              </button>
            </div>
            <div className="flex items-center justify-between mt-2">
              <p className={`text-xs ${wordError ? 'text-red-500 font-bold dark:text-red-500' : 'text-muted-foreground'}`}>
                {wordError
                  ? wordError
                  : (browserAIInputDisabled ? browserAIError || "Gemini Nano ใช้งานไม่ได้บน browser หรืออุปกรณ์นี้" : "Enter เพื่อส่ง • Shift+Enter บรรทัดใหม่")}
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default BlogAIChat;
