import React, { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";

const QUERIES = [
  { q: "search engine", cat: "tech" },
  { q: "vector database", cat: "tech" },
  { q: "tf-idf explanation", cat: "tech" },
  { q: "go programming", cat: "tech" },
  { q: "postgresql performance", cat: "tech" }
].sort((a, b) => a.cat.localeCompare(b.cat));

interface EvaluationResult {
  name: string;
  url: string;
  type: "vector" | "tfidf";
  isRelevant: boolean;
}


export default function EvaluationPage() {
  const sortedQueries = React.useMemo(() => [...QUERIES].sort((a, b) => a.cat.localeCompare(b.cat)), []);
  const [queriesList] = useState(sortedQueries);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [results, setResults] = useState<EvaluationResult[]>([]);
  const [loading, setLoading] = useState(false);

  const currentQuery = queriesList[currentIndex];
  const categoryQueries = queriesList.filter(q => q.cat === currentQuery.cat);
  const indexInCategory = categoryQueries.indexOf(currentQuery);

  useEffect(() => {
    if (currentIndex >= queriesList.length) return;
    
    setLoading(true);
    setResults([]);
    
    const currentQuery = queriesList[currentIndex].q;

    // Fetch both simultaneously
    Promise.all([
      fetch("http://localhost:9091/api/search", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ query: currentQuery, type: "vector" })
      }).then(r => r.json()).then(data => data.map((r: any) => ({ ...r, type: "vector" as const }))),
      fetch("http://localhost:9091/api/search", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ query: currentQuery, type: "tfidf" })
      }).then(r => r.json()).then(data => data.map((r: any) => ({ ...r, type: "tfidf" as const })))
    ]).then(([vectorRes, tfidfRes]) => {
      // Take top 3 from each
      const combined = [
        ...vectorRes.slice(0, 3).map((r: any) => ({ name: r.title, url: r.url, type: r.type, isRelevant: false })),
        ...tfidfRes.slice(0, 3).map((r: any) => ({ name: r.title, url: r.url, type: r.type, isRelevant: false }))
      ];
      // Shuffle
      combined.sort(() => Math.random() - 0.5);
      setResults(combined);
      setLoading(false);
    });
  }, [currentIndex, queriesList]);

  const toggleRelevant = (index: number) => {
    setResults(prev => prev.map((r, i) => i === index ? { ...r, isRelevant: !r.isRelevant } : r));
  };

  const handleSubmit = () => {
    // Submit evaluation
    fetch("http://localhost:9091/api/evaluate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        query: queriesList[currentIndex].q,
        category: queriesList[currentIndex].cat,
        results: results
      })
    }).then(() => {
      setCurrentIndex(prev => prev + 1);
    });
  };

  if (currentIndex >= queriesList.length && queriesList.length > 0) {
    return <div className="p-10 text-center">Evaluation Complete!</div>;
  }

  return (
    <div className="min-h-screen container mx-auto p-10">
      <h1 className="text-2xl mb-5">Evaluation: {currentIndex + 1} / {queriesList.length}</h1>
      <div className="flex gap-2 mb-5 items-center">
        <p className="text-lg font-bold">Category: {currentQuery.cat} ({indexInCategory + 1} / {categoryQueries.length})</p>
      </div>
      <div className="flex gap-2 mb-5">
        <p className="text-lg font-bold">Query: {currentQuery.q}</p>
      </div>
      
      {loading ? (
        <Skeleton className="h-60" />
      ) : (
        <div className="space-y-4 mb-5">
          {results.map((r, i) => (
            <div key={i} className="flex items-center gap-3 p-3 border rounded">
              <input 
                type="checkbox" 
                checked={r.isRelevant} 
                onChange={() => toggleRelevant(i)}
                className="w-5 h-5"
              />
              <a 
                href={r.url} 
                target="_blank" 
                rel="noopener noreferrer"
                className="flex-1"
              >
                <p className="font-medium text-primary hover:underline">{r.name}</p>
                <p className="text-sm text-gray-500">{r.url}</p>
              </a>
            </div>
          ))}
        </div>
      )}

      <Button onClick={handleSubmit}>Submit & Next</Button>
    </div>
  );
}
