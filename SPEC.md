# Spec: Full-Stack Calculator

Status: `ready-for-agent`

## Problem Statement

A user wants to do arithmetic in a web browser, on a desktop or on a phone. The user wants the basic operations: addition, subtraction, multiplication and division. The user also wants power, square root and percentage.

The user wants to type a full calculation, such as `2 + 3 × 4`, on a keypad that works like a phone calculator. The calculator must use the standard order of operations and brackets.

The user needs clear results and clear error messages. Bad input, division by zero, or the square root of a negative number must not show a crash, a blank screen or `NaN`.

The reviewer of this project wants to run the full application with 1 command. The reviewer also wants to read the tests, see the test coverage, and understand the design decisions.

## Solution

The solution is a web calculator with 2 parts:

- A React and TypeScript frontend. It looks and works like a phone calculator. The user enters an expression with a keypad or a keyboard, and sees the result or an error message.
- A Go REST API. The API parses and calculates each expression, validates all input, and returns JSON.

The frontend does not calculate. It sends each expression to the API and shows the response. The frontend also gets the list of operations from the API, so the 2 parts always agree.

Docker Compose runs the application as 2 services. The `api` service runs the Go API. The `web` service runs nginx. nginx serves the frontend files and forwards `/api/` requests to the `api` service, so the browser uses 1 origin. The repository also has a README, a coverage report, a CI workflow, and a record of the main AI prompts.

## User Stories

### Calculation

1. As a user, I want to add 2 numbers, so that I get their sum.
2. As a user, I want to subtract 1 number from another, so that I get the difference.
3. As a user, I want to multiply 2 numbers, so that I get the product.
4. As a user, I want to divide 1 number by another, so that I get the quotient.
5. As a user, I want to raise a base to an exponent, so that I get the power.
6. As a user, I want to get the square root of a number, so that I get a value such as `√9 = 3`.
7. As a user, I want the `%` key to work like a phone calculator, so that `15% × 80 = 12` and `80 + 15% = 92`.
8. As a user, I want to use negative numbers and decimal numbers, so that I can do real calculations.
9. As a user, I want to use numbers in exponent notation (for example `1e6`), so that I can enter very large or very small values.
10. As a user, I want to enter a calculation with more than 1 operator, such as `2 + 3 × 4 = 14`, so that I do not calculate step by step.
11. As a user, I want `×` and `÷` to come before `+` and `−`, and `^` to come before both, so that the result follows the standard order of operations.
12. As a user, I want to use brackets, such as `(2 + 3) × 4 = 20`, so that I can change the order of operations.

### Keypad and input

13. As a user, I want a keypad like a phone calculator, so that I can calculate with taps only.
14. As a user, I want to type the expression with my keyboard, so that I can work fast on a desktop.
15. As a user, I want the input to accept only calculator characters, so that I do not send bad data.
16. As a user, I want an error message when the expression is empty, so that I know what to correct.
17. As a user, I want a clear error message that shows the position of a syntax error (for example a missing bracket), so that I can correct the expression.
18. As a user, I want to delete the last character, so that I can correct a typing error.
19. As a user, I want to continue from the full last result when I press an operator, and to start a new expression when I press a digit, so that the keypad works like a phone calculator and does not lose precision (`1 ÷ 3`, then `× 3`, gives `1`). A negative result continues in brackets (`−3`, then `^ 2`, gives `(−3) ^ 2 = 9`).
20. As a user, I want the `=` key to stay disabled while a request runs, so that I do not send the same request 2 times.
21. As a user, I want to see a loading state, so that I know the app is working.

### Results and errors

22. As a user, I want to see the result clearly together with the full expression (for example `80 + 15% = 92`), so that I can check what I calculated.
23. As a user, I want a clear message for division by zero, so that I understand why there is no result.
24. As a user, I want a clear message for the square root of a negative number, so that I understand the input is outside the domain.
25. As a user, I want a clear message when a result is too large, so that I do not see `Infinity`.
26. As a user, I want a clear message when the server is not available, so that I know the problem is not my input.
27. As a user, I want results such as `0.1 + 0.2` to show as `0.3`, so that the display does not show floating-point noise.

