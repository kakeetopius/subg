# subg

A command-line tool for searching and downloading movie and TV series subtitles.

## Features

- Search for subtitles by movie/series name, IMDB ID, or other criteria
- Filter by language, release year, season, and episode
- Download subtitles in multiple formats (SRT, VTT, etc.)
- Fallback to other subtitle providers in case Open Subtitles fails. (addic7ed only so far)
- Support for both movies and TV series
- Automatic subtitle selection (not yet implemented)

## Installation

```bash
go install github.com/kakeetopius/subg@latest
```

## Quick Start

### 1. Login to OpenSubtitles

Before you can search and download subtitles specifically using Open Subtitles, you must authenticate with OpenSubtitles:

```bash
subg login --provider os --username <your_username> --password <your_password>
```

You can sign up for a free OpenSubtitles account at [opensubtitles.com](https://www.opensubtitles.com/).

### 2. Search and Download Subtitles

```bash
# Search for a movie subtitle in English
subg search "The Matrix"

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
- `--provider` - The specific provider to use. See the different providers below. (Note that if this is set no fallback will be done if any errors occur)
- `--auto` - Automatically select first result without prompting
</details>

### login

To download from some providers like opensubtitles the user must first authenticate to the subtitle provider.
For others like addic7ed authentication is not required.

```bash
subg login --provider <provider> [flags]
```

<details>
<summary>Flags</summary>

- `--provider, -p` - Provider to authenticate to
- `--username, -u` - Account username
- `--password, -P` - Account password

</details>

## Subtitle Providers

subg can download subtitles from different providers. The following is a list of supported providers in order of priority together with their codes that can be passed
via the --provider flag or in the configuration file. (See Below)

| Provider          | Code |
| ----------------- | ---- |
| opensubtitles.com | os   |
| addic7ed.com      | a7   |

> ![NOTE]
> If a provider is not specified using the flag --provider or using the configuration file (See Below), all providers are tried in the order shown above.

## Configuration

Configuration can be set via:

1. Configuration file
2. Environment variables
3. Command-line flags (highest priority)

### Configuration File

Place a `subg.toml` file in one of these locations:

**On Linux:**

- `$HOME/subg.toml`
- `$XDG_CONFIG_HOME/subg.toml` or `~/.config/subg.toml`
- `$XDG_CONFIG_HOME/subg/subg.toml` or `~/.config/subg/subg.toml`

**On Windows:**

- `%USERPROFILE%\subg.toml`
- `%APPDATA%\subg.toml`
- `%APPDATA%\subg\subg.toml`

**Example `subg.toml`:**

```toml
#if you want to only use a specific provider, you can specify here. (See above for the different codes.)
provider = "a7"

#directory to store temporary information like JWT tokens for an opensubtitles session.
cache_dir = "$HOME/.cache/subg"

[opensubtitles]
#NOTE that the opensubtitles api key is required if using opensubtitles to download. It can be set here or passed via the --api-key flag.
api_key = "your-api-key-here"
#The username and password are used when logging in to opensubtitles. They can be set here or can be passed via corresponding flags.
username = "your-username"
password = "your-password"
```

### Environment Variables

- `OPENSUBTITLES_API_KEY` - You can also set API key for OpenSubtitles using this environment variable instead of passing it via a flag or putting it in the configuration file.

## Future Plans

- Support for additional subtitle providers beyond OpenSubtitles
- Batch downloading capabilities
- Subtitle synchronization and adjustment tools
- Subtitle generation from a video.

## License

MIT  
See [LICENSE](LICENSE) file for details.
