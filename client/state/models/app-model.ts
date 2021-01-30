import { APIPayload, APIRequestResult } from "@/api/common";
import { retrieveServicesAPI } from "@/api/services";
import { retrieveAllSessionsAPI, retrieveSessionAPI } from "@/api/session";
import { types, flow, cast, Instance } from "mobx-state-tree";
import { ServiceModel, IService } from "./service-model";
import { SessionModel, ISession } from "./session-model";

export const AppModel = types.model({
    services: types.optional(types.array(ServiceModel), []),
    session: types.maybeNull(SessionModel),
    sessions: types.optional(types.array(SessionModel), []),
})
    .actions(self => {
        const retrieveServices = flow(function* retrieveServices() {
            const services: APIPayload<IService[]> = yield retrieveServicesAPI();
            if (services.result === APIRequestResult.SUCCEEDED) {
                self.services = cast(services.payload);
            }
            return services;
        });
        return { retrieveServices };
    })
    .actions(self => {

        const retrieveAllSessions = flow(function* retrieveAllSessions() {
            const sessions: APIPayload<ISession[]> = yield retrieveAllSessionsAPI();
            if (sessions.result === APIRequestResult.SUCCEEDED) {
                self.sessions = cast(sessions.payload);
            }
            return sessions;
        });

        const retrieveSession = flow(function* retrieveSession(uuid: string) {
            self.session = null
            const session: APIPayload<ISession> = yield retrieveSessionAPI(uuid);
            if (session.result == APIRequestResult.SUCCEEDED) {
                self.session = session.payload;
            }
            return session;
        });
        return { retrieveSession, retrieveAllSessions };
    })
    .views(self => ({
        get sessionsByServiceName() {

            return self.sessions.reduce<{ [serviceName: string]: ISession[] }>((accumulator, session) => {
                const serviceName = session.service.name;
                if (!accumulator[serviceName]) accumulator[serviceName] = [];
                accumulator[serviceName].push(session);
                return accumulator
            }, {});

        }
    }));

export interface IApp extends Instance<typeof AppModel> { }

export const initialAppState = AppModel.create({
    services: []
});