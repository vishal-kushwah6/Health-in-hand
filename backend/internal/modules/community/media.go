package community

import (
	"strconv"
	"strings"
)

// CloudinaryPrefix turns CLOUDINARY_URL ("cloudinary://key:secret@cloudname")
// into the only URL prefix post images may start with:
// "https://res.cloudinary.com/cloudname/". Returns "" if it cannot be parsed.
func CloudinaryPrefix(cloudinaryURL string) string {
	i := strings.LastIndex(cloudinaryURL, "@")
	if i < 0 {
		return ""
	}
	name := cloudinaryURL[i+1:]
	if j := strings.IndexAny(name, "/?"); j >= 0 {
		name = name[:j]
	}
	if name == "" {
		return ""
	}
	return "https://res.cloudinary.com/" + name + "/"
}

// publicIDFromURL recovers the Cloudinary public id from a delivery URL so the
// file can be deleted with its post:
//
//	https://res.cloudinary.com/demo/image/upload/v123/posts/user_1/abc.jpg -> posts/user_1/abc
func publicIDFromURL(u string) string {
	i := strings.Index(u, "/upload/")
	if i < 0 {
		return ""
	}
	rest := u[i+len("/upload/"):]
	if strings.HasPrefix(rest, "v") { // optional version segment "v123/"
		if j := strings.Index(rest, "/"); j > 1 {
			if _, err := strconv.Atoi(rest[1:j]); err == nil {
				rest = rest[j+1:]
			}
		}
	}
	if k := strings.LastIndex(rest, "."); k > 0 {
		rest = rest[:k]
	}
	return rest
}
