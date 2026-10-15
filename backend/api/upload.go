package api

import (
	"bytes"
	"io"
)

// multipartMemoryBytes is how much of a multipart upload is buffered in memory;
// anything larger is spooled to a temporary file by net/http. The overall
// request size is capped separately with http.MaxBytesReader.
const multipartMemoryBytes = 8 << 20 // 8 MB

// sniffAudioMIME inspects the leading bytes of an audio file and returns its
// MIME type, or "" if the content is not one of the supported formats.
// The client-supplied Content-Type is never trusted.
func sniffAudioMIME(head []byte) string {
	switch {
	case len(head) >= 12 && bytes.Equal(head[0:4], []byte("RIFF")) && bytes.Equal(head[8:12], []byte("WAVE")):
		return "audio/wav"
	case len(head) >= 4 && bytes.Equal(head[0:4], []byte("OggS")):
		return "audio/ogg"
	case len(head) >= 4 && bytes.Equal(head[0:4], []byte("fLaC")):
		return "audio/flac"
	case len(head) >= 8 && bytes.Equal(head[4:8], []byte("ftyp")):
		return "audio/mp4" // .m4a / .mp4 audio
	case len(head) >= 3 && bytes.Equal(head[0:3], []byte("ID3")):
		return "audio/mpeg" // MP3 with ID3v2 tag
	case len(head) >= 2 && head[0] == 0xFF && head[1]&0xF6 == 0xF0:
		return "audio/aac" // ADTS frame sync (MPEG-2/4 AAC, layer bits 00)
	case len(head) >= 2 && head[0] == 0xFF && head[1]&0xE0 == 0xE0:
		return "audio/mpeg" // bare MPEG audio frame sync
	}
	return ""
}

// readHead reads up to n bytes from r, returning fewer only at EOF.
func readHead(r io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	got, err := io.ReadFull(r, buf)
	if err == io.ErrUnexpectedEOF || err == io.EOF {
		err = nil
	}
	return buf[:got], err
}
