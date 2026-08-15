const API_URL = process.env.API_URL ?? "http://localhost:8080";
const BASE_URL = process.env.PLAYWRIGHT_BASE_URL ?? "http://localhost:3000";

export default async function globalSetup() {
  try {
    const apiHealth = await fetch(`${API_URL}/health`);
    if (!apiHealth.ok) {
      throw new Error(`Go API health check failed (${apiHealth.status})`);
    }
  } catch {
    throw new Error(
      `Go POS API is not reachable at ${API_URL}. Start it with:\n` +
        `  DATABASE_URL=postgres://pos:pos@localhost:5434/pos?sslmode=disable go run ./cmd/pos-api`,
    );
  }

  try {
    const webResponse = await fetch(BASE_URL);
    if (!webResponse.ok) {
      throw new Error(`Next.js health check failed (${webResponse.status})`);
    }
  } catch {
    throw new Error(
      `Next.js POS web is not reachable at ${BASE_URL}. Start it with:\n` +
        `  cd apps/web && NEXT_PUBLIC_API_URL=${API_URL} npm run dev`,
    );
  }
}
