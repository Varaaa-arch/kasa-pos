This is a [Next.js](https://nextjs.org) project bootstrapped with [`create-next-app`](https://nextjs.org/docs/app/api-reference/cli/create-next-app).

## E2E Tests (Playwright)

Real browser E2E tests live in `e2e/` and exercise the full stack:

Browser → Next.js → Go API → PostgreSQL

### Prerequisites

Start these services before running E2E tests:

**Terminal 1 — PostgreSQL** (Docker example):

```bash
docker run -d --name pos-postgres -e POSTGRES_USER=pos -e POSTGRES_PASSWORD=pos -e POSTGRES_DB=pos -p 5434:5432 postgres:17
```

**Terminal 2 — Go POS API**:

```bash
DATABASE_URL=postgres://pos:pos@localhost:5434/pos?sslmode=disable go run ./cmd/pos-api
```

**Terminal 3 — Next.js**:

```bash
cd apps/web
NEXT_PUBLIC_API_URL=http://localhost:8080 npm run dev
```

**Terminal 4 — Print Agent** (optional; physical printing is a manual hardware smoke test):

```bash
go run ./experiments/printer/print-agent
```

### Run E2E

```bash
cd apps/web
npx playwright install chromium   # first time only
npm run test:e2e
```

Debug modes:

```bash
npm run test:e2e:headed
npm run test:e2e:ui
```

### Environment variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `PLAYWRIGHT_BASE_URL` | `http://localhost:3000` | Next.js app URL |
| `API_URL` | `http://localhost:8080` | Go API URL (fixtures + global setup) |
| `NEXT_PUBLIC_API_URL` | — | Required when starting Next.js |
| `PRINT_AGENT_URL` | `http://127.0.0.1:8081` | Optional print agent status check |

### What is verified

- Happy-path checkout (UI → API → DB → stock reduction → print agent)
- Insufficient payment rejection
- Multi-item checkout with stock updates
- Client-side product search
- Transaction existence via `GET /transactions/{id}`
- Print job status in checkout response (`print_job.status`)

Physical USB printer output (BP-LITE58) is **not** automated — checkout triggers the Print Agent over HTTP (`POST /print` on port 8081), but automated tests use a mock/test print server. Verify physical receipt printing manually.

## Getting Started

First, run the development server:

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
# or
bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

You can start editing the page by modifying `app/page.tsx`. The page auto-updates as you edit the file.

This project uses [`next/font`](https://nextjs.org/docs/app/building-your-application/optimizing/fonts) to automatically optimize and load [Geist](https://vercel.com/font), a new font family for Vercel.

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
