'use client';

import Link from 'next/link';
import clsx from 'clsx';
import type { SearchResult } from '@/lib/types';

interface ResultCardProps {
  result: SearchResult;
}

export function ResultCard({ result }: ResultCardProps) {
  const relevancePercent = Math.round(result.relevance_score * 100);
  const sourceLabel = result.source === 'slack' ? '💬 Slack' : '🐙 GitHub';
  
  return (
    <Link href={`/results/${result.id}`}>
      <div className="p-4 border border-gray-200 rounded-lg hover:border-blue-500 hover:shadow-lg transition-all cursor-pointer h-full">
        {/* Source Badge */}
        <div className="flex items-center justify-between mb-3">
          <span className="text-sm font-semibold text-gray-600">{sourceLabel}</span>
          <span className={clsx(
            'text-sm font-bold px-2 py-1 rounded',
            relevancePercent >= 80 ? 'bg-green-100 text-green-800' :
            relevancePercent >= 60 ? 'bg-yellow-100 text-yellow-800' :
            'bg-gray-100 text-gray-800'
          )}>
            {relevancePercent}%
          </span>
        </div>

        {/* Title */}
        <h3 className="text-lg font-semibold text-gray-900 mb-2 line-clamp-2">
          {result.title}
        </h3>

        {/* Content Preview */}
        <p className="text-gray-600 text-sm mb-3 line-clamp-3">
          {result.content}
        </p>

        {/* Metadata Footer */}
        <div className="flex items-center justify-between text-xs text-gray-500">
          {result.metadata.author && (
            <span>By {result.metadata.author}</span>
          )}
          <span>
            {new Date(result.created_at).toLocaleDateString()}
          </span>
        </div>
      </div>
    </Link>
  );
}