### Usability

28. As a user, I want to press Enter to calculate, so that I can work with the keyboard only.
29. As a user, I want to press Escape to clear everything, so that I can start again quickly.
30. As a user, I want to see my last 10 calculations, so that I can look at earlier results.
31. As a user, I want to clear the history, so that I can start with an empty list.
32. As a user on a phone, I want the layout to fit a 360px wide screen, and I do not want the phone keyboard to open, so that I can use the app without horizontal scroll.
33. As a user of a screen reader, I want labels on the input and on all keys, and announced results and error messages, so that I can use the app without a mouse.

### API client

34. As an API client, I want to send an expression as JSON and get the result as JSON, so that I can integrate with the service.
35. As an API client, I want a list of the supported operations with their symbols and number of operands, so that I can build a UI from it.
36. As an API client, I want a stable error code in each error response, so that I can handle errors without text parsing.
37. As an API client, I want an HTTP 400 status for bad requests and HTTP 422 for math errors, so that I can tell my errors from domain errors.
38. As an operator, I want a health endpoint, so that Docker and other tools can check that the service is alive.

### Reviewer and developer

39. As a reviewer, I want to start the full app with 1 command (`docker compose up --build`), so that I can test it without Go or Node on my machine.
40. As a developer, I want to run the backend and the frontend separately in development mode, so that I get fast reload.
41. As a reviewer, I want 1 command that runs all tests and makes a coverage report, so that I can check the quality.
42. As a reviewer, I want a README with setup steps, `curl` examples and design decisions, so that I understand the project quickly.
43. As a reviewer, I want a record of the main AI prompts, so that I can see how the project was built.
44. As a reviewer, I want CI to run lint, tests and the end-to-end tests on each push, so that I know the main branch works.

## Implementation Decisions

### Repository

- One repository with 2 top-level parts: the Go backend and the React frontend.
- Each part has its own Dockerfile. The frontend part also has the nginx configuration.
- Root files: a Compose file, a Makefile, a CI workflow, the README, `PROMPTS.md` and this spec.
- The Makefile is the single entry point for these tasks: run, test, lint, coverage, e2e, and the Compose stack (up and down).

### Backend modules

- **Calculator (domain module).** Pure functions. It has no HTTP code. It has 3 parts:
  - A registry of operations. Each entry has a name, a symbol, a notation (infix, prefix or postfix), an arity (the number of operands), a precedence, an associativity and a function. To add a new operation, a developer adds 1 registry entry.
  - A parser. It reads an expression and makes a syntax tree. It uses precedence climbing, and it gets the operators, the precedence and the associativity from the registry.
  - An evaluator. It calculates the syntax tree with the registry functions.

  The public API of the module is 1 function that evaluates an expression, plus the registry list. The module returns typed domain errors. A syntax error contains the position of the error.
- **API (HTTP module).** It decodes and validates requests, calls the calculator, maps domain errors to HTTP status codes and error codes, and encodes JSON. It also contains the middleware: request logging with `log/slog`, panic recovery, CORS, and a limit on the request body size (1 MB).
- **Server (entry point).** It reads the configuration from environment variables, connects the modules, serves only the API and the health endpoint, and shuts down gracefully on SIGINT and SIGTERM.
- The backend uses the Go standard library `net/http` with method and path routing (Go 1.22 or later). It uses no web framework.

### Configuration

| Variable | Default | Use |
|---|---|---|
| `PORT` | `8080` | The HTTP port. |
| `ALLOWED_ORIGINS` | `http://localhost:5173` | A comma-separated list of CORS origins. CORS is necessary only when the frontend and the API use different origins. In the Compose stack, nginx gives 1 origin, so CORS is not used. |

### Operations

