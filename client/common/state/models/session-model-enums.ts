export enum SessionStatus {
    NONE         = '',
    STOPPED      = 'stopped',
    STARTING     = 'starting',
    STARTED      = 'started',
    START_FAILED = 'start_failed',
    STOP_FAILED  = 'stop_failed',
    STOPPING     = 'stopping',
    DEGRADED     = 'degraded',
}

export enum SessionKillReason {
    NONE = '',
    STOPPED = 'stopped',
    BUILD_FAILED = 'build_failed',
    HEALTHCHECK_FAILED = 'healthcheck_failed',
    REPLACED = 'replaced',
}