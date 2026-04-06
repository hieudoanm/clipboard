// src-tauri/src/clips.rs
use rusqlite::{params, Connection};
use serde::Serialize;
use std::path::PathBuf;

#[derive(Debug, Serialize, Clone)]
pub struct Clip {
    pub id: i64,
    pub content: String,
    pub source: String,
    pub created_at: String,
    pub pinned: bool,
}

#[derive(Debug, Serialize)]
pub struct ClipsResult {
    pub clips: Vec<Clip>,
    pub db_path: String,
}

#[derive(Debug, Serialize)]
pub struct Stats {
    pub total: i64,
    pub pinned: i64,
    pub db_path: String,
}

fn db_path() -> PathBuf {
    let home = dirs::home_dir().expect("could not find home directory");
    home.join(".clipboard").join("clipboard.db")
}

fn open_db() -> Result<Connection, String> {
    let path = db_path();
    Connection::open_with_flags(
        &path,
        rusqlite::OpenFlags::SQLITE_OPEN_READ_ONLY
            | rusqlite::OpenFlags::SQLITE_OPEN_NO_MUTEX,
    )
    .map_err(|e| format!("Failed to open {:?}: {}", path, e))
}

fn query_clips(
    conn: &Connection,
    sql: &str,
    params: impl rusqlite::Params,
) -> Result<Vec<Clip>, String> {
    let mut stmt = conn.prepare(sql).map_err(|e| e.to_string())?;

    // 👇 Key fix: bind to variable so iterator drops before stmt
    let rows = stmt
        .query_map(params, map_row)
        .map_err(|e| e.to_string())?
        .filter_map(|r| r.ok())
        .collect::<Vec<_>>();

    Ok(rows)
}

#[tauri::command]
pub fn get_clips(
    search: Option<String>,
    pinned_only: bool,
    limit: Option<i64>,
) -> Result<ClipsResult, String> {
    let conn = open_db()?;
    let limit = limit.unwrap_or(100);

    let rows: Vec<Clip> = if let Some(ref q) = search.filter(|s| !s.trim().is_empty()) {
        let pattern = format!("%{}%", q.trim());

        if pinned_only {
            query_clips(
                &conn,
                "SELECT id, content, source, created_at, pinned
                 FROM clips
                 WHERE content LIKE ?1 AND pinned = 1
                 ORDER BY pinned DESC, created_at DESC
                 LIMIT ?2",
                params![pattern, limit],
            )?
        } else {
            query_clips(
                &conn,
                "SELECT id, content, source, created_at, pinned
                 FROM clips
                 WHERE content LIKE ?1
                 ORDER BY pinned DESC, created_at DESC
                 LIMIT ?2",
                params![pattern, limit],
            )?
        }
    } else if pinned_only {
        query_clips(
            &conn,
            "SELECT id, content, source, created_at, pinned
             FROM clips
             WHERE pinned = 1
             ORDER BY created_at DESC
             LIMIT ?1",
            params![limit],
        )?
    } else {
        query_clips(
            &conn,
            "SELECT id, content, source, created_at, pinned
             FROM clips
             ORDER BY pinned DESC, created_at DESC
             LIMIT ?1",
            params![limit],
        )?
    };

    Ok(ClipsResult {
        clips: rows,
        db_path: db_path().to_string_lossy().into_owned(),
    })
}

#[tauri::command]
pub fn get_stats() -> Result<Stats, String> {
    let conn = open_db()?;

    let total: i64 = conn
        .query_row("SELECT COUNT(*) FROM clips", [], |r| r.get(0))
        .map_err(|e| e.to_string())?;

    let pinned: i64 = conn
        .query_row("SELECT COUNT(*) FROM clips WHERE pinned = 1", [], |r| r.get(0))
        .map_err(|e| e.to_string())?;

    Ok(Stats {
        total,
        pinned,
        db_path: db_path().to_string_lossy().into_owned(),
    })
}

fn map_row(row: &rusqlite::Row<'_>) -> rusqlite::Result<Clip> {
    Ok(Clip {
        id: row.get(0)?,
        content: row.get(1)?,
        source: row.get::<_, Option<String>>(2)?.unwrap_or_default(),
        created_at: row.get::<_, Option<String>>(3)?.unwrap_or_default(),
        pinned: row.get::<_, i64>(4)? != 0,
    })
}
