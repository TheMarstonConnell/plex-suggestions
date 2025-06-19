# Plex Suggestions

Automatically discover and request new movies for your Plex library using AI recommendations and Radarr integration.

## What it does

This tool connects to your Plex server, analyzes your movie collection, and uses OpenAI to suggest movies you might like. It then automatically requests those movies in Radarr so they get downloaded and added to your library.

## Quick Start

### Prerequisites

You'll need:
- A Plex server with movies
- A Radarr instance 
- An OpenAI API key
- Docker (optional but recommended)

### Setup

1. **Clone this repo**
   ```bash
   git clone https://github.com/themarstonconnell/plex-suggestions.git
   cd plex-suggestions
   ```

2. **Copy and edit the environment file**
   ```bash
   cp .env.example .env
   # Edit .env with your actual settings
   ```

3. **Run it**
   ```bash
   # With Docker (recommended)
   docker-compose up --build
   
   # Or locally
   go run .
   ```

## Configuration

Your `.env` file needs these variables:

```env
# Plex - where your movies live
PLEX_IP=192.168.1.100
PLEX_TOKEN=your-plex-token-here
PLEX_SECTION_KEY=1

# Radarr - where new movies get requested
RADARR_URL=http://192.168.1.100:7878
RADARR_API_KEY=your-radarr-api-key
RADARR_QUALITY_PROFILE=HD-1080p  # Optional: use profile name instead of ID

# OpenAI - for movie recommendations
OPENAI_API_KEY=sk-your-openai-key
```

### Getting API Keys

**Plex Token:**

See [this article](https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token/) on how to get your Plex API key.

**Radarr API Key:**
- Go to Radarr → Settings → General → Security → API Key  

**OpenAI API Key:**
- Get one from [OpenAI Platform](https://platform.openai.com/api-keys)
- You'll need some credits for API calls

## How it works

1. **Fetches your movie collection** from Plex
2. **Sends the list to OpenAI** with a prompt asking for recommendations
3. **Gets back 5 movie suggestions** you don't already have
4. **Requests each movie in Radarr** for automatic downloading

The AI prompt is designed to suggest movies that match your taste while avoiding ones you already own. You can edit the system prompt in [the prompt folder](/prompts).

## Docker
Check out the `docker-compose.yml` file for local builds.
```bash
docker-compose up
```