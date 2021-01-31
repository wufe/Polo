import { APIPayload, APIRequestResult } from "@/api/common";
import { IAPISession, retrieveSessionAgeAPI, trackSessionAPI, untrackSessionAPI } from "@/api/session";
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

export interface ISessionLog extends Instance<typeof SessionLogModel> {}

export const castAPISessionToSessionModel = (apiSession: IAPISession): ISession => {
    const { logs, ...rest } = apiSession;
    const session = rest as ISession;
    session.logs = logs.reduce<{ [k: string]: ISessionLog }>((acc, log, index) => {
        acc[index] = log;
        return acc;
    }, {}) as any;
    return session;
}

export enum SessionStatus {
    STARTING     = 'starting',
    STARTED      = 'started',
    START_FAILED = 'start_failed',
    STOPPING     = 'stopping',
}

export const SessionModel = types.model({
    uuid    : types.string,
    name    : types.string,
    target  : types.string,
    port    : types.number,
    service : ServiceModel,
    status  : types.enumeration<SessionStatus>(Object.values(SessionStatus)),
    logs    : types.map(SessionLogModel),
    checkout: types.string,
    maxAge  : types.number,
    folder  : types.string
}).actions(self => {
    const track = flow(function* track() {
        const trackRequest: APIPayload<void> = yield trackSessionAPI(self.uuid);
        return trackRequest;
    });

    const untrack = flow(function *untrack() {
        const untrack: APIPayload<void> = yield untrackSessionAPI();
        return untrack;
    });

    const retrieveAge = flow(function* retrieveAge(uuid: string) {
        const age: APIPayload<number> = yield retrieveSessionAgeAPI(uuid);
        if (age.result == APIRequestResult.SUCCEEDED) {
            self.maxAge = age.payload;
        }
        return age;
    });

    return { retrieveAge, track, untrack };
})

export interface ISession extends Instance<typeof SessionModel> {}