All numbers, intermediate values and results are IEEE 754 `float64` values.

The table shows the operations in the fixed list order. A higher precedence binds more tightly.

| Name | Symbol | Notation | Arity | Precedence | Formula | Domain rules |
|---|---|---|---|---|---|---|
| `add` | `+` | infix, left-associative | 2 | 1 | a + b | none |
| `subtract` | `−` | infix, left-associative | 2 | 1 | a − b | none |
| `multiply` | `×` | infix, left-associative | 2 | 2 | a × b | none |
| `divide` | `÷` | infix, left-associative | 2 | 2 | a ÷ b | b = 0 gives `DIVISION_BY_ZERO` |
| `power` | `^` | infix, right-associative | 2 | 4 | a ^ b | a NaN result gives `INVALID_DOMAIN` (for example a negative base with a fractional exponent). `0 ^ negative` gives `DIVISION_BY_ZERO`. |
| `sqrt` | `√` | prefix | 1 | 3 | √a | a < 0 gives `INVALID_DOMAIN` |
| `percentage` | `%` | postfix | 1 | 5 | a ÷ 100, with the phone rule below | none |

These rules apply to all operations:

- A value of +Inf or −Inf at any step gives `RESULT_OUT_OF_RANGE`.
- A result of −0 becomes 0.
- The evaluation stops at the first error.

### Expression syntax

- A number is digits with an optional decimal part and an optional exponent, for example `12`, `0.5`, `.5`, `1e6` and `2.5e-3`. A number that is not finite as a `float64` (for example `1e999`) gives `INVALID_EXPRESSION`.
- Unary minus is allowed where a number can start, for example `−5`, `2 × −3`, `2 ^ −1` and `−(2 + 3)`. Unary minus has precedence 3, the same as `√`. So `−2 ^ 2 = −4`, and `√4 ^ 2 = √16 = 4`.
- `^` is right-associative: `2 ^ 3 ^ 2 = 2 ^ 9 = 512`.
- Brackets `(` and `)` change the order of operations.
- The API also accepts the ASCII symbols `*` for `×`, `/` for `÷` and `-` for `−`.
- Spaces and tabs are allowed between tokens and have no meaning.
- **Phone percentage rule.** `x%` is `x ÷ 100`. There is 1 exception: when the right operand of `+` or `−` is directly a `%` term, the percentage is of the left operand. So `a + b%` is `a + a × b ÷ 100`. Unary minus signs before the `%` term do not stop the rule, so `a + −b%` is `a − a × b ÷ 100`. Brackets stop the rule.
  - `50% = 0.5`
  - `15% × 80 = 12`
  - `80 + 15% = 92`
  - `80 − 15% = 68`
  - `80 + −15% = 68`
  - `80 − −15% = 92`
  - `80 + (−15%) = 79.85` (brackets: the right operand is not directly a `%` term)
  - `100 + 10% + 10% = 121`
  - `200 + 10% × 2 = 200.2` (the right operand of `+` is `10% × 2`, not a `%` term)
- These are not allowed and give `INVALID_EXPRESSION`: implicit multiplication (`2(3)`, `2√9`), an empty expression, an unknown character, a missing operand, a bracket without a pair, and empty brackets `()`.
- Limits: an expression has 200 characters at most, and brackets have a depth of 32 at most. A longer or deeper expression gives `INVALID_EXPRESSION`.
- A position in an error is the 0-based index of a Unicode character (a code point) in the expression. For an opening bracket without a closing bracket, the position is the opening bracket.

### API contract

**`POST /api/v1/calculate`**

Request:

```json
{ "expression": "80+15%" }
```

Success response (200). The `expression` field is the normalized form: Unicode symbols, 1 space around each infix operator, no space after a prefix operator or before a postfix operator, and the numbers as the client sent them.

```json
{ "expression": "80 + 15%", "result": 92 }
```

Error response (4xx):

```json
{ "error": { "code": "DIVISION_BY_ZERO", "message": "cannot divide by zero" } }
```

