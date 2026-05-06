# Embedded PDF Files

A web application built with Go that allows you to upload PDF files and extract embedded files contained within them.

## Features

- Upload PDF files
- Extract embedded files
- Download files individually or all files at once as a ZIP archive
- Automatic cleanup

## Development

### Prerequisites

- Docker
- Docker Compose

### Running Locally

1. Clone the repository:
   ```bash
   git clone https://github.com/eduardopaulcs/embedded-pdf-files.git
   cd embedded-pdf-files
   ```

2. (Optional) Create a `.env` file to customize the port:
   ```
   PORT=8080
   ```

3. Start the application with Docker Compose:
   ```bash
   docker-compose up
   ```

   Or to run in background:
   ```bash
   docker-compose up -d
   ```

### Accessing the Application

Once the container is running, open your browser and navigate to:

```
http://localhost:8080
```

1. Upload a PDF file using the web interface
2. View the list of embedded files found
3. Download files individually or click "Download All (ZIP)" to get all files at once
