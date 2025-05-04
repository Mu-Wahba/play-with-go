# Go API with OpenTelemetry & Jaeger

This is a RESTful API built with Go, Gin, Postgresql and OpenTelemetry tracing (Jaeger). It supports user authentication, event management, and registration, with example API requests provided for httpie.

## Getting Started

### Prerequisites

- Docker & Docker Compose
- httpie (for API testing)

### Running the Application

Start all services (API, Jaeger, and dependencies) with:

```sh
docker compose up -d
```

The API will be available at `http://localhost:3000` and Jaeger UI at [http://localhost:16686](http://localhost:16686).

## API Endpoints

### User

- `POST /signup` — Register a new user
- `POST /login` — Login and receive a JWT token
- `GET /users` — List all users

### Events

- `POST /events` — Create a new event (auth required)
- `GET /events` — List all events
- `GET /event/:id` — Get event by ID
- `PUT /event/:id` — Update event (auth required)
- `DELETE /event/:id` — Delete event (auth required)

### Registration

- `POST /events/:id/register` — Register for an event (auth required)
- `DELETE /events/:id/register` — Unregister from an event (auth required)
- `GET /registers` — List all registrations

### Misc

- `GET /` — Home
- `DELETE /clear` — Clear all data (basic auth required)

## Example API Usage with httpie

All examples are in the `api-test/` directory. Here are some quick examples:

### Signup

```sh
http POST :3000/signup email="w3@w.w" password="123"
```

### Login

```sh
http POST :3000/login email="w3@w.w" password="123"
# Response will include a JWT token
```

### Create Event

```sh
http POST :3000/events \
  Authorization:'<JWT_TOKEN>' \
  name="Event Name" description="Event Description" location="Event Location"
```

### Get All Events

```sh
http GET :3000/events
```

### Get Event by ID

```sh
http GET :3000/event/1
```

### Update Event

```sh
http PUT :3000/event/1 \
  Authorization:'<JWT_TOKEN>' \
  name="Updated Name" description="Updated Description" location="Updated Location"
```

### Delete Event

```sh
http DELETE :3000/event/1 Authorization:'<JWT_TOKEN>'
```

## Testing

You can find ready-to-use `.http` files for httpie in the `api-test/` directory.

---

**Note:** Replace `<JWT_TOKEN>` with the token received from the login endpoint.

---

## Tracing

Access Jaeger UI at [http://localhost:16686](http://localhost:16686) to view traces.

