import makeInspectable from 'mobx-devtools-mst';
import { Instance, onPatch, onSnapshot, types } from 'mobx-state-tree';
import { AppModel, IApp, initialAppState } from './app-model';

export const RootStore = types.model({
    app: AppModel
});

export const store = RootStore.create({
    app: initialAppState as any
});

export const createStore = (state: { app: IApp }) =>
    RootStore.create(state as any);

export interface IStore extends Instance<typeof store> {}

makeInspectable(store);
window.store = store;

(function start() {
    
})();

onPatch(store.app.services, console.log);

onSnapshot(store, () => {
    // console.log(getSnapshot(store))
})

export * from './app-model';
export * from './service-model';
export * from './session-model';