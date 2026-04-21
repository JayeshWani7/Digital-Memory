'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { ResultsList } from '@/components/ResultsList';
import type { SearchResult } from '@/lib/types';

export default function KnowledgePage() {
  const [items, setItems] = useState<SearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);

  useEffect(() => {
    const fetchKnowledge = async () => {
      try {
        setIsLoading(true);
        const response = await api.listKnowledge(page, 20);
        if (page === 1) {
          setItems(response.items || []);
        } else {
          setItems(prev => [...prev, ...(response.items || [])]);
        }
        setHasMore((response.items || []).length === 20);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch knowledge');
      } finally {
        setIsLoading(false);
      }
    };

    fetchKnowledge();
  }, [page]);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-4xl font-bold text-gray-900 mb-2">
          Knowledge Base
        </h1>
        <p className="text-lg text-gray-600">
          Browse all indexed knowledge from your sources
        </p>
      </div>

      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
          <p className="text-red-800">{error}</p>
        </div>
      )}

      <ResultsList 
        results={items} 
        isLoading={isLoading && page === 1}
        isEmpty={!isLoading && items.length === 0}
      />

      {hasMore && (
        <div className="text-center">
          <button
            onClick={() => setPage(p => p + 1)}
            disabled={isLoading}
            className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 transition"
          >
            {isLoading ? 'Loading...' : 'Load More'}
          </button>
        </div>
      )}
    </div>
  );
}
