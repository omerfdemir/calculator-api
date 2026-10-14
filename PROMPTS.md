# AI Prompts

This file records the main prompts that built this project, in order. Each entry has the prompt text and 1 line about the result. Short follow-up questions are not listed.

Tool: Claude Code.

## 1. Implement the spec

```text
# Task: Implement the full-stack calculator from SPEC.md

You are a senior full-stack engineer. You work in this repository: calculator-api.

## Source of truth

- Read SPEC.md completely before you write any code.
- SPEC.md is the source of truth for behavior, the API contract, error codes,
  operations, the Docker Compose design, test seams and scope. Implement ALL
  of it, including the optional parts: power, sqrt, percentage, the Docker
  Compose stack, the coverage report, the Playwright end-to-end tests and CI.
- Do not add features that are not in SPEC.md. Do not change a decision in
  SPEC.md without asking me first.
- If SPEC.md is unclear or wrong, stop and ask me. Do not guess. If I agree to
  a change, update SPEC.md in the same commit as the code.

## Tech stack

- Backend: Go (latest stable), standard library `net/http` and `log/slog`.
  No web framework. `golangci-lint` for lint.
- Frontend: React + TypeScript (strict) with Vite. Vitest, React Testing
  Library, MSW, ESLint, Prettier. No UI component library. No `any`.
- End-to-end: Playwright + `@axe-core/playwright`, against the Compose stack.
- Docker: Docker Compose with 2 services:
  - `api`: multi-stage Go build, final image `gcr.io/distroless/static`,
    non-root, port not published to the host.
  - `web`: multi-stage Node build, final image `nginxinc/nginx-unprivileged`
    (alpine), non-root, published as localhost:8080. nginx serves the
    frontend and forwards `/api/` to `api`.
- CI: GitHub Actions.
- Use exact, pinned versions for all dependencies and base images.

## Way of work

1. **Plan first.** Show me a short plan: the file tree, the phases below,
   and any question about SPEC.md. Wait for my "go".
2. **Test first (red, green, refactor).** For each behavior in SPEC.md, write
   the failing test at the seam that SPEC.md names. Then write the code that
   makes it pass. Then refactor.
3. **Test only through the seams in SPEC.md.** Test external behavior, not
   internal functions or component structure.
4. **Small commits.** Use conventional commit messages (feat:, test:, fix:,
   docs:, chore:, ci:). Each commit leaves the repository green.
5. **Verify before you say "done".** Run the full check for the phase. Fix
   every failure. Show me the real command output, not a summary from memory.

## Phases

### Phase 1: Repository setup
- Root files: .gitignore, .editorconfig, Makefile, PROMPTS.md.
- Makefile targets: `run-backend`, `run-frontend`, `test`, `lint`,
  `coverage`, `up` (= `docker compose up --build -d --wait`), `down`,
  `logs`, `e2e`, `check` (= lint + test + up + e2e + down).
- Check: `make lint` runs without errors on the empty projects.

### Phase 2: Backend
- Calculator domain module with the operation registry, then the API module,
  then the server entry point (API and `/healthz` only, plus the
  `healthcheck` subcommand).
- Cover every row of the operations table and the error code table in SPEC.md.
- Check: `go vet ./...`, `golangci-lint run`, `go test -race -cover ./...`.
  Coverage >= 90% for the calculator and API modules.

### Phase 3: Frontend
- API client, error message map, calculator state hook, components, styles.
- The Vite dev server proxies `/api` to the backend on port 8080, so the
  dev setup uses the same paths as nginx.
- Check: `npm run lint`, `npm run typecheck`, `npm test -- --coverage`.
  Line coverage >= 85%.

### Phase 4: Docker Compose
- A Dockerfile for the backend, a Dockerfile for the frontend, the nginx
  configuration, and the Compose file, as the "Docker Compose" section of
  SPEC.md describes: internal-only `api`, published `web`, `/api/` proxy,
  `index.html` fallback, cache headers, security headers, gzip, body limit,
  health checks, and `depends_on` with `service_healthy`.
- Check:
  - `docker compose config` is valid.
  - `make up` starts both services, and both are healthy.
  - `curl -fsS localhost:8080/` returns the app.
  - `curl -fsS localhost:8080/unknown/path` returns `index.html`.
  - `curl -fsS localhost:8080/api/v1/operations` returns the list.
  - One `curl` per operation through nginx returns the correct result.
  - Port 8080 of `api` is not reachable from the host.

### Phase 5: Playwright end-to-end tests
- All flows in the end-to-end seam of SPEC.md: all 7 operations, the error
  messages, keyboard-only use, the 360x800 mobile viewport, the axe scan,
  the deep-link reload and the API request through nginx.
- Playwright runs against `BASE_URL` (default http://localhost:8080) and
  does not use mocks. `make e2e` expects the stack to run (`make up`).
- Check: `make up && make e2e && make down` passes.

### Phase 6: CI
- One workflow with the jobs in SPEC.md. The end-to-end job runs
  `make up`, `make e2e`, prints `docker compose logs` if a step fails,
  and always runs `make down`.
- Upload the coverage reports and the Playwright HTML report as artifacts.
- Check: the workflow is valid YAML, and each job runs the same Makefile
  targets that I run on my machine.

### Phase 7: Documentation
README.md must have these sections:
1. Overview and a screenshot (take it with Playwright).
2. Tech stack, and a small diagram of the 2 services:
   browser -> nginx (`web`) -> Go API (`api`).
3. Prerequisites with versions.
4. Setup, and how to run: the backend, the frontend (dev mode), and the full
   stack with `docker compose up --build`.
5. Tests: how to run each test level, and how to make the coverage report.
   Include a coverage summary table with the real numbers.
6. API reference: a `curl` example (through localhost:8080/api/...) and the
   real response for every operation and every error code.
7. Design decisions and assumptions: copy the reasons from "Further Notes" in
   SPEC.md (registry, backend calculates, percentage definition, float64
   precision, 400 vs 422, 2 services with nginx).
8. Possible future work.

Also:
- Link SPEC.md from the README.
- PROMPTS.md: list every prompt used in this project, in order, with 1 line
  about the result of each. The first entries are the prompt that made
  SPEC.md, the prompt that changed the Docker design, and this prompt.

## Definition of done

- `make check` passes on a clean clone.
- `docker compose up --build` gives a working app at http://localhost:8080,
  and both services report healthy.
- Coverage targets in SPEC.md are met, and the README shows the real numbers.
- Every user story in SPEC.md has at least 1 test. Give me a table:
  user story number -> test name(s).
- README.md and PROMPTS.md are complete.
- `git status` is clean, and the history has small, meaningful commits.
```

Result: the agent read `SPEC.md`, asked about the scope, and made a plan. After the scope change in entry 3, it built the backend, the frontend, the Compose stack, the Playwright tests, CI and the README, test first and in small commits.

## 2. `/to-spec`: write the spec

```text
/to-spec Full-stack calculator: React + TS frontend, Go backend, with all optional parts (power, sqrt, percentage, Dockerfile, coverage report)
```

Result: `SPEC.md` with the user stories, the API contract, the error codes and 4 test seams.

## 3. Add multi-operator expressions

```text
Why we decided only one operators for our system?
```

```text
Ok you can delete SPEC.md for - Expressions with more than 1 operator, operator precedence, or brackets (for example `2 + 3 × 4`). - I think we should also have this

For this ->
- **Why 1 endpoint with a registry.** The alternative is 1 endpoint per operation. A single endpoint gives 1 validation path, 1 error format, and an operations list that the frontend can read. A new operation is 1 registry entry and does not need a new route.

I am not sure if one endpoint is good enough I guess so

And also make PROMPT.md with my first prompt and in addition to these and also our /to-spec one we used
```

Result: `SPEC.md` got multi-operator expressions, brackets, a phone-style keypad, the phone percentage rule and 1 expression endpoint. This file was created.
