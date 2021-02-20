package main

import (
	"crypto/rand"
	"encoding/base64"
	"io"

	"github.com/lemon-mint/challenge-server/encryption"
	"github.com/valyala/fasthttp"
	"github.com/zeebo/blake3"
)

var server = []byte("challengeserver")
var useXFF bool = true

func main() {
	key := make([]byte, 32)
	io.ReadFull(rand.Reader, key)
	p := encryption.NewPacker(key)
	fasthttp.ListenAndServe(":59710", func(ctx *fasthttp.RequestCtx) {
		//ctx.Response.Header.SetServerBytes(server)
		switch string(ctx.Path()) {
		case "/":
			ctx.SetBody(indexHTML)
			ctx.SetContentType("text/html")
		case "/verify/cookie":
			verifyWithCookie(ctx, p)
		case "/token/new":

		default:
			ctx.SetStatusCode(404)
			ctx.SetBodyString("Error 404 Not Found")
		}
	})
}

func verifyWithCookie(ctx *fasthttp.RequestCtx, p *encryption.Packer) {
	token := ctx.Request.Header.Cookie("_go_clearance")
	if token == nil {
		ctx.SetStatusCode(403)
		return
	}
	//fmt.Println(getID(ctx))
	if !p.Verify(string(token), getID(ctx)) {
		ctx.SetStatusCode(403)
		return
	}
	ctx.SetStatusCode(200)
	ctx.SetContentType("text/plain")
	ctx.WriteString("OK 200")
}

func getID(ctx *fasthttp.RequestCtx) string {
	h := blake3.New()
	h.Write(ctx.UserAgent())
	h.Write(ctx.Host())
	h.WriteString(ctx.RemoteIP().String())
	userTrack := ctx.Request.Header.Cookie("_clearance_track")
	if userTrack == nil {
		ctx.SetStatusCode(403)
		buf := make([]byte, 8)
		io.ReadFull(rand.Reader, buf)
		uuid := base64.RawURLEncoding.EncodeToString(buf)
		h.WriteString(uuid)
		tracker := fasthttp.AcquireCookie()
		tracker.SetKey("_clearance_track")
		tracker.SetHTTPOnly(true)
		tracker.SetValue(uuid)
		tracker.SetPath("/")
		ctx.Response.Header.SetCookie(tracker)
		fasthttp.ReleaseCookie(tracker)
	} else {
		h.Write(userTrack)
	}
	if useXFF {
		xff := ctx.Request.Header.Peek("X-Forwarded-For")
		h.Write(xff)
	}
	hash := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	//fmt.Println(hash)
	return hash
}