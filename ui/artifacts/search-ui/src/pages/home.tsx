import React from "react";
import { SearchBar } from "@/components/SearchBar";

export default function Home() {
  return (
    <div className="min-h-screen w-full flex flex-col items-center justify-center p-6 relative overflow-hidden">
      {/* Subtle background element */}
      <div className="absolute inset-0 pointer-events-none opacity-[0.03]" 
           style={{ backgroundImage: 'radial-gradient(var(--primary) 1px, transparent 1px)', backgroundSize: '32px 32px' }} />
      
      <div className="w-full max-w-3xl flex flex-col items-center gap-10 relative z-10 -mt-20">
        <div className="flex flex-col items-center gap-3">
          <h1 
            className="text-6xl md:text-7xl font-serif italic text-primary tracking-tight"
            data-testid="text-logo-home"
          >
            Lumina
          </h1>
          <p className="text-muted-foreground font-medium tracking-wide text-sm md:text-base uppercase" data-testid="text-tagline">
            Precision Search
          </p>
        </div>
        
        <div className="w-full w-animate-in fade-in slide-in-from-bottom-4 duration-700">
          <SearchBar size="lg" />
        </div>
      </div>
    </div>
  );
}
