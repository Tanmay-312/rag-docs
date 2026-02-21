"use client";

import { useState, useRef, useEffect } from "react";
import { motion } from "framer-motion";
import { Send, FileSearch, Sparkles, Loader2 } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Button } from "@/components/ui/button";

interface Message {
  role: "user" | "ai";
  text: string;
  citations?: string[];
}

export function ChatPanel({ sessionId, disabled }: { sessionId: string | null; disabled: boolean }) {
  const [messages, setMessages] = useState<Message[]>([
    { role: "ai", text: "System Online. Waiting for document context..." }
  ]);
  const [input, setInput] = useState("");
  const [isTyping, setIsTyping] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!disabled && messages.length === 1) {
      setMessages([{ role: "ai", text: "Context acquired. Ready for semantic queries. How can I help you analyze this document?" }]);
    }
  }, [disabled, messages.length]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || !sessionId || disabled) return;

    const userMsg = input.trim();
    setInput("");
    setMessages(prev => [...prev, { role: "user", text: userMsg }]);
    setIsTyping(true);

    setMessages(prev => [...prev, { role: "ai", text: "", citations: [] }]);

    try {
      const res = await fetch("/api/chat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ session_id: sessionId, message: userMsg })
      });

      if (!res.body) throw new Error("No response body");

      const reader = res.body.getReader();
      const decoder = new TextDecoder();

      let currentText = "";
      let currentCitations: string[] = [];

      while (true) {
        const { value, done } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value);
        const lines = chunk.split("\n\n");

        for (const line of lines) {
          if (line.startsWith("data: ")) {
            const dataStr = line.slice(6);
            if (dataStr === "[DONE]") break;
            
            try {
              const data = JSON.parse(dataStr);
              if (data.type === "citations") {
                currentCitations = data.citations || [];
                setMessages(prev => {
                  const newMsgs = [...prev];
                  newMsgs[newMsgs.length - 1].citations = currentCitations;
                  return newMsgs;
                });
              } else if (data.type === "text") {
                currentText += data.text;
                // Typewriter effect updates state per chunk
                setMessages(prev => {
                  const newMsgs = [...prev];
                  newMsgs[newMsgs.length - 1].text = currentText;
                  return newMsgs;
                });
              }
            } catch {
                // partial chunks
            }
          }
        }
      }
    } catch {
      setMessages(prev => {
        const newMsgs = [...prev];
        newMsgs[newMsgs.length - 1].text = "Error communicating with AI engine.";
        return newMsgs;
      });
    } finally {
      setIsTyping(false);
    }
  };

  useEffect(() => {
    if (scrollRef.current) {
        scrollRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages]);

  return (
    <div className={`flex flex-col h-[700px] bg-slate-900/60 border border-slate-700 rounded-3xl backdrop-blur-xl shadow-2xl overflow-hidden transition-all duration-500 ${disabled ? 'opacity-50 grayscale pointer-events-none' : ''}`}>
      <div className="p-5 border-b border-slate-700/50 flex items-center space-x-4 bg-slate-800/40 mx-3 mt-3 rounded-2xl">
        <div className="p-2.5 bg-fuchsia-500/20 rounded-xl relative">
            <div className="absolute inset-0 bg-fuchsia-500/20 animate-pulse rounded-xl" />
            <Sparkles className="w-5 h-5 text-fuchsia-400 relative z-10" />
        </div>
        <div>
            <h2 className="font-bold text-slate-100 tracking-wide text-lg">Semantic Search Engine</h2>
            <p className="text-xs text-fuchsia-400/80 font-medium">Powered by Gemini 1.5 Flash</p>
        </div>
      </div>

      <ScrollArea className="flex-1 p-6 z-10">
        <div className="space-y-6">
          {messages.map((m, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              className={`flex ${m.role === "user" ? "justify-end" : "justify-start"}`}
            >
              <div
                className={`max-w-[85%] rounded-3xl p-5 shadow-lg leading-relaxed text-[15px] font-medium ${
                  m.role === "user"
                    ? "bg-indigo-600/90 border border-indigo-400/30 text-white rounded-br-sm"
                    : "bg-slate-800/90 border border-slate-600/50 text-slate-200 rounded-bl-sm"
                }`}
              >
                {m.role === "ai" && isTyping && i === messages.length - 1 && m.text === "" && (
                    <div className="flex items-center space-x-3 text-fuchsia-400 py-1">
                        <Loader2 className="w-5 h-5 animate-spin" />
                        <span className="animate-pulse font-semibold tracking-wide uppercase text-sm">Synthesizing...</span>
                    </div>
                )}
                {m.text && <p className="whitespace-pre-wrap">{m.text}</p>}

                {m.citations && m.citations.length > 0 && (
                  <div className="mt-5 pt-5 border-t border-slate-600/50 space-y-3">
                    <div className="flex items-center space-x-2 text-cyan-400 text-xs font-bold uppercase tracking-wider">
                        <FileSearch className="w-4 h-4" />
                        <span>Source Citations Retrieved</span>
                    </div>
                    <div className="space-y-2">
                        {m.citations.map((cit, cIdx) => (
                          <div key={cIdx} className="bg-slate-900/80 p-3.5 rounded-xl border border-white/5 text-sm text-slate-400 italic shadow-inner">
                            &quot;{cit}&quot;
                          </div>
                        ))}
                    </div>
                  </div>
                )}
              </div>
            </motion.div>
          ))}
          <div ref={scrollRef} className="h-4" />
        </div>
      </ScrollArea>

      <div className="p-4 bg-slate-900/80 border-t border-slate-700/50">
          <form onSubmit={handleSubmit} className="flex gap-4 items-center relative">
            <input
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Ask a question about the document..."
              className="flex-1 bg-slate-800/80 border border-slate-600 rounded-2xl pl-6 pr-16 py-4 text-[15px] text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-fuchsia-500/50 transition-all font-medium shadow-inner"
              disabled={disabled || isTyping}
            />
            <Button 
                type="submit" 
                disabled={disabled || isTyping || !input.trim()}
                className="absolute right-2 rounded-xl h-10 w-12 p-0 bg-fuchsia-600 hover:bg-fuchsia-500 border-none shadow-[0_0_15px_rgba(217,70,239,0.4)] transition-all disabled:opacity-50"
            >
              <Send className="w-5 h-5 text-white" />
            </Button>
          </form>
      </div>
    </div>
  );
}
