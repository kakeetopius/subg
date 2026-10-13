# subg

An automation tool for working with subtitles eg searching, downloading, generating, extracting, embedding, and converting.

## Features

- Download subtitles by movie/series name, IMDB ID, or other criteria
- Download subtitles in multiple formats (SRT, VTT, ASS, SSA, TTML, STL)
- Convert a subtitle file from one format to another
- Generate subtitles from video or audio files using Whisper
- Embed subtitles directly into video files (after downloading or generating them, or on their own)
- Extract embedded subtitles from video files
- Fallback to other subtitle providers in case the primary one fails

## Installation

```bash
go install github.com/kakeetopius/subg@latest
```

## Subtitle Providers

subg can download subtitles from different providers. The following is a list of supported providers in order of priority together with their codes that can be passed via the `--providers` flag or in the configuration file (see below).

| Provider          | Code |
| ----------------- | ---- |
| opensubtitles.org | os   |
| subdl.com         | sd   |
| addic7ed.com      | a7   |

> [!NOTE]
> If no provider is specified via the `--providers` flag or the configuration file, all providers are tried in the order shown above.  
> The order of trial can be changed by using the `--providers` flag or via the configuration file (See Below) e.g. `--providers a7,sd,os`. Providers are then tried in the order given.

## Quick Start

### 1. Download Subtitles

```bash
# Download a movie subtitle
subg download "In the Heights"

# Download a TV series subtitle
subg download "Bridge and Tunnel" --season 1 --episode 5

# Download with custom output
subg download "Grown Ups 2" --lang en --output-file GU2.srt --output-dir ./subtitles

# Download a subtitle for a video file and embed it into the video
subg download --file movie.mp4 --embed

# Download subtitles for every movie/tv show episode in a directory
subg download --dir ./movies

# Show the results and pick one yourself instead of auto-selecting
subg download "In the Heights" --no-auto
```

### 2. Generate Subtitles from Video or Audio

```bash
# Generate subtitles from a video file
subg generate movie.mp4

# Use a specific Whisper model
subg generate movie.mp4 --model large-v3

# Generate subtitles for every video/audio file in a directory
subg generate --dir ./movies

# Generate subtitles and embed them into the video
subg generate movie.mp4 --embed
```

### 3. Convert Subtitles

```bash
# Convert an SRT file to VTT
subg convert subtitle.srt --out subtitle.vtt

# Convert using the --convert-to flag (output filename is derived from input with only the extension changed)
subg convert subtitle.srt --convert-to ass
```

### 4. Embed Subtitles into a Video

```bash
# Embed a subtitle into a video (overwrites the input video by default)
subg embed movie.mp4 --subtitle movie.srt

# Write to a new file instead of overwriting
subg embed movie.mp4 --subtitle movie.srt --out movie-subbed.mp4
```

### 5. Extract Subtitles from a Video

```bash
# Extract the subtitles embedded in a video file
subg extract movie.mkv

# Extract from every video in a directory
subg extract --dir ./movies
```

## Command Reference

### download

Search and download subtitles for a movie or TV show. Aliases: `dl`, `d`.

```bash
subg download [query] [flags]
```

<details>
<summary>Flags</summary>

- `--lang, -l` - Subtitle language code (default: `"en"`)
- `--season, -s` - TV series season number
- `--episode, -e` - TV series episode number
- `--year, -y` - Release year to reduce ambiguity
- `--format, -f` - Subtitle format to download (default: `"srt"`)
- `--imdb-id` - Search using IMDB ID
- `--movie` - Specify that the query is for a movie
- `--serie` - Specify that the query is for a TV series
- `--file, -F` - Download subtitles for a movie / TV show video file
- `--dir, -d` - Download subtitles for all movies / TV shows in a directory
- `--embed` - Embed the downloaded subtitle(s) into the video file(s) given via `--file` or `--dir`
- `--keep` - Keep the subtitle file(s) even after embedding
- `--output-file, -o` - Custom output filename
- `--output-dir` - Output directory for downloaded subtitle
- `--no-auto` - Do not automatically select a subtitle; show the results and let you choose
- `--providers, -p` - The provider(s) to use as a comma-separated list (e.g. `os,sd`). If set, no fallback is done on failure.

> [!NOTE]
> Subtitles are selected automatically by default. Use `--no-auto` to be prompted instead.

</details>

### generate

