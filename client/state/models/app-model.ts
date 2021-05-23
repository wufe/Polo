import { APIPayload, APIRequestResult } from "@/api/common";
import { retrieveApplicationsAPI, retrieveFailedSessionAPI, retrieveFailedSessionLogsAPI, retrieveFailedSessionsAPI } from "@/api/applications";
import { IAPISession, retrieveAllSessionsAPI, retrieveSessionAPI } from "@/api/session";
import { values } from "mobx";
import { types, flow, cast, Instance, getType, applySnapshot, applyPatch } from "mobx-state-tree";
import { ApplicationModel, IApplication } from "./application-model";
import { SessionModel, ISession, castAPISessionToSessionModel, ISessionLog } from "./session-model";
import { initialModalState, ModalModel } from "./modal-model";
import { INotification, NotificationModel, NotificationType } from "./notification-model";
import { v1 } from 'uuid';
import { FailuresModel, initialFailuresState } from "./failures-model";

export enum SessionSubscriptionEventType {
    FAIL = 'fail',
}
export type TSubscription = {
    sessionUUID: string;
    event      : SessionSubscriptionEventType;
    cb: (session: ISession, event: SessionSubscriptionEventType) => void;
};

export type TNotificationProps = {
    text: string;
    type?: NotificationType;
    title?: string;
    expiration?: number;
    onClick?: (notification: INotification) => void;
};

export const AppModel = types.model({
    session      : types.maybeNull(SessionModel),
    sessions     : types.map(SessionModel),
    failures     : FailuresModel,
    applications : types.map(ApplicationModel),
    notifications: types.map(NotificationModel),
    modal        : ModalModel,
})
// #region Session events subscriptions
.volatile(self => {
    return {
        subscriptions: [] as TSubscription[]
    };
})
.actions(self => {
    const subscribeToSessionEvents = (sessionUUID: TSubscription['sessionUUID'], event: TSubscription['event'], cb: TSubscription['cb']) => {
        self.subscriptions.push({ sessionUUID, event, cb });
    };

    const publishSessionEvent = (s: ISession, e: SessionSubscriptionEventType) => {
        for (const { sessionUUID, event, cb } of self.subscriptions) {
            if (sessionUUID === s.uuid && event === e)
                cb(s, event);
        }
    };
    return { subscribeToSessionEvents, publishSessionEvent };
})
// #endregion
.actions(self => {
    const retrieveApplications = flow(function* retrieveApplications() {
        const applications: APIPayload<IApplication[]> = yield retrieveApplicationsAPI();
        if (applications.result === APIRequestResult.SUCCEEDED) {

            const applicationsMap = applications.payload.reduce<{[applicationName: string]: IApplication}>((acc, application) => {
                acc[application.configuration.name] = application;
                return acc;
            }, {});

            self.applications.replace(applicationsMap);
        }
        return applications;
    });

    const retrieveAllSessions = flow(function* retrieveAllSessions() {
        const sessions: APIPayload<IAPISession[]> = yield retrieveAllSessionsAPI();
        if (sessions.result === APIRequestResult.SUCCEEDED) {
            const sessionsMap = sessions.payload.reduce<{ [applicationName: string]: ISession }>((acc, session) => {
                session.beingReplaced = !!sessions.payload.find(s => s.replacesSessions && s.replacesSessions.indexOf(session.uuid) > -1);
                acc[session.uuid] = castAPISessionToSessionModel(session);
                return acc;
            }, {});

            self.sessions.replace(sessionsMap);
        }
        return sessions;
    });

    const retrieveSession = flow(function* retrieveSession(uuid: string) {
        const session: APIPayload<IAPISession> = yield retrieveSessionAPI(uuid);
        if (session.result == APIRequestResult.SUCCEEDED) {
            self.session = castAPISessionToSessionModel(session.payload);
        }
        return session;
    });

    return { retrieveSession, retrieveAllSessions, retrieveApplications };
})
.actions(self => {
    const deleteNotification = (uuid: string) => {
        self.notifications.delete(uuid);
    };
    return { deleteNotification };
})
.actions(self => {
    const addNotification = ({
        text,
        type = NotificationType.INFO,
        title = '',
        expiration = 10,
        onClick,
    }: TNotificationProps) => {

        const uuid = v1();

        const notification = NotificationModel.create({
            uuid,
            expiration,
            text,
            title,
            type
        });

        notification.addOnClick(onClick);

        self.notifications.set(uuid, notification);

        if (expiration > 0) {
            setTimeout(() => {
                self.deleteNotification(uuid);
            }, expiration * 1000);
        }
    };

    return { addNotification };
})
.views(self => ({
    get sessionsByApplicationName() {
        return (values(self.sessions) as any as ISession[])
            .reduce<{ [name: string]: ISession[] }>((accumulator, session: ISession) => {
                const applicationName = session.applicationName;
                if (!accumulator[applicationName]) accumulator[applicationName] = [];
                accumulator[applicationName].push(session);
                return accumulator
            }, {});
    }
}));

export interface IApp extends Instance<typeof AppModel> { }

export const initialAppState = AppModel.create({
    modal: initialModalState,
    failures: initialFailuresState as any,
});