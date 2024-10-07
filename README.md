# REST API song Library
Song library REST API server

This server uses next techniques:
- REST API
- middleware
- software layers
- PostgreSQL database
- database migrations
- filtering and pagination
- extended logging
- .env config file
- graceful shutdown
- Swagger/OpenApi specification (located in ./swagger/song-library.yaml)

## REST API endpoints

- GET /sons - Get songs from library with filtering and pagination
- POST /songs - Add a new song to the library
- PUT /songs - Update existing song data
- DELETE /songs - Delete a song from the library
- GET /songs/text - Get lyrics of a song with pagination
- GET /info - Get existing song data