Generate subtitles from a video or audio file using [whisper-ctranslate2](https://github.com/Softcatala/whisper-ctranslate2). Aliases: `gen`, `g`.

> [!IMPORTANT]
> This command requires [Python](https://www.python.org/) and [FFmpeg](https://ffmpeg.org/) to be installed on your system. Other dependencies are installed automatically by subg.

```bash
subg generate [file] [flags]
```

<details>
<summary>Flags</summary>

- `--language, -l` - The language of the input file(s)
- `--format, -f` - The subtitle format to save as (default: `"srt"`)
- `--dir, -d` - Generate subtitles for all video/audio files in a directory
- `--output-dir` - The directory to save the subtitle files to
- `--translate, -t` - Translate the audio to English instead of transcribing in the original language
- `--embed, -e` - Embed the generated subtitle into the video (if the input is a video file)
- `--keep, -k` - Keep the subtitle file even after embedding
- `--model, -m` - The Whisper model to use: `tiny`, `medium`, `large-v3`, `turbo`, etc. (default: `"large-v3-turbo"`). See the [model card](https://github.com/openai/whisper/blob/main/model-card.md) for more information.
- `--hf-token` - A Hugging Face access token to access transcribing models
- `--verbose` - Print extra information about what is happening (default: `true`)

</details>

### convert

Convert a subtitle file from one format to another. Alias: `c`.

Supported formats: `srt`, `vtt`, `ass`, `ssa`, `ttml`, `stl`.

```bash
subg convert file [flags]
```

<details>
<summary>Flags</summary>

- `--out, -o` - The output file name. If not given, the output filename is derived from the input filename.
- `--convert-to, -c` - The format to convert to. Required if `--out` is not given with a recognizable extension.

</details>

### embed

Embed a subtitle into a video file. Alias: `e`.

> [!IMPORTANT]
> This command requires [FFmpeg](https://ffmpeg.org/) to be installed on your system.

```bash
subg embed file --subtitle <subtitle-file> [flags]
```

<details>
<summary>Flags</summary>

- `--subtitle, -s` - The subtitle file to embed into the video
- `--out, -o` - The output file name of the new video
- `--overwrite` - Overwrite the input file if the output file is not given (default: `true`)
- `--verbose` - Print extra debug information from ffmpeg

</details>

### extract

Extract subtitles from a video file. Alias: `ex`.

> [!IMPORTANT]
> This command requires [FFmpeg](https://ffmpeg.org/) (including `ffprobe`) to be installed on your system.

```bash
subg extract [file] [flags]
```

<details>
<summary>Flags</summary>

- `--dir, -d` - Extract subtitles from all video files in a directory
- `--output-dir` - The directory to save the extracted subtitle files to
- `--format, -f` - The format to save the extracted subtitle files as (default: `"srt"`)
- `--verbose` - Print extra debug information from ffmpeg and ffprobe

</details>

### login

Log in to an API-based subtitle provider. See [Optional: Configure Authenticated Providers](#optional-configure-authenticated-providers).

```bash
subg login --provider os_api
```

## Configuration

Configuration can be set via:

1. Configuration file
2. Environment variables
3. Command-line flags (highest priority)

### Configuration File

Place a `subg.toml` file in one of these locations:

**On Linux:**

- `$XDG_CONFIG_HOME/subg/subg.toml` or `~/.config/subg/subg.toml`
- `$XDG_CONFIG_HOME/subg.toml` or `~/.config/subg.toml`

**On Windows:**

- `%APPDATA%\subg\subg.toml`
- `%APPDATA%\subg.toml`

A different file can be given using the `--config` flag.

**Example `subg.toml`:**

```toml
# Specify one or more providers to use by default. (See above for codes.)
providers = ["os", "sd"]

# Directory to store temporary information like JWT tokens for an OpenSubtitles session.
# Defaults to a "subg" folder inside your user cache directory (e.g. ~/.cache/subg on Linux).
cache_dir = "$HOME/.cache/subg"

[transcriber]
# Hugging Face access token to use when accessing Whisper transcribing models.
# This is optional but can offer higher rate limits  and possibly faster downloads if provided.
# It can be set here or passed via the --hf-token flag or via the env variable HF_TOKEN
hf_token = "your-hf-token-here"
```

### Environment Variables

- `HF_TOKEN` - Hugging Face Access Token. (It is optional, but it can offer higher rate limits and faster downloads.)
- `OPENSUBTITLES_API_KEY` - API key for the opensubtitles.com API provider (`os_api`).
- `SUBDL_API_KEY` - API key for the SubDL API provider (`sd_api`).

## Optional: Configure Authenticated Providers

By default, subg can search and download subtitles using `subdl.com`, `opensubtitles.org`, and `addic7ed.com` without requiring an account or API key.

For improved search capabilities, higher reliability, or access to the official APIs, subg also supports the API-based providers below:

<details>
<summary>opensubtitles.com</summary>

To use the official opensubtitles.com API:

1. Create an account [here](https://www.opensubtitles.com) if you don't have one already.
2. Log in using:

```bash
subg login --provider os_api
```

3. Obtain an API key from your opensubtitles.com [here.](https://www.opensubtitles.com/en/consumers)
4. Add the API key to your `subg.toml` configuration file (or set the `OPENSUBTITLES_API_KEY` environment variable):

```toml
[opensubtitles]
api_key = "your-api-key"
```

The provider code for the API-based provider is **`os_api`**. You can include it in the provider list to give it priority, for example:

```toml
providers = ["os_api", "sd", "os", "a7"]
```

</details>

<details>
<summary>subdl.com</summary>

Although subg can use subdl.com without authentication, you may also use the official SubDL API.

1. Obtain an API key from the website [here](https://subdl.com/panel/api)
2. Add it to your configuration file (or set the `SUBDL_API_KEY` environment variable):

```toml
[subdl]
api_key = "your-api-key"
```

The provider code for the API-based provider is **`sd_api`**. To prioritize it, include it in your provider list, for example:

```toml
providers = ["sd_api", "os", "a7"]
```

</details>

The API-backed providers (`os_api` and `sd_api`) are separate from the providers `os` and `sd`. This allows you to choose whether to use the official APIs, the normal providers, or both. When no providers are specified, the API-backed providers are added automatically if an API key is configured.

## Future Plans

- Support for additional subtitle providers
- Subtitle synchronization and adjustment tools

## License

MIT  
See [LICENSE](LICENSE) file for details.
