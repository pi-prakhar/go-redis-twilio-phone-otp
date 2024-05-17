## Golang Phone OTP Service with Twilio and Redis

This project provides a Golang service for sending one-time passwords (OTPs) to phone numbers for verification purposes. It utilizes Twilio as the SMS messenger service and Redis as a cache database to store OTPs.

### Prerequisites

* Golang (version 1.17 or higher recommended)
* Docker (optional, for containerized deployment)
* Redis server (local or remote)

### Installation

1. **Clone the repository:**

```bash
git clone https://github.com/pi-prakhar/go-redis-twilio-phone-otp.git
```

2. **Create a `.env` file:**

Copy the `.env.sample` file to `.env` and fill in the following environment variables:

* `TWILIO_ACCOUNT_SID`: Your Twilio Account SID
* `TWILIO_AUTHTOKEN`: Your Twilio Auth Token
* `TWILIO_SERVICES_ID`: Your Twilio Verify Service ID
* `TWILIO_PHONE_NUMBER`: Your Twilio phone number for sending OTPs
* `REDIS_DB_PASSWORD`: Password for your Redis database (if applicable)

**3. (Optional) Docker Setup:**

Build and start the service using Docker Compose:

```bash
docker-compose up -d --build
```

**4. Local Development:**

* Install Redis locally (if not using a remote server).
* Update the `redis-db-address` in `config.json` if needed (defaults to `redis-db:6379`).
* Run the service:

```bash
go build -o main ./cmd/go-redis-twilio-phone-otp
./main
```

### Configuration

The project uses a `config.json` file for basic configuration settings:

* `test-port`: Default port for the server (3000)
* `test-hostname`: Default hostname for the server (localhost)
* `redis-db-address`: Redis server domain (defaults to redis-db:6379)
* `log-level`: Log level (info, error, warn, debug)
* `otp-timeout`: OTP expiration time in seconds (defaults to 30)
* `otp-max-attempts`: Maximum number of OTP verification attempts (defaults to 5)
* `otp-lock-timeout`: Duration to lock user after exceeding attempts (defaults to 30 minutes)

You can modify these values in the `config.json` file.

### Usage

The service provides two main API endpoints for OTP management:

#### 1. `/api/send-otp` (POST)

This endpoint initiates the OTP sending process.

**Request Body:**

```json
{
  "phoneNumber": "string" // User's phone number in E.164 format (e.g., +14155552671)
}
```
**Response Body (Success):**

* **Status Code: 200 (OK):**
  * Message: "OTP send successfully."
  * * Data: "number of trials left"
* **Status Code: 403 (Forbidden):**
  * Message: "User locked out due to exceeding maximum attempts."
  * Data includes `lockout_duration` in minutes until user can send OTP again

**Response Body (Error):**

* **Status Code: 400 (Bad Request):**
  * Message: "Invalid request body" (e.g., missing or invalid phone number)
* **Status Code: 403 (Forbidden):**
  * Message: "User locked out due to exceeding maximum attempts." (data includes `lockout_duration` in minutes until user can send OTP again)

#### 2. `/api/verify-otp` (POST)

This endpoint verifies the received OTP code.

**Request Body:**

```json
{
  "user": {
    "phoneNumber": "string" // User's phone number in E.164 format
  },
  "code": "string" // The received OTP code
}
```

**Response Body (Success):**

* **Status Code: 200 (OK):**
  * Message: "OTP verified successfully."
* **Status Code: 401 (Unauthorized):**
  * Message: "Incorrect OTP/ OTP Expired"
  * Data: "number of trials left"
* **Status Code: 403 (Forbidden):**
  * Message: "User locked out due to exceeding maximum attempts."
  * Data includes `lockout_duration` in minutes until user can send OTP again

**Response Body (Error):**

* **Status Code: 400 (Bad Request):**
  * Message: "Invalid request body" (e.g., missing phone number or code)
* **Status Code: 500 (Internal Server Error):**
  * Message: "Internal server error occurred."

**Response Body (Success):**

```json
{
  "code": STATUS_CODE,
  "message": "",
  "data": {}
}
```
**Response Body (Error):**

```json
{
  "code": STATUS_CODE,
  "message": "",
}
```
