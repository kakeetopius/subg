# subg

A command-line tool for searching and downloading movie and TV series subtitles from OpenSubtitles.

## Features

- Search for subtitles by movie/series name, IMDB ID, or other criteria
- Filter by language, release year, season, and episode
- Download subtitles in multiple formats (SRT, VTT, etc.)
- Support for both movies and TV series
- Automatic subtitle selection (not yet implemented)

## Installation

```bash
go install github.com/kakeetopius/subg@latest
```

## Quick Start

### 1. Login to OpenSubtitles

Before you can search and download subtitles, you must authenticate with OpenSubtitles:

```bash
subg login --provider os --username <your_username> --password <your_password>
```

You can sign up for a free OpenSubtitles account at [opensubtitles.com](https://www.opensubtitles.com/).

### 2. Search and Download Subtitles

```bash
# Search for a movie subtitle in English
subg search "The Matrix" --lang en

# Search for a TV series subtitle
subg search "Breaking Bad" --season 1 --episode 5 --lang en

# Download with custom output
subg search "Inception" --lang en --output-file Inception.srt --output-dir ./subtitles
```

## Usage

### search

Search and download subtitles for a movie or show.

```bash
subg search <query> [flags]
```

<details>
<summary>Flags</summary>

- `--lang` - Subtitle language code (default: "en")
- `--season` - TV series season number
- `--episode` - TV series episode number
- `--format` - Subtitle format to download (default: "srt")
- `--year` - Release year to reduce ambiguity
- `--output-file` - Custom output filename
- `--output-dir` - Output directory for downloaded subtitle
- `--imdb-id` - Search using IMDB ID
- `--movie` - Specify query is a movie
- `--serie` - Specify query is a TV series
- `--auto` - Automatically select first result without prompting
</details>

### login

Authenticate to a subtitle provider.

```bash
subg login --provider <provider> [flags]
```

<details>
<summary>Flags</summary>

- `--provider, -p` - Provider to authenticate to (currently: "os" for OpenSubtitles)
- `--username, -u` - Account username
- `--password, -P` - Account password

</details>

## Configuration

Configuration can be set via:

1. Configuration file
2. Environment variables
3. Command-line flags (highest priority)

### Configuration File

Place a `subg.toml` file in one of these locations:

- `$HOME/subg.toml`
- `$XDG_CONFIG_HOME/subg.toml` or `~/.config/subg.toml`
- `$XDG_CONFIG_HOME/subg/subg.toml` or `~/.config/subg/subg.toml`

**Example `subg.toml`:**

```toml
[opensubtitles]
api_key = "your-api-key-here"
username = "your-username"
password = "your-password"

cache_dir = "$HOME/.cache/subg"
```

### Environment Variables

- `OPENSUBTITLES_API_KEY` - You can set API key for OpenSubtitles using this environment variable instead of passing it via a flag or putting it in the configuration file.

## Future Plans

- Support for additional subtitle providers beyond OpenSubtitles
- Batch downloading capabilities
- Subtitle synchronization and adjustment tools
- Subtitle generation from a video.

## License

MIT  
See [LICENSE](LICENSE) file for details.