An `INVALID_EXPRESSION` error also has a `position` field. For the expression `2×(1+2`:

```json
{ "error": { "code": "INVALID_EXPRESSION", "message": "a closing bracket is missing", "position": 2 } }
```

**`GET /api/v1/operations`** (200):

```json
{ "operations": [ { "name": "add", "symbol": "+", "notation": "infix", "arity": 2 }, { "name": "sqrt", "symbol": "√", "notation": "prefix", "arity": 1 } ] }
```

The list order is fixed. It is the order of the operations table.

**`GET /healthz`** (200): `{ "status": "ok" }`

### Error codes

| Condition | HTTP status | Code |
|---|---|---|
| Malformed JSON, empty body, unknown JSON fields, `expression` is missing or is not a string, or `Content-Type` is not JSON | 400 | `INVALID_REQUEST` |
| Request body larger than 1 MB | 413 | `REQUEST_TOO_LARGE` |
| The expression is empty, too long, too deep, has an unknown character, a syntax error or a number that is not finite | 400 | `INVALID_EXPRESSION` (with `position`) |
| Division by zero | 422 | `DIVISION_BY_ZERO` |
| Input is outside the domain of the operation | 422 | `INVALID_DOMAIN` |
| A value overflows to infinity | 422 | `RESULT_OUT_OF_RANGE` |
| Unknown path | 404 | `NOT_FOUND` |
| Wrong HTTP method | 405 | `METHOD_NOT_ALLOWED` |
| Unexpected panic | 500 | `INTERNAL_ERROR` (the response never shows internal details) |

The error codes are part of the public contract. The message text can change.

### Frontend modules

- **API client.** A typed `fetch` wrapper. It reads the base URL from `VITE_API_URL`. The default is the same origin. It stops a request after 5 seconds with `AbortController`. It returns a discriminated union: a success with the result, or a failure with an error code (and a position for `INVALID_EXPRESSION`). A network failure or a timeout becomes the code `NETWORK_ERROR`. A 502, 503 or 504 response without the JSON error shape also becomes `NETWORK_ERROR`, because it comes from a proxy that cannot reach the API.
- **Error messages.** One map from each error code to a user-friendly message. The UI never shows raw server text. The message for `INVALID_EXPRESSION` shows the 1-based character position.
- **Calculator state hook.** It holds the expression, the status (`idle`, `loading`, `success`, `error`), the result and the history. It does the client-side validation and the phone keypad rules.
- **Components.** Calculator (the container), display (the expression input and the result line), keypad, error message, and history list.
- **Styles.** Plain CSS with CSS variables. No UI component library. The layout is mobile-first.

### Frontend behavior

- At start, the app loads the operations from the API. If this request fails, the app shows an error message with a Retry button.
- **Keypad.** The keypad has a fixed grid of 4 columns, like a phone calculator:

  | | | | |
  |---|---|---|---|
  | `AC` | `⌫` | `(` | `)` |
  | `√` | `^` | `%` | `÷` |
  | `7` | `8` | `9` | `×` |
  | `4` | `5` | `6` | `−` |
  | `1` | `2` | `3` | `+` |
  | `EE` | `0` | `.` | `=` |

  The keypad shows an operation key only for an operation in the list from the API. An operation in the list that is not in this grid gets a key in an extra row after the grid. `EE` inserts `e` for exponent notation. Each key has an accessible name, for example "divide" and "square root".
