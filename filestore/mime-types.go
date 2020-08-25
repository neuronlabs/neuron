package filestore

import (
	"path/filepath"
)

// MimeTypes are the file extension well-known mime types.
var MimeTypes = map[string]string{
	// Text
	"css":  "text/css",
	"html": "text/html",
	"txt":  "text/plain",
	// Images
	"gif":  "image/gif",
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"tiff": "image/tiff",
	"ico":  "image/vnd.microsoft.icon",

	// Application
	"js":   "application/javascript",
	"json": "application/json",
	"xml":  "application/xml",
	// Audio
	"mp3": "audio/mpeg",
	"wma": "audio/x-ms-wma",
	"ra":  "audio/vnd.rn-realaudio",
	"rm":  "audio/vnd.rn-realaudio",
	"ram": "audio/vnd.rn-realaudio",
	"wav": "audio/x-wav",

	// Video
	"mpeg": "video/mpeg",
	"mp4":  "video/mp4",
	"wmv":  "video/x-ms-wmv",
	"qt":   "video/quicktime",
	"mov":  "video/quicktime",
}

// GetMimeType gets the default mime type for given file extension.
func GetMimeType(file File) string {
	return MimeTypes[filepath.Ext(file.Name())]
}
