# MailRaven Thunder Client Collection

To use these requests in VS Code with Thunder Client:
1. Copy the JSON content below
2. Open Thunder Client -> Collections
3. Click the "Burger Menu" (three lines) -> Import -> From Text/JSON
4. Paste the content

Alternatively, you can manually create requests using the details below.

## Environment Variables
Create a Thunder Client Environment with:
- `baseUrl`: `http://10.0.2.2:8080` (for Android Emulator) or `http://localhost:8080` (for Localhost)
- `token`: `{{token_from_login_response}}` (You will set this after login)

## Requests

### 1. Health Check
*   **Method**: `GET`
*   **URL**: `{{baseUrl}}/health`
*   **Description**: Verify server is running

### 2. Login
*   **Method**: `POST`
*   **URL**: `{{baseUrl}}/api/v1/auth/login`
*   **Body (JSON)**:
    ```json
    {
      "email": "admin@example.com",
      "password": "your_secure_password"
    }
    ```
*   **Tests/Post-Request**:
    *   Action: Set Environment Variable
    *   Variable: `token`
    *   Value: `json.token`

### 3. Get Messages
*   **Method**: `GET`
*   **URL**: `{{baseUrl}}/api/v1/messages`
*   **Headers**:
    *   `Authorization`: `Bearer {{token}}`

### 4. Send Message
*   **Method**: `POST`
*   **URL**: `{{baseUrl}}/api/v1/messages/send`
*   **Headers**:
    *   `Authorization`: `Bearer {{token}}`
*   **Body (JSON)**:
    ```json
    {
      "from": "admin@example.com",
      "to": ["recipient@example.com"],
      "subject": "Test Email from Thunder Client",
      "body": "This is a test message sent via the API."
    }
    ```

### 5. Get System Stats (Admin Only)
*   **Method**: `GET`
*   **URL**: `{{baseUrl}}/api/v1/admin/stats`
*   **Headers**:
    *   `Authorization`: `Bearer {{token}}`

---

## Thunder Client Collection JSON

```json
{
  "client": "Thunder Client",
  "collectionName": "MailRaven API",
  "dateExported": "2024-02-06T00:00:00.000Z",
  "version": "1.1",
  "folders": [],
  "requests": [
    {
      "containerId": "",
      "sortNum": 10,
      "headers": [
        {
          "name": "Accept",
          "value": "*/*"
        },
        {
          "name": "User-Agent",
          "value": "Thunder Client (https://www.thunderclient.com)"
        }
      ],
      "colId": "mailraven_col",
      "name": "Health Check",
      "url": "{{baseUrl}}/health",
      "method": "GET",
      "modified": "2024-02-06T00:00:00.000Z",
      "created": "2024-02-06T00:00:00.000Z",
      "_id": "req_health",
      "params": [],
      "tests": []
    },
    {
      "containerId": "",
      "sortNum": 20,
      "headers": [
        {
          "name": "Accept",
          "value": "*/*"
        },
        {
          "name": "User-Agent",
          "value": "Thunder Client (https://www.thunderclient.com)"
        },
        {
          "name": "Content-Type",
          "value": "application/json"
        }
      ],
      "colId": "mailraven_col",
      "name": "Login",
      "url": "{{baseUrl}}/api/v1/auth/login",
      "method": "POST",
      "modified": "2024-02-06T00:00:00.000Z",
      "created": "2024-02-06T00:00:00.000Z",
      "_id": "req_login",
      "params": [],
      "body": {
        "type": "json",
        "raw": "{\n  \"email\": \"admin@example.com\",\n  \"password\": \"your_secure_password\"\n}",
        "form": []
      },
      "tests": [
        {
          "type": "set-env-var",
          "custom": "json.token",
          "action": "setto",
          "value": "{{token}}"
        }
      ]
    },
    {
      "containerId": "",
      "sortNum": 30,
      "headers": [
        {
          "name": "Accept",
          "value": "*/*"
        },
        {
          "name": "User-Agent",
          "value": "Thunder Client (https://www.thunderclient.com)"
        },
        {
          "name": "Authorization",
          "value": "Bearer {{token}}"
        }
      ],
      "colId": "mailraven_col",
      "name": "List Messages",
      "url": "{{baseUrl}}/api/v1/messages",
      "method": "GET",
      "modified": "2024-02-06T00:00:00.000Z",
      "created": "2024-02-06T00:00:00.000Z",
      "_id": "req_messages",
      "params": [],
      "tests": []
    },
    {
      "containerId": "",
      "sortNum": 40,
      "headers": [
        {
          "name": "Accept",
          "value": "*/*"
        },
        {
          "name": "User-Agent",
          "value": "Thunder Client (https://www.thunderclient.com)"
        },
        {
          "name": "Content-Type",
          "value": "application/json"
        },
        {
          "name": "Authorization",
          "value": "Bearer {{token}}"
        }
      ],
      "colId": "mailraven_col",
      "name": "Send Message",
      "url": "{{baseUrl}}/api/v1/messages/send",
      "method": "POST",
      "modified": "2024-02-06T00:00:00.000Z",
      "created": "2024-02-06T00:00:00.000Z",
      "_id": "req_send",
      "params": [],
      "body": {
        "type": "json",
        "raw": "{\n  \"from\": \"admin@example.com\",\n  \"to\": [\"recipient@example.com\"],\n  \"subject\": \"Test Email\",\n  \"body\": \"This is a test message.\"\n}",
        "form": []
      },
      "tests": []
    }
  ],
  "settings": {
    "headers": [],
    "tests": [],
    "options": {
      "baseUrl": "http://localhost:8080"
    }
  }
}
```
