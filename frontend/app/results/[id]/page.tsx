'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { api } from '@/lib/api';
import type { SearchResult } from '@/lib/types';
import { ArrowLeftIcon } from '@radix-ui/react-icons';

interface ResultDetailPageProps {
  params: {
    id: string;
  };
}

export default function ResultDetail({ params }: ResultDetailPageProps) {
  const [result, setResult] = useState<SearchResult | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchResult = async () => {
      try {
        setIsLoading(true);
        const data = await api.getKnowledge(params.id);
        setResult(data);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch result');
      } finally {
        setIsLoading(false);
      }
    };

    fetchResult();
  }, [params.id]);

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="h-8 bg-gray-200 rounded w-1/2 animate-pulse" />
        <div className="space-y-2">
          <div className="h-4 bg-gray-200 rounded animate-pulse" />
          <div className="h-4 bg-gray-200 rounded w-5/6 animate-pulse" />
        </div>
      </div>
    );
  }

  if (error || !result) {
    return (
      <div className="text-center py-12">
        <p className="text-red-600 mb-4">{error || 'Result not found'}</p>
        <Link href="/" className="text-blue-500 hover:underline">
          Back to search
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <Link 
        href="/" 
        className="inline-flex items-center gap-2 text-blue-500 hover:text-blue-700 mb-4"
      >
        <ArrowLeftIcon className="w-4 h-4" />
        Back to results
      </Link>

      <div className="bg-white p-8 rounded-lg shadow-sm border border-gray-200">
        {/* Title and Badge */}
        <div className="flex items-start justify-between gap-4 mb-6">
          <div>
            <h1 className="text-4xl font-bold text-gray-900 mb-2">
              {result.title}
            </h1>
            <p className="text-gray-600">
              Relevance: {Math.round(result.relevance_score * 100)}%
            </p>
          </div>
          <span className="text-sm font-semibold px-4 py-2 bg-blue-100 text-blue-800 rounded">
            {result.source === 'slack' ? '💬 Slack' : '🐙 GitHub'}
          </span>
        </div>

        {/* Content */}
        <div className="prose prose-sm max-w-none mb-8">
          <p className="text-lg text-gray-700 leading-relaxed whitespace-pre-wrap">
            {result.content}
          </p>
        </div>

        {/* Metadata */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 py-6 border-t border-gray-200">
          {result.metadata.author && (
            <div>
              <p className="text-sm text-gray-600">Author</p>
              <p className="text-lg font-semibold text-gray-900">
                {result.metadata.author}
              </p>
            </div>
          )}
          {result.metadata.channel && (
            <div>
              <p className="text-sm text-gray-600">Channel</p>
              <p className="text-lg font-semibold text-gray-900">
                #{result.metadata.channel}
              </p>
            </div>
          )}
          <div>
            <p className="text-sm text-gray-600">Created</p>
            <p className="text-lg font-semibold text-gray-900">
              {new Date(result.created_at).toLocaleDateString('en-US', {
                year: 'numeric',
                month: 'long',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit',
              })}
            </p>
          </div>
          {result.metadata.url && (
            <div>
              <p className="text-sm text-gray-600">Source</p>
              <a 
                href={result.metadata.url} 
                target="_blank" 
                rel="noopener noreferrer"
                className="text-lg font-semibold text-blue-500 hover:underline break-all"
              >
                Open Original
              </a>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
