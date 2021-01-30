import { APIPayload } from "@/api/common";
import { trackSessionAPI } from "@/api/session";
import { flow, Instance, types } from "mobx-state-tree";
import { ServiceModel } from "./service-model";

export enum SessionLogType {
    TRACE    = 'trace',
    DEBUG    = 'debug',
    INFO     = 'info',
    WARN     = 'warn',
    ERROR    = 'error',
    CRITICAL = 'critical',
    STDOUT   = 'stdout',
    STDERR   = 'stderr',
    STDIN    = 'stdin',
}

export const SessionLogModel = types.model({
    type   : types.enumeration<SessionLogType>(Object.values(SessionLogType)),
    message: types.string,
})

export enum SessionStatus {
    STARTING     = 'starting',
    STARTED      = 'started',
    START_FAILED = 'start_failed',
    STOPPING     = 'stopping',
}

export const SessionModel = types.model({
    uuid      : types.string,
    name      : types.string,
    target    : types.string,
    port      : types.number,
    service   : ServiceModel,
    status    : types.enumeration<SessionStatus>(Object.values(SessionStatus)),
    logs      : types.optional(types.array(SessionLogModel), []),
    checkout  : types.string,
    inactiveAt: types.string,
    folder    : types.string
}).actions(self => {
    const track = flow(function* track() {
        const trackRequest: APIPayload<void> = yield trackSessionAPI(self.uuid);
        return trackRequest;
    });
    return { track };
})

export interface ISession extends Instance<typeof SessionModel> {}