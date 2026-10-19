# Agnos Hospital Middleware API

Middleware API for searching patient information across hospitals.
Hospital staff log in and can only see patients that belong to their own hospital.

**Tech stack:** Go, Gin, PostgreSQL, Docker Compose, Nginx

## Quick start

Requirements: Docker Desktop (with Docker Compose).

1. Create the environment file and set your own values:

   ```bash
   cp .env.example .env
   ```

   The `.env` file must contain these four variables:

   ```env
   POSTGRES_USER=agnos
   POSTGRES_PASSWORD=<your database password>
   POSTGRES_DB=agnos
   JWT_SECRET=<random string, at least 32 characters>
   ```

   | Variable | Description |
   |----------|-------------|
   | `POSTGRES_USER` | PostgreSQL user created on first start (example: `agnos`) |
   | `POSTGRES_PASSWORD` | Password for that user. Use letters and digits only, because it is placed inside a connection URL |
   | `POSTGRES_DB` | Database name (example: `agnos`) |
   | `JWT_SECRET` | Secret used to sign login tokens. Use a long random string of at least 32 characters and keep it private |
   | `HIS_HOSPITAL_A_URL` | Optional. Base URL of Hospital A's HIS (for example `https://hospital-a.api.co.th`). Leave it empty to use the built-in mock HIS |

   Replace the `<...>` values with your own. Never commit the real `.env` file (it is listed in `.gitignore`).

2. Build and start everything (nginx + api + postgres):

   ```bash
   docker compose up -d --build
   ```

3. Check that it works:

   ```bash
   curl http://localhost/health
   ```

   Expected: `{"message":"ok"}`

Only nginx is exposed to the host (port 80). The API and the database are reachable only inside the Docker network.

The database schema and sample data are loaded automatically from `migrations/` the first time the
database volume is created. To reset the database: `docker compose down -v` and start again.

## Sample data

| Hospital   | Patient (EN)     | National ID     |
|------------|------------------|-----------------|
| Hospital A | Somchai Jaidee   | 1100100000011   |
| Hospital A | Somying Rakdee   | 1100100000022   |
| Hospital B | Wichai Mankong   | 1100100000033   |

All data is fictional. Two hospitals exist: `Hospital A` and `Hospital B`.

The built-in mock HIS of Hospital A also knows one patient that is **not** in the database:
`Anan Sukjai`, national ID `1100100000044`, passport ID `AA1234567`. Looking it up with
`GET /patient/search/{id}` as a Hospital A staff member fetches it from the HIS and saves it locally.

## API

All request and response bodies are JSON. Errors have the form `{"error": "message"}`.

### POST /staff/create

Create a staff member. Passwords are stored as bcrypt hashes.

Request:

```json
{ "username": "nurse1", "password": "pass1234", "hospital": "Hospital A" }
```

| Status | Meaning |
|--------|---------|
| 201 | Created: `{"message":"staff created"}` |
| 400 | Invalid body, password shorter than 6 characters, or unknown hospital |
| 409 | Username already exists |

### POST /staff/login

Request:

```json
{ "username": "nurse1", "password": "pass1234", "hospital": "Hospital A" }
```

| Status | Meaning |
|--------|---------|
| 200 | `{"token":"<jwt>"}` (valid for 24 hours) |
| 400 | Invalid body |
| 401 | Wrong username, password or hospital (same message for all, so nothing about the account is revealed) |

### POST /patient/search (login required)

Send the token as `Authorization: Bearer <token>`. Every field is optional; an empty body returns
all patients of the staff member's hospital.

```json
{
  "national_id": "",
  "passport_id": "",
  "first_name": "",
  "middle_name": "",
  "last_name": "",
  "date_of_birth": "1990-05-12",
  "phone_number": "",
  "email": ""
}
```

- `national_id`, `passport_id`, `phone_number`: exact match.
- `first_name`, `middle_name`, `last_name`: partial, case-insensitive match against both the Thai and the English name.
- `email`: exact match, case-insensitive.
- `date_of_birth`: `YYYY-MM-DD`.
- Filters are combined with AND.

