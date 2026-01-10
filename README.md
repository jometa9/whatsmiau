# WhatsMiau

![logo-whatsmiau](logo.png)

WhatsMiau is a backend service for WhatsApp, built with Go. It uses the Whatsmeow library to connect to WhatsApp and provides an HTTP API to send and receive messages.

[Community Whatsapp (BR)](https://chat.whatsapp.com/FXMrTY552nOBFXU71Be8Zh)
## About The Project

This project provides a robust, scalable, and production-ready solution for integrating WhatsApp functionalities into your applications. It is extremely lightweight, consuming very little memory, making it ideal for resource-constrained environments.

It's designed to be compatible with the Evolution API, making it a flexible choice for developers familiar with that ecosystem.

## Features

- **Lightweight & Efficient:** Optimized for low memory consumption.
- **Production Ready:** Stable and reliable for use in production environments.
- **WhatsApp Integration:** Connects to WhatsApp to send and receive messages.
- **HTTP API:** Exposes an HTTP API for easy integration with other services.
- **Redis Support:** Uses Redis for session storage and caching.
- **SQLite Database:** Utilizes SQLite for persistent data storage.
- **Environment-based Configuration:** Easily configure the application using environment variables.
- **Structured Logging:** Implements structured logging with Zap for better monitoring and debugging.

## Getting Started

To get a local copy up and running follow these simple steps.

### Prerequisites

- Go 1.24 or higher
- Redis
- SQLite

### Installation

1. Clone the repo
   ```sh
   git clone https://github.com/verbeux-ai/whatsmiau.git
   ```
2. Install Go packages
   ```sh
   go mod tidy
   ```
3. Set up your environment variables by copying `.env.example` to `.env` and filling in the required values.
   ```sh
   cp .env.example .env
   ```
4. Run the application
   ```sh
   go run main.go
   ```

## Running with Docker

You can also run the application using Docker and Docker Compose.

1.  **Build and run the containers:**
    ```sh
    docker-compose up -d --build
    ```
2.  **View the logs:**
    ```sh
    docker-compose logs -f
    ```
3.  **Stop the containers:**
    ```sh
    docker-compose down
    ```

## Docker Image

Official Docker images are available on Docker Hub.

- **Latest stable release:** `impedr029/whatsmiau:vX.Y.Z` [(see versions)](https://github.com/verbeux-ai/whatsmiau/tags)
- **Development version:** `impedr029/whatsmiau:develop`

You can pull the latest stable image with (example):
```sh
docker pull impedr029/whatsmiau:vX.Y.Z
```

Or the development image with:
```sh
docker pull impedr029/whatsmiau:develop
```

## Configuration

The application is configured using environment variables. The following variables are available:

| Variable | Description | Default |
| --- | --- | --- |
| `PORT` | The port the server will run on. | `8080` |
| `DEBUG_MODE` | Enable or disable debug mode. | `false` |
| `DEBUG_WHATSMEOW` | Enable or disable debug mode for Whatsmeow. | `false` |
| `REDIS_URL` | The URL of the Redis server. | `localhost:6379` |
| `REDIS_PASSWORD` | The password for the Redis server. | `` |
| `REDIS_TLS` | Enable or disable TLS for Redis. | `false` |
| `API_KEY` | The API key to protect the service. | `` |
| `DIALECT_DB` | The database dialect to use (`sqlite3` or `postgres`). | `sqlite3` |
| `DB_URL` | The database connection URL. | `file:data.db?_foreign_keys=on` |
| `GCS_ENABLED` | Enable or disable Google Cloud Storage. | `false` |
| `GCS_BUCKET` | The GCS bucket name. | `whatsmiau` |
| `GCS_URL` | The GCS URL. | `https://storage.googleapis.com` |
| `GCL_APP_NAME` | The GCL application name. | `whatsmiau-br-1` |
| `GCL_ENABLED` | Enable or disable Google Cloud Logging. | `false` |
| `GCL_PROJECT_ID` | The GCL project ID. | `` |
| `EMITTER_BUFFER_SIZE` | The emitter buffer size. | `2048` |
| `HANDLER_SEMAPH-ORE_SIZE` | The handler semaphore size. | `512` |
| `PROXY_ADDRESSES` | A comma-separated list of proxy addresses. Example: `SOCKS5://user:pass@host:port,HTTP://host:port` | `` |
| `PROXY_STRATEGY` | The strategy to use when selecting a proxy from the list (`RANDOM`). | `RANDOM` |
| `PROXY_NO_MEDIA` | If set to `true`, media will not be sent through the proxy. | `false` |

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/verbeux-ai/whatsmiau/tags).

## Compatibility

This API is designed to be compatible with the Evolution API. This means that you can use clients and tools designed for the Evolution API with this project.

It exclusively supports webhooks in the Evolution API format, offering two distinct approaches for their implementation, providing flexibility for different use cases.

## Migration from Evolution API

WhatsMiau is designed to be a lightweight, drop-in replacement for the Evolution API. If you are running WhatsMiau on the same host and port as your previous Evolution API instance, migration is seamless.

Since WhatsMiau maintains compatibility with the Evolution API's routes, you only need to stop your Evolution API server and start the WhatsMiau server. No changes to your existing API calls are necessary.

### Example

For instance, if you were sending a text message using a `curl` command to an Evolution API server running on `localhost:8080`, the exact same command will work with WhatsMiau.

**Before (Evolution API):**
```bash
curl -X POST 'http://localhost:8080/message/sendText/my-instance' \
-H 'Content-Type: application/json' \
-H 'apikey: YOUR_API_KEY' \
-d ".{\"number\": \"1234567890\",\"textMessage\": {\"text\": \"Hello from Evolution API!\"}}"
```

**After (WhatsMiau):**

Simply point your application to the WhatsMiau server URL. The same request will be handled by WhatsMiau:
```bash
curl -X POST 'http://localhost:8080/v1/message/sendText/my-instance' \
-H 'Content-Type: application/json' \
-H 'apikey: YOUR_API_KEY' \
-d ".{\"number\": \"1234567890\",\"textMessage\": {\"text\": \"Hello from WhatsMiau!\"}}"
```

## API Routes

Same Pattern: https://www.postman.com/agenciadgcode/evolution-api/overview
| Method | Path                                      | Description                 |
|--------|-------------------------------------------|-----------------------------|
| POST   | /v1/instance                            | Create a new instance       |
| GET    | /v1/instance                            | List all instances          |
| POST   | /v1/instance/:id/connect                | Connect to an instance      |
| POST   | /v1/instance/:id/logout                 | Logout from an instance     |
| DELETE | /v1/instance/:id                        | Delete an instance          |
| GET    | /v1/instance/:id/status                 | Get instance status         |
| POST   | /v1/instance/:instance/message/text     | Send a text message         |
| POST   | /v1/instance/:instance/message/audio    | Send an audio message       |
| POST   | /v1/instance/:instance/message/document | Send a document             |
| POST   | /v1/instance/:instance/message/image    | Send an image message       |
| POST   | /v1/instance/:instance/chat/presence    | Send chat presence          |
| POST   | /v1/instance/:instance/chat/read-messages| Mark messages as read       |
| POST   | /v1/instance/:instance/chat/whatsapp-numbers| Check if a number is on WhatsApp |

### Evolution API Compatibility Routes

| Method | Path                               | Description                 |
|--- |--- |--- |
| POST   | /v1/instance/create                | Create a new instance       |
| GET    | /v1/instance/fetchInstances        | List all instances          |
| GET    | /v1/instance/connect/:id           | Connect to an instance      |
| GET    | /v1/instance/connectionState/:id   | Get instance status         |
| DELETE | /v1/instance/logout/:id            | Logout from an instance     |
| DELETE | /v1/instance/delete/:id            | Delete an instance          |
| PUT    | /v1/instance/update/:id            | Update an instance          |
| POST   | /v1/message/sendText/:instance     | Send a text message         |
| POST   | /v1/message/sendWhatsAppAudio/:instance | Send an audio message       |
| POST   | /v1/message/sendMedia/:instance    | Send a media message        |
| POST   | /v1/message/sendReaction/:instance | Send a reaction to a message |
| POST   | /v1/chat/markMessageAsRead/:instance | Mark messages as read       |
| POST   | /v1/chat/sendPresence/:instance    | Send chat presence          |
| POST   | /v1/chat/whatsappNumbers/:instance | Check if a number is on WhatsApp |

## Supported Events

The application can send webhook events for the following actions:

| Event             | Description                                         |
|-------------------|-----------------------------------------------------|
| `MESSAGES_UPSERT` | Triggered when a new message is received.           |
| `MESSAGES_UPDATE` | Triggered when a message status changes (e.g., read). |
| `CONTACTS_UPSERT` | Triggered when a contact is created or updated.     |

## Webhook Configuration

You can configure a webhook URL for each instance to receive events. When a webhook URL is configured, all events will be sent to that URL as HTTP POST requests.

### Configure Webhook for an Instance

To set or update the webhook URL for an instance, use the update endpoint:

```bash
curl -X PUT 'http://localhost:8080/v1/instance/update/my-instance' \
  -H 'Content-Type: application/json' \
  -H 'apikey: YOUR_API_KEY' \
  -d '{
    "webhook": {
      "url": "https://your-webhook-url.com/webhook",
      "base64": false,
      "events": ["MESSAGES_UPSERT", "MESSAGES_UPDATE", "CONTACTS_UPSERT"]
    }
  }'
```

### Parameters

- `url` (required): The webhook URL where events will be sent
- `base64` (optional): Whether to send media in base64 format (default: false)
- `events` (optional): Array of events to receive. If not specified or empty, all events will be sent:
  - `MESSAGES_UPSERT`: New messages
  - `MESSAGES_UPDATE`: Message status updates
  - `CONTACTS_UPSERT`: Contact updates

### Example: Configure webhook to receive all events

If you don't specify the `events` array or leave it empty, all events will be sent:

```bash
curl -X PUT 'http://localhost:8080/v1/instance/update/my-instance' \
  -H 'Content-Type: application/json' \
  -H 'apikey: YOUR_API_KEY' \
  -d '{
    "webhook": {
      "url": "https://your-webhook-url.com/webhook"
    }
  }'
```

### Example: Configure webhook to receive only specific events

```bash
curl -X PUT 'http://localhost:8080/v1/instance/update/my-instance' \
  -H 'Content-Type: application/json' \
  -H 'apikey: YOUR_API_KEY' \
  -d '{
    "webhook": {
      "url": "https://your-webhook-url.com/webhook",
      "events": ["MESSAGES_UPSERT"]
    }
  }'
```

## Heartbeat Webhook Configuration

The heartbeat webhook sends system metrics (CPU, memory, etc.) every minute. You can configure it in three ways:

### 1. Inject at Build Time (Recommended)

Inject the webhook URL when building the binary:

```bash
WEBHOOK_URL=https://your-webhook.com/metrics ./build-linux.sh docker
```

The URL will be embedded in the binary and used by default.

### 2. Set via Environment Variable

Set `WEBHOOK_URL` in your `.env` file or as a system environment variable:

```bash
WEBHOOK_URL=https://your-webhook.com/metrics
```

### 3. Update Dynamically via API

Update the heartbeat webhook URL at runtime using the API:

```bash
# Update heartbeat webhook URL
curl -X PUT 'http://localhost:8080/v1/webhook/heartbeat' \
  -H 'Content-Type: application/json' \
  -H 'apikey: YOUR_API_KEY' \
  -d '{
    "url": "https://your-new-webhook.com/metrics"
  }'

# Get current heartbeat webhook URL
curl -X GET 'http://localhost:8080/v1/webhook/heartbeat' \
  -H 'apikey: YOUR_API_KEY'
```

**Priority order:**
1. Dynamic URL (set via API) - highest priority
2. Environment variable (`WEBHOOK_URL`)
3. Build-time injected URL - lowest priority

## Did you like project?
Donate: https://buy.stripe.com/8x28wI5vKfPbe9b8ih1VK0f