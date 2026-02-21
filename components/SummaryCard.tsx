"use client";

import { motion } from "framer-motion";
import { FileText, Tag } from "lucide-react";

export function SummaryCard({ summary, entities }: { summary: string; entities: string[] }) {
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className="bg-white/5 border border-white/10 rounded-3xl p-6 backdrop-blur-xl shadow-2xl space-y-6 relative overflow-hidden group"
    >
      <div className="absolute inset-0 bg-gradient-to-br from-indigo-500/10 to-purple-500/10 opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
      <div className="relative">
        <div className="flex items-center space-x-3 text-indigo-400 mb-4">
            <div className="p-2 bg-indigo-500/20 rounded-lg">
                <FileText className="w-5 h-5" />
            </div>
            <h3 className="font-bold uppercase tracking-wider text-sm text-indigo-300">Executive Summary</h3>
        </div>
        <p className="text-slate-300 leading-relaxed text-sm font-medium">
            {summary}
        </p>
      </div>

      <div className="pt-5 border-t border-white/10 relative">
        <div className="flex items-center space-x-3 text-cyan-400 mb-4">
            <div className="p-2 bg-cyan-500/20 rounded-lg">
                <Tag className="w-5 h-5" />
            </div>
            <h3 className="font-bold uppercase tracking-wider text-sm text-cyan-300">Key Entities</h3>
        </div>
        <div className="flex flex-wrap gap-2">
            {entities.map((ent, i) => (
                <span key={i} className="px-3 py-1.5 bg-cyan-950/40 text-cyan-300 border border-cyan-500/20 rounded-full text-xs font-bold tracking-wide shadow-inner">
                    {ent}
                </span>
            ))}
        </div>
      </div>
    </motion.div>
  );
}
