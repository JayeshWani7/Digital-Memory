'use client';

import { Suspense } from 'react';
import dynamic from 'next/dynamic';
import { SearchBar } from '@/components/SearchBar';
import { ResultsList } from '@/components/ResultsList';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { useState } from 'react';
import type { SearchQuery } from '@/lib/types';

export default function Home() {
  const [query, setQuery] = useState('');
  const [topK, setTopK] = useState(10);

  const { data: searchResults, isLoading, error } = useQuery({
    queryKey: ['search', query, topK],
    queryFn: async () => {
      if (!query.trim()) return null;
      const response = await api.search({ 
        query: query, 
        top_k: topK 
      });
      return response;
    },
    enabled: !!query.trim(),
    staleTime: 1000 * 60 * 5, // 5 minutes
  });

  const handleSearch = (q: string) => {
    setQuery(q);
  };

  return (
    <div className="space-y-12">
      {/* Hero Section */}
      <div className="text-center space-y-4 py-8">
        <h1 className="text-4xl md:text-5xl font-bold text-gray-900">
          Your Knowledge Base
        </h1>
        <p className="text-xl text-gray-600 max-w-2xl mx-auto">
          Semantically search across your Slack messages and GitHub discussions.
          Find what you need instantly with AI-powered understanding.
        </p>
      </div>

      {/* Search Section */}
      <div className="bg-white p-8 rounded-lg shadow-sm border border-gray-200">
        <SearchBar 
          onSearch={handleSearch} 
          isLoading={isLoading}
          placeholder="What do you want to find? (e.g., 'database decisions', 'API design')"
        />
        
        {/* Results Controls */}
        {query && (
          <div className="mt-4 flex items-center gap-4 text-sm text-gray-600">
            <span>
              Results: {searchResults?.total_count ?? 0} found
              {searchResults?.query_time_ms ? ` in ${searchResults.query_time_ms}ms` : ''}
            </span>
            <select 
              value={topK}
              onChange={(e) => setTopK(Number(e.target.value))}
              className="px-3 py-1 border border-gray-300 rounded text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value={5}>Top 5</option>
              <option value={10}>Top 10</option>
              <option value={20}>Top 20</option>
              <option value={50}>Top 50</option>
            </select>
          </div>
        )}
      </div>

      {/* Error State */}
      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
          <p className="text-red-800">
            Error: {error instanceof Error ? error.message : 'Failed to fetch results'}
          </p>
        </div>
      )}

      {/* Results Section */}
      {query && (
        <div className="space-y-4">
          <h2 className="text-2xl font-bold text-gray-900">
            Results for "{query}"
          </h2>
          <ResultsList 
            results={searchResults?.results ?? []} 
            isLoading={isLoading}
            isEmpty={!isLoading && !error && searchResults?.results?.length === 0}
          />
        </div>
      )}

      {/* Empty State */}
      {!query && (
        <div className="text-center space-y-6 py-12">
          <div className="text-6xl">🔍</div>
          <div>
            <p className="text-xl text-gray-600 mb-2">
              Start by searching for something
            </p>
            <p className="text-gray-500">
              Try questions like "What were the key decisions?" or "Find discussions about performance"
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
