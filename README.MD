![Soundr Logo](/resources/logo.svg "Soundr Logo")

# Soundr
Soundr is a simple, open-source, cross-platform audio playing server written in go.
It aims to be simple to setup and work in many envoriments. It is also designed to be
easy to use and maintain. 
Soundr is able to play multiple audio files at the same time. It is able to intigrate well as it uses a REST endpoint.
Swagger Documentation for that endpoint is available in `apiDocs.yml`.
The software it self is written in go and uses the BEEP library. It is made to be shipped as a single executable.
Another target was a minimal dependency tree.
Initally it was written to be used with [Bitfocus Companion](https://bitfocus.io/companion) in a more professional envoriment. [An companion module is now available!](https://github.com/bitfocus/companion-module-pnh-soundr)

# Installation
Installation is as simple as it gets as it is a single executable.
Download one of the releases, drop your sounds into the /sounds folder and run the executable.

# Configuration
The configuration is done in the `conf.json` file.
This is the basic json configuration file layout:
```json
{
    "Port": 8082
}
```

> All sounds should be in the same bitrate. It will assume the bitrate of the first loaded sample.

# Usage
Drop your sounds into the /sounds. You can play them by sending a GET request to the /v1/play endpoint.
You need to know the base64 encoded file name of the sound you want to play. You can get started by querying /v1/list. It will return a list of all sounds with their respective base64 encoded file name.
Use that base64 as the `file` parameter in the request. The response also includes sample request urls.

## Vanity IDs
When playing a sound file it will get an incrementing ID. You can use that ID to stop the sound. These are not really predictable and can change at any time. You can also use a vanity ID. This is an integer that you can use to identify the sound. You can set the vanity ID by adding a `id` parameter to the request. The response will include the vanity ID. You can then use that ID to stop the sound.

# Endpoints
> All endpoints are prefixed with `/v1/`

`GET /list` - Returns a list of all sounds with their respective base64 encoded file name. <br>
`GET /play` - Plays a sound. The `file` parameter is required. It is the base64 encoded file name of the sound you want to play. Also supports `loop` to be either `true` or `false` (default). It also supports the `id` parameter to set a vanity ID.<br>
`GET /stop` - Stops a sound. The `id` parameter is required. It is the ID of the sound you want to stop.<br>
`GET /stopAll` - Stops all sounds.<br>
`GET /buffer` - Buffers a sound. The `file` parameter is required. It is the base64 encoded file name of the sound you want to buffer.<br>
`GET /bufferAll` - Buffers all sounds. **Note**: This may take a while. And should probably be run before a show.<br>
`GET /current` - Returns a list of all currently playing sounds.<br>
`GET /remaining` - Takes a `id` parameter. Returns the remaining time of the sound with the given ID.<br>

**Note**: The sounds must be in the format `*.mp3`, `*.wav`, `*.flac` or `*.ogg` (`flac` files may take longer time to buffer).
