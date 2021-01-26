import makeInspectable from 'mobx-devtools-mst';
import { cast, flow, getSnapshot, Instance, onSnapshot, SnapshotIn, SnapshotOrInstance, SnapshotOut, types } from 'mobx-state-tree';
import Axios, { AxiosResponse } from 'axios';
import { retrieveServicesAPI } from '@/api/services';
import { IService, Service } from './service';
import { ISession, Session } from './session';
import { retrieveSessionAPI } from '@/api/session';
import { APIPayload, APIRequestResult, APIResponseObject } from '@/api/common';

export const App = types.model({
    services: types.optional(types.array(Service), []),
    session: types.maybeNull(Session)
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
    const retrieveSession = flow(function* retrieveSession(uuid: string) {
        self.session = null
        const session: APIPayload<ISession> = yield retrieveSessionAPI(uuid);
        if (session.result == APIRequestResult.SUCCEEDED) {
            self.session = session.payload;
        }
        return session;
    });
    return { retrieveSession };
});

export interface IApp extends Instance<typeof App> {}

export const initialAppState = App.create({
    services: []
});

export const RootStore = types.model({
    app: App
});

export const store = RootStore.create({
    app: initialAppState
});

export interface IStore extends Instance<typeof store> {}

makeInspectable(store);
window.store = store;

(function start() {
    store.app.retrieveServices();
})();

onSnapshot(store, snapshot => {
    console.log(getSnapshot(store))
})