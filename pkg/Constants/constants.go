package Constants

import "time"

const (
	LoggerPathLinux     = "./Log.txt"
	LoggerPathWindows   = "../../Log.txt"
	TemplatePathLinux   = "./pkg/templates/"
	TemplatePathWindows = "../../pkg/templates/"
	TemplatePathDarwin  = "../../pkg/templates/"
	StaticPathLinux     = "./pkg/static/"
	StaticPathWindows   = "../../pkg/static/"
	StaticPathDarwin    = "../../pkg/static/"

	ShutdownTime = 10 * time.Second
)
