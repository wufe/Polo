import makeInspectable from 'mobx-devtools-mst';
import { Instance, onSnapshot, types } from 'mobx-state-tree';
import { AppModel, initialAppState } from './app-model';

export const RootStore = types.model({
    app: AppModel
});

export const store = RootStore.create({
    app: initialAppState
});

export interface IStore extends Instance<typeof store> {}

makeInspectable(store);
window.store = store;

(function start() {
    
})();

onSnapshot(store, () => {
    // console.log(getSnapshot(store))
})

export * from './app-model';
export * from './service-model';
export * from './session-model';