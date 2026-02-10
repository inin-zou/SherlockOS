'use client';

import { useState, useRef, useEffect, useCallback } from 'react';
import { Send, User, Bot, Loader2, Sparkles, RotateCcw, ArrowLeft } from 'lucide-react';
import Link from 'next/link';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/v1';

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  image?: string;
}

const SUGGESTIONS = [
  'Male, mid-30s, short dark hair, strong jawline, light stubble',
  'Female, early 20s, long black hair, round face, glasses',
  'Male, 50s, bald, heavy build, thick eyebrows, scar on right cheek',
  'Female, 40s, red curly hair, slim build, freckles, green eyes',
];

export default function PortraitPage() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [currentPortrait, setCurrentPortrait] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  const scrollToBottom = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [messages, scrollToBottom]);

  const handleSend = async (text?: string) => {
    const messageText = text || input.trim();
    if (!messageText || isGenerating) return;

    setError(null);
    const userMessage: Message = {
      id: crypto.randomUUID(),
      role: 'user',
      content: messageText,
    };

    const updatedMessages = [...messages, userMessage];
    setMessages(updatedMessages);
    setInput('');
    setIsGenerating(true);

    try {
      // Build API messages from conversation history
      const apiMessages = updatedMessages.map((m) => ({
        role: m.role === 'assistant' ? 'model' : 'user',
        content: m.content,
        image_base64: m.image || '',
      }));

      const response = await fetch(`${API_BASE}/portrait/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ messages: apiMessages }),
      });

      const data = await response.json();

      if (data.success && data.data) {
        const assistantMessage: Message = {
          id: crypto.randomUUID(),
          role: 'assistant',
          content: data.data.text || 'Portrait generated.',
          image: data.data.image_base64 || undefined,
        };
        setMessages((prev) => [...prev, assistantMessage]);
        if (data.data.image_base64) {
          setCurrentPortrait(data.data.image_base64);
        }
      } else {
        const errMsg = data.error?.message || 'Failed to generate portrait';
        setError(errMsg);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Network error');
    } finally {
      setIsGenerating(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleReset = () => {
    setMessages([]);
    setCurrentPortrait(null);
    setError(null);
    setInput('');
  };

  const hasConversation = messages.length > 0;

  return (
    <div className="h-screen bg-[#0a0a0c] text-[#f0f0f2] flex">
      {/* Left: Portrait Preview */}
      <div className="w-[480px] border-r border-[#1e1e24] flex flex-col">
        {/* Header */}
        <div className="h-14 border-b border-[#1e1e24] flex items-center px-5 gap-3 shrink-0">
          <Link href="/" className="text-[#606068] hover:text-[#a0a0a8] transition-colors">
            <ArrowLeft className="w-4 h-4" />
          </Link>
          <Sparkles className="w-4 h-4 text-[#3b82f6]" />
          <span className="text-sm font-semibold">Suspect Portrait</span>
        </div>

        {/* Portrait Display */}
        <div className="flex-1 flex items-center justify-center p-8">
          {currentPortrait ? (
            <div className="relative w-full max-w-[400px] aspect-square rounded-2xl overflow-hidden ring-1 ring-[#2a2a32]">
              <img
                src={`data:image/png;base64,${currentPortrait}`}
                alt="Suspect portrait"
                className="w-full h-full object-cover"
              />
              <div className="absolute bottom-3 right-3">
                <span className="text-[10px] bg-black/70 text-[#a0a0a8] px-2 py-1 rounded-full">
                  AI Generated
                </span>
              </div>
            </div>
          ) : (
            <div className="w-full max-w-[400px] aspect-square rounded-2xl border border-dashed border-[#2a2a32] flex flex-col items-center justify-center gap-4">
              <div className="w-20 h-20 rounded-full bg-[#111114] flex items-center justify-center">
                <User className="w-10 h-10 text-[#2a2a32]" />
              </div>
              <div className="text-center space-y-1">
                <p className="text-sm text-[#606068]">No portrait yet</p>
                <p className="text-xs text-[#404048]">Describe a suspect to generate</p>
              </div>
            </div>
          )}
        </div>

        {/* Portrait info */}
        {currentPortrait && (
          <div className="px-5 pb-5">
            <div className="bg-[#111114] rounded-xl p-4 space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-[11px] font-semibold text-[#606068] uppercase tracking-wider">Generation Info</span>
                <span className="text-[11px] text-[#404048]">{messages.filter(m => m.role === 'user').length} revision{messages.filter(m => m.role === 'user').length !== 1 ? 's' : ''}</span>
              </div>
              <p className="text-xs text-[#a0a0a8]">Powered by Gemini Nano Banana</p>
            </div>
          </div>
        )}
      </div>

      {/* Right: Chat */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Chat Header */}
        <div className="h-14 border-b border-[#1e1e24] flex items-center justify-between px-5 shrink-0">
          <span className="text-sm text-[#a0a0a8]">
            {hasConversation ? 'Describe changes to refine the portrait' : 'Describe the suspect to begin'}
          </span>
          {hasConversation && (
            <button
              onClick={handleReset}
              className="flex items-center gap-1.5 text-xs text-[#606068] hover:text-[#a0a0a8] transition-colors"
            >
              <RotateCcw className="w-3.5 h-3.5" />
              Start over
            </button>
          )}
        </div>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto">
          {!hasConversation ? (
            /* Empty State with Suggestions */
            <div className="h-full flex flex-col items-center justify-center px-8">
              <div className="max-w-lg w-full space-y-8">
                <div className="text-center space-y-3">
                  <div className="w-12 h-12 rounded-2xl bg-[#3b82f6]/10 flex items-center justify-center mx-auto">
                    <Sparkles className="w-6 h-6 text-[#3b82f6]" />
                  </div>
                  <h2 className="text-lg font-semibold">Suspect Portrait Generator</h2>
                  <p className="text-sm text-[#606068] leading-relaxed">
                    Describe the suspect&apos;s physical appearance in natural language.
                    After generating, you can refine details iteratively.
                  </p>
                </div>

                <div className="space-y-2">
                  <span className="text-[11px] font-semibold text-[#606068] uppercase tracking-wider">Try a description</span>
                  <div className="grid grid-cols-1 gap-2">
                    {SUGGESTIONS.map((suggestion, i) => (
                      <button
                        key={i}
                        onClick={() => handleSend(suggestion)}
                        className="text-left px-4 py-3 rounded-xl bg-[#111114] border border-[#1e1e24] text-sm text-[#a0a0a8] hover:border-[#3b82f6]/40 hover:text-[#f0f0f2] transition-all"
                      >
                        {suggestion}
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          ) : (
            /* Chat Messages */
            <div className="px-5 py-6 space-y-5">
              {messages.map((message) => (
                <div key={message.id} className="flex gap-3">
                  {/* Avatar */}
                  <div className={`w-7 h-7 rounded-lg shrink-0 flex items-center justify-center ${
                    message.role === 'user' ? 'bg-[#1f1f24]' : 'bg-[#3b82f6]/10'
                  }`}>
                    {message.role === 'user' ? (
                      <User className="w-3.5 h-3.5 text-[#a0a0a8]" />
                    ) : (
                      <Bot className="w-3.5 h-3.5 text-[#3b82f6]" />
                    )}
                  </div>

                  {/* Content */}
                  <div className="flex-1 min-w-0 space-y-3">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-semibold text-[#a0a0a8]">
                        {message.role === 'user' ? 'You' : 'SherlockOS'}
                      </span>
                    </div>
                    <p className="text-sm text-[#d0d0d4] leading-relaxed">{message.content}</p>
                    {message.image && (
                      <div className="w-64 rounded-xl overflow-hidden ring-1 ring-[#2a2a32]">
                        <img
                          src={`data:image/png;base64,${message.image}`}
                          alt="Generated portrait"
                          className="w-full aspect-square object-cover"
                        />
                      </div>
                    )}
                  </div>
                </div>
              ))}

              {/* Generating indicator */}
              {isGenerating && (
                <div className="flex gap-3">
                  <div className="w-7 h-7 rounded-lg shrink-0 flex items-center justify-center bg-[#3b82f6]/10">
                    <Bot className="w-3.5 h-3.5 text-[#3b82f6]" />
                  </div>
                  <div className="flex items-center gap-2 py-2">
                    <Loader2 className="w-4 h-4 text-[#3b82f6] animate-spin" />
                    <span className="text-sm text-[#606068]">Generating portrait...</span>
                  </div>
                </div>
              )}

              {/* Error */}
              {error && (
                <div className="mx-10 px-4 py-3 rounded-xl bg-[#ef4444]/10 border border-[#ef4444]/20 text-sm text-[#ef4444]">
                  {error}
                </div>
              )}

              <div ref={messagesEndRef} />
            </div>
          )}
        </div>

        {/* Input Area */}
        <div className="border-t border-[#1e1e24] p-4">
          <div className="relative">
            <textarea
              ref={inputRef}
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={
                hasConversation
                  ? 'Describe what to change... (e.g., "make the hair shorter", "add glasses")'
                  : 'Describe the suspect\'s appearance...'
              }
              rows={2}
              disabled={isGenerating}
              className="w-full px-4 py-3 pr-12 bg-[#111114] border border-[#2a2a32] rounded-xl text-sm text-[#f0f0f2] placeholder:text-[#404048] focus:outline-none focus:ring-1 focus:ring-[#3b82f6] resize-none disabled:opacity-50"
            />
            <button
              onClick={() => handleSend()}
              disabled={!input.trim() || isGenerating}
              className="absolute right-3 bottom-3 w-8 h-8 rounded-lg bg-[#3b82f6] flex items-center justify-center text-white hover:bg-[#2563eb] transition-colors disabled:opacity-30 disabled:hover:bg-[#3b82f6]"
            >
              {isGenerating ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Send className="w-4 h-4" />
              )}
            </button>
          </div>
          <p className="mt-2 text-[10px] text-[#404048] text-center">
            Press Enter to send · Shift+Enter for new line
          </p>
        </div>
      </div>
    </div>
  );
}
