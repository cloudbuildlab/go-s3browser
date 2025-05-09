# Geonet Open Data S3 Browser

A Fastly Compute@Edge application written in Go that lets you browse and download files from the public Geonet Open Data S3 bucket via a web interface. All file requests are proxied through your Fastly service, so users see your domain, not the S3 bucket.

## Features

- Browse folders and files in the public S3 bucket
- Clean breadcrumb navigation
- File downloads and previews are proxied through your Fastly service
- No AWS credentials required (public bucket)

## Local Development

1. **Install dependencies:**
   - [Go](https://golang.org/dl/) (1.21+ recommended)
   - [Fastly CLI](https://developer.fastly.com/reference/cli/)

2. **Build and serve locally:**

   ```sh
   fastly compute build && fastly compute serve
   ```

   - Visit [http://127.0.0.1:7676](http://127.0.0.1:7676) in your browser.

## Deploying to Fastly

1. **Authenticate with Fastly CLI:**

   ```sh
   fastly profile create
   ```

   > Note: Use `fastly profile create` to authenticate with your Fastly API token. This is the correct command for recent CLI versions.

2. **Build the project:**

   ```sh
   fastly compute build
   ```

3. **(First time only) Initialize the service:**

   ```sh
   fastly compute init
   ```

   - Follow the prompts to set the service name (e.g., `go-s3browser` or any name you like).

4. **Deploy:**

   ```sh
   fastly compute deploy
   ```

5. **Add the S3 backend:**
   When prompted by the CLI, use the following values:

   - **Backend (hostname or IP address):**

     ```plaintext
     geonet-open-data.s3-ap-southeast-2.amazonaws.com
     ```

   - **Backend name:**

     ```plaintext
     TheOrigin
     ```

     (This must match the backend name in your code and `fastly.toml`)
   - **Port:**

     ```plaintext
     443
     ```

   - **Use SSL:**

     ```plaintext
     yes
     ```

   - **When prompted for another backend, just press Enter to continue.**

6. **Test your service:**
   - Use the Fastly-provided domain (shown after deploy) to verify the app works.

## Adding a Health Check (Optional)

To monitor your S3 backend's health, you can add a health check using the Fastly CLI:

```sh
fastly healthcheck create \
  --service-id <your-service-id> \
  --version=latest \
  --name s3-health \
  --host geonet-open-data.s3-ap-southeast-2.amazonaws.com \
  --path /explorer.html \
  --method GET \
  --expected-response 200 \
  --check-interval 3600000 \
  --timeout 2000 \
  --autoclone
```

- Replace `<your-service-id>` with your actual service ID (see the Fastly UI).
- This will create a health check on the latest version of your service, auto-cloning if needed.
- **Note:** After creating a health check (or any config change), you must manually activate the new version in the Fastly UI or with:

```sh
fastly service-version activate --service-id <your-service-id> --version <version-number>
```

Otherwise, your changes will remain as a draft and will not be live.

- Adjust the path and timing as needed for your use case.

## Production Notes

- For production, set up a custom domain and TLS in the Fastly UI.
- Optionally configure logging, CORS, and other Fastly features as needed.
- No AWS credentials are required for this public bucket.

## Project Structure

- `main.go` — Main entrypoint, HTTP handler, and template rendering
- `s3browser.go` — S3 listing and utility functions
- `fastly.toml` — Fastly Compute@Edge configuration

## License

MIT (or your preferred license)

## Contact

For questions or contributions, open an issue or contact the maintainer.