- **Expression input.** The display has 1 text input with a visible label. It uses `inputmode="none"`, so a phone does not open its own keyboard. A keypad key inserts its symbol at the end of the expression.
- **Keyboard.** The user can type in the input. The input accepts only digits, `.`, `e`, `E`, the operation symbols, `(`, `)` and spaces. It ignores other characters. It shows `*`, `/` and `-` as `×`, `÷` and `−`. Enter calculates. Escape clears the expression, the result and the error. Backspace deletes 1 character.
- **After a result.** An operation key or an operation character continues the expression from the result (for example the result `14`, then `×`, gives `14 ×`). A digit, `.`, `√` or `(` starts a new expression. The input shows the result as the display shows it, but the app sends the full `float64` value to the API. A negative result goes in brackets, in the input and in the request, because `^` binds more tightly than unary minus: the result `−3`, then `^ 2`, gives `(−3) ^ 2 = 9`, not `−3 ^ 2 = −9`. So `1 ÷ 3 = 0.333333333333`, then `× 3`, gives `0.333333333333 × 3 = 1`. The app sends the full value only while the expression still starts with the display text and no digit, `.` or `e` follows it. After another edit of the result, the app sends the expression as typed. The result line and the history show the display text, not the full value.
- The client-side validation is for user experience only. The API is the source of truth. The client checks only for an empty expression, filters the characters, and stops the input at 200 characters, the same limit as the API. The client does not parse the expression and does not check domain rules. These errors come from the API, so the frontend and the backend cannot disagree.
- The display shows the result line as `<normalized expression> = <result>`, with the normalized expression from the API.
- The display rounds a result to 12 significant digits and removes trailing zeros. The API returns the full `float64` value.
- The `=` key and Enter do nothing while a request runs. The display shows a loading state.
- The result line uses `role="status"`. The error message uses `role="alert"`.
- History holds the last 10 successful calculations in memory, the newest first. It does not persist after a page reload.

### Docker Compose

The stack has 2 services on 1 private Compose network.

| Service | Image | Port | Job |
|---|---|---|---|
| `api` | Multi-stage build. Stage 1 (Go) builds a static binary (`CGO_ENABLED=0`). Stage 2 is `gcr.io/distroless/static` as a non-root user. | 8080, internal only | Serves `/api/v1/*` and `/healthz`. |
| `web` | Multi-stage build. Stage 1 (Node) builds the frontend. Stage 2 is `nginxinc/nginx-unprivileged` (alpine) as a non-root user. | 8080 in the container, published as `localhost:8080` | Serves the frontend files and forwards `/api/` to `api`. |

- Only the `web` service publishes a port to the host. The `api` service is not reachable from the host. All requests go through nginx.
- nginx configuration:
  - `location /api/` forwards the request to `http://api:8080` and keeps the path. It sets the `X-Forwarded-For` and `X-Forwarded-Proto` headers.
  - `location /` serves the static files. For an unknown path, it returns `index.html` (`try_files`), so that a page reload on a deep link works.
  - Files with a content hash in the name get a long cache time (1 year, `immutable`). `index.html` gets `no-cache`.
  - It sets security headers: `X-Content-Type-Options: nosniff`, `Referrer-Policy`, and a `Content-Security-Policy` that allows only `'self'`.
  - It turns on gzip for text, JavaScript, CSS and JSON. It sets `client_max_body_size 1m`, the same limit as the API.
- The frontend build uses no `VITE_API_URL`, so the frontend calls the API on the same origin.
- Health checks:
  - `api`: the distroless image has no shell, so the health check uses a `healthcheck` subcommand of the server binary. This subcommand calls `/healthz` and returns exit code 0 or 1.
  - `web`: the health check requests `/` with `wget` in the nginx image.
  - `web` starts only after `api` is healthy (`depends_on` with `condition: service_healthy`).
- Both services use `restart: unless-stopped`.

### Tooling

- Backend: `go vet` and `golangci-lint`.
- Frontend: strict TypeScript, ESLint and Prettier. The code does not use `any`.
- CI (GitHub Actions) runs these jobs on each push and pull request:
  1. Backend lint and tests with coverage.
  2. Frontend lint, type check and tests with coverage.
  3. Compose build and start, then the Playwright tests against the running stack, then Compose stop.
  4. Upload of the coverage reports and the Playwright report as artifacts.

## Testing Decisions

### What makes a good test

