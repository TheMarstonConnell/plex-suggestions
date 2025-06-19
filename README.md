# Plex Suggestions

A Go application that analyzes your Plex movie collection and uses AI to suggest new movies, then automatically requests them in Radarr.

## Features

- 🎬 **Plex Integration**: Fetches your movie collection from Plex
- 🤖 **AI Recommendations**: Uses OpenAI to suggest movies based on your taste
- 📥 **Radarr Integration**: Automatically requests suggested movies in Radarr
- 🔧 **Environment-based Configuration**: Secure configuration via environment variables
- 📦 **Docker Support**: Containerized for easy deployment

## Quick Start

### Using Docker (Recommended)

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/plex-suggestions.git
   cd plex-suggestions
   ```

2. **Create your `.env` file**
   ```bash
   cp .env.example .env
   # Edit .env with your actual configuration
   ```

3. **Run with Docker Compose**
   ```bash
   docker-compose up --build
   ```

### Using Docker Image

```bash
docker run --env-file .env ghcr.io/yourusername/plex-suggestions:latest
```

### Local Development

1. **Install Go 1.21+**
2. **Set up environment variables**
3. **Run the application**
   ```bash
   go run .
   ```

## Configuration

Create a `.env` file with the following variables:

```env
# Plex Configuration
PLEX_IP=your-plex-server-ip
PLEX_TOKEN=your-plex-token
PLEX_SECTION_KEY=1

# Radarr Configuration
RADARR_URL=http://your-radarr-ip:7878
RADARR_API_KEY=your-radarr-api-key

# OpenAI Configuration
OPENAI_API_KEY=your-openai-api-key
```

### Getting API Keys

- **Plex Token**: Go to Plex Web → Settings → General → Security → API Key
- **Radarr API Key**: Go to Radarr Web → Settings → General → Security → API Key
- **OpenAI API Key**: Get from [OpenAI Platform](https://platform.openai.com/api-keys)

## Docker Images

The application is automatically built and published to GitHub Container Registry:

- **Latest**: `ghcr.io/yourusername/plex-suggestions:latest`
- **Tagged releases**: `ghcr.io/yourusername/plex-suggestions:v1.0.0`

## GitHub Actions

This repository includes automated Docker builds via GitHub Actions:

- **On push to main**: Builds and publishes to `ghcr.io/yourusername/plex-suggestions:main`
- **On tags**: Builds and publishes versioned releases
- **On PRs**: Builds for testing (doesn't publish)

## Development

### Project Structure

```
plex-suggestions/
├── main.go              # Main application entry point
├── radarr.go            # Radarr API client
├── openai.go            # OpenAI API client
├── prompts/             # Embedded AI prompts
│   └── movie_recommendations.txt
├── Dockerfile           # Multi-stage Docker build
├── docker-compose.yml   # Local development setup
└── .github/workflows/   # GitHub Actions
    └── docker-publish.yml
```

### Building Locally

```bash
# Build Docker image
docker build -t plex-suggestions .

# Run container
docker run --env-file .env plex-suggestions
```

### Testing

```bash
# Test Plex connection
go run main.go

# Test Radarr connection (uncomment main function in radarr.go)
go run radarr.go
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test with Docker
5. Submit a pull request

## License

MIT License - see LICENSE file for details. 