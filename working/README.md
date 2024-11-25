# BBCapp_login

## Setup
1. Clone the repository
2. Copy `config.example.json` to `config.json`
3. Edit `config.json` and set your secure cookie secret
4. Run the application: `go run .`

## Configuration
The application requires a `config.json` file with the following structure:
```json
{
    "cookie_secret": "your_secure_random_string"
}
