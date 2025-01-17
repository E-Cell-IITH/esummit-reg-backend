# API Documentation for E-Summit 2025 Registration

## Routes
### **1. User Registration (Sign-up)**
#### **1.1. Send OTP for Sign-up**
- **Endpoint**: `/signup/otp/send`
- **Method**: `POST`
- **Description**: Sends an OTP to the provided email address for user registration.
- **Request Body**:
  ```json
  {
    "email": "user@example.com"
  }
  ```
- **Response**:
  - Sucess:
    ```json
    {
        "message": "OTP sent successfully"
    }
    ```
  - Error (User already exists):
    ```json
    {
        "error": "User already exists"
    }
    ```
  - Error (Invalid JSON)::
    ```json
    {
        "error": "Invalid JSON"
    }
    ```
    
#### **1.2. Verify OTP for Sign-up**
- **Endpoint**: `/signup/otp/verify`
- **Method**: `POST`
- **Description**: Verifies the OTP sent to the provided email address.
- **Request Body**:
  ```json
    {
        "email": "user@example.com",
        "otp": "123456"
    }
  ```
- **Response**:
  - Sucess:
    ```json
    {
        "message": "OTP verified successfully"
    }
    ```
  - Error (Invalid OTP):
    ```json
    {
        "error": "Invalid OTP"
    }
    ```
  - Error (Invalid JSON):
    ```json
    {
        "error": "Invalid JSON"
    }
    ```

#### **1.3. Register User**
- **Endpoint**: `/signup`
- **Method**: `POST`
- **Description**: Verifies the OTP sent to the provided email address.
- **Request Body**:
  ```json
    {
        "email": "user@example.com",
        "name": "John Doe",
        "contact_number": "9867666666",
        "data": "{\"college\":\"IITH\",\"roll\":\"MS22BTECH11010\"",
        "otp": "123456"
    }
  ```
  `
  Here data is basically a json string which can contain custom data, so send data whatever you want to save for a user and use this accordingly. 
  ** Create your own format on the frontend side what all data you want for users and send them to server, by doing this you can prevent null value issue while fetching data.
  `

- **Response**:
  - Sucess:
    ```json
    {
        "message": "User registered successfully",
        "id": 1
    }
    ```
     *And In headers client will receive cookie, named `session`, containing users `metadata`.*

  - Error (User already exists):
    ```json
    {
        "error": "User already exists"
    }
    ```
  - Error (OTP not verified):
    ```json
    {
        "error": "OTP not verified"
    }
    ```
  - Error (Invalid JSON):
    ```json
    {
        "error": "Invalid JSON"
    }
    ```
  - Error (Internal Server Error):
    ```json
    {
        "error": "Internal Server Error"
    }
    ``` 

### **2.User Sign-in (Login)**
#### **2.1. Send OTP for Sign-in**
- **Endpoint**: `/signin/otp/send`
- **Method**: `POST`
- **Description**: Sends an OTP to the provided email address for user sign-in.
- **Request Body**:
  ```json
  {
    "email": "user@example.com"
  }
  ```
- **Response**:
  - Sucess:
    ```json
    {
        "message": "OTP sent successfully"
    }
    ```
  - Error (User already exists):
    ```json
    {
       "error": "User does not exists"
    }
    ```
  - Error (Invalid JSON)::
    ```json
    {
        "error": "Invalid JSON"
    }
    ```
#### **2.1. Verify OTP for Sign-in**
- **Endpoint**: `/signin/otp/verify`
- **Method**: `POST`
- **Description**: Sends an OTP to the provided email address for user sign-in.
- **Request Body**:
  ```json
    {
        "email": "user@example.com",
        "otp": "123456"
    }
  ```
- **Response**:
  - Sucess:
    ```json
    {
        "message": "OTP verified successfully"
    }
    ```
    *And In headers client will receive cookie, named `session`, containing users `metadata`.*

  - Error (Invalid OTP):
    ```json
    {
       "error": "Invalid OTP"
    }
    ```
  - Error (Invalid JSON):
    ```json
    {
        "error": "Invalid JSON"
    }
    ```
  - Error (Internal Server Error):
    ```json
    {
        "error": "Internal Server Error"
    }
    ```

### Responses
For suceess the `status_code` is`200`. *In case of errors, the API returns standard error responses:*

- **Bad Request (400)**:
```
    Invalid JSON input or missing required fields.
```
- **Internal Server Error (500)**:
```
    Server-side errors, such as failure to save data or generate OTP.
```

## Flow Summary
1. **Registration:**
    - User sends a request to `/signup/otp/send` to receive an OTP.
    - User verifies the OTP with `/signup/otp/verify`.
    - User registers with `/signup` after OTP verification and this will start a new session as well.

2. **Sign-in:**
    - User sends a request to `/signin/otp/send` to receive an OTP for login.
    - User verifies the OTP with `/signin/otp/verify` and logs in [starts new session].
  

#### Note
- The email used for registration and sign-in must be valid.
- OTP verification is mandatory for both registration and sign-in processes.
- Tokens are stored as cookies for maintaining the user's session.
  
