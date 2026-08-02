package browserd

import processcapture "github.com/wsnacj/agentx-go/browser/host/browserd/internal/processcapture"

// ErrProcessOutputLimitExceeded reports that a browserd bootstrap command
// produced more output than the configured safe capture bound.
var ErrProcessOutputLimitExceeded = processcapture.ErrLimitExceeded
