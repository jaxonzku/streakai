# streakai Real-time Collaborative Voting System

### git clone [https://github.com/jaxonzku/streakai.git]

### `cd streakai`

### `docker-compose up --build`

The application will be accessible at http://localhost:8080.

API Endpoints

- POST /login

Logs a user into the system.

### `Request Body: { "username": "<username>", "password": "<password>" }`

---

- POST /register

Registers a new user.

### `Request Body: { "username": "<username>", "password": "<password>" }`

---

- POST /logout

Logs out a user from the system.

### `Request Body: { "username": "<username>" }`

---

- GET /sessions/{id}

Retrieves all voting sessions.
Authentication Required: JWT token in Authorization header

---

- POST /sessions

Creates a new voting session.

### `Request Body: { "name": "<session-name>" }`

Authentication Required: JWT token in Authorization header

---

- PATCH /sessions

Casts a vote for a session.

### `Request Body: { "id": "<session-id>", "vote": true/false }`

Authentication Required: JWT token in Authorization header
