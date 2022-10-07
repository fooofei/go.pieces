package xboom

import (
	"context"
	"io"
)

type Boomable interface {
	// LoadBullet for shoot many times, run only in main thread
	LoadBullet(waitCtx context.Context, addr string) error
	// run in threads, must keep thread safe
	// return values
	//      @error shoot success or fail
	Shoot(waitCtx context.Context) error
	// Close will do close. run only in main thread
	io.Closer
}
