// Pubprix — moteur de génération d'image.
// Microservice interne, appelé uniquement par l'API Go. Voir le cahier
// des charges technique, section "Moteur de génération d'image (Rust)".

mod background;
mod compose;
mod download;
mod error;
mod export;
mod font;
mod generate;
mod model;
mod overlay;

use axum::{routing::get, routing::post, Router};

use generate::{generate_handler, AppState};

async fn health() -> &'static str {
    "ok"
}

#[tokio::main]
async fn main() {
    let state = AppState {
        http_client: reqwest::Client::new(),
    };

    let app = Router::new()
        .route("/healthz", get(health))
        .route("/generate", post(generate_handler))
        .with_state(state);

    let listener = tokio::net::TcpListener::bind("0.0.0.0:8081")
        .await
        .expect("impossible de démarrer le moteur sur le port 8081");

    println!("Pubprix engine à l'écoute sur :8081");
    axum::serve(listener, app).await.unwrap();
}
