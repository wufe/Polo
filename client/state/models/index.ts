import { AppModel, IApp, initialAppState } from '@/state/models/app-model';
import { Instance, onPatch, types } from 'mobx-state-tree';
import { isDev } from '@/utils/env';
import makeInspectable from 'mobx-devtools-mst';

export const RootStore = types.model({
    app: AppModel
});

export const store = RootStore.create({
    app: initialAppState as any
});

export const createStore = (state: { app: IApp }) =>
    RootStore.create(state as any);

export interface IStore extends Instance<typeof store> {}

if (isDev()) {
    // onPatch(store.app, console.log);
    makeInspectable(store);
    window.store = store;
}

export * from './app-model';
export * from './application-model';
export * from './session-model';