- A good test checks external behavior through a public seam. It does not check internal functions, internal state or component structure.
- A test must stay green after a refactor that does not change behavior.
- Each acceptance case in this spec maps to at least 1 test.
- Backend tests are table-driven. Each table row has a name, an input and an expected output.

### Seams

There are 4 seams. Each seam has a different job.

1. **Backend HTTP API (primary backend seam).** Tests send real HTTP requests to the full router with `httptest`, and check the status code, the headers and the JSON body. This seam covers:
   - Every operation, with typical values.
   - Every row of the error code table, including the `position` field.
   - The normalized `expression` in the success response.
   - The JSON shape of success and error responses.
   - The operations list endpoint and the health endpoint.
   - CORS headers, the body size limit and panic recovery.
2. **Calculator domain module (parser and numeric edge cases).** Table-driven tests of the public calculator API: the evaluate function and the registry list. This seam covers the expression rules and the numeric edge cases, because they are clearer and faster without HTTP:
   - Precedence and associativity: `2 + 3 × 4`, `(2 + 3) × 4`, `2 ^ 3 ^ 2`, `−2 ^ 2`, `√4 ^ 2`, `10 − 4 − 3`, `8 ÷ 4 ÷ 2`.
   - Unary minus, nested brackets, ASCII symbols and spaces.
   - Every example of the phone percentage rule.
   - Syntax errors with their positions: missing operand, bracket without a pair, empty brackets, implicit multiplication, unknown character, empty expression, the length limit and the depth limit.
   - Overflow (`1e308 × 10`) and a number literal that is not finite (`1e999`).
   - `−0` normalization.
   - `0 ^ −1`, `(−8) ^ 0.5` and `√−1`.
   - `0.1 + 0.2` returns the raw `float64` value.
   - Very small and very large exponents.
3. **Frontend rendered app (primary frontend seam).** Vitest and React Testing Library render the full app. MSW mocks the API at the network boundary. Tests act like a user: they click keys, type, press keyboard keys and read the screen. This seam covers:
   - Each operation flow with the keypad.
   - A multi-operator expression and brackets.
   - Keyboard typing, the character filter and the `*`, `/`, `-` replacement.
   - `⌫`, `AC`, and the rules after a result (operator continues from the full value, a negative result goes in brackets, digit starts new, an edited result is sent as typed).
   - The keypad shows only the operations from the API.
   - The loading state and the disabled `=` key.
   - Each API error code shows the correct message, including the position for `INVALID_EXPRESSION`.
   - Network failure and timeout.
   - Operations load failure, and Retry.
   - Client-side validation (empty expression).
   - Display rounding.
   - History limit and clear.
   - Enter and Escape keys.

   The display formatter and the API client can also have small direct tests when a behavior is hard to reach through the UI. An example is the timeout.
4. **End-to-end (Playwright).** Playwright runs against the full Compose stack at `localhost:8080`. It uses no mocks. The tests cover these flows:
   - The page loads and shows a key for each of the 7 operations from the real API.
   - A happy path for each operation with the keypad.
   - A multi-operator expression with brackets, a phone percentage (`80 + 15% = 92`), a continuation from the full result (`1 ÷ 3`, then `× 3`, gives `1`), and a negative result in brackets before a power (`−3`, then `^ 2`, gives `9`).
   - Division by zero, the square root of a negative number and a syntax error show the correct messages.
   - A keyboard-only flow.
   - A mobile viewport (360 × 800) has no horizontal scroll, and the main flow works.
   - An accessibility scan with `@axe-core/playwright` finds no serious or critical violations.
   - nginx: a page reload on a deep link (for example `/unknown/path`) shows the app, a request to `/api/v1/operations` through nginx returns the operations list, and an unknown API path (for example `/api/v1/unknown`) returns the JSON `NOT_FOUND` error.

### Coverage targets

- Backend: 90% or more of the statements in the calculator and API modules.
- Frontend: 85% or more of the lines.
- The coverage command makes HTML and text reports for both parts. The README contains a summary table.

