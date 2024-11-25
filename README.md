# GOapp_login

This is a simple login session handling app written with GO.
It is a good start for creating your own web app where login functionality is required.
The app by default stores user data using SQLite database.

## Setup

Clone the repository with `git clone https://github.com/Minionan/GOapp_login.git`

### Setting session_key

1. Copy `session_key.example.txt` to `session_key.txt`
2. Edit `session_key.txt` and set your secure cookie secret
3. Run the application: `go run .`

### Initialising user database

1. Run `init_db.go` script by typing in terminal `go run init_db.go`
2. Verify if SQLite database file was created in db.


