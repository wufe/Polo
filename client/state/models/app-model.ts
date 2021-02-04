import { APIPayload, APIRequestResult } from "@/api/common";
import { retrieveApplicationsAPI } from "@/api/applications";
import { IAPISession, retrieveAllSessionsAPI, retrieveSessionAPI } from "@/api/session";
import { values } from "mobx";
import { types, flow, cast, Instance, getType, applySnapshot, applyPatch } from "mobx-state-tree";
import { ApplicationModel, IApplication } from "./application-model";
import { SessionModel, ISession, castAPISessionToSessionModel } from "./session-model";

export const AppModel = types.model({
    session     : types.maybeNull(SessionModel),
    sessions    : types.map(SessionModel),
    applications: types.map(ApplicationModel)
})
.actions(self => {
    const retrieveApplications = flow(function* retrieveApplications() {
        const applications: APIPayload<IApplication[]> = yield retrieveApplicationsAPI();
        if (applications.result === APIRequestResult.SUCCEEDED) {

            const applicationsMap = applications.payload.reduce<{[applicationName: string]: IApplication}>((acc, application) => {
                acc[application.name] = application;
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
.views(self => ({
    get sessionsByApplicationName() {
        return (values(self.sessions) as any as ISession[])
            .reduce<{ [name: string]: ISession[] }>((accumulator, session: ISession) => {
                const applicationName = session.application.name;
                if (!accumulator[applicationName]) accumulator[applicationName] = [];
                accumulator[applicationName].push(session);
                return accumulator
            }, {});
    }
}));

export interface IApp extends Instance<typeof AppModel> { }

export const initialAppState = AppModel.create({
    
});