# post-video-go

An exploration into back-end systems with Golang. This project contains back-end services for video uploading and processing.

## Compression

### Implementation & Motivation

Three common video codecs used are AV1, H.265, and H.264. These three are compared against their speed, quality, and device compatibilty, for video compression.

A brief summary of findings:

- Speed:
  1. x264 is the fastest
  2. x265 is 2-15x slower
  3. AV1 is 15-30x slower
- Quality (equal intervals between ranks):
  1. AV1
  2. x265
  3. x264
- Browser compatibility
  1. x264 (virtually all browsers)
  2. AV1 (Edge, Firefox, Chrome, Opera)
  3. x265 (Chrome, Safari)

### Verdict

The project currently uses x264 codec for its speed and compatibilty. But the other codecs may be considered for some uncommon yet to be discovered scenarios where the compression time differences are negligable and so quality can be prioritised.

### Sources:

- https://www.winxdvd.com/convert-hevc-video/av1-vs-hevc.htm#compatibility
- https://www.lambdatest.com/web-technologies/hevc
- https://www.wowza.com/blog/h265-codec-high-efficiency-video-coding-hevc-explained

## Other Notes

### PubSub with Redis

You will need to run a redis server and point publications and subscriptions to the server/ your local redis server (https://redis.io/docs/getting-started/#installing-redis-more-properly).
To run locally install redis-cli and run redis-server. Each service currently points to the local redis server at localhost:6379.

Note that the PubSub buffer limit by default may be lower than what's required to contain large videos in the message queue. To increase the pubsub buffer limit, run the following command in the terminal:

> redis-cli config set client-output-buffer-limit "pubsub <hard-limit> <soft-limit> 0"

Replacing <hard-limit> and <soft-limit> with your maximum byte limit. For example for 300MB:

> redis-cli config set client-output-buffer-limit "pubsub 314572800 314572800 60"

For more information about this:

- https://rtfm.co.ua/en/en-draft-redis-psync-scheduled-to-be-closed-asap-for-overcoming-of-output-buffer-limits-i-client-output-buffer-limit/
- https://gist.github.com/amgorb/c38a77599c1f30e853cc9babf596d634

### FFMPEG

Most if not all services use ffmpeg (https://ffmpeg.org/) and so it will need to installed on each services server.
