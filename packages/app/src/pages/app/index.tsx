import { useClips, useStats, type Clip } from '@clipboard/hooks/useClips';
import { NextPage } from 'next';
import Head from 'next/head';
import { useCallback, useEffect, useRef, useState } from 'react';

// ── helpers ──────────────────────────────────────────────────────────────────

function timeAgo(dateStr: string): string {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  const diff = Date.now() - d.getTime();
  const s = Math.floor(diff / 1000);
  if (s < 60) return `${s}s ago`;
  const m = Math.floor(s / 60);
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h ago`;
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

// ── ClipRow ───────────────────────────────────────────────────────────────────

function ClipRow({
  clip,
  onCopy,
  copiedId,
}: {
  clip: Clip;
  onCopy: (id: number, content: string) => void;
  copiedId: number | null;
}) {
  const [expanded, setExpanded] = useState(false);
  const lines = clip.content.split('\n');
  const isLong = lines.length > 3 || clip.content.length > 140;
  const displayText =
    isLong && !expanded
      ? clip.content.slice(0, 140).replace(/\n/g, ' ↵ ') + '…'
      : clip.content;

  const isCopied = copiedId === clip.id;

  return (
    <div
      className={`card card-compact bg-base-200 border transition-colors duration-150 ${
        clip.pinned
          ? 'border-warning/40'
          : 'border-base-300 hover:border-base-content/20'
      }`}>
      <div className="card-body gap-2">
        {/* top row */}
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-base-content/30 font-mono text-xs">
            #{clip.id}
          </span>
          {clip.pinned && (
            <span className="badge badge-warning badge-xs">📌 pinned</span>
          )}
          <div className="ml-auto flex items-center gap-2">
            <span className="text-base-content/30 text-xs">
              {timeAgo(clip.created_at)}
            </span>
            {clip.source && (
              <span className="badge badge-ghost badge-xs font-mono">
                {clip.source}
              </span>
            )}
          </div>
        </div>

        {/* content */}
        <pre className="text-base-content/80 font-mono text-xs leading-relaxed break-all whitespace-pre-wrap">
          {displayText}
        </pre>

        {/* footer */}
        <div className="mt-1 flex items-center justify-between">
          {isLong ? (
            <button
              className="btn btn-ghost btn-xs text-base-content/40 font-mono"
              onClick={() => setExpanded((e) => !e)}>
              {expanded ? '▲ collapse' : `▼ ${lines.length} lines`}
            </button>
          ) : (
            <span />
          )}
          <button
            className={`btn btn-xs font-mono ${
              isCopied
                ? 'btn-success'
                : 'btn-ghost border-base-300 hover:border-primary hover:text-primary border'
            }`}
            onClick={() => onCopy(clip.id, clip.content)}>
            {isCopied ? '✓ copied' : '⎘ copy'}
          </button>
        </div>
      </div>
    </div>
  );
}

// ── Page ──────────────────────────────────────────────────────────────────────

const AppPage: NextPage = () => {
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [pinnedOnly, setPinnedOnly] = useState(false);
  const [copiedId, setCopiedId] = useState<number | null>(null);
  const searchRef = useRef<HTMLInputElement>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const { clips, dbPath, loading, error, refetch } = useClips({
    search: debouncedSearch,
    pinnedOnly,
    limit: 200,
  });
  const stats = useStats();

  const handleSearch = (val: string) => {
    setSearch(val);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => setDebouncedSearch(val), 220);
  };

  // '/' focuses search
  useEffect(() => {
    const h = (e: KeyboardEvent) => {
      if (e.key === '/' && document.activeElement !== searchRef.current) {
        e.preventDefault();
        searchRef.current?.focus();
      }
      if (e.key === 'Escape') searchRef.current?.blur();
    };
    window.addEventListener('keydown', h);
    return () => window.removeEventListener('keydown', h);
  }, []);

  const handleCopy = useCallback(async (id: number, content: string) => {
    try {
      await navigator.clipboard.writeText(content);
      setCopiedId(id);
      setTimeout(() => setCopiedId(null), 1600);
    } catch {
      /* ignore */
    }
  }, []);

  const pinned = clips.filter((c) => c.pinned);
  const recent = clips.filter((c) => !c.pinned);

  return (
    <>
      <Head>
        <title>Clipboard</title>
      </Head>

      <div
        className="bg-base-100 flex h-screen overflow-hidden"
        data-theme="luxury">
        {/* ── Sidebar ── */}
        <aside className="border-base-300 bg-base-200 flex w-52 flex-shrink-0 flex-col overflow-hidden border-r">
          {/* logo */}
          <div className="border-base-300 border-b px-4 pt-5 pb-4">
            <h1 className="text-primary font-mono text-lg font-bold tracking-widest uppercase">
              Clipboard
            </h1>
            {dbPath && (
              <p
                className="text-base-content/30 mt-1 truncate font-mono text-[9px]"
                title={dbPath}>
                {dbPath.replace(/^.*\//, '…/')}
              </p>
            )}
          </div>

          {/* stats */}
          {stats && (
            <div className="divide-base-300 border-base-300 grid grid-cols-2 divide-x border-b">
              <div className="flex flex-col items-center py-3">
                <span className="text-base-content text-2xl font-bold">
                  {stats.total}
                </span>
                <span className="text-base-content/30 text-[9px] tracking-widest uppercase">
                  total
                </span>
              </div>
              <div className="flex flex-col items-center py-3">
                <span className="text-warning text-2xl font-bold">
                  {stats.pinned}
                </span>
                <span className="text-base-content/30 text-[9px] tracking-widest uppercase">
                  pinned
                </span>
              </div>
            </div>
          )}

          {/* nav */}
          <nav className="flex flex-1 flex-col gap-1 p-2">
            <button
              className={`btn btn-sm justify-start gap-2 font-mono text-xs ${
                !pinnedOnly
                  ? 'btn-primary btn-outline'
                  : 'btn-ghost text-base-content/40'
              }`}
              onClick={() => setPinnedOnly(false)}>
              ◈ All clips
            </button>
            <button
              className={`btn btn-sm justify-start gap-2 font-mono text-xs ${
                pinnedOnly
                  ? 'btn-warning btn-outline'
                  : 'btn-ghost text-base-content/40'
              }`}
              onClick={() => setPinnedOnly(true)}>
              📌 Pinned only
            </button>
          </nav>
        </aside>

        {/* ── Main ── */}
        <div className="flex min-w-0 flex-1 flex-col overflow-hidden">
          {/* toolbar */}
          <div className="border-base-300 bg-base-100 flex flex-shrink-0 items-center gap-2 border-b px-4 py-3">
            <label className="input input-sm input-bordered bg-base-200 flex flex-1 items-center gap-2 font-mono">
              <span className="text-base-content/30 text-xs select-none">
                /
              </span>
              <input
                ref={searchRef}
                type="text"
                className="grow text-xs"
                placeholder="search clips…"
                value={search}
                onChange={(e) => handleSearch(e.target.value)}
              />
              {search && (
                <button
                  className="text-base-content/30 hover:text-base-content transition-colors"
                  onClick={() => handleSearch('')}>
                  ✕
                </button>
              )}
            </label>
            <button
              className="btn btn-sm btn-ghost btn-square text-base-content/40 hover:text-base-content text-lg"
              onClick={refetch}
              title="Refresh (r)">
              ↺
            </button>
          </div>

          {/* clip list */}
          <div className="flex-1 space-y-4 overflow-y-auto p-4">
            {loading && (
              <div className="text-base-content/40 flex h-48 items-center justify-center gap-3 font-mono text-sm">
                <span className="loading loading-spinner loading-sm" />
                reading database…
              </div>
            )}

            {!loading && error && (
              <div className="alert alert-error mx-auto mt-10 max-w-lg font-mono text-sm">
                <div className="flex flex-col gap-1">
                  <p className="font-bold">Could not open database</p>
                  <p className="text-xs break-all opacity-80">{error}</p>
                  <p className="mt-1 text-xs opacity-60">
                    Ensure{' '}
                    <code className="rounded bg-black/20 px-1">
                      ~/.clipboard/clipboard.db
                    </code>{' '}
                    exists and{' '}
                    <code className="rounded bg-black/20 px-1">rusqlite</code>{' '}
                    is in Cargo.toml.
                  </p>
                </div>
              </div>
            )}

            {!loading && !error && clips.length === 0 && (
              <div className="text-base-content/30 flex h-48 items-center justify-center font-mono text-sm">
                {debouncedSearch
                  ? `no results for "${debouncedSearch}"`
                  : 'no clips yet'}
              </div>
            )}

            {!loading && !error && clips.length > 0 && (
              <>
                {pinned.length > 0 && (
                  <section>
                    <p className="text-base-content/30 mb-2 ml-1 font-mono text-[9px] tracking-widest uppercase">
                      pinned — {pinned.length}
                    </p>
                    <div className="space-y-2">
                      {pinned.map((c) => (
                        <ClipRow
                          key={c.id}
                          clip={c}
                          onCopy={handleCopy}
                          copiedId={copiedId}
                        />
                      ))}
                    </div>
                  </section>
                )}

                {recent.length > 0 && (
                  <section>
                    {pinned.length > 0 && (
                      <p className="text-base-content/30 mb-2 ml-1 font-mono text-[9px] tracking-widest uppercase">
                        recent — {recent.length}
                      </p>
                    )}
                    <div className="space-y-2">
                      {recent.map((c) => (
                        <ClipRow
                          key={c.id}
                          clip={c}
                          onCopy={handleCopy}
                          copiedId={copiedId}
                        />
                      ))}
                    </div>
                  </section>
                )}

                <p className="text-base-content/20 py-4 text-center font-mono text-[10px]">
                  {clips.length} clips
                </p>
              </>
            )}
          </div>
        </div>
      </div>
    </>
  );
};

export default AppPage;
