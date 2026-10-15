package api

import (
	"bytes"
	"testing"
)

func TestSniffAudioMIME(t *testing.T) {
	cases := []struct {
		name string
		head []byte
		want string
	}{
		{"wav", []byte("RIFF\x24\x08\x00\x00WAVEfmt "), "audio/wav"},
		{"ogg", []byte("OggS\x00\x02"), "audio/ogg"},
		{"flac", []byte("fLaC\x00\x00\x00\x22"), "audio/flac"},
		{"m4a", []byte("\x00\x00\x00\x20ftypM4A "), "audio/mp4"},
		{"mp3 id3", []byte("ID3\x04\x00"), "audio/mpeg"},
		{"mp3 mpeg1 layer3", []byte{0xFF, 0xFB, 0x90, 0x64}, "audio/mpeg"},
		{"mp3 mpeg2 layer3", []byte{0xFF, 0xF3, 0x90, 0x64}, "audio/mpeg"},
		{"aac adts mpeg4", []byte{0xFF, 0xF1, 0x50, 0x80}, "audio/aac"},
		{"aac adts mpeg2", []byte{0xFF, 0xF9, 0x50, 0x80}, "audio/aac"},
		{"html", []byte("<!doctype html><script>"), ""},
		{"riff not wave", []byte("RIFF\x24\x08\x00\x00AVI LIST"), ""},
		{"empty", nil, ""},
	}
	for _, c := range cases {
		if got := sniffAudioMIME(c.head); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestReadHeadShortInput(t *testing.T) {
	head, err := readHead(bytes.NewReader([]byte("abc")), 512)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(head) != "abc" {
		t.Fatalf("got %q, want %q", head, "abc")
	}
}