### Prior art

The repository is new, so there is no prior art. Follow the standard patterns:

- Go: table-driven tests and `net/http/httptest`.
- React: Testing Library queries by role and label, and MSW request handlers.
- Playwright: the page object pattern is not necessary for this small number of tests. Use role-based locators.

## Out of Scope

- Implicit multiplication (`2(3)`), variables, constants such as `π`, and an `Ans` key.
- Arbitrary-precision or decimal arithmetic. The app uses `float64`.
- Memory keys (M+, M−, MR), scientific functions other than those listed, and unit conversion.
- A `±` key. Unary minus with the `−` key does the same job.
- Persistent history: no database, no `localStorage`.
- User accounts, authentication and rate limiting.
- Internationalization. The UI is in English, and numbers use `.` as the decimal separator.
- Deployment to a cloud provider. The Compose stack is the deliverable.
- TLS in nginx. A real deployment puts TLS in front of the `web` service.

## Further Notes

- **Why 1 expression endpoint with a registry.** There are 3 alternatives: 1 endpoint per operation, 1 endpoint for 1 operation with its operands, or 2 endpoints (1 for 1 operation, 1 for an expression). The keypad sends a full expression, so an expression endpoint is necessary. A second endpoint for 1 operation adds code and tests, but the frontend does not use it, and an expression such as `2 + 3` does the same job. So there is 1 endpoint. It gives 1 validation path and 1 error format. The registry holds the symbol, the precedence and the function of each operation, and the parser reads them from the registry. A new operation is 1 registry entry and does not need a new route. The cost: an API client sends a string, not structured operands.
- **Why the backend does all calculations.** The assignment says that the frontend consumes the API. If both parts calculate, they can disagree. So the parser and the calculation are in the backend only. The frontend only filters characters and checks for an empty expression.
- **Percentage definition.** The assignment does not define "percentage". This spec uses the rule of a phone calculator: `x%` is `x ÷ 100`, and `a + b%` is `a + b% of a`. So "15% of 80" is `15% × 80 = 12`, and a 15% increase of 80 is `80 + 15% = 92`. The README must state this assumption.
- **Order of operations.** The calculator uses the standard mathematical order, as a modern phone calculator does: `2 + 3 × 4 = 14`, not 20. `−2 ^ 2 = −4`, because `^` binds more tightly than unary minus. The README must state this.
- **Floating-point precision.** The API returns the raw `float64` result, so an API client gets the true value. The UI rounds to 12 significant digits only for display. When the user continues from a result, the UI sends the full value, so the rounding does not change the next result. The README must explain this decision with the `0.1 + 0.2` example.
- **Status 422 compared to 400.** 400 means that the request is malformed, and this includes a syntax error in the expression. 422 means that the expression is valid, but there is no mathematical result. This lets an API client tell its own errors from domain errors.
- **Expression limits.** The limits of 200 characters and a bracket depth of 32 stop very long or very deep input. Deep input can use too much stack in a recursive parser.
- **Keypad with a text input.** A text input with `inputmode="none"` gives a visible label, screen reader support, normal keyboard typing and Enter, and it does not open the phone keyboard. The keypad buttons add the touch layout of a phone calculator.
- **2 services with Docker Compose.** Each service has 1 job. nginx is good at static files, caching, compression and security headers. The Go service does only the API. The 2 parts build, scale and update independently. nginx forwards `/api/` to the API, so the browser sees 1 origin and the stack needs no CORS. The cost is 1 more image and an nginx configuration file to maintain. A single Go binary that embeds the frontend is simpler, but it mixes 2 jobs in 1 service.
- **Time budget.** The assignment gives 2–4 hours. Priority order: the backend with tests (with the parser), the frontend with tests, Docker, the Playwright tests, CI, and then the documentation polish.
- **AI prompts.** `PROMPTS.md` records the main prompts in order, with 1 line about the result of each prompt. The first entry is the implementation prompt.
