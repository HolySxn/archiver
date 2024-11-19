# Archiver api

---

## Overview
This project is designed to develop a REST API that allows users to:

1. Retrieve information about an archive file.
2. Create an archive from a list of files.
3. Send a file to multiple email recipients.

## Endpoints

### 1. Retrieve Archive Information
#### Endpoint
`POST /api/archive/information`

#### Description
This route accepts a ZIP file as input and returns detailed information about the contents of the archive, including file paths, sizes, and MIME types.

#### Request
**Method:** `POST`  
**Content-Type:** `multipart/form-data`  
**Form Data:**
- `file`: ZIP file to be analyzed.

#### Response
**Status Code:** `200 OK`  
**Content-Type:** `application/json`

Example response:
```json
{
    "filename": "my_archive.zip",
    "archive_size": 4102029.312,
    "total_size": 6836715.52,
    "total_files": 2,
    "files": [
        {
            "file_path": "photo.jpg",
            "size": 2516582.4,
            "mimetype": "image/jpeg"
        },
        {
            "file_path": "directory/document.docx",
            "size": 4320133.12,
            "mimetype": "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
        }
    ]
}
```

#### Error Handling
- If the provided file is not a valid archive (ZIP), return an appropriate error message.

### 2. Create Archive
#### Endpoint
`POST /api/archive/files`

#### Description
This route accepts a list of files and combines them into a ZIP archive. Only files with specific allowed MIME types can be included.

#### Request
**Method:** `POST`  
**Content-Type:** `multipart/form-data`  
**Form Data:**
- `files[]`: A list of files to include in the archive.

Allowed MIME types:
- `application/vnd.openxmlformats-officedocument.wordprocessingml.document`
- `application/xml`
- `image/jpeg`
- `image/png`

#### Response
**Status Code:** `200 OK`  
**Content-Type:** `application/zip`

Binary data of the ZIP file.

Example response:
```zip
{Binary data of ZIP file}
```

#### Error Handling
- If any file in the `files[]` list is not one of the allowed MIME types, return an appropriate error message.

### 3. Send File to Multiple Recipients
#### Endpoint
`POST /api/mail/file`

#### Description
This route accepts a file and a list of email addresses, sending the file to all specified recipients.

#### Request
**Method:** `POST`  
**Content-Type:** `multipart/form-data`  
**Form Data:**
- `file`: The file to be sent.
- `emails`: A comma-separated list of email addresses to receive the file.

Allowed MIME types:
- `application/vnd.openxmlformats-officedocument.wordprocessingml.document`
- `application/pdf`

#### Response
**Status Code:** `200 OK`

Example response:
```text
200 OK
```

#### Error Handling
- If any file in the `file` section is not one of the allowed MIME types, return an appropriate error message.

## Environment Variables
- `FROM_EMAIL`: Sender’s email address.
- `FROM_EMAIL_PASSWORD`: Password for the sender’s email account.
- `HOST`: SMTP server host (e.g., `smtp.gmail.com`).

## Installation & Setup
1. Clone the repository:
    ```bash
    git clone <repository_url>
    cd doodocs-backend-challenge
    ```

2. Set environment variables:
    ```bash
    export FROM_EMAIL=<your_email>
    export FROM_EMAIL_PASSWORD=<your_password>
    export HOST=<smtp_server_host>
    ```

3. Install dependencies:
    ```bash
    go mod tidy
    ```

4. Run the application:
    ```bash
    go run main.go
    ```