Response `200`: an array of patients (empty array when nothing matches). Each patient has
`first_name_th`, `middle_name_th`, `last_name_th`, `first_name_en`, `middle_name_en`, `last_name_en`,
`date_of_birth`, `patient_hn`, `national_id`, `passport_id`, `phone_number`, `email`, `gender` (`M` / `F`).

| Status | Meaning |
|--------|---------|
| 200 | Matching patients of the same hospital |
| 400 | Invalid body or date format |
| 401 | Missing, invalid or expired token |

### GET /patient/search/{id} (login required)

`id` is a national ID or a passport ID. Returns one patient of the staff member's hospital.

Lookup order:

1. The local database (only patients of the staff member's hospital).
2. If not found and the staff member's hospital has a HIS (currently Hospital A), the HIS is called with
   `GET {HIS_HOSPITAL_A_URL}/patient/search/{id}`. A patient found there is saved to the database and returned.

| Status | Meaning |
|--------|---------|
| 200 | The patient |
| 401 | Missing, invalid or expired token |
| 404 | Not found in the database or the HIS (also returned for patients of other hospitals) |
| 502 | The hospital's HIS could not be reached or returned an error |

### Example

```bash
curl -X POST http://localhost/staff/create -H "Content-Type: application/json" \
  -d '{"username":"nurse1","password":"pass1234","hospital":"Hospital A"}'

TOKEN=$(curl -s -X POST http://localhost/staff/login -H "Content-Type: application/json" \
  -d '{"username":"nurse1","password":"pass1234","hospital":"Hospital A"}' | sed 's/.*"token":"\([^"]*\)".*/\1/')

curl -X POST http://localhost/patient/search -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" -d '{"national_id":"1100100000011"}'

curl http://localhost/patient/search/1100100000011 -H "Authorization: Bearer $TOKEN"

# Not in the database: fetched from Hospital A's (mock) HIS
curl http://localhost/patient/search/1100100000044 -H "Authorization: Bearer $TOKEN"
```

## Tests

```bash
go test ./... -cover
```

Unit tests cover the handlers, the JWT middleware, the services and the HIS client, including negative cases
(bad input, duplicate username, wrong password or hospital, expired or forged token, other hospital's patient,
HIS not found or unavailable).
Dependencies are replaced by fakes, so no database is needed.

## Project structure

```
.
├── cmd/server/main.go        # entry point: config, DB pool, routes
├── internal/
│   ├── config/               # reads environment variables
│   ├── model/                # Patient, PatientSearch, Staff, shared errors
│   ├── repository/           # SQL, talks to PostgreSQL
│   ├── service/              # business logic: bcrypt, JWT
│   ├── handler/              # Gin HTTP handlers
│   ├── his/                  # HIS client: HTTP client + built-in mock
│   └── middleware/           # JWT validation, puts hospital_id in the request context
├── migrations/001_init.sql   # schema + sample data (loaded by postgres on first start)
├── nginx/nginx.conf          # reverse proxy to the api service
├── Dockerfile                # multi-stage build of the Go service
└── docker-compose.yml        # nginx + api + postgres
```

Request flow: `client -> nginx :80 -> api :8080 -> postgres :5432`.

## Design notes and assumptions

- **Hospital A HIS:** the assignment names an external API (`GET https://hospital-a.api.co.th/patient/search/{id}`).
  It is not publicly reachable, so the service talks to it through a `HISClient` interface with two
  implementations: a real HTTP client (used when `HIS_HOSPITAL_A_URL` is set) and a built-in mock (default).
  Results from the HIS are saved to the `patients` table, and `GET /patient/search/{id}` follows the same contract.
  Only Hospital A staff trigger a Hospital A HIS lookup.
- **Hospital isolation:** the hospital is read from the signed JWT, never from the request body, and is
  always part of the SQL `WHERE` clause.
- `GET /patient/search/{id}` also requires login, because it returns patient data.
- `/patient/search` is a `POST` because it takes many optional fields.

## Known limitations

- `/staff/create` is open, as written in the assignment. In production it should be restricted to administrators.
- No rate limiting on login and no token refresh or revocation.
- The repository (SQL) layer has no automated tests yet; it would need integration tests against a real PostgreSQL.
- `POST /patient/search` searches the local database only; the HIS is used for the lookup by ID.
