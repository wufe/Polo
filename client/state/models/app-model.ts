import { APIPayload, APIRequestResult } from "@/api/common";
import { retrieveServicesAPI } from "@/api/services";
import { IAPISession, retrieveAllSessionsAPI, retrieveSessionAPI } from "@/api/session";
import { values } from "mobx";
import { types, flow, cast, Instance, getType, applySnapshot, applyPatch } from "mobx-state-tree";
import { ServiceModel, IService } from "./service-model";
import { SessionModel, ISession, castAPISessionToSessionModel } from "./session-model";

export const AppModel = types.model({
    session: types.maybeNull(SessionModel),
    sessions: types.map(SessionModel),
    services: types.map(ServiceModel)
})
.actions(self => {
    const retrieveServices = flow(function* retrieveServices() {
        const services: APIPayload<IService[]> = yield retrieveServicesAPI();
        if (services.result === APIRequestResult.SUCCEEDED) {

            const servicesMap = services.payload.reduce<{[serviceName: string]: IService}>((acc, service) => {
                acc[service.name] = service;
                return acc;
            }, {});

            self.services.replace(servicesMap);
        }
        return services;
    });

    const retrieveAllSessions = flow(function* retrieveAllSessions() {
        const sessions: APIPayload<IAPISession[]> = yield retrieveAllSessionsAPI();
        if (sessions.result === APIRequestResult.SUCCEEDED) {
            const sessionsMap = sessions.payload.reduce<{ [serviceName: string]: ISession }>((acc, session) => {
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
    return { retrieveSession, retrieveAllSessions, retrieveServices };
})
.views(self => ({
    get sessionsByServiceName() {
        return (values(self.sessions) as any as ISession[])
            .reduce<{ [serviceName: string]: ISession[] }>((accumulator, session: ISession) => {
                const serviceName = session.service.name;
                if (!accumulator[serviceName]) accumulator[serviceName] = [];
                accumulator[serviceName].push(session);
                return accumulator
            }, {});
    }
}));

export interface IApp extends Instance<typeof AppModel> { }

export const initialAppState = AppModel.create({
    
});