package core

import "fmt"

// LollipopHeaderName is the name of the custom Lollipop header
const LollipopHeaderName = "X-Lollipop"

// LollipopVersion is the version of the build
var LollipopVersion = "undefined"

// LollipopHeaderValue is the value of the custom Lollipop header
var LollipopHeaderValue = fmt.Sprintf("Version %s", LollipopVersion)

// LollipopUserAgent is the value of the user agent header sent to the backends
var LollipopUserAgent = fmt.Sprintf("Lollipop Version %s", LollipopVersion)
