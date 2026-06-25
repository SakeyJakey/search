import React, { useEffect, useState } from "react";
import { Link, useSearch as useWouterSearch } from "wouter";
import { SearchBar } from "@/components/SearchBar";
import { Skeleton } from "@/components/ui/skeleton";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Globe, ArrowRight } from "lucide-react";
import { useSearch, setBaseUrl } from "@workspace/api-client-react";

setBaseUrl("http://localhost:9091");

interface SearchResult {
  name: string;
  url: string;
}

export default function Search() {
  const searchString = useWouterSearch();
  const searchParams = new URLSearchParams(searchString);
  const query = searchParams.get("q") || "";

  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchType, setSearchType] = useState<"vector" | "tfidf">("vector");
  
  const searchMutation = useSearch();

  useEffect(() => {
    if (!query) return;
    
    setLoading(true);
    setError(null);
    
    // Call the new API
    searchMutation.mutate(
      { data: { query, type: searchType } },
      {
        onSuccess: (data) => {
          // Map API response (title) to UI expected format (name)
          setResults(data.map(r => ({ name: r.title, url: r.url })));
          setLoading(false);
        },
        onError: (err) => {
          console.error(err);
          setError("Unable to connect to search server. Please try again later.");
          setLoading(false);
        }
      }
    );
  }, [query, searchType]);

  return (
    <div className="min-h-screen w-full flex flex-col">
      {/* Header */}
      <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80">
        <div className="container mx-auto px-4 md:px-6 h-20 flex items-center gap-6 md:gap-10">
          <Link href="/" className="shrink-0 transition-opacity hover:opacity-80" data-testid="link-home">
            <h1 className="text-3xl font-serif italic text-primary tracking-tight">
              Lumina
            </h1>
          </Link>
          
            <div className="flex-1 max-w-3xl flex items-center gap-4">
              <div className="flex-1">
                <SearchBar initialQuery={query} />
              </div>
              <RadioGroup 
                value={searchType} 
                onValueChange={(value: "vector" | "tfidf") => setSearchType(value)}
                className="flex items-center gap-2 bg-muted/30 p-1 rounded-full border border-border"
              >
                <div className="flex items-center">
                  <RadioGroupItem value="vector" id="vector" className="sr-only" />
                  <Label 
                    htmlFor="vector" 
                    className={`cursor-pointer px-3 py-1 text-xs rounded-full transition-colors ${searchType === "vector" ? "bg-background shadow-sm text-foreground" : "text-muted-foreground"}`}
                  >
                    Vector
                  </Label>
                </div>
                <div className="flex items-center">
                  <RadioGroupItem value="tfidf" id="tfidf" className="sr-only" />
                  <Label 
                    htmlFor="tfidf" 
                    className={`cursor-pointer px-3 py-1 text-xs rounded-full transition-colors ${searchType === "tfidf" ? "bg-background shadow-sm text-foreground" : "text-muted-foreground"}`}
                  >
                    TF-IDF
                  </Label>
                </div>
              </RadioGroup>
            </div>

        </div>
      </header>

      {/* Main Content */}
      <main className="flex-1 container mx-auto px-4 md:px-6 py-8 md:py-12">
        <div className="max-w-3xl mx-auto w-full">
          {!query ? (
            <div className="py-20 text-center text-muted-foreground" data-testid="empty-query">
              Enter a search query to see results.
            </div>
          ) : loading ? (
            <div className="space-y-8" data-testid="loading-state">
              {[1, 2, 3, 4].map((i) => (
                <div key={i} className="space-y-3">
                  <Skeleton className="h-5 w-3/4 max-w-[400px]" />
                  <Skeleton className="h-4 w-1/2 max-w-[300px]" />
                  <Skeleton className="h-4 w-full" />
                </div>
              ))}
            </div>
          ) : error ? (
            <div className="py-12 px-6 rounded-2xl bg-destructive/5 border border-destructive/10 text-center" data-testid="error-state">
              <p className="text-destructive font-medium mb-2">{error}</p>
              <p className="text-muted-foreground text-sm">Make sure the local API server is running on port 9091.</p>
            </div>
          ) : results.length === 0 ? (
            <div className="py-20 text-center" data-testid="empty-results">
              <p className="text-lg text-foreground font-medium mb-2">No results found for "{query}"</p>
              <p className="text-muted-foreground">Try adjusting your search terms or using broader keywords.</p>
            </div>
          ) : (
            <div className="space-y-10" data-testid="search-results">
              <p className="text-sm text-muted-foreground" data-testid="text-results-count">
                Showing {results.length} result{results.length !== 1 ? 's' : ''} for <span className="font-medium text-foreground">{query}</span>
              </p>
              
              <div className="space-y-8">
                {results.map((result, index) => (
                  <article 
                    key={index} 
                    className="group relative flex flex-col gap-1 w-animate-in fade-in slide-in-from-bottom-2"
                    style={{ animationDelay: `${index * 50}ms` }}
                    data-testid={`card-result-${index}`}
                  >
                    <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
                      <Globe size={14} className="text-primary/60" />
                      <span className="truncate">{result.url}</span>
                    </div>
                    
                    <a 
                      href={result.url} 
                      target="_blank" 
                      rel="noopener noreferrer"
                      className="inline-block text-xl font-medium text-foreground group-hover:text-primary transition-colors focus-visible:outline-none focus-visible:underline underline-offset-4"
                      data-testid={`link-result-${index}`}
                    >
                      {result.name}
                    </a>
                    
                    <div className="flex items-center gap-2 text-primary/0 group-hover:text-primary transition-colors text-sm font-medium mt-1">
                      <span>Visit site</span>
                      <ArrowRight size={14} className="group-hover:translate-x-1 transition-transform" />
                    </div>
                  </article>
                ))}
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
