// Prevents additional console window on Windows in release
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod clips;

fn main() {
    tauri::Builder::default()
        .invoke_handler(tauri::generate_handler![
            clips::get_clips,
            clips::get_stats,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
