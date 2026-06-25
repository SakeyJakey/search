import React, { useState } from "react";
import { useLocation } from "wouter";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Search } from "lucide-react";

interface SearchBarProps {
  initialQuery?: string;
  size?: "lg" | "default";
}

export function SearchBar({ initialQuery = "", size = "default" }: SearchBarProps) {
  const [query, setQuery] = useState(initialQuery);
  const [, setLocation] = useLocation();

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (query.trim()) {
      setLocation(`/search?q=${encodeURIComponent(query.trim())}`);
    }
  };

  return (
    <form 
      onSubmit={handleSearch} 
      className={`relative flex items-center w-full max-w-2xl mx-auto shadow-sm rounded-full bg-card border transition-all duration-300 focus-within:ring-2 focus-within:ring-primary/20 focus-within:border-primary/50 ${size === "lg" ? "h-14" : "h-12"}`}
      data-testid="form-search"
    >
      <div className="pl-4 pr-2 text-muted-foreground">
        <Search size={size === "lg" ? 22 : 18} />
      </div>
      <Input
        type="search"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search the web..."
        className="border-0 bg-transparent shadow-none focus-visible:ring-0 px-2 text-base w-full h-full placeholder:text-muted-foreground/60 rounded-full"
        data-testid="input-search"
      />
      <div className="pr-2">
        <Button 
          type="submit" 
          size="sm" 
          className="rounded-full px-5 font-medium transition-transform hover:scale-105 active:scale-95"
          data-testid="button-submit-search"
        >
          Search
        </Button>
      </div>
    </form>
  );
}
