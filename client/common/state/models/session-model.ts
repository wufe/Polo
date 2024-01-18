import { APIPayload, APIRequestResult } from '../../api/common';
import {
    getLogsWSURL,
    IAPISession,
    IAPISessionIntegrations,
    IAPISessionLogsAndStatus, IApiSessionStatus,
    killSessionAPI,
    retrieveLogsAndStatusAPI,
    retrieveSessionIntegrationsStatusAPI,
    retrieveSessionStatusAPI,
    trackSessionAPI,
    untrackSessionAPI
} from '../../api/session';
import { flow, IAnyModelType, Instance, types } from "mobx-state-tree";
import { SessionStatus, SessionKillReason } from "./session-model-enums";

export const SessionConfigurationModel = types.model({
    isDefault: types.boolean,
});

//#region Integrations
export const TiltDashboardModel = types.model({
    id: types.string,
    url: types.string,
});

export const TiltSessionIntegrationModel = types.model({
    dashboards: types.array(TiltDashboardModel),
});

export const SessionIntegrationsModel = types.model({
    tilt: TiltSessionIntegrationModel,
});
//#endregion

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
        session.logs = Array.from(logs.values()).reduce<{ [k: string]: ISessionLog }>((acc, log) => {
            acc[log.uuid] = log;
            return acc;
        }, {}) as any;
    } else {
        session.logs = {} as any;
    }
    
    return session;
}

export const SessionModel = types.model({
    uuid             : types.string,
    displayName      : types.string,
    alias            : types.string,
    target           : types.string,
    port             : types.number,
    applicationName  : types.string,
    status           : types.enumeration<SessionStatus>(Object.values(SessionStatus)),
    createdAt        : types.string,
    commitID         : types.string,
    commitMessage    : types.string,
    commitAuthorName : types.string,
    commitAuthorEmail: types.string,
    commitDate       : types.string,
    logs             : types.map(SessionLogModel),
    checkout         : types.string,
    age              : types.number,
    folder           : types.string,
    replacesSessions : types.array(types.string),
    beingReplacedBy  : types.maybe(types.late((): IAnyModelType => SessionModel)),
    configuration    : SessionConfigurationModel,
    killReason       : types.enumeration<SessionKillReason>(Object.values(SessionKillReason)),
    replacedBy       : types.string,
    permalink        : types.string,
    smartURL         : types.string,
    integrations     : SessionIntegrationsModel,
}).views(self => ({
    get beingReplacedBySession() {
        return self.beingReplacedBy as ISession;
    }
})).actions(self => {
    const track = flow(function* track() {
        const trackRequest: APIPayload<void> = yield trackSessionAPI(self.uuid);
        return trackRequest;
    });

    const untrack = flow(function *untrack() {
        const untrack: APIPayload<void> = yield untrackSessionAPI();
        return untrack;
    });

    const retrieveAge = flow(function* retrieveAge() {
        const sess: APIPayload<IApiSessionStatus> = yield retrieveSessionStatusAPI(self.uuid);
        if (sess.result === APIRequestResult.SUCCEEDED) {
            self.age = sess.payload.age;
        }
    });

    const retrieveStatus = flow(function* retrieveAge() {
        const age: APIPayload<IApiSessionStatus> = yield retrieveSessionStatusAPI(self.uuid);
        if (age.result === APIRequestResult.SUCCEEDED) {
            self.status = age.payload.status;
            return self.status;
        }
        return SessionStatus.NONE;
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
    });

    const retrieveIntegrationsStatus = flow(function* retrieveIntegrations() {
        const integrations: APIPayload<IAPISessionIntegrations> = yield retrieveSessionIntegrationsStatusAPI(self.uuid);

        if (integrations.result === APIRequestResult.SUCCEEDED) {
            self.integrations = {
                ...self.integrations,
                tilt: TiltSessionIntegrationModel.create({
                    dashboards: integrations.payload.tilt.dashboards.map(dashboard => TiltDashboardModel.create({
                        id: dashboard.id,
                        url: dashboard.url,
                    })),
                }),
            };
        }
    });

    const kill = flow(function* kill() {
        const kill: APIPayload<void> = yield killSessionAPI(self.uuid);
        return kill;
    })

    return { retrieveAge, retrieveStatus, track, untrack, kill, retrieveLogsAndStatus, retrieveIntegrationsStatus };
});

export interface ISession extends Instance<typeof SessionModel> {}