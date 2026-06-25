import React, { useEffect, useState, useRef } from "react";
import { Link } from "wouter";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Globe, ArrowRight } from "lucide-react";

const QUERIES = ["search engine", "vector database", "tf-idf explanation", "go programming", "postgresql performance"];

interface SearchResult {
  name: string;
  url: string;
}

export default function EvaluationPage() {
  const [queryQueue, setQueryQueue] = useState<{query: string, type: "vector" | "tfidf"}[]>([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [score, setScore] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    // Shuffle queries and assign random type
    const shuffled = [...QUERIES].sort(() => Math.random() - 0.5);
    const queue = shuffled.map(q => ({
      query: q,
      type: Math.random() > 0.5 ? "vector" : "tfidf" as "vector" | "tfidf"
    }));
    setQueryQueue(queue);
  }, []);

  useEffect(() => {
    if (queryQueue.length === 0 || currentIndex >= queryQueue.length) return;
    
    setLoading(true);
    setResults([]);
    
    // Fetch results
    fetch("http://localhost:9091/api/search", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ query: queryQueue[currentIndex].query, type: queryQueue[currentIndex].type })
    })
      .then(res => res.json())
      .then(data => {
        setResults(data.map((r: any) => ({ name: r.title, url: r.url })).slice(0, 3));
        setLoading(false);
        setTimeout(() => inputRef.current?.focus(), 100);
      })
      .catch(err => {
        console.error(err);
        setLoading(false);
      });
  }, [currentIndex, queryQueue]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const scoreVal = parseInt(score);
    if (isNaN(scoreVal) || scoreVal < 0 || scoreVal > 3) return;

    // Submit evaluation
    fetch("http://localhost:9091/api/evaluate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        query: queryQueue[currentIndex].query,
        type: queryQueue[currentIndex].type,
        score: scoreVal
      })
    }).then(() => {
      setScore("");
      setCurrentIndex(prev => prev + 1);
    });
  };

  if (currentIndex >= queryQueue.length && queryQueue.length > 0) {
    return <div className="p-10 text-center">Evaluation Complete!</div>;
  }

  return (
    <div className="min-h-screen container mx-auto p-10">
      <h1 className="text-2xl mb-5">Evaluation: {currentIndex + 1} / {queryQueue.length}</h1>
      <p className="text-lg mb-5 font-bold">Query: {queryQueue[currentIndex]?.query}</p>
      
      {loading ? (
        <Skeleton className="h-40" />
      ) : (
        <div className="space-y-4 mb-5">
          {results.map((r, i) => (
            <a 
              key={i} 
              href={r.url} 
              target="_blank" 
              rel="noopener noreferrer"
              className="block p-3 border rounded hover:bg-muted/50 transition-colors"
            >
              <p className="font-medium text-primary hover:underline">{r.name}</p>
              <p className="text-sm text-gray-500">{r.url}</p>
            </a>
          ))}
        </div>
      )}

      <form onSubmit={handleSubmit} className="flex gap-2">
        <Input 
          ref={inputRef}
          type="number" 
          value={score} 
          onChange={e => setScore(e.target.value)}
          placeholder="Score (0-3)"
          className="w-20"
          min="0"
          max="3"
        />
        <Button type="submit">Submit & Next</Button>
      </form>
    </div>
  );
}
