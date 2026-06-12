# subg

A command-line tool for searching, downloading, generating, and converting subtitles.

## Features

- Search for subtitles by movie/series name, IMDB ID, or other criteria
- Filter by language, release year, season, and episode
- Download subtitles in multiple formats (SRT, VTT, ASS, SSA, TTML, STL)
- Convert subtitles from one format to another
- Generate subtitles from video or audio files using Whisper
- Fallback to other subtitle providers in case the primary one fails
- Support for both movies and TV series
- Automatic subtitle selection

## Installation

```bash
go install github.com/kakeetopius/subg@latest
```

## Subtitle Providers

subg can download subtitles from different providers. The following is a list of supported providers so far in order of priority together with their codes that can be passed
via the `--providers` flag or in the configuration file (see below).

| Provider          | Code |
| ----------------- | ---- |
| opensubtitles.com | os   |
| subdl.com         | sd   |
| addic7ed.com      | a7   |

> [!NOTE]
> If no provider is specified via the `--providers` flag or the configuration file, all providers are tried in the order shown above.  
> Multiple providers can be specified at once (e.g. --providers os,sd). Providers are then tried in the order given.

## Quick Start

### 1. Login to Provider

Before you can search and download subtitles specifically using Open Subtitles, you must authenticate with OpenSubtitles:

```bash
subg login --providers os --username <your_username> --password <your_password>
```

- You can sign up for a free OpenSubtitles account at [opensubtitles.com](https://www.opensubtitles.com/).
- For subdl.com only an API key is required. It can be obtained at [subdl.com](https://subdl.com/).
- For addic7ed.com no login or API key is required.

### 2. Download Subtitles

```bash
# Download a movie subtitle in English
subg download "In the Heights"

# Download a TV series subtitle
subg download "Bridge and Tunnel" --season 1 --episode 5 --lang en

# Download with custom output
subg download "Grown Ups 2" --lang en --output-file GU2.srt --output-dir ./subtitles
```

### 3. Generate Subtitles from Video or Audio

```bash
# Generate subtitles from a video file
subg generate movie.mp4

# Generate and translate to English
subg generate movie.mp4 --translate

# Use a specific Whisper model
subg generate movie.mp4 --model large-v3
```

### 4. Convert Subtitles

```bash
# Convert an SRT file to VTT
subg convert --in subtitle.srt --out subtitle.vtt

# Convert using the --convert-to flag (output filename is derived from input with only the extension changed)
subg convert --in subtitle.srt --convert-to ass
```

## Sample Usage

### download

Search and download subtitles for a movie or TV show. Aliases: `dl`, `d`.

```bash
subg download <query> [flags]
```

<details>
<summary>Flags</summary>

- `--lang, -l` - Subtitle language code (default: `"en"`)
- `--season, -s` - TV series season number
- `--episode, -e` - TV series episode number
- `--format, -f` - Subtitle format to download (default: `"srt"`)
- `--year, -y` - Release year to reduce ambiguity
- `--output-file, -o` - Custom output filename
- `--output-dir` - Output directory for downloaded subtitle
- `--imdb-id` - Search using IMDB ID
- `--movie` - Specify that the query is for a movie
- `--serie` - Specify that the query is for a TV series
- `--providers, -p` - The provider(s) to use as a comma-separated list (e.g. `os,sd`). If set, no fallback is done on failure.
- `--auto` - Automatically select the first result without prompting
</details>

### generate

Generate subtitles from a video or audio file using [whisper-ctranslate2](https://github.com/Softcatala/whisper-ctranslate2). Aliases: `gen`, `g`.

> [!IMPORTANT]
> This command requires [Python](https://www.python.org/) and [FFmpeg](https://ffmpeg.org/) to be installed on your system. Other dependencies are installed automatically by subg.

```bash
subg generate files... [flags]
```

<details>
<summary>Flags</summary>

- `--lang, -l` - The language of the input file(s) (default: `"en"`)
- `--format, -f` - The subtitle format to save as (default: `"srt"`)
- `--output-dir` - The directory to save the subtitle files to
- `--translate, -t` - Translate the audio to English instead of transcribing in the original language
- `--model, -m` - The Whisper model to use: `tiny`, `medium`, `large-v3`, `turbo`, etc. (default: `"turbo"`). See the [model card](https://github.com/openai/whisper/blob/main/model-card.md) for more information.
- `--hf-token` - A Hugging Face access token to access transcribing models
- `--verbose` - Print extra information about what is happening
</details>

### convert

Convert a subtitle file from one format to another. Alias: `c`.

Supported formats: `srt`, `vtt`, `ass`, `ssa`, `ttml`, `stl`.

```bash
subg convert --in <input-file> [flags]
```

<details>
<summary>Flags</summary>

- `--in, -i` - The input subtitle file **(required)**
- `--out, -o` - The output file name. If not given, the output filename is derived from the input filename.
- `--convert-to, -c` - The format to convert to. Required if `--out` is not given with a recognizable extension.
</details>

### login

Authenticate to a subtitle provider. Alias: `l`.

```bash
subg login --providers <provider> [flags]
```

<details>
<summary>Flags</summary>

- `--providers, -p` - Provider(s) to authenticate to
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
# Specify one or more providers to use by default. (See above for codes.)
providers = ["os", "sd"]

# Directory to store temporary information like JWT tokens for an OpenSubtitles session.
cache_dir = "$HOME/.cache/subg"

[opensubtitles]
# The OpenSubtitles API key is required when using OpenSubtitles.
# It can be set here or passed via the --api-key flag.
api_key = "your-api-key-here"
# Username and password used when logging in to OpenSubtitles.
# They can be set here or passed via corresponding flags.
username = "your-username"
password = "your-password"

[subdl]
# For subdl.com only the API key is required.
# It can be set here or passed via the --api-key flag.
api_key = "subdl_api_key"

[transcriber]
# Hugging Face access token for accessing Whisper transcribing models.
# It can be set here or passed via the --hf-token flag or via the env variable HF_TOKEN
hf_token = "your-hf-token-here"
```

### Environment Variables

- `OPENSUBTITLES_API_KEY` - API key for OpenSubtitles
- `SUBDL_API_KEY` - API key for subdl.com
- `HF_TOKEN` - Hugging Face Access Token

## Future Plans

- Support for additional subtitle providers
- Batch downloading capabilities
- Subtitle synchronization and adjustment tools

## License

MIT  
See [LICENSE](LICENSE) file for details.
