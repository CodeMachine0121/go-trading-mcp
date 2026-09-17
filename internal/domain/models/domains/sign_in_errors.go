package domains

import "errors"

// ErrSignInRequired is what an ability that must know who is asking says when nobody
// has said.
var ErrSignInRequired = errors.New("請先登入")

// ErrSignInExpired is what a signing-in past saving says.
//
// It is a different sentence from ErrSignInRequired on purpose, even though both end
// with the person signing in: one means "you never did", the other means "you did and
// it ran out". Telling somebody who signed in that they never did is how you get a
// report saying this connector forgets people.
var ErrSignInExpired = errors.New("登入已失效，請重新登入")
