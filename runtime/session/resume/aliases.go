package resume

// Worker coordinates one validated continuation readback and Host wake dispatch.
type Worker = ObjectiveRuntimeSchedulerResumeWorker

// WorkerConfig supplies display-safe Host references used by Worker.
type WorkerConfig = ObjectiveRuntimeSchedulerResumeWorkerConfig

// Daemon coordinates queue enqueue and one bounded tick execution.
type Daemon = ObjectiveRuntimeSchedulerResumeDaemon

// DaemonConfig selects the portable scheduler lane and Host references.
type DaemonConfig = ObjectiveRuntimeSchedulerResumeDaemonConfig

// Service executes bounded Daemon cycles without owning a system process.
type Service = ObjectiveRuntimeSchedulerResumeDaemonService

// ServiceWaitFunc is the Host-provided wait seam between bounded cycles.
type ServiceWaitFunc = ObjectiveRuntimeSchedulerResumeDaemonServiceWaitFunc

// EnqueueRequest is the display-safe input for one resume tick.
type EnqueueRequest = ObjectiveRuntimeSchedulerResumeTickEnqueueInput

// EnqueueResult reports whether a resume tick was accepted.
type EnqueueResult = ObjectiveRuntimeSchedulerResumeTickEnqueueReport

// ServiceInput bounds one service-loop invocation.
type ServiceInput = ObjectiveRuntimeSchedulerResumeDaemonServiceInput

// ServiceReport summarizes one bounded service-loop invocation.
type ServiceReport = ObjectiveRuntimeSchedulerResumeDaemonServiceReport
