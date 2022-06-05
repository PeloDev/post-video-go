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
