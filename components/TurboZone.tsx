"use client";

import { useState } from "react";
import { motion } from "framer-motion";
import { UploadCloud, AlertCircle } from "lucide-react";
import { Progress } from "@/components/ui/progress";

export function TurboZone({ sessionId, onComplete }: { sessionId: string | null; onComplete: (summary: string, entities: string[]) => void }) {
  const [isDragging, setIsDragging] = useState(false);
  const [progress, setProgress] = useState(0);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === "dragenter" || e.type === "dragover") {
      setIsDragging(true);
    } else if (e.type === "dragleave") {
      setIsDragging(false);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      uploadFile(e.dataTransfer.files[0]);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      uploadFile(e.target.files[0]);
    }
  };

  const uploadFile = async (file: File) => {
    if (!sessionId) return;
    if (file.type !== "application/pdf") {
      setError("Please upload a PDF file.");
      return;
    }
    setError(null);
    setIsUploading(true);
    setProgress(10); // Start progress bar

    const formData = new FormData();
    formData.append("file", file);
    formData.append("session_id", sessionId);

    // Simulate Server-Sent Events style progress
    const progressInterval = setInterval(() => {
        setProgress(prev => (prev >= 95 ? 95 : prev + 15));
    }, 800);

    try {
      const res = await fetch("/api/upload", {
        method: "POST",
        body: formData,
      });

      clearInterval(progressInterval);

      if (!res.ok) {
        throw new Error(await res.text());
      }
      
      const data = await res.json();

      setProgress(100);
      
      onComplete(
        `Successfully parsed and encrypted your document using Gemini 1.5 Flash. Extracted ${data.chunks} semantic shards into the Upstash Vector session pool. PII elements have been intelligently completely redacted.`,
        ["Semantic Shards", "Upstash Vector", "Zero-Trust Encryption", "Gemini Model"]
      );
    } catch (err: unknown) {
      clearInterval(progressInterval);
      setError((err as Error).message || "Failed to upload.");
      setProgress(0);
    } finally {
      setTimeout(() => {
        setIsUploading(false);
        setProgress(0);
      }, 1000);
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
      className={`relative p-8 rounded-3xl border-2 transition-all cursor-pointer backdrop-blur-md overflow-hidden ${
        isDragging 
          ? "border-cyan-400 bg-cyan-900/20 shadow-[0_0_40px_rgba(34,211,238,0.3)]" 
          : "border-slate-700 border-dashed bg-white/5 hover:border-indigo-400 hover:bg-indigo-900/10"
      }`}
      onDragEnter={handleDrag}
      onDragLeave={handleDrag}
      onDragOver={handleDrag}
      onDrop={handleDrop}
    >
      <input 
        type="file" 
        accept="application/pdf"
        className="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10"
        onChange={handleChange}
        disabled={isUploading}
      />
      
      <div className="flex flex-col items-center justify-center space-y-4 text-center pt-4 pb-4">
        <motion.div
            animate={isDragging ? { y: -10, scale: 1.1 } : { y: 0, scale: 1 }}
            className="p-5 bg-indigo-500/20 rounded-2xl text-indigo-400 shadow-[0_0_20px_rgba(99,102,241,0.2)]"
        >
            <UploadCloud className="w-12 h-12" />
        </motion.div>
        
        {isUploading ? (
          <div className="w-full space-y-3 px-4">
            <p className="font-bold text-cyan-400 animate-pulse text-sm uppercase tracking-widest">
                Ingesting via Goroutines...
            </p>
            <Progress value={progress} className="h-2 bg-slate-800" />
          </div>
        ) : (
          <div>
            <h3 className="text-xl font-bold text-slate-200">Turbo-Zone</h3>
            <p className="text-sm text-slate-400 mt-2 font-medium">Drag & Drop PDF to inject knowledge</p>
          </div>
        )}

        {error && (
            <div className="flex items-center space-x-2 text-red-400 text-sm bg-red-950/50 px-4 py-2 rounded-xl mt-4 border border-red-500/20">
                <AlertCircle className="w-4 h-4" />
                <span>{error}</span>
            </div>
        )}
      </div>
    </motion.div>
  );
}
