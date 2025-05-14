# Geonet Open Data S3 Browser

A Fastly Compute@Edge application written in Go that lets you browse and download files from the public Geonet Open Data S3 bucket via a web interface. All file requests are proxied through your Fastly service, so users see your domain, not the S3 bucket.

## Features

- Browse folders and files in the public S3 bucket
- Clean breadcrumb navigation
- File downloads and previews are proxied through your Fastly service
- No AWS credentials required (public bucket)
- Staging and production environments with GitHub Actions deployment

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

### Initial Setup

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

4. **Deploy to Staging:**

   ```sh
   fastly compute deploy --service-id <your-service-id>
   ```

   - This will automatically create and activate a new version in staging
   - Access your app at the staging domain provided by Fastly

5. **Deploy to Production:**

   ```sh
   fastly compute deploy --service-id <your-service-id>
   fastly service-version activate --service-id <your-service-id> --version <version-number>
   ```

   - This will deploy to your production domain
   - You must manually activate the new version for production

6. **Add the S3 backend:**
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

7. **Test your service:**
   - Use the Fastly-provided domain (shown after deploy) to verify the app works.

### Setting Up Staging

Fastly provides a built-in staging environment for Compute services. You do NOT need to manually add a staging domain.

1. **Enable Staging:**
   - In the Fastly UI, go to your Compute service and click "Opt in to Staging" (banner at the top of the service page)
   - After enabling, Fastly will provide a special staging domain (e.g., `your-service.staging.edgecompute.app`)

2. **Configure GitHub Secrets:**
