"use client";

import { useEffect, useState } from "react";
import { v4 as uuidv4 } from "uuid";
import { TurboZone } from "@/components/TurboZone";
import { ChatPanel } from "@/components/ChatPanel";
import { SummaryCard } from "@/components/SummaryCard";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";

export default function Home() {
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [isUploaded, setIsUploaded] = useState(false);
  const [summaryData, setSummaryData] = useState<{
    summary: string;
    entities: string[];
  } | null>(null);

  useEffect(() => {
    let id = localStorage.getItem("rag_session_id");
    if (!id) {
      id = uuidv4();
      localStorage.setItem("rag_session_id", id);
    }
    // Using setTimeout to defer state update until after render phase
    setTimeout(() => {
        setSessionId(id);
    }, 0);
  }, []);

  const handleWipe = async () => {
    if (!sessionId) return;
    try {
      await fetch("/api/wipe", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ session_id: sessionId }),
      });
      localStorage.removeItem("rag_session_id");
      const newId = uuidv4();
      localStorage.setItem("rag_session_id", newId);
      setSessionId(newId);
      setIsUploaded(false);
      setSummaryData(null);
      toast.success("Nuclear Wipe Complete! Session reset.");
    } catch {
      toast.error("Failed to wipe data");
    }
  };

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-900 via-indigo-950 to-slate-900 text-slate-100 p-8 font-sans">
      <div className="max-w-6xl mx-auto space-y-8">
        <header className="flex justify-between items-center backdrop-blur-xl bg-white/5 border border-white/10 p-6 rounded-3xl shadow-2xl">
          <div>
            <h1 className="text-4xl font-extrabold bg-clip-text text-transparent bg-gradient-to-r from-cyan-400 to-indigo-400">
              Turbo RAG
            </h1>
            <p className="text-slate-400 mt-1 font-medium">
              Zero-Login AI PDF Assistant
            </p>
          </div>
          <Button
            onClick={handleWipe}
            variant="destructive"
            className="bg-red-500/20 text-red-400 hover:bg-red-500 hover:text-white border border-red-500/30 transition-all font-bold tracking-wider uppercase shadow-[0_0_15px_rgba(239,68,68,0.3)] hover:shadow-[0_0_25px_rgba(239,68,68,0.6)]"
          >
            ☢️ Nuclear Wipe
          </Button>
        </header>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <div className="lg:col-span-1 space-y-6">
            <TurboZone
              sessionId={sessionId}
              onComplete={(summary: string, entities: string[]) => {
                setIsUploaded(true);
                setSummaryData({ summary, entities });
                toast.success("PDF processing complete!");
              }}
            />
            {isUploaded && summaryData && (
              <SummaryCard
                summary={summaryData.summary}
                entities={summaryData.entities}
              />
            )}
          </div>
          <div className="lg:col-span-2">
            <ChatPanel sessionId={sessionId} disabled={!isUploaded} />
          </div>
        </div>
      </div>
    </main>
  );
}
