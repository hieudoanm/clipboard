// lib/useClips.ts
import { invoke } from '@tauri-apps/api/core';
import { useCallback, useEffect, useState } from 'react';

export type Clip = {
  id: number;
  content: string;
  source: string;
  created_at: string;
  pinned: boolean;
};

type ClipsResult = {
  clips: Clip[];
  db_path: string;
};

type Stats = {
  total: number;
  pinned: number;
  db_path: string;
};

type Options = {
  search?: string;
  pinnedOnly?: boolean;
  limit?: number;
};

export const useClips = (options: Options = {}) => {
  const { search = '', pinnedOnly = false, limit = 100 } = options;

  const [clips, setClips] = useState<Clip[]>([]);
  const [dbPath, setDbPath] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const fetch = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const result = await invoke<ClipsResult>('get_clips', {
        search: search || null,
        pinnedOnly,
        limit,
      });
      setClips(result.clips);
      setDbPath(result.db_path);
    } catch (e: unknown) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }, [search, pinnedOnly, limit]);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { clips, dbPath, loading, error, refetch: fetch };
};

export function useStats() {
  const [stats, setStats] = useState<Stats | null>(null);

  useEffect(() => {
    invoke<Stats>('get_stats').then(setStats).catch(console.error);
  }, []);

  return stats;
}
