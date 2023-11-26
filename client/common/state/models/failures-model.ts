import { IAPIFailedSessions, markFailedSessionAsAcknowledgedAPI, retrieveFailedSessionAPI, retrieveFailedSessionLogsAPI } from "@polo/common/api/applications";
import { APIPayload, APIRequestResult } from "@polo/common/api/common";
import { values } from "mobx";
import { flow, getParent, hasParent, types } from "mobx-state-tree";
import { IApp, SessionSubscriptionEventType } from ".";
import { ISession, ISessionLog, SessionModel } from "./session-model";

export enum FailureStatus {
    ACK   = 'acknowledged',
    UNACK = 'unacknowledged',
}

export type TFailuresByApplicationDictionary = {
    [appName: string]: TFailuresDictionary;
};

export type TFailuresDictionary = {
    [k in FailureStatus]: ISession[];
};

export const FailuresModel = types.model({
    acknowledged  : types.map(SessionModel),
    unacknowledged: types.map(SessionModel),
})
.actions(self => {
    const storeFailure = (session: ISession, status: FailureStatus) => {

        let addMap = self.acknowledged;
        let removeMap = self.unacknowledged;
        if (status === FailureStatus.UNACK) {
            addMap = self.unacknowledged;
            removeMap = self.acknowledged;
        }

        if (!addMap.has(session.uuid)) {
            addMap.set(session.uuid, session);
            removeMap.delete(session.uuid);
            if (status === FailureStatus.UNACK && hasParent(self)) {
                (getParent(self) as IApp).publishSessionEvent(session, SessionSubscriptionEventType.FAIL);
            }
        }
    }

    const markFailedSessionAsAcknowledged = flow(function* markFailedSessionAsAcknowledged(uuid: string) {
        const request = yield markFailedSessionAsAcknowledgedAPI(uuid);
        return request;
    })

    const retrieveFailedSession = flow(function* retrieveFailedSession(uuid: string, markAsSeen = true) {
        const request: APIPayload<ISession> = yield retrieveFailedSessionAPI(uuid);
        if (markAsSeen) markFailedSessionAsAcknowledged(uuid);
        return request;
    });

    const retrieveFailedSessionLogs = flow(function* retrieveFailedSessionLogs(uuid: string) {
        const request: APIPayload<ISessionLog[]> = yield retrieveFailedSessionLogsAPI(uuid);
        return request;
    });

    return { retrieveFailedSession, retrieveFailedSessionLogs, markFailedSessionAsAcknowledged, storeFailure };
})
.views(self => {
    const sessionsToMap = (sessions: ISession[], status: FailureStatus, accumulator: TFailuresByApplicationDictionary = {}): TFailuresByApplicationDictionary => {
        return sessions
            .reduce<TFailuresByApplicationDictionary>((accumulator, session: ISession) => {
                const applicationName = session.applicationName;
                if (!accumulator[applicationName]) {
                    accumulator[applicationName] = {
                        [FailureStatus.ACK]: [],
                        [FailureStatus.UNACK]: []
                    };
                }
                if (status === FailureStatus.ACK)
                    accumulator[applicationName].acknowledged.push(session);
                else
                    accumulator[applicationName].unacknowledged.push(session);
                return accumulator;
            }, accumulator);
    };

    return {
        get byApplicationName(): TFailuresByApplicationDictionary {
            return sessionsToMap(
                    Array.from(self.unacknowledged.values()),
                    FailureStatus.UNACK,
                    sessionsToMap(
                        Array.from(self.acknowledged.values()),
                        FailureStatus.ACK
                    )
            );
        }
    }
});

export const initialFailuresState = FailuresModel.create({});