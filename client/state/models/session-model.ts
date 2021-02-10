import { APIPayload, APIRequestResult } from "@/api/common";
import { IAPISession, IAPISessionLogsAndStatus, killSessionAPI, retrieveLogsAndStatusAPI, retrieveSessionAgeAPI, trackSessionAPI, untrackSessionAPI } from "@/api/session";
import { flow, Instance, types } from "mobx-state-tree";
import { ApplicationModel } from "./application-model";

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
    when   : types.string,
    uuid   : types.string,
    type   : types.enumeration<SessionLogType>(Object.values(SessionLogType)),
    message: types.string,
})

export interface ISessionLog extends Instance<typeof SessionLogModel> {}

export const castAPISessionToSessionModel = (apiSession: IAPISession): ISession => {
    const { logs, ...rest } = apiSession;
    const session = rest as ISession;
    if (logs) {
        session.logs = logs.reduce<{ [k: string]: ISessionLog }>((acc, log) => {
            acc[log.uuid] = log;
            return acc;
        }, {}) as any;
    } else {
        session.logs = {} as any;
    }
    
    return session;
}

export enum SessionStatus {
    STARTING     = 'starting',
    STARTED      = 'started',
    START_FAILED = 'start_failed',
    STOPPING     = 'stopping',
    DEGRADED     = 'degraded',
}

export const SessionModel = types.model({
    uuid       : types.string,
    name       : types.string,
    target     : types.string,
    port       : types.number,
    application: ApplicationModel,
    status     : types.enumeration<SessionStatus>(Object.values(SessionStatus)),
    logs       : types.map(SessionLogModel),
    checkout   : types.string,
    maxAge     : types.number,
    folder     : types.string
}).actions(self => {
    const track = flow(function* track() {
        const trackRequest: APIPayload<void> = yield trackSessionAPI(self.uuid);
        return trackRequest;
    });

    const untrack = flow(function *untrack() {
        const untrack: APIPayload<void> = yield untrackSessionAPI();
        return untrack;
    });

    const retrieveAge = flow(function* retrieveAge() {
        const age: APIPayload<number> = yield retrieveSessionAgeAPI(self.uuid);
        if (age.result === APIRequestResult.SUCCEEDED) {
            self.maxAge = age.payload;
        }
        return age;
    });

    const retrieveLogsAndStatus = flow(function* retrieveLogsAndStatus(lastLogUUID?: string) {
        const logsAndStatus: APIPayload<IAPISessionLogsAndStatus> = yield retrieveLogsAndStatusAPI(self.uuid, lastLogUUID);

        if (logsAndStatus.result === APIRequestResult.SUCCEEDED) {
            self.status = logsAndStatus.payload.status;
            for (const log of logsAndStatus.payload.logs) {
                self.logs.set(log.uuid, log);
            }
        }

        return logsAndStatus;
    })

    const kill = flow(function* kill() {
        const kill: APIPayload<void> = yield killSessionAPI(self.uuid);
        return kill;
    })

    return { retrieveAge, track, untrack, kill, retrieveLogsAndStatus };
})

export interface ISession extends Instance<typeof SessionModel> {}