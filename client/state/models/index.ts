import makeInspectable from 'mobx-devtools-mst';
import { cast, flow, getSnapshot, Instance, onSnapshot, SnapshotIn, SnapshotOrInstance, SnapshotOut, types } from 'mobx-state-tree';
import Axios from 'axios';
import { retrieveServicesAPI } from '@/api/services';

export const Service = types.model({
    name  : types.string,
    remote: types.string,
    target: types.string,
});

export interface IService extends Instance<typeof Service> {}
export interface IServiceSnapshotOut extends SnapshotOut<typeof Service>{}
export interface IServiceSnapshotIn extends SnapshotIn<typeof Service> { }

export const App = types.model({
    services: types.optional(types.array(Service), [])
})
.actions(self => {
    const retrieveServices = flow(function* retrieveServices() {
        const services: IService[] = yield retrieveServicesAPI();
        self.services = cast(services);
    });
    return { retrieveServices };
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