# Frontend

React dashboard for the auth and task-tracker services.

## Gateway

The frontend talks only to the gateway. Auth validates login credentials, then
Kong validates access tokens for protected routes and forwards trusted identity
headers to services.

Choose one gateway in `.env`:

```env
VITE_API_URL=http://localhost:8000
```

Zipkin UI:

```env
VITE_ZIPKIN_URL=http://localhost:9411
```

## Run

```sh
npm run dev
```

Restart Vite after changing `.env`.

## Build

```sh
npm run build
```
