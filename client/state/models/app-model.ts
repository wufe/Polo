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

export const AppModel = types.model({
    session       : types.maybeNull(SessionModel),
    sessions      : types.map(SessionModel),
    failedSessions: types.map(SessionModel),
    applications  : types.map(ApplicationModel),
    notifications : types.map(NotificationModel),
    modal         : ModalModel,
})
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
                session.beingReplaced = !!sessions.payload.find(s => s.replacesSession === session.uuid);
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

    const retrieveFailedSessions = flow(function* retrieveFailedSessions() {
        const request: APIPayload<ISession[]> = yield retrieveFailedSessionsAPI();
        if (request.result === APIRequestResult.SUCCEEDED) {
            for (const session of request.payload) {
                self.failedSessions.set(session.uuid, session);
            }
        }
        return request;
    });

    const retrieveFailedSession = flow(function *retrieveFailedSession(uuid: string) {
        const request: APIPayload<ISession> = yield retrieveFailedSessionAPI(uuid);
        return request
    })

    const retrieveFailedSessionLogs = flow(function *retrieveFailedSessionLogs(uuid: string) {
        const request: APIPayload<ISessionLog[]> = yield retrieveFailedSessionLogsAPI(uuid);
        return request;
    })

    return { retrieveSession, retrieveAllSessions, retrieveApplications, retrieveFailedSessions, retrieveFailedSession, retrieveFailedSessionLogs };
})
.actions(self => {
    const deleteNotification = (uuid: string) => {
        self.notifications.delete(uuid);
    };
    return { deleteNotification };
})
.actions(self => {

    type TNotificationProps = {
        text  : string;
        type? : NotificationType;
        title?: string;
        expiration?: number;
        onClick?: (notification: INotification) => void;
    };
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
    },
    get failedSessionsByApplicationName() {
        return (values(self.failedSessions) as any as ISession[])
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
    modal: initialModalState
});